package health

import (
	"encoding/json"
	"net/http"
)

func Check_Health(w http.ResponseWriter, r *http.Request) {
	var healthCheckResponse = struct {
		Status string
	}{
		Status: "Still rockin'!",
	}
	w.WriteHeader(200)
	json.NewEncoder(w).Encode(healthCheckResponse)
}
