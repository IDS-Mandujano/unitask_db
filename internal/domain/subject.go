package domain

type Subject struct {
	ID           int    `json:"id"`
	UserID       int    `json:"user_id"`
	Name         string `json:"name"`
	TeacherName  string `json:"teacher_name,omitempty"`
	TeacherEmail string `json:"teacher_email,omitempty"`
}

type SubjectRepository interface {
	Create(s *Subject) error
	GetAllByUserID(userID int) ([]Subject, error)
	GetByID(id int, userID int) (*Subject, error)
	Update(s *Subject) error
	Delete(id int, userID int) error
}
