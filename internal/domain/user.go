package domain

import (
	"context"
	"time"
)

// User — доменная сущность пользователя
type User struct {
	ID           string
	PhoneNumber  string
	Username     string
	DisplayName  string
	Bio          *string
	PasswordHash string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// DTO для входных и выходных данных
type RegisterInput struct {
	PhoneNumber string `json:"phone_number"`
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
	Password    string `json:"password"`
}

type LoginInput struct {
	PhoneNumber string `json:"phone_number"`
	Password    string `json:"password"`
}

type UserResponse struct {
	ID          string `json:"id"`
	PhoneNumber string `json:"phone_number"`
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
}

type AuthResponse struct {
	Token string       `json:"token"`
	User  UserResponse `json:"user"`
}

// UserRepository — контракт хранилища пользователей (Domain/Port)
type UserRepository interface {
	Create(ctx context.Context, user *User) error
	GetByPhone(ctx context.Context, phone string) (*User, error)
	GetByID(ctx context.Context, id string) (*User, error)
}

// AuthUsecase — контракт сценариев авторизации и регистрации
type AuthUsecase interface {
	Register(ctx context.Context, input RegisterInput) (*AuthResponse, error)
	Login(ctx context.Context, input LoginInput) (*AuthResponse, error)
}
