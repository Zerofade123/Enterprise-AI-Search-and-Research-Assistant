package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/Zerofade123/Enterprise-AI-Search-and-Research-Assistant/backend/internal/auth/domain"
	platformErrors "github.com/Zerofade123/Enterprise-AI-Search-and-Research-Assistant/backend/internal/platform/errors"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO users (id, email, password_hash, first_name, last_name, mfa_enabled, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
	`, user.ID, user.Email, user.PasswordHash, user.FirstName, user.LastName, user.MFAEnabled, user.CreatedAt, user.UpdatedAt)
	if err != nil {
		return platformErrors.Wrap("userRepo.Create", platformErrors.CodeInternal, "failed to create user", err)
	}
	return nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, email, password_hash, first_name, last_name, mfa_enabled, created_at, updated_at, last_login_at
		FROM users WHERE email = $1 AND deleted_at IS NULL
	`, email)

	var user domain.User
	var lastLogin sql.NullTime
	if err := row.Scan(&user.ID, &user.Email, &user.PasswordHash, &user.FirstName, &user.LastName, &user.MFAEnabled, &user.CreatedAt, &user.UpdatedAt, &lastLogin); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, platformErrors.Wrap("userRepo.GetByEmail", platformErrors.CodeInternal, "failed to get user", err)
	}
	if lastLogin.Valid {
		user.LastLoginAt = &lastLogin.Time
	}
	return &user, nil
}

func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, email, password_hash, first_name, last_name, mfa_enabled, created_at, updated_at, last_login_at
		FROM users WHERE id = $1 AND deleted_at IS NULL
	`, id)

	var user domain.User
	var lastLogin sql.NullTime
	if err := row.Scan(&user.ID, &user.Email, &user.PasswordHash, &user.FirstName, &user.LastName, &user.MFAEnabled, &user.CreatedAt, &user.UpdatedAt, &lastLogin); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, platformErrors.Wrap("userRepo.GetByID", platformErrors.CodeInternal, "failed to get user", err)
	}
	if lastLogin.Valid {
		user.LastLoginAt = &lastLogin.Time
	}
	return &user, nil
}

func (r *UserRepository) UpdateLastLogin(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE users SET last_login_at = $1, updated_at = $2 WHERE id = $3
	`, time.Now().UTC(), time.Now().UTC(), id)
	if err != nil {
		return platformErrors.Wrap("userRepo.UpdateLastLogin", platformErrors.CodeInternal, "failed to update last login", err)
	}
	return nil
}
