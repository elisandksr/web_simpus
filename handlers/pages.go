package handlers

import (
	"fmt"
	"html/template"
	"latihan_cloud8/middleware"
	"latihan_cloud8/store"
	"latihan_cloud8/utils"
	"log"
	"net/http"
	"time"
)

type PageHandler struct {
	Store *store.MySQLStore
}

func NewPageHandler(s *store.MySQLStore) *PageHandler {
	return &PageHandler{Store: s}
}

func (h *PageHandler) ShowLoginPage(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/login.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, nil)
}

// Helper to render protected pages with layout
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
	}

	// Parse layout AND the specific content template
	tmpl, err := template.ParseFiles("templates/layout.html", "templates/"+tmplName)
	if err != nil {
		log.Println("Template Error:", err)
		http.Error(w, "Template Error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	err = tmpl.ExecuteTemplate(w, "layout", data)
	if err != nil {
		log.Println("Execute Error:", err)
	}
}

func (h *PageHandler) ShowDashboard(w http.ResponseWriter, r *http.Request) {
	v := r.Context().Value(middleware.UserCtxKey)
	if v == nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	claims := v.(*utils.Claims)

	// Get Real-time Date
	// Format: "Senin, 02 Januari 2006"
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
		myActiveLoans, _ := h.Store.CountActiveLoansByUser(claims.Username) // Assuming username is ID, or need to fetch ID?
		// Wait, claims.Username might not be the UUID ID. Let's check logic.
		// The store uses UUID for IDs. claims has Username.
		// I need to fetch the User to get the ID if CreateUser uses UUID.
		// Let's check CreateUser in store.
		// CreateUser: id := uuid.NewString().
		// So checking by Username in CountActiveLoansByUser won't work if it expects UUID.
		// But wait, the handler ShowProfile fetches User to get details.

		// Let's check `CountActiveLoansByUser` implementation I just wrote.
		// `SELECT COUNT(*) FROM loans WHERE user_id = ? ...`
		// `loans` table `user_id` refers to `users(id)`.
		// So I need the UUID.

		user, err := h.Store.GetByUsername(claims.Username)
		if err == nil {
			myActiveLoans, _ = h.Store.CountActiveLoansByUser(user.ID)
		}
		data["MyActiveLoans"] = myActiveLoans
		data["Status"] = "Aktif" // Just static for now as per requirement
	}

	tmpl, err := template.ParseFiles("templates/layout.html", "templates/dashboard.html")
	if err != nil {
		log.Println("Template Error:", err)
		http.Error(w, "Template Error", http.StatusInternalServerError)
		return
	}
	tmpl.ExecuteTemplate(w, "layout", data)
}

func (h *PageHandler) ShowAdminBooks(w http.ResponseWriter, r *http.Request) {
	render(w, r, "admin_books.html", "Manajemen Buku", "books")
}

func (h *PageHandler) ShowAdminMembers(w http.ResponseWriter, r *http.Request) {
	render(w, r, "admin_members.html", "Manajemen Anggota", "members")
}

func (h *PageHandler) ShowAdminTransactions(w http.ResponseWriter, r *http.Request) {
	render(w, r, "admin_transactions.html", "Transaksi", "transactions")
}

func (h *PageHandler) ShowLandingPage(w http.ResponseWriter, r *http.Request) {
	// If already logged in, maybe redirect to dashboard?
	// For now just show landing.
	data := map[string]interface{}{
		"Title": "Selamat Datang",
	}
	utils.RenderTemplate(w, "landing.html", data)
}

func (h *PageHandler) ShowCatalog(w http.ResponseWriter, r *http.Request) {
	render(w, r, "member_catalog.html", "Katalog Buku", "catalog")
}

func (h *PageHandler) ShowMyLoans(w http.ResponseWriter, r *http.Request) {
	render(w, r, "member_loans.html", "Peminjaman Saya", "loans")
}

func (h *PageHandler) ShowProfile(w http.ResponseWriter, r *http.Request) {
	// Custom render to include user data
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
		"Title":      "Profil Saya",
		"ActivePage": "profile",
		"User":       user,
	}

	tmpl, err := template.ParseFiles("templates/layout.html", "templates/profile.html")
	if err != nil {
		log.Println("Template Error:", err)
		http.Error(w, "Template Error", http.StatusInternalServerError)
		return
	}
	tmpl.ExecuteTemplate(w, "layout", data)
}

func (h *PageHandler) ShowAdminReports(w http.ResponseWriter, r *http.Request) {
	render(w, r, "admin_reports.html", "Laporan & Statistik", "reports")
}
