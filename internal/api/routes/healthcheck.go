package routes

import (
	"encoding/json"
	"net/http"
)

type HealthCheckHandler struct {
}

func (h *HealthCheckHandler) Ping(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]string{"message": "pong"})
}
