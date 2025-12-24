package main

import (
	"latihan_cloud8/handlers"
	"latihan_cloud8/middleware"
	"latihan_cloud8/store"
	"latihan_cloud8/workers" // Import workers
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

	// Initialize Schema (SIMPUS Tables)
	if err := st.InitSchema(); err != nil {
		log.Fatalf("Failed to upload schema: %v", err)
	}

	log.Println("✅ Successfully connected to MySQL database")

	// ============================================
	// HANDLERS INITIALIZATION
	// ============================================	// Init Handlers
	// Init Handlers
	authHandler := handlers.NewAuthHandler(st)
	bookHandler := handlers.NewBookHandler(st)
	loanHandler := handlers.NewLoanHandler(st)

	categoryHandler := handlers.NewCategoryHandler(st)
	pageHandler := handlers.NewPageHandler(st)          // Inject store
	notifHandler := handlers.NewNotificationHandler(st) // Init Notification Handler

	// Start Background Worker
	notifier := workers.NewNotifier(st)
	notifier.Start()

	// ============================================
	// ROUTES SETUP
	// ============================================
	mux := http.NewServeMux()

	// Static files
	fs := http.FileServer(http.Dir("static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	// Uploads
	up := http.FileServer(http.Dir("upload"))
	mux.Handle("/upload/", http.StripPrefix("/upload/", up))

	// Public Routes
	mux.HandleFunc("/", pageHandler.ShowLandingPage)
	mux.HandleFunc("/login", pageHandler.ShowLoginPage)
	mux.HandleFunc("/api/login", authHandler.Login) // distinct from page
	mux.HandleFunc("/register", authHandler.Register)
	mux.HandleFunc("/api/books", bookHandler.GetBooks)

	mux.Handle("/api/profile/update", middleware.AuthMiddleware(http.HandlerFunc(authHandler.UpdateSelf)))

	// Protected Page Routes (UI)
	// Common
	mux.Handle("/dashboard", middleware.AuthMiddleware(http.HandlerFunc(pageHandler.ShowDashboard)))
	mux.Handle("/profile", middleware.AuthMiddleware(http.HandlerFunc(pageHandler.ShowProfile)))

	// Admin UI
	mux.Handle("/admin", middleware.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
	}))) // Redirect /admin to dashboard
	mux.Handle("/admin/books", middleware.AuthMiddleware(middleware.RequireRole("admin")(http.HandlerFunc(pageHandler.ShowAdminBooks))))
	mux.Handle("/admin/members", middleware.AuthMiddleware(middleware.RequireRole("admin")(http.HandlerFunc(pageHandler.ShowAdminMembers))))
	mux.Handle("/admin/transactions", middleware.AuthMiddleware(middleware.RequireRole("admin")(http.HandlerFunc(pageHandler.ShowAdminTransactions))))
	mux.Handle("/admin/reports", middleware.AuthMiddleware(middleware.RequireRole("admin")(http.HandlerFunc(pageHandler.ShowAdminReports))))

	// Member UI
	mux.Handle("/catalog", middleware.AuthMiddleware(http.HandlerFunc(pageHandler.ShowCatalog)))
	mux.Handle("/loans", middleware.AuthMiddleware(http.HandlerFunc(pageHandler.ShowMyLoans)))

	// API Routes (JSON)
	mux.Handle("/api/profile", middleware.AuthMiddleware(http.HandlerFunc(authHandler.Profile)))
	mux.Handle("/api/users", middleware.AuthMiddleware(middleware.RequireRole("admin")(http.HandlerFunc(authHandler.GetUsers))))
	mux.Handle("/api/users/update", middleware.AuthMiddleware(middleware.RequireRole("admin")(http.HandlerFunc(authHandler.UpdateUser))))
	mux.Handle("/api/users/delete", middleware.AuthMiddleware(middleware.RequireRole("admin")(http.HandlerFunc(authHandler.DeleteUser))))

	mux.Handle("/api/books/create", middleware.AuthMiddleware(middleware.RequireRole("admin")(http.HandlerFunc(bookHandler.CreateBook))))
	mux.Handle("/api/books/update", middleware.AuthMiddleware(middleware.RequireRole("admin")(http.HandlerFunc(bookHandler.UpdateBook))))
	mux.Handle("/api/books/delete", middleware.AuthMiddleware(middleware.RequireRole("admin")(http.HandlerFunc(bookHandler.DeleteBook))))

	mux.Handle("/api/categories", middleware.AuthMiddleware(http.HandlerFunc(categoryHandler.GetCategories)))
	mux.Handle("/api/categories/create", middleware.AuthMiddleware(middleware.RequireRole("admin")(http.HandlerFunc(categoryHandler.CreateCategory))))
	mux.Handle("/api/categories/delete", middleware.AuthMiddleware(middleware.RequireRole("admin")(http.HandlerFunc(categoryHandler.DeleteCategory))))

	mux.Handle("/api/loans", middleware.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			loanHandler.Borrow(w, r)
		} else {
			loanHandler.ListLoans(w, r)
		}
	})))
	mux.Handle("/api/loans/return", middleware.AuthMiddleware(middleware.RequireRole("admin")(http.HandlerFunc(loanHandler.Return))))

	// Notification Routes
	mux.Handle("/notifications", middleware.AuthMiddleware(http.HandlerFunc(notifHandler.ShowNotificationsPage)))
	mux.Handle("/api/notifications", middleware.AuthMiddleware(http.HandlerFunc(notifHandler.GetNotifications)))
	mux.Handle("/api/notifications/read", middleware.AuthMiddleware(http.HandlerFunc(notifHandler.MarkRead)))
	mux.Handle("/api/notifications/delete", middleware.AuthMiddleware(http.HandlerFunc(notifHandler.DeleteNotification)))
	mux.Handle("/api/notifications/send", middleware.AuthMiddleware(middleware.RequireRole("admin")(http.HandlerFunc(notifHandler.SendNotification))))

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
