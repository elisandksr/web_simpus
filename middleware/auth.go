package middleware

import (
	"context"
	"log"
	"net/http"
	"strings"

	"latihan_cloud8/utils"
)

type ctxKey string

const UserCtxKey ctxKey = "user"

// AuthMiddleware mengelola autentikasi pengguna via token (Header/Cookie).
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var token string

		// Cek token di Header Authorization
		auth := r.Header.Get("Authorization")
		if auth != "" {
			parts := strings.SplitN(auth, " ", 2)
			if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
				token = parts[1]
				log.Println("Token dari Authorization Header:", token[:20]+"...")
			}
		}

		// Jika tidak ada header, cek Cookie
		if token == "" {
			cookie, err := r.Cookie("token")
			if err == nil {
				token = cookie.Value
				log.Println("Token dari Cookie:", token[:20]+"...")
			}
		}

		// Jika token kosong, tangani error atau redirect
		if token == "" {
			if strings.HasPrefix(r.URL.Path, "/api") || strings.Contains(r.Header.Get("Accept"), "application/json") {
				log.Println("No token found for API request")
				http.Error(w, "missing Authorization header or token cookie", http.StatusUnauthorized)
				return
			}

			// Redirect ke login jika bukan API
			log.Println("No token found, redirecting to login")
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		// Validasi token JWT
		claims, err := utils.ParseToken(token)
		if err != nil {
			log.Println("Invalid token:", err.Error())
			http.Error(w, "invalid token: "+err.Error(), http.StatusUnauthorized)
			return
		}

		log.Println("Token valid! Username:", claims.Username, "Role:", claims.Role)

		// Simpan data user (claims) ke context
		ctx := context.WithValue(r.Context(), UserCtxKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireRole memvalidasi peran pengguna sebelum akses diizinkan.
func RequireRole(role string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			val := r.Context().Value(UserCtxKey)
			if val == nil {
				log.Println("No user in context")
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			claims, ok := val.(*utils.Claims)
			if !ok {
				log.Println("Invalid auth context type")
				http.Error(w, "invalid auth context", http.StatusUnauthorized)
				return
			}

			if claims.Role != role {
				log.Println("Insufficient role. Need:", role, "Got:", claims.Role)
				http.Error(w, "forbidden: insufficient role", http.StatusForbidden)
				return
			}

			log.Println("Role check passed for user:", claims.Username)
			next.ServeHTTP(w, r)
		})
	}
}
