package ssh

import (
	"fmt"
	"strings"
	"sync"
)

// BatchCommand 批量命令
type BatchCommand struct {
	Command  string `json:"command"`  // 命令模板（支持 {{host}}, {{user}}, {{port}} 等变量）
	Targets  []string `json:"targets"` // 目标连接 ID 列表
}

// BatchResult 批量执行结果
type BatchResult struct {
	ConnID  string `json:"connId"`
	Host    string `json:"host"`
	Command string `json:"command"`
	Output  string `json:"output"`
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

// TemplateVars 模板变量
type TemplateVars struct {
	Host     string
	Port     int
	Username string
	Name     string
}

// ExpandTemplate 展开命令模板变量
// 支持的变量：{{host}}, {{port}}, {{user}}, {{name}}
func ExpandTemplate(template string, vars TemplateVars) string {
	result := template
	result = strings.ReplaceAll(result, "{{host}}", vars.Host)
	result = strings.ReplaceAll(result, "{{port}}", fmt.Sprintf("%d", vars.Port))
	result = strings.ReplaceAll(result, "{{user}}", vars.Username)
	result = strings.ReplaceAll(result, "{{name}}", vars.Name)
	return result
}

// ExecuteBatch 批量执行命令
func (s *SSHService) ExecuteBatch(commands []BatchCommand) []BatchResult {
	var results []BatchResult
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, batch := range commands {
		for _, connID := range batch.Targets {
			wg.Add(1)
			go func(connID, cmdTmpl string) {
				defer wg.Done()

				// 获取连接信息用于模板变量
				s.mu.RLock()
				client, ok := s.clients[connID]
				var vars TemplateVars
				if ok {
					vars = TemplateVars{
						Host:     client.config.Host,
						Port:     client.config.Port,
						Username: client.config.Username,
						Name:     client.config.Name,
					}
				}
				s.mu.RUnlock()

				if !ok {
					mu.Lock()
					results = append(results, BatchResult{
						ConnID:  connID,
						Command: cmdTmpl,
						Success: false,
						Error:   "连接不存在",
					})
					mu.Unlock()
					return
				}

				// 展开模板变量
				cmd := ExpandTemplate(cmdTmpl, vars)

				// 执行命令
				result, err := client.ExecuteCommand(cmd)

				batchResult := BatchResult{
					ConnID:  connID,
					Host:    vars.Host,
					Command: cmd,
				}

				if err != nil {
					batchResult.Success = false
					batchResult.Error = err.Error()
				} else if result != nil {
					batchResult.Success = result.Success
					batchResult.Output = result.Stdout
					if !result.Success {
						batchResult.Error = result.Stderr
					}
				}

				// 记录审计日志
				s.auditLogger.Log(connID, vars.Host, vars.Username, "batch-command", cmd, batchResult.Success, batchResult.Error)

				mu.Lock()
				results = append(results, batchResult)
				mu.Unlock()
			}(connID, batch.Command)
		}
	}

	wg.Wait()
	return results
}

// CompareResults 对比批量执行结果
func CompareResults(results []BatchResult) map[string][]BatchResult {
	grouped := make(map[string][]BatchResult)
	for _, r := range results {
		key := r.Command
		grouped[key] = append(grouped[key], r)
	}
	return grouped
}
