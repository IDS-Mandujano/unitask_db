package domain

import "time"

// Entidad de usuario para el MVP 01 [cite: 84, 85]
type User struct {
    ID           int       `json:"id"`
    Username     string    `json:"username"`
    Email        string    `json:"email"`
    Password     string    `json:"password,omitempty"`
    PasswordHash string    `json:"-"`
    Career       string    `json:"career"`     // Nuevo campo para el registro [cite: 109]
    University   string    `json:"university"` // Nuevo campo para el registro [cite: 109]
    CreatedAt    time.Time `json:"created_at"`
}

// Interfaz del repositorio (Abstracción) 
type UserRepository interface {
    Create(user *User) error
    GetByEmail(email string) (*User, error)
    GetByID(id int) (*User, error)
}