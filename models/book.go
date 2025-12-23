package models

import "time"

type Book struct {
	ID            int       `json:"id" db:"id"`
	Title         string    `json:"title" db:"title"`
	Author        string    `json:"author" db:"author"`
	Category      string    `json:"category" db:"category"`
	Stock         int       `json:"stock" db:"stock"`
	ImageURL      string    `json:"image_url" db:"image_url"`
	PublishedYear int       `json:"published_year" db:"published_year"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
}

type BookRequest struct {
	Title         string `json:"title"`
	Author        string `json:"author"`
	Category      string `json:"category"`
	Stock         int    `json:"stock"`
	PublishedYear int    `json:"published_year"`
}
