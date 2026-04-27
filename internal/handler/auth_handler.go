package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
	"unitask-api/internal/domain"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	Repo domain.UserRepository
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var u domain.User
	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	// Usamos el campo 'Password' que acabamos de agregar al struct
	if u.Password == "" {
		http.Error(w, "Password es requerido", http.StatusBadRequest)
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Error al procesar password", http.StatusInternalServerError)
		return
	}
	u.PasswordHash = string(hash)

	if err := h.Repo.Create(&u); err != nil {
		// ESTO ES CLAVE: Mira tu terminal cuando tires el curl
		fmt.Printf("ERROR MYSQL REAL: %v\n", err)
		http.Error(w, "Error al crear usuario", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated) // 201 [cite: 135]
	json.NewEncoder(w).Encode(map[string]string{"message": "Usuario registrado"})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var credentials struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&credentials); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	user, err := h.Repo.GetByEmail(credentials.Email)
	if err != nil || bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(credentials.Password)) != nil {
		http.Error(w, "Credenciales incorrectas", http.StatusUnauthorized)
		return
	}

	// Generar el token real usando la función de abajo
	token, err := h.generateToken(user.ID)
	if err != nil {
		http.Error(w, "Error al generar sesión", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"token": token,
		"user":  user,
	})
}

func (h *AuthHandler) generateToken(userID int) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(os.Getenv("JWT_SECRET")))
}

func (h *AuthHandler) Profile(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "No autorizado", http.StatusUnauthorized)
		return
	}

	user, err := h.Repo.GetByID(userID)
	if err != nil {
		http.Error(w, "Usuario no encontrado", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}
