package ai

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// ServerInfoProvider 从 connID 解析出 serverKey (host:port)
type ServerInfoProvider interface {
	GetServerKey(connID string) string
}

// SSHCommandRunner 执行 SSH 命令
type SSHCommandRunner interface {
	RunCommand(connID, command string) (string, error)
}

// AIService AI 服务（Wails 暴露给前端）
type AIService struct {
	mu       sync.RWMutex
	app      *application.App
	config   *ConfigManager
	store    *ChatStore
	logger   *AILogger
	sessions map[string]*session  // serverKey → 共享会话（消息历史）
	conns    map[string]*connState // connId → 独立状态（队列/处理中）

	// 工具审批/执行通道
	approvals map[string]chan bool   // callID → approve/deny
	results   map[string]chan string // callID → result

	// 外部依赖
	serverInfo ServerInfoProvider
	sshRunner  SSHCommandRunner
}

// session 每个服务器的会话状态（按 serverKey 共享）
type session struct {
	mu       sync.Mutex
	messages []map[string]interface{} // 统一消息源（API 上下文 + 前端显示）
}

// connState 每个连接的独立状态（按 connId 隔离）
type connState struct {
	mu         sync.Mutex
	processing bool
	cancel     chan struct{}
	queue      []*ChatRequest
}

