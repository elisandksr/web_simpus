package main

import (
	"latihan_cloud8/handlers"
	"latihan_cloud8/middleware"
	"latihan_cloud8/store"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	// ============================================
	// DATABASE CONFIGURATION
	// ============================================
	dbUser := os.Getenv("DB_USER")
	if dbUser == "" {
		dbUser = "root"
	}

	dbPass := os.Getenv("DB_PASS")
	if dbPass == "" {
		dbPass = ""
	}

	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		dbHost = "localhost"
	}

	dbPort := os.Getenv("DB_PORT")
	if dbPort == "" {
		dbPort = "3306"
	}

	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		dbName = "jwt_auth_db"
	}

	// DSN format: user:password@tcp(host:port)/dbname?parseTime=true
	dsn := dbUser + ":" + dbPass + "@tcp(" + dbHost + ":" + dbPort + ")/" + dbName + "?parseTime=true"

	// Initialize MySQL store
	st, err := store.NewMySQLStore(dsn)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer st.Close()

	log.Println("✅ Successfully connected to MySQL database")

	// ============================================
	// HANDLERS INITIALIZATION
	// ============================================
	authHandler := handlers.NewAuthHandler(st)
	pageHandler := handlers.NewPageHandler()

	// ============================================
	// ROUTES SETUP
	// ============================================
	mux := http.NewServeMux()

	// Static files
	fs := http.FileServer(http.Dir("static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	// ============================================
	// PUBLIC ROUTES (Tidak perlu auth)
	// ============================================
	log.Println("Setting up public routes...")
	mux.HandleFunc("/", pageHandler.ShowLoginPage)              // GET - Halaman login
	mux.HandleFunc("/login", authHandler.Login)                // POST - API login
	mux.HandleFunc("/register", authHandler.Register)          // POST - API register

	// ============================================
	// PROTECTED ROUTES (Perlu auth)
	// ============================================
	log.Println("Setting up protected routes...")

	// PENTING: Gunakan middleware.AuthMiddleware untuk protect routes
	// Middleware akan cek token dari Authorization header ATAU cookie
	
	// Admin page - Protected
	mux.Handle("/admin", middleware.AuthMiddleware(http.HandlerFunc(pageHandler.ShowAdminPage)))

	// API Profile - Protected
	mux.Handle("/api/profile", middleware.AuthMiddleware(http.HandlerFunc(authHandler.Profile)))

	// Admin only routes
	adminUsersHandler := middleware.RequireRole("admin")(http.HandlerFunc(authHandler.GetUsers))
	mux.Handle("/api/users", middleware.AuthMiddleware(adminUsersHandler))

	// ============================================
	// APPLY LOGGING MIDDLEWARE GLOBALLY
	// ============================================
	handler := middleware.Logging(mux)

	// ============================================
	// SERVER CONFIGURATION
	// ============================================
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	log.Printf("🚀 Server running on http://localhost:%s\n", port)
	log.Printf("🔗 Login: http://localhost:%s/\n", port)
	log.Printf("📊 Admin: http://localhost:%s/admin (after login)\n", port)
	log.Fatal(srv.ListenAndServe())
}