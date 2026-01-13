package handlers

import (
	"encoding/json"
	"fmt"
	"latihan_cloud8/middleware"
	"latihan_cloud8/models"
	"latihan_cloud8/store"
	"latihan_cloud8/utils"
	"net/http"
	"time"
)

type LoanHandler struct {
	Store *store.MySQLStore
}

func NewLoanHandler(store *store.MySQLStore) *LoanHandler {
	return &LoanHandler{Store: store}
}

// Borrow endpoint.
// Menangani proses peminjaman buku oleh pengguna.
func (h *LoanHandler) Borrow(w http.ResponseWriter, r *http.Request) {
	// Ambil ID pengguna dari token
	claims, ok := r.Context().Value(middleware.UserCtxKey).(*utils.Claims)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	user, err := h.Store.GetByUsername(claims.Username)
	if err != nil {
		http.Error(w, "User not found", http.StatusInternalServerError)
		return
	}

	// Cek batasan jumlah pinjaman
	settings, err := h.Store.GetSettings()
	if err != nil {
		// Gunakan default jika setting gagal dimuat
		settings = &models.Settings{MaxLoanBooks: 3}
	}

	activeLoans, err := h.Store.GetLoansByUserID(user.ID)
	if err != nil {
		http.Error(w, "Error checking loans", http.StatusInternalServerError)
		return
	}

	activeCount := 0
	for _, l := range activeLoans {
		if l.Status == "borrowed" {
			activeCount++
		}
	}

	if activeCount >= settings.MaxLoanBooks {
		http.Error(w, fmt.Sprintf("Batas maksimal peminjaman tercapai (%d buku)", settings.MaxLoanBooks), http.StatusBadRequest)
		return
	}

	var payload models.LoanRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	// Validasi durasi pinjaman
	duration := payload.Duration
	if duration <= 0 {
		duration = settings.LoanDuration // Default dari setting
	}
	// Pastikan durasi tidak melebihi batas maksimal
	if duration > settings.LoanDuration {
		http.Error(w, fmt.Sprintf("Durasi maksimal peminjaman adalah %d hari", settings.LoanDuration), http.StatusBadRequest)
		return
	}

	loan, err := h.Store.BorrowBook(user.ID, payload.BookID, duration)
	if err == store.ErrBookNotFound {
		http.Error(w, "Book not found", http.StatusNotFound)
		return
	}
	if err == store.ErrOutOfStock {
		http.Error(w, "Out of stock", http.StatusBadRequest)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Buat notifikasi peminjaman
	// Ambil judul buku
	book, _ := h.Store.GetBookByID(payload.BookID)
	title := "Buku"
	if book != nil {
		title = book.Title
	}
	msg := fmt.Sprintf("Peminjaman berhasil: %s. Batas waktu: %s", title, loan.DueDate.Format("02 Jan 2006"))
	h.Store.CreateNotification(user.ID, msg)

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(loan)
}

// Return endpoint (khusus admin).
// Menangani pengembalian buku dan perhitungan denda.
func (h *LoanHandler) Return(w http.ResponseWriter, r *http.Request) {
	// Proses pengembalian buku (biasanya oleh admin)
	var payload struct {
		LoanID int `json:"loan_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	loan, err := h.Store.ReturnBook(payload.LoanID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Buat notifikasi pengembalian
	// Ambil judul buku untuk pesan notifikasi
	book, _ := h.Store.GetBookByID(loan.BookID)
	title := "Buku"
	if book != nil {
		title = book.Title
	}
	msg := fmt.Sprintf("Pengembalian berhasil: %s. Denda: Rp %d", title, loan.Fine)
	h.Store.CreateNotification(loan.UserID, msg)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Book returned successfully"})
}

// ListLoans endpoint.
// Menampilkan daftar peminjaman (semua untuk admin, milik sendiri untuk user).
func (h *LoanHandler) ListLoans(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(middleware.UserCtxKey).(*utils.Claims)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var loans []models.Loan
	var err error

	if claims.Role == "admin" {
		startDateStr := r.URL.Query().Get("start_date")
		endDateStr := r.URL.Query().Get("end_date")

		if startDateStr != "" && endDateStr != "" {
			layout := "2006-01-02"
			start, err1 := time.Parse(layout, startDateStr)
			end, err2 := time.Parse(layout, endDateStr)

			if err1 == nil && err2 == nil {
				// Set waktu akhir ke penghujung hari
				end = end.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
				loans, err = h.Store.GetLoansFiltered(start, end)
			} else {
				// Jika format tanggal salah, ambil semua data
				loans, err = h.Store.GetAllLoans()
			}
		} else {
			loans, err = h.Store.GetAllLoans()
		}
	} else {
		// Ambil data user dahulu
		user, uErr := h.Store.GetByUsername(claims.Username)
		if uErr != nil {
			http.Error(w, "User error", http.StatusInternalServerError)
			return
		}
		loans, err = h.Store.GetLoansByUserID(user.ID)
	}

	if err != nil {
		http.Error(w, "Error fetching loans", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(loans)
}
