package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"musuhi-api/internal/service"
)

// SystemOverviewHandler は /api/v1/system-overviews のハンドラ
type SystemOverviewHandler struct {
	svc service.SystemOverviewService
}

// NewSystemOverviewHandler は SystemOverviewHandler を生成する
func NewSystemOverviewHandler(svc service.SystemOverviewService) *SystemOverviewHandler {
	return &SystemOverviewHandler{svc: svc}
}

type createSystemOverviewRequest struct {
	Content string `json:"content"`
}

type updateSystemOverviewRequest struct {
	Content string `json:"content"`
}

type systemOverviewResponse struct {
	ID        string `json:"id"`
	Content   string `json:"content"`
	CreatedAt string `json:"createdAt"`
}

type dataEnvelope struct {
	Data any `json:"data"`
}

type errorDetail struct {
	Field   string `json:"field,omitempty"`
	Message string `json:"message"`
}

type errorBody struct {
	Code    string        `json:"code"`
	Message string        `json:"message"`
	Details []errorDetail `json:"details,omitempty"`
}

type errorEnvelope struct {
	Error errorBody `json:"error"`
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, code, message string, details []errorDetail) {
	writeJSON(w, status, errorEnvelope{Error: errorBody{Code: code, Message: message, Details: details}})
}

// Create は POST /api/v1/system-overviews を処理する
func (h *SystemOverviewHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req createSystemOverviewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "リクエストの形式が不正です", nil)
		return
	}

	m, err := h.svc.Create(r.Context(), req.Content)
	if err != nil {
		if errors.Is(err, service.ErrValidation) {
			writeError(w, http.StatusUnprocessableEntity, "VALIDATION_ERROR", err.Error(),
				[]errorDetail{{Field: "content", Message: err.Error()}})
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "サーバーエラーが発生しました", nil)
		return
	}

	writeJSON(w, http.StatusCreated, dataEnvelope{Data: systemOverviewResponse{
		ID:        m.ID.String(),
		Content:   m.Content,
		CreatedAt: m.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}})
}

// Update は PUT /api/v1/system-overviews/{id} を処理する
func (h *SystemOverviewHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "id は必須です", nil)
		return
	}

	var req updateSystemOverviewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "リクエストの形式が不正です", nil)
		return
	}

	m, err := h.svc.Update(r.Context(), id, req.Content)
	if err != nil {
		if errors.Is(err, service.ErrValidation) {
			writeError(w, http.StatusUnprocessableEntity, "VALIDATION_ERROR", err.Error(),
				[]errorDetail{{Field: "content", Message: err.Error()}})
			return
		}
		if errors.Is(err, service.ErrNotFound) {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "指定されたIDのシステム概要が存在しません", nil)
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "サーバーエラーが発生しました", nil)
		return
	}

	writeJSON(w, http.StatusOK, dataEnvelope{Data: systemOverviewResponse{
		ID:        m.ID.String(),
		Content:   m.Content,
		CreatedAt: m.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}})
}

// GetByID は GET /api/v1/system-overviews/{id} を処理する
func (h *SystemOverviewHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	// パスから id を取得（Go 1.22 の ServeMux パターン {id} を使用）
	id := r.PathValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "id は必須です", nil)
		return
	}

	m, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "指定されたIDのシステム概要が存在しません", nil)
			return
		}
		if errors.Is(err, service.ErrValidation) {
			writeError(w, http.StatusUnprocessableEntity, "VALIDATION_ERROR", err.Error(),
				[]errorDetail{{Field: "id", Message: err.Error()}})
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "サーバーエラーが発生しました", nil)
		return
	}

	writeJSON(w, http.StatusOK, dataEnvelope{Data: systemOverviewResponse{
		ID:        m.ID.String(),
		Content:   m.Content,
		CreatedAt: m.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}})
}
