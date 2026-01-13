package handlers

import (
	"encoding/json"
	"latihan_cloud8/store"
	"net/http"
	"strconv"
)

type CategoryHandler struct {
	Store *store.MySQLStore
}

func NewCategoryHandler(store *store.MySQLStore) *CategoryHandler {
	return &CategoryHandler{Store: store}
}

// GetCategories endpoint.
// Mengambil daftar semua kategori buku.
func (h *CategoryHandler) GetCategories(w http.ResponseWriter, r *http.Request) {
	cats, err := h.Store.GetAllCategories()
	if err != nil {
		http.Error(w, "Error fetching categories", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cats)
}

// CreateCategory endpoint (khusus admin).
// Menambahkan kategori baru.
func (h *CategoryHandler) CreateCategory(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	if payload.Name == "" {
		http.Error(w, "Name required", http.StatusBadRequest)
		return
	}

	if err := h.Store.CreateCategory(payload.Name); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Category created"})
}

// DeleteCategory endpoint (khusus admin).
// Menghapus kategori berdasarkan ID.
func (h *CategoryHandler) DeleteCategory(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	if err := h.Store.DeleteCategory(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Category deleted"})
}
