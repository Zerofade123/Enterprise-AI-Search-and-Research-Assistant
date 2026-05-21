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

type APIKeyRepository struct {
	pool *pgxpool.Pool
}

func NewAPIKeyRepository(pool *pgxpool.Pool) *APIKeyRepository {
	return &APIKeyRepository{pool: pool}
}

func (r *APIKeyRepository) Create(ctx context.Context, key *domain.APIKey) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO api_keys (id, user_id, name, key_hash, created_at, last_used_at, expires_at, revoked_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
	`, key.ID, key.UserID, key.Name, key.KeyHash, key.CreatedAt, key.LastUsedAt, key.ExpiresAt, key.RevokedAt)
	if err != nil {
		return platformErrors.Wrap("apiKeyRepo.Create", platformErrors.CodeInternal, "failed to create api key", err)
	}
	return nil
}

func (r *APIKeyRepository) GetByHash(ctx context.Context, hash string) (*domain.APIKey, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, user_id, name, key_hash, created_at, last_used_at, expires_at, revoked_at
		FROM api_keys WHERE key_hash = $1
	`, hash)

	var key domain.APIKey
	var lastUsed, expires, revoked sql.NullTime
	if err := row.Scan(&key.ID, &key.UserID, &key.Name, &key.KeyHash, &key.CreatedAt, &lastUsed, &expires, &revoked); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, platformErrors.Wrap("apiKeyRepo.GetByHash", platformErrors.CodeInternal, "failed to get api key", err)
	}
	if lastUsed.Valid {
		key.LastUsedAt = &lastUsed.Time
	}
	if expires.Valid {
		key.ExpiresAt = &expires.Time
	}
	if revoked.Valid {
		key.RevokedAt = &revoked.Time
	}
	return &key, nil
}

func (r *APIKeyRepository) UpdateLastUsed(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE api_keys SET last_used_at = $1 WHERE id = $2
	`, time.Now().UTC(), id)
	if err != nil {
		return platformErrors.Wrap("apiKeyRepo.UpdateLastUsed", platformErrors.CodeInternal, "failed to update last used", err)
	}
	return nil
}

func (r *APIKeyRepository) Revoke(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE api_keys SET revoked_at = $1 WHERE id = $2
	`, time.Now().UTC(), id)
	if err != nil {
		return platformErrors.Wrap("apiKeyRepo.Revoke", platformErrors.CodeInternal, "failed to revoke api key", err)
	}
	return nil
}
