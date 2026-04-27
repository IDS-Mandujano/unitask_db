package handler

import (
	"context" // <--- Importante para pasar datos entre funciones
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const userIDContextKey contextKey = "user_id"

func UserIDFromContext(ctx context.Context) (int, bool) {
	userID, ok := ctx.Value(userIDContextKey).(int)
	return userID, ok
}

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Se requiere token de autenticación", http.StatusUnauthorized)
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if token.Method.Alg() != jwt.SigningMethodHS256.Alg() {
				return nil, fmt.Errorf("unexpected signing method")
			}
			return []byte(os.Getenv("JWT_SECRET")), nil
		})

		// Si el token falla, aquí es donde podrías poner un log para tu consola
		if err != nil || !token.Valid {
			http.Error(w, "Token inválido o expirado", http.StatusUnauthorized)
			return
		}

		// --- ESTO ES LO NUEVO ---
		// Extraemos el user_id de los claims del token
		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			rawUserID, ok := claims["user_id"]
			if !ok {
				http.Error(w, "Token mal formado", http.StatusUnauthorized)
				return
			}

			var userID int
			switch value := rawUserID.(type) {
			case float64:
				userID = int(value)
			case int:
				userID = value
			default:
				http.Error(w, "Token mal formado", http.StatusUnauthorized)
				return
			}

			// Metemos el userID en el contexto de la petición
			ctx := context.WithValue(r.Context(), userIDContextKey, userID)

			// Pasamos la petición con el nuevo contexto al siguiente handler
			next.ServeHTTP(w, r.WithContext(ctx))
		} else {
			http.Error(w, "Token mal formado", http.StatusUnauthorized)
		}
	})
}
