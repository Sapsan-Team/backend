package usecase

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"stack/internal/domain"
	"stack/pkg/hasher"
	"stack/pkg/token"
)

var (
	phoneRegex    = regexp.MustCompile(`^\+[1-9]\d{7,14}$`)    // Формат E.164 (+77011234567)
	usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_]{3,30}$`) // 3-30 символов, a-z, 0-9, _
)

type AuthUsecase struct {
	userRepo     domain.UserRepository
	hasher       hasher.PasswordHasher
	tokenManager token.TokenManager
	tokenTTL     time.Duration
}

func NewAuthUsecase(
	userRepo domain.UserRepository,
	hasher hasher.PasswordHasher,
	tokenManager token.TokenManager,
	tokenTTL time.Duration,
) *AuthUsecase {
	return &AuthUsecase{
		userRepo:     userRepo,
		hasher:       hasher,
		tokenManager: tokenManager,
		tokenTTL:     tokenTTL,
	}
}

func (u *AuthUsecase) Register(ctx context.Context, input domain.RegisterInput) (*domain.AuthResponse, error) {
	input.PhoneNumber = strings.TrimSpace(input.PhoneNumber)
	input.Username = strings.TrimSpace(input.Username)
	input.DisplayName = strings.TrimSpace(input.DisplayName)

	if !phoneRegex.MatchString(input.PhoneNumber) {
		return nil, fmt.Errorf("%w: некорректный номер телефона (формат E.164, например: +77011234567)", domain.ErrInvalidInput)
	}
	if !usernameRegex.MatchString(input.Username) {
		return nil, fmt.Errorf("%w: имя пользователя должно быть от 3 до 30 символов (латиница, цифры и _)", domain.ErrInvalidInput)
	}
	if len(input.Password) < 8 {
		return nil, fmt.Errorf("%w: пароль должен содержать не менее 8 символов", domain.ErrInvalidInput)
	}

	hashedPassword, err := u.hasher.Hash(input.Password)
	if err != nil {
		return nil, fmt.Errorf("ошибка хеширования пароля: %w", err)
	}

	user := &domain.User{
		PhoneNumber:  input.PhoneNumber,
		Username:     input.Username,
		DisplayName:  input.DisplayName,
		PasswordHash: hashedPassword,
	}

	if err := u.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	tokenStr, err := u.tokenManager.Generate(user.ID, u.tokenTTL)
	if err != nil {
		return nil, fmt.Errorf("ошибка генерации токена: %w", err)
	}

	return &domain.AuthResponse{
		Token: tokenStr,
		User: domain.UserResponse{
			ID:          user.ID,
			PhoneNumber: user.PhoneNumber,
			Username:    user.Username,
			DisplayName: user.DisplayName,
		},
	}, nil
}

func (u *AuthUsecase) Login(ctx context.Context, input domain.LoginInput) (*domain.AuthResponse, error) {
	input.PhoneNumber = strings.TrimSpace(input.PhoneNumber)
	if input.PhoneNumber == "" || input.Password == "" {
		return nil, fmt.Errorf("%w: номер телефона и пароль обязательны", domain.ErrInvalidInput)
	}

	user, err := u.userRepo.GetByPhone(ctx, input.PhoneNumber)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return nil, domain.ErrInvalidCredentials
		}
		return nil, err
	}

	if err := u.hasher.Compare(user.PasswordHash, input.Password); err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	tokenStr, err := u.tokenManager.Generate(user.ID, u.tokenTTL)
	if err != nil {
		return nil, fmt.Errorf("ошибка генерации токена: %w", err)
	}

	return &domain.AuthResponse{
		Token: tokenStr,
		User: domain.UserResponse{
			ID:          user.ID,
			PhoneNumber: user.PhoneNumber,
			Username:    user.Username,
			DisplayName: user.DisplayName,
		},
	}, nil
}
