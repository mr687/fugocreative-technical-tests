package response

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

type apiError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
}

type errorResponse struct {
	Error apiError `json:"error"`
}

func Success(w http.ResponseWriter, status int, payload any) {
	JSON(w, status, payload)
}

func Error(w http.ResponseWriter, status int, code, message string, details any) {
	JSON(w, status, errorResponse{Error: apiError{Code: code, Message: message, Details: details}})
}

func JSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		slog.Error("write JSON response failed", "error", err)
	}
}
