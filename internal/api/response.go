package api

import (
	"encoding/json"
	"net/http"
	"time"
)

// ProblemDetails implements RFC 7807 structured HTTP error schema.
type ProblemDetails struct {
	Type      string `json:"type"`
	Title     string `json:"title"`
	Status    int    `json:"status"`
	Detail    string `json:"detail"`
	Instance  string `json:"instance"`
	Code      string `json:"code"`
	Timestamp string `json:"timestamp"`
}

// WriteJSON writes a standard JSON response with HTTP status code.
func WriteJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

// WriteError writes an RFC 7807 error payload.
func WriteError(w http.ResponseWriter, r *http.Request, status int, code, title, detail string) {
	problem := ProblemDetails{
		Type:      "https://lunar-navigator.nasa/errors/" + code,
		Title:     title,
		Status:    status,
		Detail:    detail,
		Instance:  r.URL.Path,
		Code:      code,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
	WriteJSON(w, status, problem)
}
