package store

import (
	"database/sql"
	"errors"
	"fmt"
	"latihan_cloud8/models"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
)

var (
	ErrUserExists   = errors.New("user already exists")
	ErrUserNotFound = errors.New("user not found")
	ErrBookNotFound = errors.New("book not found")
	ErrOutOfStock   = errors.New("book out of stock")
)

type MySQLStore struct {
	db *sql.DB
}

// NewMySQLStore menginisialisasi koneksi database MySQL baru.
// Fungsi ini membuka koneksi dan melakukan ping untuk memastikan database aktif.
func NewMySQLStore(dsn string) (*MySQLStore, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &MySQLStore{db: db}, nil
}

// Close menutup koneksi database.
func (s *MySQLStore) Close() error {
	return s.db.Close()
}

// InitSchema membuat tabel-tabel database jika belum ada.
// Tabel meliputi: users, books, loans, categories, settings, notifications.
func (s *MySQLStore) InitSchema() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id VARCHAR(36) PRIMARY KEY,
			username VARCHAR(255) NOT NULL UNIQUE,
			password VARCHAR(255) NOT NULL,
			role VARCHAR(50) NOT NULL,
			fullname VARCHAR(255),
			nip VARCHAR(50),
			contact VARCHAR(255),
			created_at DATETIME
		)`,
		`CREATE TABLE IF NOT EXISTS books (
			id INT AUTO_INCREMENT PRIMARY KEY,
			title VARCHAR(255) NOT NULL,
			author VARCHAR(255) NOT NULL,
			category VARCHAR(255),
			stock INT DEFAULT 0,
			image_url VARCHAR(255),
			published_year INT,
			created_at DATETIME
		)`,
		`CREATE TABLE IF NOT EXISTS loans (
			id INT AUTO_INCREMENT PRIMARY KEY,
			user_id VARCHAR(36) NOT NULL,
			book_id INT NOT NULL,
			loan_date DATETIME NOT NULL,
			due_date DATETIME NOT NULL,
			return_date DATETIME,
			status VARCHAR(50) NOT NULL,
			fine INT DEFAULT 0,
			FOREIGN KEY (user_id) REFERENCES users(id),
			FOREIGN KEY (book_id) REFERENCES books(id)
		)`,
		`CREATE TABLE IF NOT EXISTS settings (
			id INT PRIMARY KEY,
			max_loan_books INT DEFAULT 3,
			loan_duration INT DEFAULT 7,
			fine_per_day INT DEFAULT 5000
		)`,
		`CREATE TABLE IF NOT EXISTS categories (
			id INT AUTO_INCREMENT PRIMARY KEY,
			name VARCHAR(100) NOT NULL UNIQUE
		)`,
		`CREATE TABLE IF NOT EXISTS notifications (
			id INT AUTO_INCREMENT PRIMARY KEY,
			user_id VARCHAR(36) NOT NULL,
			message TEXT NOT NULL,
			is_read BOOLEAN DEFAULT FALSE,
			created_at DATETIME,
			FOREIGN KEY (user_id) REFERENCES users(id)
		)`,
	}

	for _, query := range queries {
		if _, err := s.db.Exec(query); err != nil {
			return fmt.Errorf("failed to execute query: %v, error: %w", query, err)
		}
	}

	// Migrasi untuk tabel yang sudah ada (menambah kolom baru dsb)
	s.db.Exec("ALTER TABLE users DROP COLUMN member_type") // Explicitly remove member_type
	s.db.Exec("ALTER TABLE users ADD COLUMN fullname VARCHAR(255)")
	s.db.Exec("ALTER TABLE users ADD COLUMN fullname VARCHAR(255)")
	// s.db.Exec("ALTER TABLE users ADD COLUMN member_type VARCHAR(50)") -- Removed
	s.db.Exec("ALTER TABLE users ADD COLUMN nip VARCHAR(50)")
	s.db.Exec("ALTER TABLE users ADD COLUMN nip VARCHAR(50)")
	s.db.Exec("ALTER TABLE users ADD COLUMN contact VARCHAR(255)")
	s.db.Exec("ALTER TABLE users ADD COLUMN contact VARCHAR(255)")
	s.db.Exec("ALTER TABLE books ADD COLUMN image_url VARCHAR(255)")
	s.db.Exec("ALTER TABLE books ADD COLUMN published_year INT")

	// Masukkan pengaturan default jika belum ada
	var settingsCount int
	s.db.QueryRow("SELECT COUNT(*) FROM settings").Scan(&settingsCount)
	if settingsCount == 0 {
		s.db.Exec("INSERT INTO settings (id, max_loan_books, loan_duration, fine_per_day) VALUES (1, 3, 7, 5000)")
	}

	return nil
}

// ==========================================
// USER
// ==========================================

// CreateUser menambahkan pengguna baru ke database.
// Melakukan pengecekan duplikasi username sebelum insert.
func (s *MySQLStore) CreateUser(username, hashedPassword, role, fullname string) (*models.User, error) {
	// Check if user exists
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM users WHERE username = ?", username).Scan(&count)
	if err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, ErrUserExists
	}

	// Create new user
	id := uuid.NewString()
	_, err = s.db.Exec(
		"INSERT INTO users (id, username, password, role, fullname, nip, contact, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		id, username, hashedPassword, role, fullname, "", "", time.Now(), // Default empty NIP/Contact for now, or update signature
	)
	if err != nil {
		return nil, err
	}

	return &models.User{
		ID:        id,
		Username:  username,
		Password:  hashedPassword,
		Role:      role,
		Fullname:  fullname,
		NIP:       "",
		Contact:   "",
		CreatedAt: time.Now(),
	}, nil
}

// GetByUsername mencari pengguna berdasarkan username.
func (s *MySQLStore) GetByUsername(username string) (*models.User, error) {
	user := &models.User{}
	var fullname, nip, contact sql.NullString // Handle potential nulls
	err := s.db.QueryRow(
		"SELECT id, username, password, role, fullname, nip, contact, created_at FROM users WHERE username = ?",
		username,
	).Scan(&user.ID, &user.Username, &user.Password, &user.Role, &fullname, &nip, &contact, &user.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}

	user.Fullname = fullname.String
	user.NIP = nip.String
	user.Contact = contact.String
	return user, nil
}

// GetAllUsers mengambil semua data pengguna.
func (s *MySQLStore) GetAllUsers() ([]models.User, error) {
	rows, err := s.db.Query("SELECT id, username, role, fullname, nip, contact, created_at FROM users")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		var fullname, nip, contact sql.NullString
		if err := rows.Scan(&user.ID, &user.Username, &user.Role, &fullname, &nip, &contact, &user.CreatedAt); err != nil {
			return nil, err
		}
		user.Fullname = fullname.String
		user.NIP = nip.String
		user.Contact = contact.String
		users = append(users, user)
	}
	return users, nil
}

// ==========================================
// USER
// ==========================================

// UpdateUser memperbarui data pengguna (nama, role, NIP, kontak).
func (s *MySQLStore) UpdateUser(user *models.User) error {
	_, err := s.db.Exec("UPDATE users SET fullname=?, role=?, nip=?, contact=? WHERE id=?",
		user.Fullname, user.Role, user.NIP, user.Contact, user.ID)
	return err
}

// DeleteUser menghapus data pengguna beserta data terkait (notifikasi, histori pinjaman).
func (s *MySQLStore) DeleteUser(id string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 1. Delete Notifications
	if _, err := tx.Exec("DELETE FROM notifications WHERE user_id = ?", id); err != nil {
		return fmt.Errorf("failed to delete notifications: %v", err)
	}

	// 2. Delete Loans (History)
	if _, err := tx.Exec("DELETE FROM loans WHERE user_id = ?", id); err != nil {
		return fmt.Errorf("failed to delete loans: %v", err)
	}

	// 3. Delete User
	if _, err := tx.Exec("DELETE FROM users WHERE id=?", id); err != nil {
		return fmt.Errorf("failed to delete user: %v", err)
	}

	return tx.Commit()
}

// SearchUsers mencari pengguna berdasarkan username, nama lengkap, atau NIP.
func (s *MySQLStore) SearchUsers(query string) ([]models.User, error) {
	q := "%" + query + "%"
	// Also search by NIP or Contact? Let's check Name, Username, NIP.
	rows, err := s.db.Query("SELECT id, username, role, fullname, nip, contact, created_at FROM users WHERE username LIKE ? OR fullname LIKE ? OR nip LIKE ?", q, q, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		var fullname, nip, contact sql.NullString
		if err := rows.Scan(&user.ID, &user.Username, &user.Role, &fullname, &nip, &contact, &user.CreatedAt); err != nil {
			return nil, err
		}
		user.Fullname = fullname.String
		user.NIP = nip.String
		user.Contact = contact.String
		users = append(users, user)
	}
	return users, nil
}

// ... existing CreateUser, GetByUsername, GetAllUsers ...

// ==========================================
// BOOKS
// ==========================================

// SearchBooks mencari buku berdasarkan judul, penulis, atau kategori.
func (s *MySQLStore) SearchBooks(query string) ([]models.Book, error) {
	q := "%" + query + "%"
	rows, err := s.db.Query("SELECT id, title, author, category, stock, image_url, published_year, created_at FROM books WHERE title LIKE ? OR author LIKE ? OR category LIKE ? ORDER BY created_at DESC", q, q, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var books []models.Book
	for rows.Next() {
		var b models.Book
		var imageURL sql.NullString
		var pubYear sql.NullInt64
		if err := rows.Scan(&b.ID, &b.Title, &b.Author, &b.Category, &b.Stock, &imageURL, &pubYear, &b.CreatedAt); err != nil {
			return nil, err
		}
		b.ImageURL = imageURL.String
		b.PublishedYear = int(pubYear.Int64)
		books = append(books, b)
	}
	return books, nil
}

// CreateBook menambahkan buku baru ke database.
func (s *MySQLStore) CreateBook(book *models.Book) error {
	res, err := s.db.Exec("INSERT INTO books (title, author, category, stock, image_url, published_year, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)",
		book.Title, book.Author, book.Category, book.Stock, book.ImageURL, book.PublishedYear, time.Now())
	if err != nil {
		return err
	}
	id, _ := res.LastInsertId()
	book.ID = int(id)
	return nil
}

// GetAllBooks mengambil semua daftar buku diurutkan dari yang terbaru.
func (s *MySQLStore) GetAllBooks() ([]models.Book, error) {
	rows, err := s.db.Query("SELECT id, title, author, category, stock, image_url, published_year, created_at FROM books ORDER BY created_at DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var books []models.Book
	for rows.Next() {
		var b models.Book
		var imageURL sql.NullString
		var pubYear sql.NullInt64
		if err := rows.Scan(&b.ID, &b.Title, &b.Author, &b.Category, &b.Stock, &imageURL, &pubYear, &b.CreatedAt); err != nil {
			return nil, err
		}
		b.ImageURL = imageURL.String
		b.PublishedYear = int(pubYear.Int64)
		books = append(books, b)
	}
	return books, nil
}

// GetBookByID mengambil detail buku berdasarkan ID.
func (s *MySQLStore) GetBookByID(id int) (*models.Book, error) {
	var b models.Book
	var imageURL sql.NullString
	var pubYear sql.NullInt64
	err := s.db.QueryRow("SELECT id, title, author, category, stock, image_url, published_year, created_at FROM books WHERE id = ?", id).
		Scan(&b.ID, &b.Title, &b.Author, &b.Category, &b.Stock, &imageURL, &pubYear, &b.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, ErrBookNotFound
	}
	if err != nil {
		return nil, err
	}
	b.ImageURL = imageURL.String
	b.PublishedYear = int(pubYear.Int64)
	return &b, nil
}

// UpdateBook memperbarui informasi buku.
func (s *MySQLStore) UpdateBook(book *models.Book) error {
	_, err := s.db.Exec("UPDATE books SET title=?, author=?, category=?, stock=?, image_url=?, published_year=? WHERE id=?",
		book.Title, book.Author, book.Category, book.Stock, book.ImageURL, book.PublishedYear, book.ID)
	return err
}

// DeleteBook menghapus buku berdasarkan ID.
func (s *MySQLStore) DeleteBook(id int) error {
	_, err := s.db.Exec("DELETE FROM books WHERE id=?", id)
	return err
}

// ==========================================
// LOANS
// ==========================================

// BorrowBook memproses peminjaman buku (cek stok, kurangi stok, buat record).
func (s *MySQLStore) BorrowBook(userID string, bookID, duration int) (*models.Loan, error) {
	// Mulai transaksi database
	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Cek ketersediaan stok buku
	var stock int
	err = tx.QueryRow("SELECT stock FROM books WHERE id = ?", bookID).Scan(&stock)
	if err == sql.ErrNoRows {
		return nil, ErrBookNotFound
	}
	if err != nil {
		return nil, err
	}
	if stock <= 0 {
		return nil, ErrOutOfStock
	}

	// Kurangi stok buku
	_, err = tx.Exec("UPDATE books SET stock = stock - 1 WHERE id = ?", bookID)
	if err != nil {
		return nil, err
	}

	// Buat record peminjaman
	loanDate := time.Now()
	dueDate := loanDate.AddDate(0, 0, duration)
	res, err := tx.Exec("INSERT INTO loans (user_id, book_id, loan_date, due_date, status) VALUES (?, ?, ?, ?, ?)",
		userID, bookID, loanDate, dueDate, "borrowed")
	if err != nil {
		return nil, err
	}

	loanID, _ := res.LastInsertId()

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &models.Loan{
		ID:       int(loanID),
		UserID:   userID,
		BookID:   bookID,
		LoanDate: loanDate,
		DueDate:  dueDate,
		Status:   "borrowed",
	}, nil
}

// ReturnBook memproses pengembalian buku (hitung denda, update status, tambah stok).
func (s *MySQLStore) ReturnBook(loanID int) (*models.Loan, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Ambil data peminjaman
	var l models.Loan
	var dueDate time.Time
	err = tx.QueryRow("SELECT id, user_id, book_id, status, due_date FROM loans WHERE id = ?", loanID).
		Scan(&l.ID, &l.UserID, &l.BookID, &l.Status, &dueDate)
	if err != nil {
		return nil, err
	}

	if l.Status == "returned" {
		return nil, errors.New("book already returned")
	}

	// Hitung Denda (Dinamis berdasarkan settings)
	var finePerDay int
	err = tx.QueryRow("SELECT fine_per_day FROM settings WHERE id=1").Scan(&finePerDay)
	if err != nil {
		finePerDay = 5000 // Default fallback
	}

	returnDate := time.Now()
	fine := 0
	if returnDate.After(dueDate) {
		daysLate := int(returnDate.Sub(dueDate).Hours() / 24)
		if daysLate < 1 {
			if returnDate.Day() != dueDate.Day() || returnDate.Month() != dueDate.Month() || returnDate.Year() != dueDate.Year() {
				daysLate = int(returnDate.Sub(dueDate).Hours() / 24)
				if daysLate == 0 {
					daysLate = 1
				}
			} else {
				daysLate = 0
			}
		}

		if daysLate > 0 {
			fine = daysLate * finePerDay
		}
	}

	// Update data peminjaman
	_, err = tx.Exec("UPDATE loans SET return_date=?, status=?, fine=? WHERE id=?",
		returnDate, "returned", fine, loanID)
	if err != nil {
		return nil, err
	}

	// Tambah stok buku kembali
	_, err = tx.Exec("UPDATE books SET stock = stock + 1 WHERE id = ?", l.BookID)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	// Isi struct hasil pengembalian
	l.Status = "returned"
	l.Fine = fine
	l.ReturnDate = &returnDate
	l.DueDate = dueDate // Ensure DueDate is set for the returned loan object
	// Power User Method for Worker
	return &l, nil
}

// GetAllBorrowedLoans mengambil semua peminjaman yang statusnya masih 'borrowed'.
func (s *MySQLStore) GetAllBorrowedLoans() ([]models.Loan, error) {
	rows, err := s.db.Query("SELECT id, user_id, book_id, loan_date, due_date, status FROM loans WHERE status = 'borrowed'")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var loans []models.Loan
	for rows.Next() {
		var l models.Loan
		var dueDate time.Time
		var loanDate time.Time
		if err := rows.Scan(&l.ID, &l.UserID, &l.BookID, &loanDate, &dueDate, &l.Status); err != nil {
			return nil, err
		}
		l.DueDate = dueDate
		l.LoanDate = loanDate
		loans = append(loans, l)
	}
	return loans, nil
}

// GetAllLoans mengambil semua riwayat peminjaman (join dengan user dan buku).
func (s *MySQLStore) GetAllLoans() ([]models.Loan, error) {
	// Join tables untuk tampilan lengkap
	query := `
		SELECT l.id, l.user_id, l.book_id, l.loan_date, l.due_date, l.return_date, l.status, l.fine,
		       b.title, u.username
		FROM loans l
		JOIN books b ON l.book_id = b.id
		JOIN users u ON l.user_id = u.id
		ORDER BY l.loan_date DESC
	`
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var loans []models.Loan
	for rows.Next() {
		var l models.Loan
		var returnDate sql.NullTime
		var bookTitle, username string

		err := rows.Scan(&l.ID, &l.UserID, &l.BookID, &l.LoanDate, &l.DueDate, &returnDate, &l.Status, &l.Fine, &bookTitle, &username)
		if err != nil {
			return nil, err
		}

		if returnDate.Valid {
			t := returnDate.Time
			l.ReturnDate = &t
		}

		l.Book = &models.Book{ID: l.BookID, Title: bookTitle}
		l.User = &models.User{ID: l.UserID, Username: username}

		loans = append(loans, l)
	}
	return loans, nil
}

// GetLoansFiltered mengambil riwayat peminjaman berdasarkan rentang tanggal.
func (s *MySQLStore) GetLoansFiltered(startDate, endDate time.Time) ([]models.Loan, error) {
	// Join tables untuk tampilan lengkap
	query := `
		SELECT l.id, l.user_id, l.book_id, l.loan_date, l.due_date, l.return_date, l.status, l.fine,
		       b.title, u.username
		FROM loans l
		JOIN books b ON l.book_id = b.id
		JOIN users u ON l.user_id = u.id
		WHERE l.loan_date >= ? AND l.loan_date <= ?
		ORDER BY l.loan_date DESC
	`
	rows, err := s.db.Query(query, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var loans []models.Loan
	for rows.Next() {
		var l models.Loan
		var returnDate sql.NullTime
		var bookTitle, username string

		err := rows.Scan(&l.ID, &l.UserID, &l.BookID, &l.LoanDate, &l.DueDate, &returnDate, &l.Status, &l.Fine, &bookTitle, &username)
		if err != nil {
			return nil, err
		}

		if returnDate.Valid {
			t := returnDate.Time
			l.ReturnDate = &t
		}

		l.Book = &models.Book{ID: l.BookID, Title: bookTitle}
		l.User = &models.User{ID: l.UserID, Username: username}

		loans = append(loans, l)
	}
	return loans, nil
}

// GetLoansByUserID mengambil riwayat peminjaman milik user tertentu.
func (s *MySQLStore) GetLoansByUserID(userID string) ([]models.Loan, error) {
	query := `
		SELECT l.id, l.user_id, l.book_id, l.loan_date, l.due_date, l.return_date, l.status, l.fine,
		       b.title
		FROM loans l
		JOIN books b ON l.book_id = b.id
		WHERE l.user_id = ?
		ORDER BY l.loan_date DESC
	`
	rows, err := s.db.Query(query, userID)
	if err != nil {
		log.Println("Error querying loans:", err)
		return nil, err
	}
	defer rows.Close()

	var loans []models.Loan
	for rows.Next() {
		var l models.Loan
		var returnDate sql.NullTime
		var bookTitle string

		err := rows.Scan(&l.ID, &l.UserID, &l.BookID, &l.LoanDate, &l.DueDate, &returnDate, &l.Status, &l.Fine, &bookTitle)
		if err != nil {
			log.Println("Error scanning loan:", err)
			return nil, err
		}

		if returnDate.Valid {
			t := returnDate.Time
			l.ReturnDate = &t
		}

		l.Book = &models.Book{ID: l.BookID, Title: bookTitle}
		loans = append(loans, l)
	}
	return loans, nil
}

// GetOverdueLoans mengambil daftar peminjaman yang terlambat dan belum dikembalikan.
func (s *MySQLStore) GetOverdueLoans(userID string) ([]models.Loan, error) {
	query := `
		SELECT l.id, l.book_id, l.due_date, b.title 
		FROM loans l
		JOIN books b ON l.book_id = b.id
		WHERE l.user_id = ? AND l.status = 'borrowed' AND l.due_date < NOW()
	`
	rows, err := s.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var loans []models.Loan
	for rows.Next() {
		var l models.Loan
		var bookTitle string
		if err := rows.Scan(&l.ID, &l.BookID, &l.DueDate, &bookTitle); err != nil {
			return nil, err
		}
		l.Book = &models.Book{ID: l.BookID, Title: bookTitle}
		loans = append(loans, l)
	}
	return loans, nil
}

// GetNotifications mengambil daftar notifikasi untuk user tertentu.
func (s *MySQLStore) GetNotifications(userID string) ([]models.Notification, error) {
	rows, err := s.db.Query("SELECT id, user_id, message, is_read, created_at FROM notifications WHERE user_id = ? ORDER BY created_at DESC", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notifs []models.Notification
	for rows.Next() {
		var n models.Notification
		if err := rows.Scan(&n.ID, &n.UserID, &n.Message, &n.IsRead, &n.CreatedAt); err != nil {
			return nil, err
		}
		notifs = append(notifs, n)
	}
	return notifs, nil
}

// MarkNotificationRead menandai notifikasi sebagai sudah dibaca.
func (s *MySQLStore) MarkNotificationRead(id int) error {
	_, err := s.db.Exec("UPDATE notifications SET is_read = TRUE WHERE id = ?", id)
	return err
}

// CreateNotification membuat notifikasi baru.
// Mencegah duplikasi pesan yang sama untuk user yang sama.
func (s *MySQLStore) CreateNotification(userID, message string) error {
	// Deduplikasi: Cek jika pesan yang sama sudah ada
	// Ini mencegah spamming notifikasi yang sama (misal dari worker).
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM notifications WHERE user_id = ? AND message = ?", userID, message).Scan(&count)
	if err != nil {
		return err
	}
	if count > 0 {
		return nil // Duplicate found (even if read), skip
	}

	_, err = s.db.Exec("INSERT INTO notifications (user_id, message, is_read, created_at) VALUES (?, ?, FALSE, ?)", userID, message, time.Now())
	return err
}

// DeleteNotification menghapus notifikasi.
func (s *MySQLStore) DeleteNotification(id int) error {
	_, err := s.db.Exec("DELETE FROM notifications WHERE id = ?", id)
	return err
}

// GetSettings mengambil pengaturan aplikasi.
func (s *MySQLStore) GetSettings() (*models.Settings, error) {
	var set models.Settings
	err := s.db.QueryRow("SELECT max_loan_books, loan_duration, fine_per_day FROM settings WHERE id = 1").
		Scan(&set.MaxLoanBooks, &set.LoanDuration, &set.FinePerDay)
	if err == sql.ErrNoRows {
		return &models.Settings{MaxLoanBooks: 3, LoanDuration: 7, FinePerDay: 5000}, nil // Default
	}
	if err != nil {
		return nil, err
	}
	return &set, nil
}

// Category Methods

// CreateCategory menambah kategori buku baru.
func (s *MySQLStore) CreateCategory(name string) error {
	_, err := s.db.Exec("INSERT INTO categories (name) VALUES (?)", name)
	return err
}

// GetAllCategories mengambil semua kategori buku.
func (s *MySQLStore) GetAllCategories() ([]models.Category, error) {
	rows, err := s.db.Query("SELECT id, name FROM categories ORDER BY name")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []models.Category
	for rows.Next() {
		var c models.Category
		if err := rows.Scan(&c.ID, &c.Name); err != nil {
			return nil, err
		}
		categories = append(categories, c)
	}
	return categories, nil
}

// DeleteCategory menghapus kategori.
func (s *MySQLStore) DeleteCategory(id int) error {
	_, err := s.db.Exec("DELETE FROM categories WHERE id=?", id)
	return err
}

// ==========================================
// DASHBOARD STATS
// ==========================================

// CountUsers menghitung total pengguna terdaftar.
func (s *MySQLStore) CountUsers() (int, error) {
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	return count, err
}

// CountBooks menghitung total buku di database.
func (s *MySQLStore) CountBooks() (int, error) {
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM books").Scan(&count)
	return count, err
}

// CountTotalActiveLoans menghitung total peminjaman yang masih aktif.
func (s *MySQLStore) CountTotalActiveLoans() (int, error) {
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM loans WHERE status = 'borrowed'").Scan(&count)
	return count, err
}

// CountActiveLoansByUser menghitung peminjaman aktif milik user tertentu.
func (s *MySQLStore) CountActiveLoansByUser(userID string) (int, error) {
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM loans WHERE user_id = ? AND status = 'borrowed'", userID).Scan(&count)
	return count, err
}
