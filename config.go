package main

import (
	"fmt"
	"os"
	"strconv"
)

// Config holds the application configuration from environment variables
type Config struct {
	// WeChat Enterprise configuration
	CorpID      string // WECHAT_CORP_ID - 企业ID
	AgentID     int    // WECHAT_AGENT_ID - 应用ID
	AgentSecret string // WECHAT_AGENT_SECRET - 应用Secret

	// Message recipients
	ToUser  string // WECHAT_TO_USER - 接收用户，默认 @all
	ToParty string // WECHAT_TO_PARTY - 接收部门
	ToTag   string // WECHAT_TO_TAG - 接收标签

	// Server configuration
	ServerPort string // SERVER_PORT - 服务监听端口，默认 8080
}

// LoadConfig loads configuration from environment variables
func LoadConfig() (*Config, error) {
	config := &Config{}

	// Required fields
	config.CorpID = os.Getenv("WECHAT_CORP_ID")
	if config.CorpID == "" {
		return nil, fmt.Errorf("WECHAT_CORP_ID is required")
	}

	agentIDStr := os.Getenv("WECHAT_AGENT_ID")
	if agentIDStr == "" {
		return nil, fmt.Errorf("WECHAT_AGENT_ID is required")
	}
	agentID, err := strconv.Atoi(agentIDStr)
	if err != nil {
		return nil, fmt.Errorf("WECHAT_AGENT_ID must be a valid integer: %v", err)
	}
	config.AgentID = agentID

	config.AgentSecret = os.Getenv("WECHAT_AGENT_SECRET")
	if config.AgentSecret == "" {
		return nil, fmt.Errorf("WECHAT_AGENT_SECRET is required")
	}

	// Optional fields with defaults
	config.ToUser = os.Getenv("WECHAT_TO_USER")
	if config.ToUser == "" {
		config.ToUser = "@all"
	}

	config.ToParty = os.Getenv("WECHAT_TO_PARTY")
	config.ToTag = os.Getenv("WECHAT_TO_TAG")

	config.ServerPort = os.Getenv("SERVER_PORT")
	if config.ServerPort == "" {
		config.ServerPort = "8080"
	}

	return config, nil
}
