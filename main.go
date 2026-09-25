package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
)

var (
	priority   = os.Getenv("ROUTER_PRIORITY")
	routerURL  = os.Getenv("ROUTER_ENDPOINT")
	localModel = os.Getenv("LOCAL_MODEL")
)

type chatRequest struct {
	Prompt string `json:"prompt"`
}
type chatReply struct {
	Reply   string `json:"reply"`
	Path    string `json:"path"`
	Model   string `json:"model"`
	Backend string `json:"backend"`
}

func handle(w http.ResponseWriter, req *http.Request) {
	var in chatRequest
	if err := json.NewDecoder(req.Body).Decode(&in); err != nil {
		http.Error(w, `{"error":"bad request"}`, 400)
		return
	}
	json.NewEncoder(w).Encode(chatReply{
		Reply:   "delegate to finops router: " + in.Prompt,
		Path:    priority,
		Model:   localModel,
		Backend: routerURL,
	})
}

func health(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

func main() {
	http.HandleFunc("/chat", handle)
	http.HandleFunc("/healthz", health)
	log.Println("managed-llm-service listening :8080 priority=" + priority)
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
