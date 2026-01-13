package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"latihan_cloud8/models"
	"latihan_cloud8/store"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

type BookHandler struct {
	Store *store.MySQLStore
}

func NewBookHandler(store *store.MySQLStore) *BookHandler {
	return &BookHandler{Store: store}
}

// GetBooks endpoint.
// Mengambil daftar buku, bisa dengan filter pencarian.
func (h *BookHandler) GetBooks(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	var books []models.Book
	var err error

	if query != "" {
		books, err = h.Store.SearchBooks(query)
	} else {
		books, err = h.Store.GetAllBooks()
	}

	if err != nil {
		http.Error(w, "Error fetching books", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(books)
}

// ... existing code ...

// CreateBook endpoint (khusus admin).
// Menambah buku baru beserta upload gambar sampul.
func (h *BookHandler) CreateBook(w http.ResponseWriter, r *http.Request) {
	// Parse data form multipart (limit 10MB)
	if err := r.ParseMultipartForm(10 << 20); err != nil { // 10MB limit
		http.Error(w, "Error parsing form", http.StatusBadRequest)
		return
	}

	title := r.FormValue("title")
	author := r.FormValue("author")
	category := r.FormValue("category")
	stockStr := r.FormValue("stock")
	stock, _ := strconv.Atoi(stockStr)
	yearStr := r.FormValue("published_year")
	year, _ := strconv.Atoi(yearStr)

	if title == "" || stock < 0 {
		http.Error(w, "Invalid data. Title required, stock >= 0", http.StatusBadRequest)
		return
	}

	// Proses upload gambar jika ada
	imageURL := ""
	file, handler, err := r.FormFile("image")
	if err == nil {
		defer file.Close()

		// Buat nama file unik
		ext := filepath.Ext(handler.Filename)
		filename := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)

		// Buat folder jika belum ada
		uploadDir := filepath.Join("upload", "books")
		os.MkdirAll(uploadDir, os.ModePerm)

		fpath := filepath.Join(uploadDir, filename)

		// Simpan file ke disk
		dst, err := os.Create(fpath)
		if err != nil {
			http.Error(w, "Error saving file", http.StatusInternalServerError)
			return
		}
		defer dst.Close()
		if _, err := io.Copy(dst, file); err != nil {
			http.Error(w, "Error saving file content", http.StatusInternalServerError)
			return
		}
		imageURL = "/upload/books/" + filename
	}

	book := &models.Book{
		Title:         title,
		Author:        author,
		Category:      category,
		Stock:         stock,
		PublishedYear: year,
		ImageURL:      imageURL,
	}

	if err := h.Store.CreateBook(book); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(book)
}

// UpdateBook endpoint (khusus admin).
// Memperbarui data buku yang ada.
func (h *BookHandler) UpdateBook(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	// Parse data form multipart
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "Error parsing form", http.StatusBadRequest)
		return
	}

	// Ambil data buku lama
	book, err := h.Store.GetBookByID(id)
	if err != nil {
		http.Error(w, "Book not found", http.StatusNotFound)
		return
	}

	// Update field jika ada perubahan
	if title := r.FormValue("title"); title != "" {
		book.Title = title
	}
	if author := r.FormValue("author"); author != "" {
		book.Author = author
	}
	if category := r.FormValue("category"); category != "" {
		book.Category = category
	}
	if stockStr := r.FormValue("stock"); stockStr != "" {
		if stock, err := strconv.Atoi(stockStr); err == nil {
			book.Stock = stock
		}
	}
	if yearStr := r.FormValue("published_year"); yearStr != "" {
		if year, err := strconv.Atoi(yearStr); err == nil {
			book.PublishedYear = year
		}
	}

	// Update gambar jika ada upload baru
	file, handler, err := r.FormFile("image")
	if err == nil {
		defer file.Close()
		ext := filepath.Ext(handler.Filename)
		filename := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)

		// Pastikan folder ada
		uploadDir := filepath.Join("upload", "books")
		os.MkdirAll(uploadDir, os.ModePerm)

		fpath := filepath.Join(uploadDir, filename)

		dst, err := os.Create(fpath)
		if err != nil {
			http.Error(w, "Error saving file", http.StatusInternalServerError)
			return
		}
		defer dst.Close()
		if _, err := io.Copy(dst, file); err == nil {
			book.ImageURL = "/upload/books/" + filename
		}
	}

	if err := h.Store.UpdateBook(book); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Book updated"})
}

// DeleteBook endpoint (khusus admin).
// Menghapus buku dari database berdasarkan ID.
func (h *BookHandler) DeleteBook(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	if err := h.Store.DeleteBook(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Book deleted"})
}
