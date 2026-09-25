package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

var (
	priority  = os.Getenv("ROUTER_PRIORITY")
	routerURL = os.Getenv("ROUTER_ENDPOINT")
	client    = &http.Client{Timeout: 120 * time.Second}
)

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func forward(w http.ResponseWriter, req *http.Request) {
	body, err := io.ReadAll(io.LimitReader(req.Body, 1<<20))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
		return
	}
	messages, _ := payload["messages"].([]any)
	prompt, _ := payload["prompt"].(string)
	if len(messages) == 0 && prompt == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "messages or prompt is required"})
		return
	}
	if len(messages) == 0 {
		payload["messages"] = []map[string]string{{"role": "user", "content": prompt}}
	}
	proxyBody, err := json.Marshal(payload)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}
	endpoint := strings.TrimRight(routerURL, "/") + "/v1/chat/completions"
	proxyReq, err := http.NewRequestWithContext(req.Context(), http.MethodPost, endpoint, bytes.NewReader(proxyBody))
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": "invalid router endpoint"})
		return
	}
	proxyReq.Header.Set("Content-Type", "application/json")
	proxyReq.Header.Set("X-Router-Priority", priority)
	resp, err := client.Do(proxyReq)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error()})
		return
	}
	defer resp.Body.Close()
	responseBody, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": "invalid router response"})
		return
	}
	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/json"
	}
	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(resp.StatusCode)
	_, _ = w.Write(responseBody)
}

func health(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

func main() {
	if priority == "" {
		priority = "cost-optimized"
	}
	if routerURL == "" {
		routerURL = "http://finops-arbitrage-router.finops-arbitrage-router.svc.cluster.local:8080"
	}
	http.HandleFunc("/v1/chat/completions", forward)
	http.HandleFunc("/chat", forward)
	http.HandleFunc("/healthz", health)
	log.Printf("managed-llm-service listening on :8080 with priority %s", priority)
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
