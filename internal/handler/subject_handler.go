package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"unitask-api/internal/domain"
)

type SubjectHandler struct {
	Repo domain.SubjectRepository
}

func (h *SubjectHandler) HandleSubjects(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "No autorizado", http.StatusUnauthorized)
		return
	}

	switch r.Method {
	case http.MethodPost:
		// Crear materia [cite: 91]
		var s domain.Subject
		if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
			http.Error(w, "JSON inválido", http.StatusBadRequest)
			return
		}
		s.UserID = userID
		if err := h.Repo.Create(&s); err != nil {
			http.Error(w, "Error al crear materia", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated) // 201 [cite: 135]

	case http.MethodGet:
		// Ver lista de materias [cite: 92]
		subjects, err := h.Repo.GetAllByUserID(userID)
		if err != nil {
			http.Error(w, "Error al obtener materias", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(subjects)

	default:
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
	}
}

func (h *SubjectHandler) HandleSubjectByID(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "No autorizado", http.StatusUnauthorized)
		return
	}

	idStr := r.URL.Path[len("/subjects/"):]
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		subject, err := h.Repo.GetByID(id, userID)
		if err != nil {
			http.Error(w, "Materia no encontrada", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(subject)
	case http.MethodPut:
		var subject domain.Subject
		if err := json.NewDecoder(r.Body).Decode(&subject); err != nil {
			http.Error(w, "JSON inválido", http.StatusBadRequest)
			return
		}
		subject.ID = id
		subject.UserID = userID
		if err := h.Repo.Update(&subject); err != nil {
			http.Error(w, "Error al actualizar materia", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	case http.MethodDelete:
		if err := h.Repo.Delete(id, userID); err != nil {
			http.Error(w, "Error al eliminar materia", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	default:
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
	}
}
