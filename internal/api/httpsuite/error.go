package httpsuite

import (
	"encoding/json"
	"net/http"
)

// ErrorResponse documents the envelope of every failed request (Response without a body). Used by the Swagger annotations only.
type ErrorResponse struct {
	Message string `json:"message" example:"forbidden"`
	Code    int    `json:"code" example:"403"`
}

type APIError struct {
	Error string `json:"error"`
}

func WriteJSONError(w http.ResponseWriter, msg string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(APIError{Error: msg})
}
