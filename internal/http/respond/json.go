package respond

import (
	"encoding/json"
	"net/http"
)

func JSON(w http.ResponseWriter, status int, body any) {
	Write(w, status, "application/json", body)
}

func Write(w http.ResponseWriter, status int, contentType string, body any) {
	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(status)
	// Headers are already committed, so interrupted write cannot be retried
	_ = json.NewEncoder(w).Encode(body)
}
