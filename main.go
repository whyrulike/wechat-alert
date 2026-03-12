package main

import (
	"encoding/json"
	"log"
	"net/http"
)

func main() {
	// Load configuration
	config, err := LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Create WeChat client
	wechatClient := NewWeChatClient(config)

	// Create webhook handler
	webhookHandler := NewWebhookHandler(wechatClient)

	// Register routes
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	http.Handle("/webhook", webhookHandler)

	// Start server
	addr := ":" + config.ServerPort
	log.Printf("Starting wechat-alert server on %s", addr)
	log.Printf("  - Health check: GET /health")
	log.Printf("  - Webhook endpoint: POST /webhook")

	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
