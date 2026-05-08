package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"musuhi-api/internal/service"
)

// ProjectHandler は FR-002 の /api/v1/projects 系ハンドラ。
type ProjectHandler struct {
	svc service.ProjectService
}

// NewProjectHandler は ProjectHandler を生成する。
func NewProjectHandler(svc service.ProjectService) *ProjectHandler {
	return &ProjectHandler{svc: svc}
}

type overviewIDRequest struct {
	OverviewID string `json:"overviewId"`
}

type initDirectoryRequest struct {
	ProjectName string `json:"projectName"`
	LocalPath   string `json:"localPath"`
	Template    string `json:"template"`
}

// ExtractFeatures は POST /api/v1/projects/extract-features を処理する。
func (h *ProjectHandler) ExtractFeatures(w http.ResponseWriter, r *http.Request) {
	var req overviewIDRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "リクエストの形式が不正です", nil)
		return
	}

	result, err := h.svc.ExtractFeatures(r.Context(), req.OverviewID)
	if err != nil {
		handleProjectServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, dataEnvelope{Data: result})
}

// SuggestName は POST /api/v1/projects/suggest-name を処理する。
func (h *ProjectHandler) SuggestName(w http.ResponseWriter, r *http.Request) {
	var req overviewIDRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "リクエストの形式が不正です", nil)
		return
	}

	result, err := h.svc.SuggestName(r.Context(), req.OverviewID)
	if err != nil {
		handleProjectServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, dataEnvelope{Data: result})
}

// InitDirectory は POST /api/v1/projects/init-directory を処理する。
func (h *ProjectHandler) InitDirectory(w http.ResponseWriter, r *http.Request) {
	var req initDirectoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "リクエストの形式が不正です", nil)
		return
	}

	result, err := h.svc.InitDirectory(r.Context(), req.ProjectName, req.LocalPath, req.Template)
	if err != nil {
		handleProjectServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, dataEnvelope{Data: result})
}

func handleProjectServiceError(w http.ResponseWriter, err error) {
	if errors.Is(err, service.ErrValidation) {
		writeError(w, http.StatusUnprocessableEntity, "VALIDATION_ERROR", err.Error(), nil)
		return
	}
	if errors.Is(err, service.ErrNotFound) {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "指定されたIDのリソースが存在しません", nil)
		return
	}
	writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "サーバーエラーが発生しました", nil)
}
