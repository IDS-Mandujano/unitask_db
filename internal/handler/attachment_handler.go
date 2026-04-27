package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"unitask-api/internal/domain"
)

type AttachmentHandler struct {
	Repo domain.AttachmentRepository
}

type registerAttachmentRequest struct {
	SubjectID      *int   `json:"subject_id"`
	TaskID         *int   `json:"task_id"`
	FileName       string `json:"file_name"`
	MimeType       string `json:"mime_type"`
	StoragePath    string `json:"storage_path"`
	DownloadURL    string `json:"download_url"`
	AttachmentType string `json:"attachment_type"`
}

func (h *AttachmentHandler) RegisterAttachment(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "No autorizado", http.StatusUnauthorized)
		return
	}

	var req registerAttachmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	if req.FileName == "" || req.StoragePath == "" || req.DownloadURL == "" {
		http.Error(w, "file_name, storage_path y download_url son requeridos", http.StatusBadRequest)
		return
	}

	attachment := &domain.Attachment{
		UserID:         userID,
		SubjectID:      req.SubjectID,
		TaskID:         req.TaskID,
		FileName:       req.FileName,
		MimeType:       req.MimeType,
		StoragePath:    req.StoragePath,
		DownloadURL:    req.DownloadURL,
		AttachmentType: req.AttachmentType,
	}

	if err := h.Repo.CreateAttachment(attachment); err != nil {
		http.Error(w, "Error guardando adjunto", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(attachment)
}

func (h *AttachmentHandler) ListAttachments(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "No autorizado", http.StatusUnauthorized)
		return
	}

	if rawTaskID := r.URL.Query().Get("task_id"); rawTaskID != "" {
		taskID, err := strconv.Atoi(rawTaskID)
		if err != nil || taskID <= 0 {
			http.Error(w, "task_id inválido", http.StatusBadRequest)
			return
		}
		attachments, err := h.Repo.GetAttachmentsByTaskID(taskID, userID)
		if err != nil {
			http.Error(w, "Error al obtener adjuntos", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(attachments)
		return
	}

	if rawSubjectID := r.URL.Query().Get("subject_id"); rawSubjectID != "" {
		subjectID, err := strconv.Atoi(rawSubjectID)
		if err != nil || subjectID <= 0 {
			http.Error(w, "subject_id inválido", http.StatusBadRequest)
			return
		}
		attachments, err := h.Repo.GetAttachmentsBySubjectID(subjectID, userID)
		if err != nil {
			http.Error(w, "Error al obtener adjuntos", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(attachments)
		return
	}

	attachments, err := h.Repo.GetAttachmentsByUserID(userID)
	if err != nil {
		http.Error(w, "Error al obtener adjuntos", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(attachments)
}

func (h *AttachmentHandler) DeleteAttachment(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "No autorizado", http.StatusUnauthorized)
		return
	}

	id, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil || id <= 0 {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}

	if err := h.Repo.DeleteAttachment(id, userID); err != nil {
		http.Error(w, "Error al eliminar adjunto", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
