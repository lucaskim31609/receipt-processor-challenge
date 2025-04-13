package main

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

// Helper to write JSON responses
func jsonResponse(w http.ResponseWriter, status int, data interface{}, logger *slog.Logger) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		logger.Error("Failed to encode JSON response", slog.Any("error", err))
	}
}

// Helper to write standard error messages
func errorResponse(w http.ResponseWriter, status int, message string, logger *slog.Logger) {
	type ErrorMsg struct {
		Error string `json:"error"`
	}
	logger.Warn("Responding with error", slog.Int("status", status), slog.String("message", message))
	jsonResponse(w, status, ErrorMsg{Error: message}, logger)
}