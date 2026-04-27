package domain

import "time"

type DeviceToken struct {
	UserID      int       `json:"user_id"`
	DeviceToken string    `json:"device_token"`
	Platform    string    `json:"platform"`
	DeviceID    string    `json:"device_id"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at,omitempty"`
	UpdatedAt   time.Time `json:"updated_at,omitempty"`
}

type DueSoonTask struct {
	ID      int       `json:"id"`
	Title   string    `json:"title"`
	DueDate time.Time `json:"due_date"`
}

type NotificationRepository interface {
	UpsertDeviceToken(token *DeviceToken) error
	DeactivateDeviceToken(userID int, deviceToken string) error
	GetActiveDeviceTokensByUserID(userID int) ([]string, error)
	GetActiveUserIDs() ([]int, error)
	GetPendingTasksCountByUserID(userID int) (int, error)
	GetDueSoonTasksByUserID(userID int, hours int) ([]DueSoonTask, error)
	HasReminderBeenSent(userID int, taskID int, reminderType string, windowHours int) (bool, error)
	MarkReminderAsSent(userID int, taskID int, reminderType string, windowHours int) error
}

type NotificationSender interface {
	SendToTokens(tokens []string, title string, body string, data map[string]string) error
}
