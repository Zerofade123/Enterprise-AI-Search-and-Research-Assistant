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

type SessionRepository struct {
	pool *pgxpool.Pool
}

func NewSessionRepository(pool *pgxpool.Pool) *SessionRepository {
	return &SessionRepository{pool: pool}
}

func (r *SessionRepository) Create(ctx context.Context, session *domain.Session) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO auth_sessions (id, user_id, refresh_token_hash, ip_address, user_agent, expires_at, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7)
	`, session.ID, session.UserID, session.RefreshTokenHash, session.IPAddress, session.UserAgent, session.ExpiresAt, session.CreatedAt)
	if err != nil {
		return platformErrors.Wrap("sessionRepo.Create", platformErrors.CodeInternal, "failed to create session", err)
	}
	return nil
}

func (r *SessionRepository) GetByRefreshHash(ctx context.Context, hash string) (*domain.Session, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, user_id, refresh_token_hash, ip_address, user_agent, expires_at, created_at, revoked_at
		FROM auth_sessions WHERE refresh_token_hash = $1
	`, hash)

	var session domain.Session
	var revokedAt sql.NullTime
	if err := row.Scan(&session.ID, &session.UserID, &session.RefreshTokenHash, &session.IPAddress, &session.UserAgent, &session.ExpiresAt, &session.CreatedAt, &revokedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, platformErrors.Wrap("sessionRepo.GetByRefreshHash", platformErrors.CodeInternal, "failed to get session", err)
	}
	if revokedAt.Valid {
		session.RevokedAt = &revokedAt.Time
	}
	return &session, nil
}

func (r *SessionRepository) Revoke(ctx context.Context, sessionID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE auth_sessions SET revoked_at = $1 WHERE id = $2
	`, time.Now().UTC(), sessionID)
	if err != nil {
		return platformErrors.Wrap("sessionRepo.Revoke", platformErrors.CodeInternal, "failed to revoke session", err)
	}
	return nil
}

func (r *SessionRepository) RevokeAllForUser(ctx context.Context, userID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE auth_sessions SET revoked_at = $1 WHERE user_id = $2
	`, time.Now().UTC(), userID)
	if err != nil {
		return platformErrors.Wrap("sessionRepo.RevokeAll", platformErrors.CodeInternal, "failed to revoke sessions", err)
	}
	return nil
}
