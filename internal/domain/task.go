package domain

import "time"

// La estructura que ya tenías [cite: 33, 34]
type Task struct {
	ID          int       `json:"id"`
	SubjectID   int       `json:"subject_id"`
	SubjectName string    `json:"subject_name,omitempty"`
	TeacherName string    `json:"teacher_name,omitempty"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	DueDate     time.Time `json:"due_date"`
	IsCompleted bool      `json:"is_completed"`
}

// La interfaz para que el Handler pueda llamar a los métodos [cite: 94, 95]
type TaskRepository interface {
	CreateTask(t *Task) error
	GetAllTasks() ([]Task, error)
	GetTasksByUserID(userID int, subjectID int) ([]Task, error)
	SubjectBelongsToUser(subjectID int, userID int) (bool, error)
	TaskBelongsToUser(taskID int, userID int) (bool, error)
	MarkAsCompleted(id string) error
	MarkAsPending(id string) error
	UpdateTask(t *Task) error
	DeleteTask(id string) error
}
