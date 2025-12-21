package entity

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type Admin struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Password  string    `json:"password"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func NewAdmin(username, name, email, password string) (*Admin, error) {
	if username == "" {
		return nil, errors.New("username is required")
	}
	if email == "" {
		return nil, errors.New("email is required")
	}
	if password == "" {
		return nil, errors.New("password is required")
	}

	return &Admin{
		ID:        uuid.New().String(),
		Name:      name,
		Username:  username,
		Email:     email,
		Password:  password,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil
}
