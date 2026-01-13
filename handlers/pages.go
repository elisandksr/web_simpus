package handlers

import (
	"fmt"
	"html/template"
	"latihan_cloud8/middleware"
	"latihan_cloud8/store"
	"latihan_cloud8/utils"
	"net/http"
	"time"
)

type PageHandler struct {
	Store *store.MySQLStore
}

func NewPageHandler(s *store.MySQLStore) *PageHandler {
	return &PageHandler{Store: s}
}

// ShowLoginPage handler.
// Menampilkan halaman login.
func (h *PageHandler) ShowLoginPage(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/login.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, nil)
}

// Helper untuk merender halaman yang dilindungi (butuh login)
func render(w http.ResponseWriter, r *http.Request, tmplName string, title string, activePage string) {
	v := r.Context().Value(middleware.UserCtxKey)
	if v == nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	claims, ok := v.(*utils.Claims)
	if !ok {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	data := map[string]interface{}{
		"Username":   claims.Username,
		"Role":       claims.Role,
		"Title":      title,
		"ActivePage": activePage,
		"Subtitle":   "Selamat Datang di Sistem Manajemen Perpustakaan Terpadu",
	}

	if title == "Riwayat Peminjaman" {
		data["Subtitle"] = "Riwayat Peminjaman"
	}

	// Render template dengan layout
	utils.RenderTemplate(w, tmplName, data)
}

// ShowDashboard handler.
// Menampilkan dashboard utama dengan statistik yang sesuai peran pengguna.
func (h *PageHandler) ShowDashboard(w http.ResponseWriter, r *http.Request) {
	// Render dengan data user tambahan
	v := r.Context().Value(middleware.UserCtxKey)
	if v == nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	claims := v.(*utils.Claims)

	// Ambil Tanggal Real-time
	now := time.Now()
	days := []string{"Minggu", "Senin", "Selasa", "Rabu", "Kamis", "Jumat", "Sabtu"}
	months := []string{"", "Januari", "Februari", "Maret", "April", "Mei", "Juni", "Juli", "Agustus", "September", "Oktober", "November", "Desember"}

	dayName := days[now.Weekday()]
	monthName := months[now.Month()]
	currentDate := fmt.Sprintf("%s, %d %s %d", dayName, now.Day(), monthName, now.Year())

	data := map[string]interface{}{
		"Username":    claims.Username,
		"Role":        claims.Role,
		"Title":       "Dashboard",
		"ActivePage":  "dashboard",
		"CurrentDate": currentDate,
	}

	if claims.Role == "admin" {
		usersCount, _ := h.Store.CountUsers()
		booksCount, _ := h.Store.CountBooks()
		activeLoansCount, _ := h.Store.CountTotalActiveLoans()

		data["TotalUsers"] = usersCount
		data["TotalBooks"] = booksCount
		data["ActiveLoans"] = activeLoansCount
	} else {
		user, err := h.Store.GetByUsername(claims.Username)
		myActiveLoans := 0
		if err == nil {
			myActiveLoans, _ = h.Store.CountActiveLoansByUser(user.ID)
		}
		data["MyActiveLoans"] = myActiveLoans
		data["Status"] = "Aktif"
	}

	utils.RenderTemplate(w, "dashboard.html", data)
}

// ShowAdminBooks handler.
// Menampilkan halaman manajemen buku (admin).
func (h *PageHandler) ShowAdminBooks(w http.ResponseWriter, r *http.Request) {
	render(w, r, "admin_books.html", "Manajemen Buku", "books")
}

// ShowAdminMembers handler.
// Menampilkan halaman manajemen anggota (admin).
func (h *PageHandler) ShowAdminMembers(w http.ResponseWriter, r *http.Request) {
	render(w, r, "admin_members.html", "Manajemen Anggota", "members")
}

// ShowAdminTransactions handler.
// Menampilkan halaman transaksi (admin).
func (h *PageHandler) ShowAdminTransactions(w http.ResponseWriter, r *http.Request) {
	render(w, r, "admin_transactions.html", "Transaksi", "transactions")
}

// ShowLandingPage handler.
// Menampilkan halaman landing (depan).
func (h *PageHandler) ShowLandingPage(w http.ResponseWriter, r *http.Request) {
	// Tampilkan halaman landing
	data := map[string]interface{}{
		"Title": "Selamat Datang",
	}
	utils.RenderTemplate(w, "landing.html", data)
}

// ShowCatalog handler.
// Menampilkan katalog buku untuk anggota.
func (h *PageHandler) ShowCatalog(w http.ResponseWriter, r *http.Request) {
	render(w, r, "member_catalog.html", "Katalog Buku", "catalog")
}

// ShowMyLoans handler.
// Menampilkan halaman peminjaman anggota (aktif maupun riwayat).
func (h *PageHandler) ShowMyLoans(w http.ResponseWriter, r *http.Request) {
	title := "Peminjaman"
	if r.URL.Query().Get("view") == "history" {
		title = "Riwayat Peminjaman"
	}
	render(w, r, "member_loans.html", title, "loans")
}

// ShowProfile handler.
// Menampilkan halaman profil pengguna.
func (h *PageHandler) ShowProfile(w http.ResponseWriter, r *http.Request) {
	// Render dengan data user lengkap
	v := r.Context().Value(middleware.UserCtxKey)
	if v == nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	claims := v.(*utils.Claims)

	user, err := h.Store.GetByUsername(claims.Username)
	if err != nil {
		http.Error(w, "User not found", http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"Username":   claims.Username,
		"Role":       claims.Role,
		"Title":      "Profil",
		"ActivePage": "profile",
		"User":       user,
	}

	utils.RenderTemplate(w, "profile.html", data)
}

// ShowAdminReports handler.
// Menampilkan halaman laporan untuk admin.
func (h *PageHandler) ShowAdminReports(w http.ResponseWriter, r *http.Request) {
	render(w, r, "admin_reports.html", "Laporan & Statistik", "reports")
}