// ChatMessage 前端显示用消息
type ChatMessage struct {
	ID        string    `json:"id"`
	Role      string    `json:"role"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

// ChatRequest 前端发来的消息请求
type ChatRequest struct {
	ConnID       string `json:"connId"`
	Message      string `json:"message"`
	DeepThinking bool   `json:"deepThinking,omitempty"` // 深度思考（仅部分模型支持）
}

// ChatResponse 返回给前端的响应
type ChatResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

// --- 初始化 ---

func NewAIService() *AIService {
	dataDir := getDataDir()

	store, err := NewChatStore(dataDir)
	if err != nil {
		fmt.Printf("[AI] 存储初始化失败: %v\n", err)
	}

	return &AIService{
		config:    NewConfigManager(),
		store:     store,
		logger:    NewAILogger(dataDir),
		sessions:  make(map[string]*session),
		conns:     make(map[string]*connState),
		approvals: make(map[string]chan bool),
		results:   make(map[string]chan string),
	}
}

func (s *AIService) SetApp(app *application.App) {
	s.app = app

	// 监听前端提交工具结果的事件
	app.Event.On("ai:submit-tool-result", func(event *application.CustomEvent) {
		data, ok := event.Data.(map[string]interface{})
		if !ok {
			fmt.Printf("[AI] ⚠️ 无效的 ai:submit-tool-result 事件数据\n")
			return
		}
		callID, _ := data["callId"].(string)
		result, _ := data["result"].(string)
		if callID == "" {
			fmt.Printf("[AI] ⚠️ ai:submit-tool-result: callId 为空\n")
			return
		}
		fmt.Printf("[AI] 收到工具结果提交: callId=%s, result长度=%d\n", callID, len(result))
		s.SubmitToolResult(callID, result)
	})
}

// InitDeps 初始化外部依赖（包级函数，不生成 Wails 绑定）
func InitDeps(s *AIService, info ServerInfoProvider, runner SSHCommandRunner) {
	s.serverInfo = info
	s.sshRunner = runner
}

// --- 配置 ---

func (s *AIService) GetConfig() *AIConfig   { return s.config.GetConfig() }
func (s *AIService) IsConfigured() bool     { return s.config.IsConfigured() }
func (s *AIService) SaveConfig(c *AIConfig) error { return s.config.SaveConfig(c) }

// FetchModels 从 API 端点获取可用模型列表（使用已保存的配置）
func (s *AIService) FetchModels() []map[string]string {
	cfg := s.config.GetConfig()
	return s.fetchModelsFrom(cfg.APIEndpoint, cfg.APIKey)
}

// FetchModelsWithParams 从指定的 API 端点获取可用模型列表（使用传入的参数，不依赖已保存的配置）
func (s *AIService) FetchModelsWithParams(endpoint, key string) []map[string]string {
	return s.fetchModelsFrom(endpoint, key)
}

func (s *AIService) fetchModelsFrom(apiEndpoint, apiKey string) []map[string]string {
	if apiEndpoint == "" || apiKey == "" {
		return nil
	}

	endpoint := strings.TrimRight(apiEndpoint, "/")
	endpoint = strings.TrimSuffix(endpoint, "/chat/completions")
	endpoint = strings.TrimSuffix(endpoint, "/chat")
	endpoint = strings.TrimRight(endpoint, "/")
	endpoint += "/models"

	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))
	req.Header.Set("api-key", apiKey)

	resp, err := client.Do(req)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil
	}

	data, ok := result["data"].([]interface{})
	if !ok {
		return nil
	}

	var models []map[string]string
	for _, item := range data {
		m, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		id, _ := m["id"].(string)
		owner, _ := m["owned_by"].(string)
		if id != "" {
			models = append(models, map[string]string{"id": id, "owner": owner})
		}
	}
	return models
}

// --- 核心：发送消息 ---

func (s *AIService) SendMessage(req *ChatRequest) *ChatResponse {
	if !s.config.IsConfigured() {
		return &ChatResponse{Success: false, Error: "请先配置 AI API"}
	}

	key := s.resolveKey(req.ConnID)
	sess := s.getOrCreateSession(key)
	cs := s.getOrCreateConnState(req.ConnID)

	cs.mu.Lock()
	if cs.processing {
		cs.queue = append(cs.queue, req)
		cs.mu.Unlock()
		return &ChatResponse{Success: true, Message: "已排队等待处理"}
	}
	cs.processing = true
	cs.cancel = make(chan struct{})
	cs.mu.Unlock()

	// 添加用户消息到共享会话
	userMsg := map[string]interface{}{
		"role":    "user",
		"content": req.Message,
	}
	sess.mu.Lock()
	sess.messages = append(sess.messages, userMsg)
	sess.mu.Unlock()

	s.emitStatus(req.ConnID, "step", "分析用户意图...")
	s.logger.Log(key, "收到消息: %s", truncate(req.Message, 100))

	go func() {
		s.processWithTools(req.ConnID, key, sess, cs)
		s.processQueue(req.ConnID, key, sess, cs)
	}()

	return &ChatResponse{Success: true, Message: "已发送"}
}

// CancelProcessing 取消当前处理
func (s *AIService) CancelProcessing(connID string) {
	key := s.resolveKey(connID)
	cs := s.getOrCreateConnState(connID)

	cs.mu.Lock()
	defer cs.mu.Unlock()

	if cs.cancel != nil {
		close(cs.cancel)
		cs.cancel = nil
	}
	cs.queue = nil
	cs.processing = false
	s.logger.Log(key, "用户取消处理")
}

// processQueue 处理排队的消息（按 connId 隔离）
func (s *AIService) processQueue(connID, key string, sess *session, cs *connState) {
	for {
		cs.mu.Lock()
		if len(cs.queue) == 0 {
			cs.processing = false
			cs.cancel = nil
			cs.mu.Unlock()
			return
		}
		req := cs.queue[0]
		cs.queue = cs.queue[1:]
		cs.cancel = make(chan struct{})
		cs.mu.Unlock()

		// 添加用户消息
		userMsg := map[string]interface{}{
			"role":    "user",
			"content": req.Message,
		}
		sess.mu.Lock()
		sess.messages = append(sess.messages, userMsg)
		sess.mu.Unlock()

		s.emitStatus(connID, "step", "分析用户意图...")
		s.logger.Log(key, "处理排队消息: %s", truncate(req.Message, 100))

		s.processWithTools(connID, key, sess, cs)
	}
}

// --- 工具调用循环 ---

func (s *AIService) processWithTools(connID, key string, sess *session, cs *connState) {
	cfg := s.config.GetConfig()
	maxRounds := 5

	for round := 0; round < maxRounds; round++ {
		// 检查是否已取消
		if s.isCancelled(cs) {
			s.logger.Log(key, "处理已取消")
			return
		}
		// 构建消息
		messages := s.buildMessages(sess, cfg)

		if round > 0 {
			s.emitStatus(connID, "step", fmt.Sprintf("分析工具结果... (第%d轮)", round+1))
		} else {
			s.emitStatus(connID, "step", "正在调用 AI 模型...")
		}
		s.logger.Log(key, "API 请求: 模型=%s, 消息数=%d, 轮次=%d", cfg.Model, len(messages), round+1)

		// 调用 API
		requestBody := map[string]interface{}{
			"model":                cfg.Model,
			"messages":             messages,
			"temperature":          cfg.Temperature,
			"max_completion_tokens": cfg.MaxTokens,
		}
		if cfg.TopP > 0 {
			requestBody["top_p"] = cfg.TopP
		}
		// 工具定义
		requestBody["tools"] = GetToolDefinitions()

		resp, err := s.callAPI(cfg, requestBody, connID)
		if err != nil {
			s.emitError(connID, key, sess, fmt.Sprintf("AI 调用失败: %v", err))
			s.logger.Log(key, "API 调用失败: %v", err)
			return
		}

		choice, err := s.parseChoice(resp)
		if err != nil {
			s.emitError(connID, key, sess, fmt.Sprintf("解析响应失败: %v", err))
			return
		}

		message, _ := choice["message"].(map[string]interface{})

		// 检查是否有工具调用
		toolCalls, hasTools := message["tool_calls"].([]interface{})
		if hasTools && len(toolCalls) > 0 {
			s.emitStatus(connID, "step", fmt.Sprintf("需要执行 %d 个工具", len(toolCalls)))

			// 保存 assistant 消息（含 tool_calls）
			assistantMsg := map[string]interface{}{
				"role":       "assistant",
				"content":    message["content"],
				"tool_calls": toolCalls,
			}
			if rc, ok := message["reasoning_content"].(string); ok && rc != "" {
				assistantMsg["reasoning_content"] = rc
			}
			sess.mu.Lock()
			sess.messages = append(sess.messages, assistantMsg)
			sess.mu.Unlock()

			// 逐个工具：审批 → 执行
			allApproved := true
			for _, tc := range toolCalls {
				tcMap, ok := tc.(map[string]interface{})
				if !ok {
					continue
				}

				callID, _ := tcMap["id"].(string)
				if callID == "" {
					callID = fmt.Sprintf("call_%d", time.Now().UnixNano())
				}

				toolName, argsJSON := "", ""
				if fn, ok := tcMap["function"].(map[string]interface{}); ok {
					toolName, _ = fn["name"].(string)
					argsJSON, _ = fn["arguments"].(string)
				}

				cmdDesc := GetCommandDescription(toolName, argsJSON)
				s.emitStatus(connID, "step", fmt.Sprintf("需要执行: %s", truncate(cmdDesc, 40)))
				s.logger.Log(key, "工具请求: %s → %s", toolName, cmdDesc)

				// 发送审批请求
				approved := s.waitForApproval(connID, callID, toolName, cmdDesc, cs)

				if !approved {
					allApproved = false
					s.logger.Log(key, "用户拒绝: %s", cmdDesc)
					sess.mu.Lock()
					sess.messages = append(sess.messages, map[string]interface{}{
						"role":        "tool",
						"tool_call_id": callID,
						"content":     "用户拒绝执行此命令。",
					})
					sess.mu.Unlock()
					continue
				}

				// 通过前端执行
				s.logger.Log(key, "用户批准: %s", cmdDesc)
				s.emitStatus(connID, "step", fmt.Sprintf("执行: %s ...", truncate(cmdDesc, 40)))

				result := s.executeViaFrontend(connID, callID, toolName, cmdDesc, argsJSON, cs)
				s.logger.Log(key, "执行完成，输出 %d 字符", len(result))

				sess.mu.Lock()
				sess.messages = append(sess.messages, map[string]interface{}{
					"role":        "tool",
					"tool_call_id": callID,
					"content":     result,
				})
				sess.mu.Unlock()

				// 通知前端显示结果
				s.emitToolResult(connID, callID, toolName, cmdDesc, result)
			}

			if !allApproved {
				// 有工具被拒绝，AI 需要知道
			}

			continue // 下一轮，让 AI 处理工具结果
		}

		// 没有工具调用 — 最终回复
		s.emitStatus(connID, "step", "正在生成回复...")
		content, _ := message["content"].(string)
		if content == "" {
			content = "(AI 返回了空回复)"
		}

		// 追加联网搜索引用
		if annotations, ok := message["annotations"].([]interface{}); ok && len(annotations) > 0 {
			refs := "\n\n---\n**参考来源：**\n"
			for i, ann := range annotations {
				if a, ok := ann.(map[string]interface{}); ok {
					title, _ := a["title"].(string)
					url, _ := a["url"].(string)
					if title != "" && url != "" {
						refs += fmt.Sprintf("%d. [%s](%s)\n", i+1, title, url)
					}
				}
			}
			content += refs
		}

		// 保存到 messages
		finalMsg := map[string]interface{}{
			"role":    "assistant",
			"content": content,
		}
		if rc, ok := message["reasoning_content"].(string); ok && rc != "" {
			finalMsg["reasoning_content"] = rc
		}
		sess.mu.Lock()
		sess.messages = append(sess.messages, finalMsg)
		sess.mu.Unlock()

		// 持久化
		s.persistMessages(key, sess)

		// 通知前端
		aiMsg := &ChatMessage{
			ID:        fmt.Sprintf("msg_%d", time.Now().UnixNano()),
			Role:      "assistant",
			Content:   content,
			Timestamp: time.Now(),
		}
		if s.app != nil {
			s.app.Event.Emit("ai:message", map[string]interface{}{
				"connId":  connID,
				"message": aiMsg,
			})
		}
		s.logger.Log(key, "回复完成，长度 %d 字符", len(content))
		return
	}

	// 工具调用轮数用尽，让 AI 基于已收集数据生成总结
	s.emitStatus(connID, "step", "正在基于已收集数据生成回复...")
	s.logger.Log(key, "工具调用轮数达上限，请求 AI 总结")

	// 追加一条提示让 AI 总结
	sess.mu.Lock()
	sess.messages = append(sess.messages, map[string]interface{}{
		"role":    "user",
		"content": "你已经收集了足够的数据，请基于以上所有工具返回的结果，直接给出完整的分析和回复。不要再调用任何工具。",
	})
	sess.mu.Unlock()

	// 最后一轮：不带工具定义，强制 AI 生成文本回复
	finalCfg := *cfg
	finalMessages := s.buildMessages(sess, &finalCfg)
	finalBody := map[string]interface{}{
		"model":                finalCfg.Model,
		"messages":             finalMessages,
		"temperature":          finalCfg.Temperature,
		"max_completion_tokens": finalCfg.MaxTokens,
	}
	resp, err := s.callAPI(&finalCfg, finalBody, connID)
	if err != nil {
		s.emitError(connID, key, sess, fmt.Sprintf("生成总结失败: %v", err))
		return
	}
	choice, err := s.parseChoice(resp)
	if err != nil {
		s.emitError(connID, key, sess, fmt.Sprintf("解析总结失败: %v", err))
		return
	}
	msg, _ := choice["message"].(map[string]interface{})
	content, _ := msg["content"].(string)
	if content == "" {
		content = "已收集到服务器数据，但未能生成总结回复。请查看上方的工具执行结果。"
	}
	aiMsg := &ChatMessage{
		ID:        fmt.Sprintf("msg_%d", time.Now().UnixNano()),
		Role:      "assistant",
		Content:   content,
		Timestamp: time.Now(),
	}
	sess.mu.Lock()
	sess.messages = append(sess.messages, map[string]interface{}{"role": "assistant", "content": content})
	sess.mu.Unlock()
	s.persistMessages(key, sess)
	if s.app != nil {
		s.app.Event.Emit("ai:message", map[string]interface{}{"connId": connID, "message": aiMsg})
	}
	s.logger.Log(key, "总结回复完成，长度 %d 字符", len(content))
}

// --- 消息构建 ---

func (s *AIService) buildMessages(sess *session, cfg *AIConfig) []map[string]interface{} {
	// TTS 模型不支持 system role
	isTTS := strings.Contains(cfg.Model, "tts")
	var systemMsg map[string]interface{}
	systemTokens := 0
	if !isTTS {
		prompt := cfg.SystemPrompt
		// 追加工具使用规则（确保 AI 始终知道如何正确使用工具）
		if !strings.Contains(prompt, "execute_command") {
			prompt += `

【可用工具】
- execute_command: 执行 shell 命令（自动复用已有终端）
- get_server_info: 获取服务器基本信息
- get_system_status: 获取服务器实时状态
- open_terminal: 开新终端（仅用于多任务并行）
- close_terminal: 关闭 AI 自己的终端

【终端规则】
- 简单命令用 execute_command，不要开新终端
- 只有同时监控多个独立任务时才用 open_terminal
- 任务完成必须 close_terminal 释放资源
- 绝不关闭用户的终端`
		}
		systemMsg = map[string]interface{}{"role": "system", "content": prompt}
		systemTokens = estimateTokens(prompt)
	}

	// 上下文预算
	maxInput := cfg.MaxTokens * 4
	if maxInput < 4000 {
		maxInput = 4000
	}
	available := maxInput - systemTokens - 200

	sess.mu.Lock()
	raw := make([]map[string]interface{}, len(sess.messages))
	copy(raw, sess.messages)
	sess.mu.Unlock()

	if len(raw) == 0 {
		return []map[string]interface{}{systemMsg}
	}

	// 从最新消息向前累加 token
	total := 0
	start := len(raw)
	for i := len(raw) - 1; i >= 0; i-- {
		tokens := estimateMessageTokens(raw[i])
		if total+tokens > available {
			break
		}
		total += tokens
		start = i
	}

	// 构建消息列表，确保 tool_calls 和 tool 结果配对
	var msgs []map[string]interface{}
	if systemMsg != nil {
		msgs = append(msgs, systemMsg)
	}
	for i := start; i < len(raw); i++ {
		msgs = append(msgs, raw[i])
	}

	// 移除开头孤立的 tool 或 assistant+tool_calls 消息
	for len(msgs) > 0 {
		role, _ := msgs[0]["role"].(string)
		if role == "tool" {
			msgs = msgs[1:]
		} else if role == "assistant" {
			if _, hasTC := msgs[0]["tool_calls"]; hasTC {
				// assistant+tool_calls 如果后面没有对应的 tool 结果，则跳过
				if len(msgs) > 1 {
					nextRole, _ := msgs[1]["role"].(string)
					if nextRole == "tool" {
						break // 有配对，保留
					}
				}
				msgs = msgs[1:] // 无配对，跳过
			} else {
				break
			}
		} else {
			break
		}
	}

	// 末尾裁剪：如果最后是 assistant+tool_calls 但没有 tool 结果，移除
	for len(msgs) > 0 {
		last := msgs[len(msgs)-1]
		role, _ := last["role"].(string)
		if role == "assistant" {
			if _, hasTC := last["tool_calls"]; hasTC {
				msgs = msgs[:len(msgs)-1]
				continue
			}
		}
		break
	}

	return msgs
}

// --- API 调用（流式） ---

// streamResult 流式返回的结果
type streamResult struct {
	Content          string
	ReasoningContent string
	ToolCalls        []interface{}
	Annotations      []interface{} // 联网搜索引用
	FinishReason     string
	Error            error
}

// callAPI 兼容旧调用：内部使用流式，返回聚合结果
func (s *AIService) callAPI(cfg *AIConfig, body map[string]interface{}, connID string) (map[string]interface{}, error) {
	result := s.callAPIStream(cfg, body, connID)
	if result.Error != nil {
		return nil, result.Error
	}

	// 构造等价的非流式响应格式
	msg := map[string]interface{}{
		"role":             "assistant",
		"content":          result.Content,
		"reasoning_content": result.ReasoningContent,
	}
	if len(result.ToolCalls) > 0 {
		msg["tool_calls"] = result.ToolCalls
	}
	if len(result.Annotations) > 0 {
		msg["annotations"] = result.Annotations
	}
	resp := map[string]interface{}{
		"choices": []interface{}{
			map[string]interface{}{
				"message":      msg,
				"finish_reason": result.FinishReason,
			},
		},
	}
	return resp, nil
}

// callAPIStream 流式调用 API，通过 SSE 逐步返回内容
func (s *AIService) callAPIStream(cfg *AIConfig, body map[string]interface{}, connID string) *streamResult {
	endpoint := buildEndpoint(cfg.APIEndpoint)
	body["stream"] = true

	jsonData, err := json.Marshal(body)
	if err != nil {
		return &streamResult{Error: fmt.Errorf("序列化请求失败: %v", err)}
	}
	fmt.Printf("[AI] 流式请求: %s, 模型: %v\n", endpoint, body["model"])

	client := &http.Client{} // 无超时，由用户手动取消

	for attempt := 1; attempt <= 3; attempt++ {
		req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(jsonData))
		if err != nil {
			return &streamResult{Error: fmt.Errorf("创建请求失败: %v", err)}
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", cfg.APIKey))
		req.Header.Set("api-key", cfg.APIKey)

		resp, err := client.Do(req)
		if err != nil {
			if attempt < 3 {
				time.Sleep(time.Duration(attempt*2) * time.Second)
				continue
			}
			return &streamResult{Error: fmt.Errorf("请求失败: %v", err)}
		}

		if resp.StatusCode != http.StatusOK {
			bodyBytes, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			if resp.StatusCode == 429 && attempt < 3 {
				time.Sleep(time.Duration(attempt*2) * time.Second)
				continue
			}
			return &streamResult{Error: fmt.Errorf("API 错误 (状态码 %d): %s", resp.StatusCode, string(bodyBytes))}
		}

		// 解析 SSE 流
		result := s.parseSSEStream(resp, connID)
		return result
	}

	return &streamResult{Error: fmt.Errorf("达到最大重试次数")}
}

// parseSSEStream 解析 SSE 流式响应
func (s *AIService) parseSSEStream(resp *http.Response, connID string) *streamResult {
	defer resp.Body.Close()

	result := &streamResult{}
	toolCallsMap := map[int]map[string]interface{}{} // index → merged tool_call
	scanner := bufio.NewScanner(resp.Body)

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" || !strings.HasPrefix(line, "data: ") {
			continue
		}

		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			break
		}

		var chunk map[string]interface{}
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			continue
		}

		choices, ok := chunk["choices"].([]interface{})
		if !ok || len(choices) == 0 {
			continue
		}
		choice, ok := choices[0].(map[string]interface{})
		if !ok {
			continue
		}
		delta, ok := choice["delta"].(map[string]interface{})
		if !ok {
			continue
		}

		// 提取 reasoning_content
		if rc, ok := delta["reasoning_content"].(string); ok && rc != "" {
			result.ReasoningContent += rc
			if s.app != nil {
				s.app.Event.Emit("ai:stream-chunk", map[string]interface{}{
					"connId": connID, "type": "reasoning", "content": rc,
				})
			}
		}

		// 提取 content
		if c, ok := delta["content"].(string); ok && c != "" {
			result.Content += c
			if s.app != nil {
				s.app.Event.Emit("ai:stream-chunk", map[string]interface{}{
					"connId": connID, "type": "content", "content": c,
				})
			}
		}

		// 提取 tool_calls（按 index 合并分片）
		if tc, ok := delta["tool_calls"].([]interface{}); ok {
			for _, item := range tc {
				tcChunk, ok := item.(map[string]interface{})
				if !ok {
					continue
				}
				idx, _ := tcChunk["index"].(float64)
				i := int(idx)

				if existing, exists := toolCallsMap[i]; exists {
					// 合并：更新 function.arguments
					if fn, ok := tcChunk["function"].(map[string]interface{}); ok {
						if existingFn, ok := existing["function"].(map[string]interface{}); ok {
							if args, ok := fn["arguments"].(string); ok {
								existingArgs, _ := existingFn["arguments"].(string)
								existingFn["arguments"] = existingArgs + args
							}
							if name, ok := fn["name"].(string); ok && name != "" {
								existingFn["name"] = name
							}
						}
					}
					// 合并其他字段
					if id, ok := tcChunk["id"].(string); ok && id != "" {
						existing["id"] = id
					}
					if typ, ok := tcChunk["type"].(string); ok && typ != "" {
						existing["type"] = typ
					}
				} else {
					// 新的 tool_call
					newTC := make(map[string]interface{})
					for k, v := range tcChunk {
						newTC[k] = v
					}
					// 确保 function.arguments 是字符串
					if fn, ok := newTC["function"].(map[string]interface{}); ok {
						if args, ok := fn["arguments"].(string); ok {
							fn["arguments"] = args
						} else {
							fn["arguments"] = ""
						}
					}
					toolCallsMap[i] = newTC
				}
			}
		}

		// 提取 annotations
		if ann, ok := delta["annotations"].([]interface{}); ok && len(ann) > 0 {
			result.Annotations = append(result.Annotations, ann...)
		}

		// 提取 finish_reason
		if fr, ok := choice["finish_reason"].(string); ok && fr != "" {
			result.FinishReason = fr
		}
	}

	// 将合并后的 tool_calls 转为有序数组
	if len(toolCallsMap) > 0 {
		maxIdx := 0
		for idx := range toolCallsMap {
			if idx > maxIdx {
				maxIdx = idx
			}
		}
		result.ToolCalls = make([]interface{}, 0, len(toolCallsMap))
		for i := 0; i <= maxIdx; i++ {
			if tc, ok := toolCallsMap[i]; ok {
				result.ToolCalls = append(result.ToolCalls, tc)
			}
		}
	}

	return result
}

func (s *AIService) parseChoice(resp map[string]interface{}) (map[string]interface{}, error) {
	choices, ok := resp["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		return nil, fmt.Errorf("响应中无 choices")
	}
	choice, ok := choices[0].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("choices 格式错误")
	}
	return choice, nil
}

// --- 工具审批/执行 ---

func (s *AIService) waitForApproval(connID, callID, tool, command string, cs *connState) bool {
	ch := make(chan bool, 1)
	s.mu.Lock()
	s.approvals[callID] = ch
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		delete(s.approvals, callID)
		s.mu.Unlock()
	}()

	if s.app != nil {
		s.app.Event.Emit("ai:tool-approval", map[string]interface{}{
			"connId": connID, "callId": callID, "tool": tool, "command": command,
		})
	}

	cs.mu.Lock()
	cancelCh := cs.cancel
	cs.mu.Unlock()

	select {
	case ok := <-ch:
		return ok
	case <-cancelCh:
		return false
	}
}

func (s *AIService) ApproveTool(callID string) {
	s.mu.RLock()
	ch, ok := s.approvals[callID]
	s.mu.RUnlock()
	if ok {
		ch <- true
	}
}

func (s *AIService) DenyTool(callID string) {
	s.mu.RLock()
	ch, ok := s.approvals[callID]
	s.mu.RUnlock()
	if ok {
		ch <- false
	}
}

func (s *AIService) executeViaFrontend(connID, callID, tool, command, argsJSON string, cs *connState) string {
	ch := make(chan string, 1)
	s.mu.Lock()
	s.results[callID] = ch
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		delete(s.results, callID)
		s.mu.Unlock()
	}()

	if s.app != nil {
		s.app.Event.Emit("ai:execute-tool", map[string]interface{}{
			"connId": connID, "callId": callID, "tool": tool, "command": command, "args": argsJSON,
		})
	}

	cs.mu.Lock()
	cancelCh := cs.cancel
	cs.mu.Unlock()

	select {
	case r := <-ch:
		return r
	case <-cancelCh:
		return "用户取消了操作"
	}
}

func (s *AIService) SubmitToolResult(callID, result string) {
	s.mu.RLock()
	ch, ok := s.results[callID]
	s.mu.RUnlock()
	if ok {
		ch <- result
	}
}

// --- 事件发送 ---

func (s *AIService) emitStatus(connID, status, text string) {
	if s.app != nil {
		s.app.Event.Emit("ai:status", map[string]interface{}{
			"connId": connID, "status": status, "text": text,
		})
	}
}

func (s *AIService) emitError(connID, key string, sess *session, msg string) {
	errMsg := map[string]interface{}{
		"role":    "assistant",
		"content": fmt.Sprintf("❌ %s", msg),
	}
	sess.mu.Lock()
	sess.messages = append(sess.messages, errMsg)
	sess.mu.Unlock()

	aiMsg := &ChatMessage{
		ID:        fmt.Sprintf("msg_%d", time.Now().UnixNano()),
		Role:      "assistant",
		Content:   errMsg["content"].(string),
		Timestamp: time.Now(),
	}
	if s.app != nil {
		s.app.Event.Emit("ai:message", map[string]interface{}{
			"connId": connID, "message": aiMsg,
		})
	}
	s.logger.Log(key, "错误: %s", msg)
}

func (s *AIService) emitToolResult(connID, callID, tool, command, result string) {
	if s.app != nil {
		s.app.Event.Emit("ai:tool-result", map[string]interface{}{
			"connId": connID, "callId": callID, "tool": tool, "command": command, "result": result,
		})
	}
}

// --- 历史管理 ---

func (s *AIService) GetChatHistory(connID string) []*ChatMessage {
	key := s.resolveKey(connID)
	sess := s.getOrCreateSession(key)

	sess.mu.Lock()
	defer sess.mu.Unlock()

	// 如果内存中没有，从磁盘加载
	if len(sess.messages) == 0 && s.store != nil {
		msgs, err := s.store.LoadMessages(key)
		if err == nil && msgs != nil {
			sess.messages = msgs
		}
	}

	var result []*ChatMessage
	for _, m := range sess.messages {
		role, _ := m["role"].(string)
		content, _ := m["content"].(string)
		if role == "user" || role == "assistant" {
			result = append(result, &ChatMessage{
				ID:        fmt.Sprintf("msg_%d", len(result)),
				Role:      role,
				Content:   content,
				Timestamp: time.Now(),
			})
		}
	}
	return result
}

func (s *AIService) ClearChatHistory(connID string) {
	key := s.resolveKey(connID)
	sess := s.getOrCreateSession(key)

	sess.mu.Lock()
	sess.messages = nil
	sess.mu.Unlock()

	if s.store != nil {
		s.store.DeleteMessages(key)
	}

	if s.app != nil {
		s.app.Event.Emit("ai:chat-cleared", map[string]interface{}{"connId": connID})
	}
	s.logger.Log(key, "清除聊天历史")
}

func (s *AIService) persistMessages(key string, sess *session) {
	if s.store == nil {
		return
	}
	sess.mu.Lock()
	msgs := make([]map[string]interface{}, len(sess.messages))
	copy(msgs, sess.messages)
	sess.mu.Unlock()

	if err := s.store.SaveMessages(key, msgs); err != nil {
		s.logger.Log(key, "持久化失败: %v", err)
	}
}

// --- 辅助 ---

func (s *AIService) resolveKey(connID string) string {
	if s.serverInfo != nil {
		if key := s.serverInfo.GetServerKey(connID); key != "" {
			return key
		}
	}
	return connID
}

func (s *AIService) getOrCreateSession(key string) *session {
	s.mu.Lock()
	defer s.mu.Unlock()

	if sess, ok := s.sessions[key]; ok {
		return sess
	}
	sess := &session{}
	s.sessions[key] = sess

	// 从磁盘加载历史
	if s.store != nil {
		msgs, err := s.store.LoadMessages(key)
		if err == nil && msgs != nil {
			sess.messages = msgs
		}
	}

	return sess
}

// getOrCreateConnState 获取或创建连接独立状态
func (s *AIService) getOrCreateConnState(connID string) *connState {
	s.mu.Lock()
	defer s.mu.Unlock()

	if cs, ok := s.conns[connID]; ok {
		return cs
	}
	cs := &connState{}
	s.conns[connID] = cs
	return cs
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}

func estimateTokens(text string) int {
	if text == "" {
		return 0
	}
	ascii, nonAscii := 0, 0
	for _, r := range text {
		if r < 128 {
			ascii++
		} else {
			nonAscii++
		}
	}
	return ascii/4 + int(float64(nonAscii)*1.5)
}

func estimateMessageTokens(msg map[string]interface{}) int {
	total := 4
	if c, ok := msg["content"].(string); ok {
		total += estimateTokens(c)
	}
	if tc, ok := msg["tool_calls"].([]interface{}); ok {
		for _, t := range tc {
			if m, ok := t.(map[string]interface{}); ok {
				if fn, ok := m["function"].(map[string]interface{}); ok {
					if args, ok := fn["arguments"].(string); ok {
						total += estimateTokens(args)
					}
				}
			}
		}
	}
	return total
}

// isCancelled 检查是否已取消
func (s *AIService) isCancelled(cs *connState) bool {
	cs.mu.Lock()
	cancelCh := cs.cancel
	cs.mu.Unlock()

	if cancelCh == nil {
		return false
	}

	select {
	case <-cancelCh:
		return true
	default:
		return false
	}
}

// buildEndpoint 构建完整的 API 端点 URL
// 支持用户填 base URL（如 /v1）或完整 URL（如 /v1/chat/completions）
func buildEndpoint(apiEndpoint string) string {
	endpoint := strings.TrimRight(apiEndpoint, "/")

	// 已经是完整路径，直接使用
	if strings.HasSuffix(endpoint, "/chat/completions") {
		return endpoint
	}

	// 已经包含 /v1/chat 等路径，追加 /completions
	if strings.Contains(endpoint, "/chat") {
		return endpoint + "/completions"
	}

	// base URL，追加完整路径
	return endpoint + "/chat/completions"
}

