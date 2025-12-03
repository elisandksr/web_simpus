package handlers

import (
	"html/template"
	"log"
	"latihan_cloud8/middleware"
	"latihan_cloud8/utils"
	"net/http"
)

type PageHandler struct{}

func NewPageHandler() *PageHandler {
	return &PageHandler{}
}

// ShowLoginPage menampilkan halaman login
func (h *PageHandler) ShowLoginPage(w http.ResponseWriter, r *http.Request) {
	log.Println("ShowLoginPage called")
	
	tmpl, err := template.ParseFiles("templates/login.html")
	if err != nil {
		log.Println("Error loading template:", err)
		http.Error(w, "Error loading template: "+err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, nil)
}

// ShowAdminPage menampilkan halaman admin (protected)
// ============================================
// PENTING: Halaman ini HARUS dipanggil setelah middleware.AuthMiddleware
// ============================================
func (h *PageHandler) ShowAdminPage(w http.ResponseWriter, r *http.Request) {
	log.Println("ShowAdminPage called")
	
	// STEP 1: Get user claims dari context (yang disimpan oleh AuthMiddleware)
	v := r.Context().Value(middleware.UserCtxKey)
	if v == nil {
		log.Println("ERROR: No user in context - redirecting to login")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// STEP 2: Type assert claims
	claims, ok := v.(*utils.Claims)
	if !ok {
		log.Println("ERROR: Invalid auth context type - redirecting to login")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	log.Println("User authenticated:", claims.Username, "Role:", claims.Role)

	// STEP 3: Prepare data untuk template
	data := map[string]interface{}{
		"Username": claims.Username,
		"Role":     claims.Role,
	}

	// STEP 4: Parse dan execute template
	tmpl, err := template.ParseFiles("templates/admin.html")
	if err != nil {
		log.Println("Error loading admin template:", err)
		http.Error(w, "Error loading template: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	tmpl.Execute(w, data)
}