package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

// GrafanaWebhook represents the Grafana v9+ alert webhook payload
type GrafanaWebhook struct {
	Receiver          string            `json:"receiver"`
	Status            string            `json:"status"`
	Alerts            []Alert           `json:"alerts"`
	GroupLabels       map[string]string `json:"groupLabels"`
	CommonLabels      map[string]string `json:"commonLabels"`
	CommonAnnotations map[string]string `json:"commonAnnotations"`
	ExternalURL       string            `json:"externalURL"`
}

// Alert represents a single alert in the webhook
type Alert struct {
	Status       string            `json:"status"`
	Labels       map[string]string `json:"labels"`
	Annotations  map[string]string `json:"annotations"`
	StartsAt     string            `json:"startsAt"`
	EndsAt       string            `json:"endsAt"`
	GeneratorURL string            `json:"generatorURL"`
}

// WebhookHandler handles incoming Grafana webhook requests
type WebhookHandler struct {
	wechatClient *WeChatClient
}

// NewWebhookHandler creates a new webhook handler
func NewWebhookHandler(client *WeChatClient) *WebhookHandler {
	return &WebhookHandler{
		wechatClient: client,
	}
}

// ServeHTTP handles the webhook POST request
func (h *WebhookHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Read the request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading request body: %v", err)
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Log the raw webhook JSON for debugging
	log.Printf("Received Grafana webhook: %s", string(body))

	// Parse the webhook payload
	var webhook GrafanaWebhook
	if err := json.Unmarshal(body, &webhook); err != nil {
		log.Printf("Error parsing webhook JSON: %v", err)
		http.Error(w, "Failed to parse webhook JSON", http.StatusBadRequest)
		return
	}

	// Format and send message
	message := h.formatMessage(&webhook)
	if err := h.wechatClient.SendMessage(message); err != nil {
		log.Printf("Error sending WeChat message: %v", err)
		http.Error(w, "Failed to send WeChat message", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}

// formatMessage formats the Grafana webhook into a WeChat text message
func (h *WebhookHandler) formatMessage(webhook *GrafanaWebhook) string {
	var messages []string

	for _, alert := range webhook.Alerts {
		msg := h.formatAlert(&alert)
		messages = append(messages, msg)
	}

	return strings.Join(messages, "\n————————————\n")
}

// formatAlert formats a single alert into text
func (h *WebhookHandler) formatAlert(alert *Alert) string {
	var sb strings.Builder

	sb.WriteString("[Grafana告警]\n")

	// Status
	sb.WriteString(fmt.Sprintf("状态: %s\n", alert.Status))

	// Alert name
	if alertName, ok := alert.Labels["alertname"]; ok {
		sb.WriteString(fmt.Sprintf("告警名称: %s\n", alertName))
	}

	// Severity
	if severity, ok := alert.Labels["severity"]; ok {
		sb.WriteString(fmt.Sprintf("级别: %s\n", severity))
	}

	// Instance
	if instance, ok := alert.Labels["instance"]; ok {
		sb.WriteString(fmt.Sprintf("实例: %s\n", instance))
	}

	// Summary
	if summary, ok := alert.Annotations["summary"]; ok {
		sb.WriteString(fmt.Sprintf("摘要: %s\n", summary))
	}

	// Description
	if description, ok := alert.Annotations["description"]; ok {
		sb.WriteString(fmt.Sprintf("详情: %s\n", description))
	}

	// Start time
	if alert.StartsAt != "" {
		startTime := h.formatTime(alert.StartsAt)
		sb.WriteString(fmt.Sprintf("开始时间: %s", startTime))
	}

	return sb.String()
}

// formatTime converts ISO 8601 time to a readable format
func (h *WebhookHandler) formatTime(isoTime string) string {
	t, err := time.Parse(time.RFC3339, isoTime)
	if err != nil {
		return isoTime
	}
	// Convert to Asia/Shanghai timezone
	loc, _ := time.LoadLocation("Asia/Shanghai")
	return t.In(loc).Format("2006-01-02 15:04:05")
}
