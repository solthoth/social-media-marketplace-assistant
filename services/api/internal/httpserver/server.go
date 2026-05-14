package httpserver

import (
	"encoding/json"
	"net/http"
	"time"
)

type healthResponse struct {
	Status  string `json:"status"`
	Service string `json:"service"`
	Time    string `json:"time"`
}

func NewRouter() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", healthz)
	mux.HandleFunc("GET /", root)
	return mux
}

func root(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"service": "social-media-marketplace-assistant-api",
		"status":  "ok",
	})
}

func healthz(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, healthResponse{
		Status:  "ok",
		Service: "social-media-marketplace-assistant-api",
		Time:    time.Now().UTC().Format(time.RFC3339),
	})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}
