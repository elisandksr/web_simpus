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

func (h *LoanHandler) Borrow(w http.ResponseWriter, r *http.Request) {
	// Get User ID
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

	// CHECK MAX LOANS
	settings, err := h.Store.GetSettings()
	if err != nil {
		// Fallback if settings fail
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

	// Duration Validation
	duration := payload.Duration
	if duration <= 0 {
		duration = settings.LoanDuration // Default if 0
	}
	// Optional: Enforce Max Duration from Settings if we consider 'LoanDuration' as Max
	// For now, let's treat Settings.LoanDuration as the standard limit.
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

	// NOTIFICATION
	// Fetch book title
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

func (h *LoanHandler) Return(w http.ResponseWriter, r *http.Request) {
	// Admin only usually, or user if allowed. Requirement says "Manajemen .. Peminjaman & pengembalian".
	// Let's assume Admin handles return.
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

	// NOTIFICATION
	// Need to fetch book title again because ReturnBook only gives IDs in the struct usually (unless we joined, but we simple queried)
	// We can add title to ReturnBook return or just fetch it here.
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
				// Adjust end date to end of day
				end = end.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
				loans, err = h.Store.GetLoansFiltered(start, end)
			} else {
				// Fallback or bad request? Let's just return all or error.
				// For robustness, let's fallback to current month or just error.
				// Let's fallback to all loans but maybe invalid date format.
				loans, err = h.Store.GetAllLoans()
			}
		} else {
			loans, err = h.Store.GetAllLoans()
		}
	} else {
		// Get User ID
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

func (h *LoanHandler) Extend(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(middleware.UserCtxKey).(*utils.Claims)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var payload struct {
		LoanID int `json:"loan_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	if err := h.Store.ExtendLoan(payload.LoanID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// NOTIFICATION
	// Get loan to know user and book
	// This is inefficient (multiple queries) but works.
	// We can't easily get info from ExtendLoan without changing its signature too.
	// For now, let's just send generic notification or skip.
	// Requirement: "Peringatan jatuh tempo" etc.
	// "User langsung mendapatkan notifikasi peminjaman berhasil" implies Borrow.
	// Let's assume Extend also notifies.
	// I won't implement Extend detail fetch right now to save time/complexity, or just fetch loan?
	// Let's fetch loan details.
	// Actually ExtendLoan updates due date.
	// I'll skip fetching detailed title for Extend to be safe/fast.
	// Or better:
	// h.Store.CreateNotification(claims.UserID, "Perpanjangan buku berhasil.") (Using claims username -> id lookup?)
	// Actually we have valid user session.
	user, _ := h.Store.GetByUsername(claims.Username)
	if user != nil {
		h.Store.CreateNotification(user.ID, "Perpanjangan peminjaman buku berhasil.")
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Loan extended successfully"})
}
