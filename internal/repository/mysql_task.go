package repository

import (
	"database/sql"
	"unitask-api/internal/domain"
)

type mysqlTask struct {
	db              *sql.DB
	hasTeacherFields bool
}

// Tu constructor con el nombre que pediste
func NewmysqlTask(db *sql.DB) domain.TaskRepository {
	return &mysqlTask{db: db, hasTeacherFields: taskHasTeacherFields(db)}
}

func taskHasTeacherFields(db *sql.DB) bool {
	var count int
	query := `SELECT COUNT(*) FROM information_schema.columns
		WHERE table_schema = DATABASE()
		AND table_name = 'subjects'
		AND column_name IN ('teacher_name', 'teacher_email')`
	if err := db.QueryRow(query).Scan(&count); err != nil {
		return false
	}
	return count == 2
}

func (r *mysqlTask) CreateTask(t *domain.Task) error {
	query := "INSERT INTO tasks (subject_id, title, description, due_date, is_completed) VALUES (?, ?, ?, ?, ?)"
	_, err := r.db.Exec(query, t.SubjectID, t.Title, t.Description, t.DueDate, t.IsCompleted)
	return err
}

func (r *mysqlTask) GetAllTasks() ([]domain.Task, error) {
	query := `SELECT t.id, t.subject_id, s.name, t.title, t.description, t.due_date, t.is_completed
	          FROM tasks t
	          INNER JOIN subjects s ON t.subject_id = s.id`
	if r.hasTeacherFields {
		query = `SELECT t.id, t.subject_id, s.name, s.teacher_name, t.title, t.description, t.due_date, t.is_completed
	          FROM tasks t
	          INNER JOIN subjects s ON t.subject_id = s.id`
	}
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []domain.Task
	for rows.Next() {
		var t domain.Task
		if r.hasTeacherFields {
			if err := rows.Scan(&t.ID, &t.SubjectID, &t.SubjectName, &t.TeacherName, &t.Title, &t.Description, &t.DueDate, &t.IsCompleted); err != nil {
				return nil, err
			}
		} else {
			if err := rows.Scan(&t.ID, &t.SubjectID, &t.SubjectName, &t.Title, &t.Description, &t.DueDate, &t.IsCompleted); err != nil {
				return nil, err
			}
		}
		tasks = append(tasks, t)
	}
	return tasks, nil
}

func (r *mysqlTask) GetTasksByUserID(userID int, subjectID int) ([]domain.Task, error) {
	query := `SELECT t.id, t.subject_id, s.name, t.title, t.description, t.due_date, t.is_completed
              FROM tasks t
              INNER JOIN subjects s ON t.subject_id = s.id
			  WHERE s.user_id = ?`
	if r.hasTeacherFields {
		query = `SELECT t.id, t.subject_id, s.name, s.teacher_name, t.title, t.description, t.due_date, t.is_completed
              FROM tasks t
              INNER JOIN subjects s ON t.subject_id = s.id
			  WHERE s.user_id = ?`
	}
	args := []any{userID}
	if subjectID > 0 {
		query += " AND s.id = ?"
		args = append(args, subjectID)
	}
	query += " ORDER BY t.due_date ASC, t.id DESC"
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []domain.Task
	for rows.Next() {
		var t domain.Task
		if r.hasTeacherFields {
			if err := rows.Scan(&t.ID, &t.SubjectID, &t.SubjectName, &t.TeacherName, &t.Title, &t.Description, &t.DueDate, &t.IsCompleted); err != nil {
				return nil, err
			}
		} else {
			if err := rows.Scan(&t.ID, &t.SubjectID, &t.SubjectName, &t.Title, &t.Description, &t.DueDate, &t.IsCompleted); err != nil {
				return nil, err
			}
		}
		tasks = append(tasks, t)
	}
	return tasks, nil
}

func (r *mysqlTask) SubjectBelongsToUser(subjectID int, userID int) (bool, error) {
	var exists int
	query := "SELECT 1 FROM subjects WHERE id = ? AND user_id = ? LIMIT 1"
	err := r.db.QueryRow(query, subjectID, userID).Scan(&exists)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (r *mysqlTask) TaskBelongsToUser(taskID int, userID int) (bool, error) {
	var exists int
	query := `SELECT 1
              FROM tasks t
              INNER JOIN subjects s ON t.subject_id = s.id
              WHERE t.id = ? AND s.user_id = ?
              LIMIT 1`
	err := r.db.QueryRow(query, taskID, userID).Scan(&exists)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (r *mysqlTask) MarkAsCompleted(id string) error {
	query := "UPDATE tasks SET is_completed = true WHERE id = ?"
	_, err := r.db.Exec(query, id)
	return err
}

func (r *mysqlTask) MarkAsPending(id string) error {
	query := "UPDATE tasks SET is_completed = false WHERE id = ?"
	_, err := r.db.Exec(query, id)
	return err
}

func (r *mysqlTask) UpdateTask(t *domain.Task) error {
	query := `UPDATE tasks SET subject_id = ?, title = ?, description = ?, due_date = ?, is_completed = ? 
              WHERE id = ?`
	_, err := r.db.Exec(query, t.SubjectID, t.Title, t.Description, t.DueDate, t.IsCompleted, t.ID)
	return err
}

func (r *mysqlTask) DeleteTask(id string) error {
	query := "DELETE FROM tasks WHERE id = ?"
	_, err := r.db.Exec(query, id)
	return err
}
