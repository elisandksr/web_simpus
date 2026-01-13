package models

// Settings merepresentasikan pengaturan global aplikasi.
type Settings struct {
	ID           int `json:"id" db:"id"`
	MaxLoanBooks int `json:"max_loan_books" db:"max_loan_books"`
	LoanDuration int `json:"loan_duration" db:"loan_duration"` // Dalam hari
	FinePerDay   int `json:"fine_per_day" db:"fine_per_day"`
}
