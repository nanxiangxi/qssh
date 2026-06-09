package ai

import (
	"encoding/json"
	"fmt"
)

// GetToolDefinitions 返回 OpenAI function calling 格式的工具定义
func GetToolDefinitions() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"type": "function",
			"function": map[string]interface{}{
				"name":        "execute_command",
				"description": "在终端执行单条 shell 命令并返回结果。自动复用已有终端，没有则创建。适用于单个任务。如果需要同时执行多条互不依赖的命令，请使用 multi_execute 而非本工具。注意：对于 ping 等持续运行的命令，必须加 -c 参数限制次数（如 ping -c 4 baidu.com），否则会超时。",
				"parameters": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"command": map[string]interface{}{
							"type":        "string",
							"description": "要执行的完整 shell 命令",
						},
						"targetTerminal": map[string]interface{}{
							"type":        "string",
							"description": "指定在哪个终端执行，填终端名称（如\"终端 1\"、\"AI 终端-1\"）。留空则自动使用 AI 终端。",
						},
					},
					"required": []string{"command"},
				},
			},
		},
		{
			"type": "function",
			"function": map[string]interface{}{
				"name":        "get_server_info",
				"description": "获取当前 SSH 服务器的基本信息（主机名、系统、内核版本、运行时间）。无需参数。",
				"parameters": map[string]interface{}{
					"type":       "object",
					"properties": map[string]interface{}{},
				},
			},
		},
		{
			"type": "function",
			"function": map[string]interface{}{
				"name":        "get_system_status",
				"description": "获取服务器实时状态（CPU、内存、磁盘、负载）。无需参数。",
				"parameters": map[string]interface{}{
					"type":       "object",
					"properties": map[string]interface{}{},
				},
			},
		},
		{
			"type": "function",
			"function": map[string]interface{}{
				"name":        "open_terminal",
				"description": "打开一个新的终端面板。仅用于需要长期保持的终端。对于一次性并行任务，请直接使用 multi_execute（它会自动创建和关闭终端），不要先调用 open_terminal。",
				"parameters": map[string]interface{}{
					"type":       "object",
					"properties": map[string]interface{}{},
				},
			},
		},
		{
			"type": "function",
			"function": map[string]interface{}{
				"name":        "close_terminal",
				"description": "已废弃，不要调用。终端由用户手动关闭，AI 不再自动关闭终端。",
				"parameters": map[string]interface{}{
					"type":       "object",
					"properties": map[string]interface{}{},
				},
			},
		},
		{
			"type": "function",
			"function": map[string]interface{}{
				"name":        "multi_execute",
				"description": "【并行任务首选工具】在多个终端中同时执行不同命令。自动复用已有终端，只在不够时补创新终端，执行后保留所有终端供后续使用（不自动关闭）。当你需要同时做多件事时（如同时 ping 两个网站、同时查看磁盘和进程），必须使用此工具而非多次调用 execute_command。无需先调用 open_terminal。注意：持续运行的命令（如 ping）必须加 -c 参数限制次数。",
				"parameters": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"commands": map[string]interface{}{
							"type":        "array",
							"description": "要并行执行的命令列表，每个命令在独立终端运行",
							"items": map[string]interface{}{
								"type": "object",
								"properties": map[string]interface{}{
									"label": map[string]interface{}{
										"type":        "string",
										"description": "命令的简短描述（如 '检查磁盘'、'查看进程'）",
									},
									"command": map[string]interface{}{
										"type":        "string",
										"description": "要执行的 shell 命令",
									},
								},
								"required": []string{"command"},
							},
						},
					},
					"required": []string{"commands"},
				},
			},
		},
	}
}

// ParseToolArgs 用 encoding/json 解析工具参数
func ParseToolArgs(argsJSON string) map[string]string {
	result := make(map[string]string)
	if argsJSON == "" {
		return result
	}
	var raw map[string]interface{}
	if err := json.Unmarshal([]byte(argsJSON), &raw); err != nil {
		return result
	}
	for k, v := range raw {
		if s, ok := v.(string); ok {
			result[k] = s
		}
	}
	return result
}

// GetCommandDescription 获取工具调用的人类可读描述
func GetCommandDescription(toolName, argsJSON string) string {
	switch toolName {
	case "execute_command":
		args := ParseToolArgs(argsJSON)
		if cmd := args["command"]; cmd != "" {
			return cmd
		}
		return "(空命令)"
	case "get_server_info":
		return "hostname && uname -a && uptime && cat /proc/version"
	case "get_system_status":
		return "top -bn1 | head -5 && free -h && df -h && cat /proc/loadavg"
	case "open_terminal":
		return "打开终端面板"
	case "multi_execute":
		args := ParseToolArgs(argsJSON)
		if cmd := args["commands"]; cmd != "" {
			return fmt.Sprintf("并行执行: %s", cmd)
		}
		return "并行执行多条命令"
	default:
		return toolName
	}
}
