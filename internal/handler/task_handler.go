package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"
	"unitask-api/internal/domain"
)

type TaskHandler struct {
	Repo domain.TaskRepository // Asegúrate de definir esta interfaz en domain
}

type taskUpsertRequest struct {
	ID          int    `json:"id"`
	SubjectID   int    `json:"subject_id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	DueDate     string `json:"due_date"`
	IsCompleted bool   `json:"is_completed"`
}

func parseDueDate(raw string) (time.Time, error) {
	loc := time.Now().Location()

	if parsed, err := time.Parse(time.RFC3339, raw); err == nil {
		return parsed.In(loc), nil
	}

	layouts := []string{
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
		"2006-01-02T15:04",
		"2006-01-02 15:04",
	}
	for _, layout := range layouts {
		if parsed, err := time.ParseInLocation(layout, raw, loc); err == nil {
			return parsed, nil
		}
	}
	return time.Time{}, strconv.ErrSyntax
}

func (h *TaskHandler) CreateTask(w http.ResponseWriter, r *http.Request) {
	var req taskUpsertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	dueDate, err := parseDueDate(req.DueDate)
	if err != nil {
		http.Error(w, "due_date inválido", http.StatusBadRequest)
		return
	}
	if dueDate.Before(time.Now()) {
		http.Error(w, "La fecha y hora no pueden ser anteriores a la actual", http.StatusBadRequest)
		return
	}

	t := domain.Task{
		SubjectID:   req.SubjectID,
		Title:       req.Title,
		Description: req.Description,
		DueDate:     dueDate,
		IsCompleted: req.IsCompleted,
	}

	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "No autorizado", http.StatusUnauthorized)
		return
	}

	allowed, err := h.Repo.SubjectBelongsToUser(t.SubjectID, userID)
	if err != nil || !allowed {
		http.Error(w, "La materia no pertenece al usuario", http.StatusForbidden)
		return
	}

	// Lógica para insertar en MySQL [cite: 93]
	if t.DueDate.IsZero() {
		http.Error(w, "due_date es requerido", http.StatusBadRequest)
		return
	}

	if err := h.Repo.CreateTask(&t); err != nil {
		http.Error(w, "Error al crear tarea", 500)
		return
	}
	w.WriteHeader(201) // [cite: 135]
}

// Esta función es la que te faltaba para que el GET no truene
func (h *TaskHandler) GetTasks(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "No autorizado", http.StatusUnauthorized)
		return
	}

	subjectID := 0
	if rawSubjectID := r.URL.Query().Get("subject_id"); rawSubjectID != "" {
		parsed, err := strconv.Atoi(rawSubjectID)
		if err != nil || parsed <= 0 {
			http.Error(w, "subject_id inválido", http.StatusBadRequest)
			return
		}
		subjectID = parsed
	}

	tasks, err := h.Repo.GetTasksByUserID(userID, subjectID)
	if err != nil {
		http.Error(w, "Error al obtener tareas", 500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tasks)
}

func (h *TaskHandler) CompleteTask(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id") // PATCH /tasks/{id}/complete [cite: 42, 98]
	taskID, err := strconv.Atoi(id)
	if err != nil || taskID <= 0 {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}

	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "No autorizado", http.StatusUnauthorized)
		return
	}

	allowed, err := h.Repo.TaskBelongsToUser(taskID, userID)
	if err != nil || !allowed {
		http.Error(w, "La tarea no pertenece al usuario", http.StatusForbidden)
		return
	}

	if err := h.Repo.MarkAsCompleted(id); err != nil {
		http.Error(w, "Error al actualizar", 500)
		return
	}
	w.WriteHeader(200)
}

func (h *TaskHandler) MarkPendingTask(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	taskID, err := strconv.Atoi(id)
	if err != nil || taskID <= 0 {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}

	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "No autorizado", http.StatusUnauthorized)
		return
	}

	allowed, err := h.Repo.TaskBelongsToUser(taskID, userID)
	if err != nil || !allowed {
		http.Error(w, "La tarea no pertenece al usuario", http.StatusForbidden)
		return
	}

	if err := h.Repo.MarkAsPending(id); err != nil {
		http.Error(w, "Error al actualizar", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *TaskHandler) UpdateTask(w http.ResponseWriter, r *http.Request) {
	var req taskUpsertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "JSON inválido", 400)
		return
	}

	dueDate, err := parseDueDate(req.DueDate)
	if err != nil {
		http.Error(w, "due_date inválido", http.StatusBadRequest)
		return
	}
	if dueDate.Before(time.Now()) {
		http.Error(w, "La fecha y hora no pueden ser anteriores a la actual", http.StatusBadRequest)
		return
	}

	t := domain.Task{
		ID:          req.ID,
		SubjectID:   req.SubjectID,
		Title:       req.Title,
		Description: req.Description,
		DueDate:     dueDate,
		IsCompleted: req.IsCompleted,
	}

	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "No autorizado", http.StatusUnauthorized)
		return
	}

	allowed, err := h.Repo.TaskBelongsToUser(t.ID, userID)
	if err != nil || !allowed {
		http.Error(w, "La tarea no pertenece al usuario", http.StatusForbidden)
		return
	}

	if err := h.Repo.UpdateTask(&t); err != nil {
		http.Error(w, "Error al actualizar tarea", 500)
		return
	}
	w.WriteHeader(200)
}

// 5. ELIMINAR TAREA (DELETE)
func (h *TaskHandler) DeleteTask(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "ID requerido", 400)
		return
	}

	taskID, err := strconv.Atoi(id)
	if err != nil || taskID <= 0 {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}

	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "No autorizado", http.StatusUnauthorized)
		return
	}

	allowed, err := h.Repo.TaskBelongsToUser(taskID, userID)
	if err != nil || !allowed {
		http.Error(w, "La tarea no pertenece al usuario", http.StatusForbidden)
		return
	}

	if err := h.Repo.DeleteTask(id); err != nil {
		http.Error(w, "Error al eliminar tarea", 500)
		return
	}
	w.WriteHeader(200)
}
