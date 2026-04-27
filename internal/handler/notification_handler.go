package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"unitask-api/internal/domain"
)

type NotificationHandler struct {
	Repo   domain.NotificationRepository
	Sender domain.NotificationSender
}

type registerDeviceTokenRequest struct {
	DeviceToken string `json:"device_token"`
	Platform    string `json:"platform"`
	DeviceID    string `json:"device_id"`
}

type unregisterDeviceTokenRequest struct {
	DeviceToken string `json:"device_token"`
}

type sendTestNotificationRequest struct {
	Title string            `json:"title"`
	Body  string            `json:"body"`
	Data  map[string]string `json:"data"`
}

func (h *NotificationHandler) RegisterDeviceToken(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "No autorizado", http.StatusUnauthorized)
		return
	}

	var req registerDeviceTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	if req.DeviceToken == "" {
		http.Error(w, "device_token es requerido", http.StatusBadRequest)
		return
	}

	token := &domain.DeviceToken{
		UserID:      userID,
		DeviceToken: req.DeviceToken,
		Platform:    req.Platform,
		DeviceID:    req.DeviceID,
	}

	if err := h.Repo.UpsertDeviceToken(token); err != nil {
		http.Error(w, "Error guardando token de dispositivo", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Token registrado correctamente"})
}

func (h *NotificationHandler) UnregisterDeviceToken(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "No autorizado", http.StatusUnauthorized)
		return
	}

	var req unregisterDeviceTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	if req.DeviceToken == "" {
		http.Error(w, "device_token es requerido", http.StatusBadRequest)
		return
	}

	if err := h.Repo.DeactivateDeviceToken(userID, req.DeviceToken); err != nil {
		http.Error(w, "Error desactivando token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Token desactivado correctamente"})
}

func (h *NotificationHandler) SendPendingReminder(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "No autorizado", http.StatusUnauthorized)
		return
	}

	tokens, err := h.Repo.GetActiveDeviceTokensByUserID(userID)
	if err != nil {
		http.Error(w, "Error obteniendo tokens", http.StatusInternalServerError)
		return
	}

	if len(tokens) == 0 {
		http.Error(w, "No hay dispositivos registrados para este usuario", http.StatusBadRequest)
		return
	}

	pendingCount, err := h.Repo.GetPendingTasksCountByUserID(userID)
	if err != nil {
		http.Error(w, "Error obteniendo tareas pendientes", http.StatusInternalServerError)
		return
	}

	if pendingCount == 0 {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"message": "No tienes tareas pendientes"})
		return
	}

	title := "Tienes tareas pendientes"
	body := fmt.Sprintf("Aún tienes %d tarea(s) por completar.", pendingCount)

	if err := h.Sender.SendToTokens(tokens, title, body, map[string]string{
		"type":          "pending_tasks",
		"pending_count": strconv.Itoa(pendingCount),
	}); err != nil {
		log.Printf("notifications pending send error user_id=%d: %v", userID, err)
		http.Error(w, "Error enviando notificación: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Notificación de pendientes enviada"})
}

func (h *NotificationHandler) SendDueSoonReminder(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "No autorizado", http.StatusUnauthorized)
		return
	}

	hours := 24
	if rawHours := r.URL.Query().Get("hours"); rawHours != "" {
		parsed, err := strconv.Atoi(rawHours)
		if err != nil || parsed <= 0 {
			http.Error(w, "hours inválido", http.StatusBadRequest)
			return
		}
		hours = parsed
	}

	tokens, err := h.Repo.GetActiveDeviceTokensByUserID(userID)
	if err != nil {
		http.Error(w, "Error obteniendo tokens", http.StatusInternalServerError)
		return
	}

	if len(tokens) == 0 {
		http.Error(w, "No hay dispositivos registrados para este usuario", http.StatusBadRequest)
		return
	}

	dueSoonTasks, err := h.Repo.GetDueSoonTasksByUserID(userID, hours)
	if err != nil {
		http.Error(w, "Error obteniendo tareas próximas", http.StatusInternalServerError)
		return
	}

	if len(dueSoonTasks) == 0 {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"message": "No tienes tareas por vencer próximamente"})
		return
	}

	title := "Tareas próximas a vencer"
	body := fmt.Sprintf("Tienes %d tarea(s) por vencer en las próximas %d horas.", len(dueSoonTasks), hours)

	if err := h.Sender.SendToTokens(tokens, title, body, map[string]string{
		"type":            "due_soon",
		"due_soon_count":  strconv.Itoa(len(dueSoonTasks)),
		"window_in_hours": strconv.Itoa(hours),
	}); err != nil {
		log.Printf("notifications due-soon send error user_id=%d: %v", userID, err)
		http.Error(w, "Error enviando notificación: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Notificación de vencimientos enviada"})
}

func (h *NotificationHandler) SendTestNotification(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "No autorizado", http.StatusUnauthorized)
		return
	}

	var req sendTestNotificationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	if req.Title == "" || req.Body == "" {
		http.Error(w, "title y body son requeridos", http.StatusBadRequest)
		return
	}

	tokens, err := h.Repo.GetActiveDeviceTokensByUserID(userID)
	if err != nil {
		http.Error(w, "Error obteniendo tokens", http.StatusInternalServerError)
		return
	}

	if len(tokens) == 0 {
		http.Error(w, "No hay dispositivos registrados para este usuario", http.StatusBadRequest)
		return
	}

	if req.Data == nil {
		req.Data = map[string]string{}
	}

	if err := h.Sender.SendToTokens(tokens, req.Title, req.Body, req.Data); err != nil {
		log.Printf("notifications test send error user_id=%d: %v", userID, err)
		http.Error(w, "Error enviando notificación: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Notificación de prueba enviada"})
}
