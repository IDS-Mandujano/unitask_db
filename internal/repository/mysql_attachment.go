package repository

import (
	"database/sql"
	"fmt"
	"unitask-api/internal/domain"
)

type mysqlAttachmentRepository struct {
	db *sql.DB
}

func NewMySQLAttachmentRepository(db *sql.DB) domain.AttachmentRepository {
	return &mysqlAttachmentRepository{db: db}
}

func (r *mysqlAttachmentRepository) CreateAttachment(a *domain.Attachment) error {
	query := `INSERT INTO attachments
		(user_id, subject_id, task_id, file_name, mime_type, storage_path, download_url, attachment_type)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := r.db.Exec(query, a.UserID, a.SubjectID, a.TaskID, a.FileName, a.MimeType, a.StoragePath, a.DownloadURL, a.AttachmentType)
	return err
}

func (r *mysqlAttachmentRepository) GetAttachmentsByUserID(userID int) ([]domain.Attachment, error) {
	return r.listAttachments(`WHERE user_id = ?`, userID)
}

func (r *mysqlAttachmentRepository) GetAttachmentsBySubjectID(subjectID int, userID int) ([]domain.Attachment, error) {
	return r.listAttachments(`WHERE subject_id = ? AND user_id = ?`, subjectID, userID)
}

func (r *mysqlAttachmentRepository) GetAttachmentsByTaskID(taskID int, userID int) ([]domain.Attachment, error) {
	return r.listAttachments(`WHERE task_id = ? AND user_id = ?`, taskID, userID)
}

func (r *mysqlAttachmentRepository) DeleteAttachment(id int, userID int) error {
	query := "DELETE FROM attachments WHERE id = ? AND user_id = ?"
	_, err := r.db.Exec(query, id, userID)
	return err
}

func (r *mysqlAttachmentRepository) listAttachments(where string, args ...any) ([]domain.Attachment, error) {
	query := fmt.Sprintf(`SELECT id, user_id, subject_id, task_id, file_name, mime_type, storage_path, download_url, attachment_type, created_at FROM attachments %s ORDER BY created_at DESC`, where)
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	attachments := make([]domain.Attachment, 0)
	for rows.Next() {
		var attachment domain.Attachment
		var subjectID sql.NullInt64
		var taskID sql.NullInt64
		if err := rows.Scan(&attachment.ID, &attachment.UserID, &subjectID, &taskID, &attachment.FileName, &attachment.MimeType, &attachment.StoragePath, &attachment.DownloadURL, &attachment.AttachmentType, &attachment.CreatedAt); err != nil {
			return nil, err
		}
		if subjectID.Valid {
			value := int(subjectID.Int64)
			attachment.SubjectID = &value
		}
		if taskID.Valid {
			value := int(taskID.Int64)
			attachment.TaskID = &value
		}
		attachments = append(attachments, attachment)
	}
	return attachments, nil
}
