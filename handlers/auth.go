package handlers

import (
	"encoding/json"
	"latihan_cloud8/middleware"
	"latihan_cloud8/models"
	"latihan_cloud8/store"
	"latihan_cloud8/utils"
	"net/http"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	Store *store.MySQLStore
}

func NewAuthHandler(store *store.MySQLStore) *AuthHandler {
	return &AuthHandler{Store: store}
}

// Register endpoint (untuk membuat user baru)
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var payload struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Role     string `json:"role"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	if payload.Username == "" || payload.Password == "" {
		http.Error(w, "Username & password required", http.StatusBadRequest)
		return
	}

	// Hash password
	hashed, err := bcrypt.GenerateFromPassword([]byte(payload.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Error hashing password", http.StatusInternalServerError)
		return
	}

	role := payload.Role
	if role == "" {
		role = "user"
	}

	_, err = h.Store.CreateUser(payload.Username, string(hashed), role)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "User created successfully",
	})
}

// Login endpoint (untuk form login)
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Invalid request format",
		})
		return
	}

	// Validate input
	if req.Username == "" || req.Password == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Username and password are required",
		})
		return
	}

	// Get user from database
	user, err := h.Store.GetByUsername(req.Username)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Invalid username or password",
		})
		return
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Invalid username or password",
		})
		return
	}

	// Generate JWT token
	token, err := utils.GenerateToken(user.Username, user.Role, time.Hour*24) // 24 hours TTL
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Could not generate token",
		})
		return
	}

	// Return success response
	resp := models.LoginResponse{
		Token:    token,
		Username: user.Username,
		Role:     user.Role,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

// Profile endpoint (protected)
func (h *AuthHandler) Profile(w http.ResponseWriter, r *http.Request) {
	v := r.Context().Value(middleware.UserCtxKey)
	if v == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	claims, ok := v.(*utils.Claims)
	if !ok {
		http.Error(w, "Invalid auth context", http.StatusUnauthorized)
		return
	}

	resp := map[string]string{
		"username": claims.Username,
		"role":     claims.Role,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// GetUsers endpoint (admin only)
func (h *AuthHandler) GetUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.Store.GetAllUsers()
	if err != nil {
		http.Error(w, "Error fetching users", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}