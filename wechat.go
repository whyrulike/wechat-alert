package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

const (
	// WeChat API endpoints
	getTokenURL   = "https://qyapi.weixin.qq.com/cgi-bin/gettoken"
	sendMessageURL = "https://qyapi.weixin.qq.com/cgi-bin/message/send"

	// Token refresh buffer (refresh 5 minutes before expiry)
	tokenRefreshBuffer = 300
)

// WeChatClient handles communication with WeChat Enterprise API
type WeChatClient struct {
	corpID      string
	agentID     int
	agentSecret string
	toUser      string
	toParty     string
	toTag       string

	// Token cache
	accessToken string
	tokenExpiry time.Time
	tokenMutex  sync.Mutex
}

// TokenResponse represents the response from gettoken API
type TokenResponse struct {
	ErrCode     int    `json:"errcode"`
	ErrMsg      string `json:"errmsg"`
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}

// SendMessageRequest represents the request body for sending messages
type SendMessageRequest struct {
	ToUser  string      `json:"touser"`
	ToParty string      `json:"toparty"`
	ToTag   string      `json:"totag"`
	MsgType string      `json:"msgtype"`
	AgentID int         `json:"agentid"`
	Text    TextContent `json:"text"`
}

// TextContent represents the text message content
type TextContent struct {
	Content string `json:"content"`
}

// SendMessageResponse represents the response from send message API
type SendMessageResponse struct {
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}

// NewWeChatClient creates a new WeChat client
func NewWeChatClient(config *Config) *WeChatClient {
	return &WeChatClient{
		corpID:      config.CorpID,
		agentID:     config.AgentID,
		agentSecret: config.AgentSecret,
		toUser:      config.ToUser,
		toParty:     config.ToParty,
		toTag:       config.ToTag,
	}
}

// GetAccessToken returns a valid access token, refreshing if necessary
func (c *WeChatClient) GetAccessToken() (string, error) {
	c.tokenMutex.Lock()
	defer c.tokenMutex.Unlock()

	// Check if token is still valid (with buffer time)
	if c.accessToken != "" && time.Now().Add(tokenRefreshBuffer*time.Second).Before(c.tokenExpiry) {
		return c.accessToken, nil
	}

	// Refresh token
	log.Println("Refreshing access_token...")

	url := fmt.Sprintf("%s?corpid=%s&corpsecret=%s", getTokenURL, c.corpID, c.agentSecret)
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to get access_token: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %v", err)
	}

	var tokenResp TokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return "", fmt.Errorf("failed to parse token response: %v", err)
	}

	if tokenResp.ErrCode != 0 {
		return "", fmt.Errorf("WeChat API error: %d - %s", tokenResp.ErrCode, tokenResp.ErrMsg)
	}

	c.accessToken = tokenResp.AccessToken
	c.tokenExpiry = time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)

	log.Printf("Access token refreshed successfully, expires in %d seconds", tokenResp.ExpiresIn)
	return c.accessToken, nil
}

// SendMessage sends a text message to WeChat Enterprise
func (c *WeChatClient) SendMessage(content string) error {
	token, err := c.GetAccessToken()
	if err != nil {
		return err
	}

	reqBody := SendMessageRequest{
		ToUser:  c.toUser,
		ToParty: c.toParty,
		ToTag:   c.toTag,
		MsgType: "text",
		AgentID: c.agentID,
		Text: TextContent{
			Content: content,
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %v", err)
	}

	url := fmt.Sprintf("%s?access_token=%s", sendMessageURL, token)
	resp, err := http.Post(url, "application/json", strings.NewReader(string(jsonData)))
	if err != nil {
		return fmt.Errorf("failed to send message: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %v", err)
	}

	var sendResp SendMessageResponse
	if err := json.Unmarshal(body, &sendResp); err != nil {
		return fmt.Errorf("failed to parse send response: %v", err)
	}

	if sendResp.ErrCode != 0 {
		return fmt.Errorf("WeChat send message error: %d - %s", sendResp.ErrCode, sendResp.ErrMsg)
	}

	log.Println("Message sent successfully")
	return nil
}
