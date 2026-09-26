package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"stack/internal/domain"
)

type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
	query := `
		INSERT INTO users (phone_number, username, display_name, password_hash)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at;
	`

	err := r.pool.QueryRow(ctx, query,
		user.PhoneNumber,
		user.Username,
		user.DisplayName,
		user.PasswordHash,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			if strings.Contains(pgErr.Detail, "phone_number") {
				return domain.ErrUserAlreadyExists
			}
			if strings.Contains(pgErr.Detail, "username") {
				return domain.ErrUsernameAlreadyExists
			}
		}
		return fmt.Errorf("ошибка создания пользователя: %w", err)
	}

	return nil
}

func (r *UserRepository) GetByPhone(ctx context.Context, phone string) (*domain.User, error) {
	query := `
		SELECT id, phone_number, username, display_name, bio, password_hash, created_at, updated_at
		FROM users
		WHERE phone_number = $1;
	`

	var user domain.User
	err := r.pool.QueryRow(ctx, query, phone).Scan(
		&user.ID,
		&user.PhoneNumber,
		&user.Username,
		&user.DisplayName,
		&user.Bio,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("ошибка получения пользователя по телефону: %w", err)
	}

	return &user, nil
}

func (r *UserRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	query := `
		SELECT id, phone_number, username, display_name, bio, password_hash, created_at, updated_at
		FROM users
		WHERE id = $1;
	`

	var user domain.User
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.PhoneNumber,
		&user.Username,
		&user.DisplayName,
		&user.Bio,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("ошибка получения пользователя по ID: %w", err)
	}

	return &user, nil
}
