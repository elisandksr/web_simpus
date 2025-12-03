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

// ============================================
// AuthMiddleware - Cek token dari Header ATAU Cookie
// ============================================
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var token string

		// STEP 1: Coba ambil token dari Authorization Header
		auth := r.Header.Get("Authorization")
		if auth != "" {
			parts := strings.SplitN(auth, " ", 2)
			if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
				token = parts[1]
				log.Println("Token dari Authorization Header:", token[:20]+"...")
			}
		}

		// STEP 2: Jika tidak ada di header, coba ambil dari Cookie
		if token == "" {
			cookie, err := r.Cookie("token")
			if err == nil {
				token = cookie.Value
				log.Println("Token dari Cookie:", token[:20]+"...")
			}
		}

		// STEP 3: Jika masih tidak ada token, return error
		if token == "" {
			log.Println("No token found in header or cookie")
			http.Error(w, "missing Authorization header or token cookie", http.StatusUnauthorized)
			return
		}

		// STEP 4: Parse dan validate token
		claims, err := utils.ParseToken(token)
		if err != nil {
			log.Println("Invalid token:", err.Error())
			http.Error(w, "invalid token: "+err.Error(), http.StatusUnauthorized)
			return
		}

		log.Println("Token valid! Username:", claims.Username, "Role:", claims.Role)

		// STEP 5: Simpan claims di context
		ctx := context.WithValue(r.Context(), UserCtxKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// ============================================
// RequireRole - Middleware untuk cek role
// ============================================
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