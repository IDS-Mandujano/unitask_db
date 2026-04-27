package domain

import "time"

type Attachment struct {
	ID              int       `json:"id"`
	UserID          int       `json:"user_id"`
	SubjectID       *int      `json:"subject_id,omitempty"`
	TaskID          *int      `json:"task_id,omitempty"`
	FileName        string    `json:"file_name"`
	MimeType        string    `json:"mime_type,omitempty"`
	StoragePath     string    `json:"storage_path"`
	DownloadURL     string    `json:"download_url"`
	AttachmentType  string    `json:"attachment_type,omitempty"`
	CreatedAt       time.Time `json:"created_at,omitempty"`
}

type AttachmentRepository interface {
	CreateAttachment(a *Attachment) error
	GetAttachmentsByUserID(userID int) ([]Attachment, error)
	GetAttachmentsBySubjectID(subjectID int, userID int) ([]Attachment, error)
	GetAttachmentsByTaskID(taskID int, userID int) ([]Attachment, error)
	DeleteAttachment(id int, userID int) error
}