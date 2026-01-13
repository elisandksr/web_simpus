package models

import "time"

// Loan merepresentasikan data peminjaman buku.
type Loan struct {
	ID         int        `json:"id" db:"id"`
	UserID     string     `json:"user_id" db:"user_id"`
	BookID     int        `json:"book_id" db:"book_id"`
	User       *User      `json:"user,omitempty"`
	Book       *Book      `json:"book,omitempty"`
	LoanDate   time.Time  `json:"loan_date" db:"loan_date"`
	DueDate    time.Time  `json:"due_date" db:"due_date"`
	ReturnDate *time.Time `json:"return_date" db:"return_date"`
	Status     string     `json:"status" db:"status"` // "borrowed", "returned", "late"
	Fine       int        `json:"fine" db:"fine"`
}

// LoanRequest adalah payload untuk membuat peminjaman baru.
type LoanRequest struct {
	BookID   int `json:"book_id"`
	Duration int `json:"duration"` // Durasi pinjam (hari)
}
