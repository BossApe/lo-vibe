package handler

import "net/http"

// HealthHandler は GET /health を処理する
func HealthHandler(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
