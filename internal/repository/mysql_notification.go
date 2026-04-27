package repository

import (
	"database/sql"
	"unitask-api/internal/domain"
)

type mysqlNotificationRepository struct {
	db *sql.DB
}

func NewMySQLNotificationRepository(db *sql.DB) domain.NotificationRepository {
	return &mysqlNotificationRepository{db: db}
}

func (r *mysqlNotificationRepository) UpsertDeviceToken(token *domain.DeviceToken) error {
	// SIMPLIFICADO: Solo activar el token registrado, sin desactivar otros
	// Esto permite múltiples tokens activos por usuario (ej: navegador + mobile)
	query := `
		INSERT INTO user_device_tokens (user_id, device_token, platform, device_id, is_active)
		VALUES (?, ?, ?, ?, 1)
		ON DUPLICATE KEY UPDATE
			user_id = VALUES(user_id),
			platform = VALUES(platform),
			device_id = VALUES(device_id),
			is_active = 1,
			updated_at = CURRENT_TIMESTAMP`

	_, err := r.db.Exec(query, token.UserID, token.DeviceToken, token.Platform, token.DeviceID)
	return err
}

func (r *mysqlNotificationRepository) DeactivateDeviceToken(userID int, deviceToken string) error {
	query := `
		UPDATE user_device_tokens
		SET is_active = 0, updated_at = CURRENT_TIMESTAMP
		WHERE user_id = ? AND device_token = ?`

	_, err := r.db.Exec(query, userID, deviceToken)
	return err
}

func (r *mysqlNotificationRepository) GetActiveDeviceTokensByUserID(userID int) ([]string, error) {
	query := `
		SELECT device_token
		FROM user_device_tokens
		WHERE user_id = ? AND is_active = 1`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tokens := make([]string, 0)
	for rows.Next() {
		var token string
		if err := rows.Scan(&token); err != nil {
			return nil, err
		}
		tokens = append(tokens, token)
	}

	return tokens, nil
}

func (r *mysqlNotificationRepository) GetActiveUserIDs() ([]int, error) {
	query := `
		SELECT DISTINCT user_id
		FROM user_device_tokens
		WHERE is_active = 1`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	userIDs := make([]int, 0)
	for rows.Next() {
		var userID int
		if err := rows.Scan(&userID); err != nil {
			return nil, err
		}
		userIDs = append(userIDs, userID)
	}

	return userIDs, nil
}

func (r *mysqlNotificationRepository) GetPendingTasksCountByUserID(userID int) (int, error) {
	query := `
		SELECT COUNT(1)
		FROM tasks t
		INNER JOIN subjects s ON t.subject_id = s.id
		WHERE s.user_id = ? AND t.is_completed = 0`

	var count int
	err := r.db.QueryRow(query, userID).Scan(&count)
	return count, err
}

func (r *mysqlNotificationRepository) GetDueSoonTasksByUserID(userID int, hours int) ([]domain.DueSoonTask, error) {
	query := `
		SELECT t.id, t.title, t.due_date
		FROM tasks t
		INNER JOIN subjects s ON t.subject_id = s.id
		WHERE s.user_id = ?
			AND t.is_completed = 0
			AND t.due_date IS NOT NULL
			AND t.due_date BETWEEN NOW() AND DATE_ADD(NOW(), INTERVAL ? HOUR)
		ORDER BY t.due_date ASC
		LIMIT 5`

	rows, err := r.db.Query(query, userID, hours)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasks := make([]domain.DueSoonTask, 0)
	for rows.Next() {
		var task domain.DueSoonTask
		if err := rows.Scan(&task.ID, &task.Title, &task.DueDate); err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}

	return tasks, nil
}

func (r *mysqlNotificationRepository) HasReminderBeenSent(userID int, taskID int, reminderType string, windowHours int) (bool, error) {
	query := `
		SELECT 1
		FROM notification_reminder_logs
		WHERE user_id = ? AND task_id = ? AND reminder_type = ? AND window_hours = ?
		LIMIT 1`

	var exists int
	err := r.db.QueryRow(query, userID, taskID, reminderType, windowHours).Scan(&exists)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (r *mysqlNotificationRepository) MarkReminderAsSent(userID int, taskID int, reminderType string, windowHours int) error {
	query := `
		INSERT INTO notification_reminder_logs (user_id, task_id, reminder_type, window_hours)
		VALUES (?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE sent_at = CURRENT_TIMESTAMP`
	_, err := r.db.Exec(query, userID, taskID, reminderType, windowHours)
	return err
}
