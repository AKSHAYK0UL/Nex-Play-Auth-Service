package response

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

type Envelope struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
	Message string `json:"message,omitempty"`
	Data    any    `json:"data,omitempty"`
}

// Helper
func write(w http.ResponseWriter, status int, v any) {

	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(v); err != nil {

		slog.Error("failed to encode response", "error", err)
	}
}

// JSON writes a successful response with Payload
func JSON(w http.ResponseWriter, status int, data any) {

	write(w, status, Envelope{Success: true, Data: data})
}

// Error writes an error response.
func Error(w http.ResponseWriter, status int, msg string) {

	write(w, status, Envelope{Success: false, Message: msg})
}

// Message writes a successful response with a plain message
func Message(w http.ResponseWriter, status int, msg string) {

	write(w, status, Envelope{Success: true, Message: msg})
}
