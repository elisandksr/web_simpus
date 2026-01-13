package models

// Category merepresentasikan data kategori buku.
type Category struct {
	ID   int    `json:"id" db:"id"`
	Name string `json:"name" db:"name"`
}
