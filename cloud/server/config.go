package server

import "changeme/cloud/server/crypto"

// Config 云端配置
type Config struct {
	Port  int    `json:"port"`
	Token string `json:"token"` // 原始令牌
}

// DefaultConfig 默认配置
func DefaultConfig() *Config {
	return &Config{
		Port:  9527,
		Token: crypto.GenerateToken(),
	}
}
