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

// Register endpoint (untuk membuat user baru).
// Fungsi ini menangani pendaftaran pengguna baru dengan menerima username, password, dan role.
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var payload models.RegisterRequest

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	if payload.Username == "" || payload.Password == "" {
		http.Error(w, "Username & password required", http.StatusBadRequest)
		return
	}

	// Hash password sebelum disimpan
	hashed, err := bcrypt.GenerateFromPassword([]byte(payload.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Error hashing password", http.StatusInternalServerError)
		return
	}

	// Gunakan role dari payload, default "mahasiswa" per permintaan
	role := payload.Role
	if role == "" {
		role = "mahasiswa"
	}

	// Validasi Role
	validRoles := map[string]bool{
		"admin":     true,
		"mahasiswa": true,
		"guru":      true,
		"karyawan":  true,
	}
	if !validRoles[role] {
		http.Error(w, "Invalid role. Must be one of: admin, mahasiswa, guru, karyawan", http.StatusBadRequest)
		return
	}

	// Buat user baru di database
	_, err = h.Store.CreateUser(payload.Username, string(hashed), role, payload.Fullname)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "User created successfully",
	})
}

// Login endpoint (untuk form login).
// Fungsi ini memverifikasi username dan password, serta menghasilkan token JWT jika kredensial valid.
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

	// Validasi input
	if req.Username == "" || req.Password == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Username and password are required",
		})
		return
	}

	// Ambil data user dari database
	user, err := h.Store.GetByUsername(req.Username)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Invalid username or password",
		})
		return
	}

	// Verifikasi password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Invalid username or password",
		})
		return
	}

	// Buat token JWT (berlaku 24 jam)
	token, err := utils.GenerateToken(user.Username, user.Role, time.Hour*24) // 24 hours TTL
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Could not generate token",
		})
		return
	}

	// Kirim respon sukses dengan token
	resp := models.LoginResponse{
		Token:    token,
		Username: user.Username,
		Role:     user.Role,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

// Profile endpoint (protected).
// Fungsi ini mengambil informasi profil pengguna yang sedang login berdasarkan token autentikasi.
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

// GetUsers endpoint (admin only).
// Fungsi ini digunakan oleh admin untuk melihat daftar semua pengguna atau mencari pengguna tertentu.
func (h *AuthHandler) GetUsers(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	var users []models.User
	var err error

	if query != "" {
		users, err = h.Store.SearchUsers(query)
	} else {
		users, err = h.Store.GetAllUsers()
	}

	if err != nil {
		http.Error(w, "Error fetching users", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

// UpdateUser endpoint (khusus admin).
// Fungsi ini digunakan admin untuk memperbarui data pengguna lain.
func (h *AuthHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	var user models.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	// Pastikan ID tersedia
	if user.ID == "" {
		http.Error(w, "User ID required", http.StatusBadRequest)
		return
	}

	// Validasi role jika ada perubahan
	if user.Role != "" {
		validRoles := map[string]bool{
			"admin":     true,
			"mahasiswa": true,
			"guru":      true,
			"karyawan":  true,
		}
		if !validRoles[user.Role] {
			http.Error(w, "Invalid role. Must be one of: admin, mahasiswa, guru, karyawan", http.StatusBadRequest)
			return
		}
	}

	// Update user ke database
	if err := h.Store.UpdateUser(&user); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "User updated"})
}

// UpdateSelf endpoint.
// Fungsi ini memungkinkan pengguna untuk memperbarui data profil mereka sendiri.
func (h *AuthHandler) UpdateSelf(w http.ResponseWriter, r *http.Request) {
	// Ambil ID dari token (Claims)
	claims, ok := r.Context().Value(middleware.UserCtxKey).(*utils.Claims)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	user, err := h.Store.GetByUsername(claims.Username)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	var payload struct {
		Fullname string `json:"fullname"`
		NIP      string `json:"nip"`
		Contact  string `json:"contact"`
		Password string `json:"password,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	// Update field nama, nip, kontak
	user.Fullname = payload.Fullname
	user.NIP = payload.NIP
	user.Contact = payload.Contact

	// Update password jika ada (opsional)

	if err := h.Store.UpdateUser(user); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Simpan perubahan profil

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Profile updated successfully"})
}

// DeleteUser endpoint (khusus admin).
// Fungsi ini menghapus pengguna dari database berdasarkan ID yang diberikan.
func (h *AuthHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "ID required", http.StatusBadRequest)
		return
	}

	if err := h.Store.DeleteUser(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "User deleted"})
}
