package usecase

import (
	"context"
	"testing"
	"time"

	"stack/internal/domain"
	"stack/pkg/hasher"
	"stack/pkg/token"
)

type mockUserRepo struct {
	usersByPhone map[string]*domain.User
}

func newMockUserRepo() *mockUserRepo {
	return &mockUserRepo{usersByPhone: make(map[string]*domain.User)}
}

func (m *mockUserRepo) Create(ctx context.Context, user *domain.User) error {
	if _, exists := m.usersByPhone[user.PhoneNumber]; exists {
		return domain.ErrUserAlreadyExists
	}
	user.ID = "generated-uuid-123"
	m.usersByPhone[user.PhoneNumber] = user
	return nil
}

func (m *mockUserRepo) GetByPhone(ctx context.Context, phone string) (*domain.User, error) {
	u, exists := m.usersByPhone[phone]
	if !exists {
		return nil, domain.ErrUserNotFound
	}
	return u, nil
}

func (m *mockUserRepo) GetByID(ctx context.Context, id string) (*domain.User, error) {
	for _, u := range m.usersByPhone {
		if u.ID == id {
			return u, nil
		}
	}
	return nil, domain.ErrUserNotFound
}

func TestAuthUsecase_RegisterAndLogin(t *testing.T) {
	repo := newMockUserRepo()
	pwdHasher := hasher.NewBcryptHasher(4) // fast for testing
	tokenMgr := token.NewJWTManager("test_secret")
	uc := NewAuthUsecase(repo, pwdHasher, tokenMgr, 1*time.Hour)

	ctx := context.Background()

	// 1. Invalid phone number
	_, err := uc.Register(ctx, domain.RegisterInput{
		PhoneNumber: "invalid-phone",
		Username:    "user123",
		DisplayName: "User One",
		Password:    "password123",
	})
	if err == nil {
		t.Fatal("expected error for invalid phone number")
	}

	// 2. Successful Register
	regResp, err := uc.Register(ctx, domain.RegisterInput{
		PhoneNumber: "+77011234567",
		Username:    "testuser",
		DisplayName: "Test User",
		Password:    "password123",
	})
	if err != nil {
		t.Fatalf("unexpected error during register: %v", err)
	}
	if regResp.Token == "" || regResp.User.ID == "" {
		t.Fatalf("expected valid token and user id, got %+v", regResp)
	}

	// 3. Duplicate Register
	_, err = uc.Register(ctx, domain.RegisterInput{
		PhoneNumber: "+77011234567",
		Username:    "testuser",
		DisplayName: "Test User",
		Password:    "password123",
	})
	if err != domain.ErrUserAlreadyExists {
		t.Fatalf("expected ErrUserAlreadyExists, got %v", err)
	}

	// 4. Successful Login
	loginResp, err := uc.Login(ctx, domain.LoginInput{
		PhoneNumber: "+77011234567",
		Password:    "password123",
	})
	if err != nil {
		t.Fatalf("unexpected error during login: %v", err)
	}
	if loginResp.Token == "" {
		t.Fatal("expected non-empty token")
	}

	// 5. Wrong Password Login
	_, err = uc.Login(ctx, domain.LoginInput{
		PhoneNumber: "+77011234567",
		Password:    "wrong_password",
	})
	if err != domain.ErrInvalidCredentials {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}
}
