package models

import "time"

type User struct {
	ID        string    `json:"id" db:"id"`
	Username  string    `json:"username" db:"username"`
	Password  string    `json:"-" db:"password"` // hashed password
	Role      string    `json:"role" db:"role"`  // "admin" or "member"
	Fullname  string    `json:"fullname" db:"fullname"`
	NIP       string    `json:"nip" db:"nip"`         // Nomor Induk (NIM/NIP)
	Contact   string    `json:"contact" db:"contact"` // HP/Email
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token    string `json:"token"`
	Username string `json:"username"`
	Role     string `json:"role"`
}

type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Fullname string `json:"fullname"`
	NIP      string `json:"nip"`
	Contact  string `json:"contact"`
	Role     string `json:"role"` // Optional, default "member"
}
