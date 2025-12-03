package store

import (
	"database/sql"
	"errors"
	"latihan_cloud8/models"

	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
)

var (
	ErrUserExists   = errors.New("user already exists")
	ErrUserNotFound = errors.New("user not found")
)

type MySQLStore struct {
	db *sql.DB
}

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

func (s *MySQLStore) Close() error {
	return s.db.Close()
}

func (s *MySQLStore) CreateUser(username, hashedPassword, role string) (*models.User, error) {
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
		"INSERT INTO users (id, username, password, role) VALUES (?, ?, ?, ?)",
		id, username, hashedPassword, role,
	)
	if err != nil {
		return nil, err
	}

	return &models.User{
		ID:       id,
		Username: username,
		Password: hashedPassword,
		Role:     role,
	}, nil
}

func (s *MySQLStore) GetByUsername(username string) (*models.User, error) {
	user := &models.User{}
	err := s.db.QueryRow(
		"SELECT id, username, password, role, created_at FROM users WHERE username = ?",
		username,
	).Scan(&user.ID, &user.Username, &user.Password, &user.Role, &user.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *MySQLStore) GetAllUsers() ([]models.User, error) {
	rows, err := s.db.Query("SELECT id, username, role, created_at FROM users")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		if err := rows.Scan(&user.ID, &user.Username, &user.Role, &user.CreatedAt); err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, nil
}