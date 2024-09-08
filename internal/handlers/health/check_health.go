package health

import (
	"encoding/json"
	"net/http"
)

func checkHealth(w http.ResponseWriter, r *http.Request) {
	var healthCheckResponse = struct {
		Success bool `json:"success"`
		Status  string
	}{
		Success: true,
		Status:  "Still rockin'!",
	}

	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(200)

	json.NewEncoder(w).Encode(healthCheckResponse)
}
