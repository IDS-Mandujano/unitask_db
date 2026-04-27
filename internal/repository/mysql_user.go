package repository

import (
    "database/sql"
    "unitask-api/internal/domain"
)

type mysqlUserRepository struct {
    db *sql.DB
}

func NewMySQLUserRepository(db *sql.DB) domain.UserRepository {
    return &mysqlUserRepository{db}
}

func (r *mysqlUserRepository) Create(u *domain.User) error {
    // ACTUALIZADO: Ahora insertamos carrera y universidad
    query := "INSERT INTO users (username, email, password_hash, career, university) VALUES (?, ?, ?, ?, ?)"
    _, err := r.db.Exec(query, u.Username, u.Email, u.PasswordHash, u.Career, u.University)
    return err
}

func (r *mysqlUserRepository) GetByEmail(email string) (*domain.User, error) {
    u := &domain.User{}
    // ACTUALIZADO: Traemos todos los campos académicos
    query := "SELECT id, username, email, password_hash, career, university, created_at FROM users WHERE email = ?"
    err := r.db.QueryRow(query, email).Scan(
        &u.ID, &u.Username, &u.Email, &u.PasswordHash, &u.Career, &u.University, &u.CreatedAt,
    )
    return u, err
}

func (r *mysqlUserRepository) GetByID(id int) (*domain.User, error) {
    u := &domain.User{}
    // ACTUALIZADO: Para que el perfil se vea completo
    query := "SELECT id, username, email, career, university FROM users WHERE id = ?"
    err := r.db.QueryRow(query, id).Scan(&u.ID, &u.Username, &u.Email, &u.Career, &u.University)
    return u, err
}