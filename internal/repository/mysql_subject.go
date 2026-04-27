package repository

import (
	"database/sql"
	"strings"
	"unitask-api/internal/domain"
)

type mysqlSubject struct {
	db              *sql.DB
	hasTeacherFields bool
}

// Constructor que usará tu main.go
func NewmysqlSubject(db *sql.DB) domain.SubjectRepository {
	return &mysqlSubject{db: db, hasTeacherFields: subjectHasTeacherFields(db)}
}

func subjectHasTeacherFields(db *sql.DB) bool {
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

func (r *mysqlSubject) Create(s *domain.Subject) error {
	if r.hasTeacherFields {
		query := "INSERT INTO subjects (user_id, name, teacher_name, teacher_email) VALUES (?, ?, ?, ?)"
		_, err := r.db.Exec(query, s.UserID, s.Name, s.TeacherName, s.TeacherEmail)
		return err
	}
	query := "INSERT INTO subjects (user_id, name) VALUES (?, ?)"
	_, err := r.db.Exec(query, s.UserID, s.Name)
	return err
}

func (r *mysqlSubject) GetAllByUserID(userID int) ([]domain.Subject, error) {
	query := "SELECT id, user_id, name FROM subjects WHERE user_id = ?"
	if r.hasTeacherFields {
		query = "SELECT id, user_id, name, teacher_name, teacher_email FROM subjects WHERE user_id = ?"
	}
	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subjects []domain.Subject
	for rows.Next() {
		var s domain.Subject
		if r.hasTeacherFields {
			var teacherName sql.NullString
			var teacherEmail sql.NullString
			if err := rows.Scan(&s.ID, &s.UserID, &s.Name, &teacherName, &teacherEmail); err != nil {
				return nil, err
			}
			s.TeacherName = strings.TrimSpace(teacherName.String)
			s.TeacherEmail = strings.TrimSpace(teacherEmail.String)
		} else {
			if err := rows.Scan(&s.ID, &s.UserID, &s.Name); err != nil {
				return nil, err
			}
		}
		subjects = append(subjects, s)
	}
	return subjects, nil
}

func (r *mysqlSubject) GetByID(id int, userID int) (*domain.Subject, error) {
	s := &domain.Subject{}
	query := "SELECT id, user_id, name FROM subjects WHERE id = ? AND user_id = ?"
	if r.hasTeacherFields {
		query = "SELECT id, user_id, name, teacher_name, teacher_email FROM subjects WHERE id = ? AND user_id = ?"
	}
	if r.hasTeacherFields {
		var teacherName sql.NullString
		var teacherEmail sql.NullString
		err := r.db.QueryRow(query, id, userID).Scan(&s.ID, &s.UserID, &s.Name, &teacherName, &teacherEmail)
		if err != nil {
			return nil, err
		}
		s.TeacherName = strings.TrimSpace(teacherName.String)
		s.TeacherEmail = strings.TrimSpace(teacherEmail.String)
		return s, nil
	}
	err := r.db.QueryRow(query, id, userID).Scan(&s.ID, &s.UserID, &s.Name)
	return s, err
}

func (r *mysqlSubject) Update(s *domain.Subject) error {
	if r.hasTeacherFields {
		query := "UPDATE subjects SET name = ?, teacher_name = ?, teacher_email = ? WHERE id = ? AND user_id = ?"
		_, err := r.db.Exec(query, s.Name, s.TeacherName, s.TeacherEmail, s.ID, s.UserID)
		return err
	}
	query := "UPDATE subjects SET name = ? WHERE id = ? AND user_id = ?"
	_, err := r.db.Exec(query, s.Name, s.ID, s.UserID)
	return err
}

func (r *mysqlSubject) Delete(id int, userID int) error {
	query := "DELETE FROM subjects WHERE id = ? AND user_id = ?"
	_, err := r.db.Exec(query, id, userID)
	return err
}
