package service

import (
	"fmt"
	"log"
	"time"

	"unitask-api/internal/domain"
)

type ReminderScheduler struct {
	repo   domain.NotificationRepository
	sender domain.NotificationSender
	window int
	loc    *time.Location
}

func NewReminderScheduler(repo domain.NotificationRepository, sender domain.NotificationSender, windowHours int, loc *time.Location) *ReminderScheduler {
	if loc == nil {
		loc = time.Local
	}
	return &ReminderScheduler{repo: repo, sender: sender, window: windowHours, loc: loc}
}

func (s *ReminderScheduler) Start(interval time.Duration) {
	go func() {
		log.Printf("reminder scheduler: iniciado (interval=%s, ventana=%dh, tz=%s)", interval.String(), s.window, s.loc.String())
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		s.runOnce()
		for range ticker.C {
			s.runOnce()
		}
	}()
}

func (s *ReminderScheduler) runOnce() {
	cycleStart := time.Now()
	log.Printf("reminder scheduler: ciclo iniciado")

	userIDs, err := s.repo.GetActiveUserIDs()
	if err != nil {
		log.Printf("reminder scheduler: no se pudieron obtener usuarios activos: %v", err)
		return
	}
	if len(userIDs) == 0 {
		log.Printf("reminder scheduler: sin usuarios con tokens activos")
		return
	}

	totalTasksEvaluated := 0
	totalSent := 0

	for _, userID := range userIDs {
		tokens, err := s.repo.GetActiveDeviceTokensByUserID(userID)
		if err != nil || len(tokens) == 0 {
			if err != nil {
				log.Printf("reminder scheduler: user_id=%d error obteniendo tokens: %v", userID, err)
			}
			continue
		}

		tasks, err := s.repo.GetDueSoonTasksByUserID(userID, s.window)
		if err != nil || len(tasks) == 0 {
			if err != nil {
				log.Printf("reminder scheduler: user_id=%d error obteniendo tareas por vencer: %v", userID, err)
			}
			continue
		}
		totalTasksEvaluated += len(tasks)
		log.Printf("reminder scheduler: user_id=%d tokens=%d tareas_due_soon=%d", userID, len(tokens), len(tasks))

		for _, task := range tasks {
			sent, err := s.repo.HasReminderBeenSent(userID, task.ID, "due_soon", s.window)
			if err != nil {
				log.Printf("reminder scheduler: error revisando log de tarea %d: %v", task.ID, err)
				continue
			}
			if sent {
				continue
			}

			title := "Tarea próxima a vencer"
			dueDateLocal := task.DueDate.In(s.loc)
			body := fmt.Sprintf("%s vence el %s", task.Title, dueDateLocal.Format("02/01 15:04"))
			err = s.sender.SendToTokens(tokens, title, body, map[string]string{
				"type":         "due_soon",
				"task_id":      fmt.Sprintf("%d", task.ID),
				"window_hours": fmt.Sprintf("%d", s.window),
			})
			if err != nil {
				log.Printf("reminder scheduler: error enviando tarea %d: %v", task.ID, err)
				continue
			}
			totalSent++
			log.Printf("reminder scheduler: enviado user_id=%d task_id=%d due_date_local=%s due_date_utc=%s", userID, task.ID, dueDateLocal.Format(time.RFC3339), task.DueDate.UTC().Format(time.RFC3339))

			if err := s.repo.MarkReminderAsSent(userID, task.ID, "due_soon", s.window); err != nil {
				log.Printf("reminder scheduler: no se pudo registrar envío de tarea %d: %v", task.ID, err)
			}
		}
	}

	log.Printf(
		"reminder scheduler: ciclo finalizado users=%d tareas_evaluadas=%d enviados=%d duracion=%s",
		len(userIDs), totalTasksEvaluated, totalSent, time.Since(cycleStart).String(),
	)
}
