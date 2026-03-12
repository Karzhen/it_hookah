package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/karzhen/restaurant-lk/internal/domain"
)

type RefreshTokenRepository struct {
	db *pgxpool.Pool
}

func NewRefreshTokenRepository(db *pgxpool.Pool) *RefreshTokenRepository {
	return &RefreshTokenRepository{db: db}
}

func (r *RefreshTokenRepository) Create(ctx context.Context, token *domain.RefreshToken) error {
	query := `
		INSERT INTO refresh_tokens (
			id, user_id, token_hash, expires_at, created_at, revoked_at, user_agent, ip
		)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
	`

	_, err := r.db.Exec(
		ctx,
		query,
		token.ID,
		token.UserID,
		token.TokenHash,
		token.ExpiresAt,
		token.CreatedAt,
		token.RevokedAt,
		token.UserAgent,
		token.IP,
	)
	return err
}

func (r *RefreshTokenRepository) GetByTokenHash(ctx context.Context, tokenHash string) (*domain.RefreshToken, error) {
	query := `
		SELECT id, user_id, token_hash, expires_at, created_at, revoked_at, user_agent, ip
		FROM refresh_tokens
		WHERE token_hash = $1
	`

	token, err := scanRefreshToken(r.db.QueryRow(ctx, query, tokenHash))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return token, nil
}

func (r *RefreshTokenRepository) RevokeByID(ctx context.Context, tokenID uuid.UUID) error {
	query := `
		UPDATE refresh_tokens
		SET revoked_at = now()
		WHERE id = $1 AND revoked_at IS NULL
	`

	result, err := r.db.Exec(ctx, query, tokenID)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

func (r *RefreshTokenRepository) RevokeByTokenHash(ctx context.Context, tokenHash string) error {
	query := `
		UPDATE refresh_tokens
		SET revoked_at = now()
		WHERE token_hash = $1 AND revoked_at IS NULL
	`

	result, err := r.db.Exec(ctx, query, tokenHash)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

func scanRefreshToken(row pgx.Row) (*domain.RefreshToken, error) {
	var (
		token              domain.RefreshToken
		revokedAt          sql.NullTime
		userAgent, ipValue sql.NullString
	)

	err := row.Scan(
		&token.ID,
		&token.UserID,
		&token.TokenHash,
		&token.ExpiresAt,
		&token.CreatedAt,
		&revokedAt,
		&userAgent,
		&ipValue,
	)
	if err != nil {
		return nil, err
	}

	if revokedAt.Valid {
		value := revokedAt.Time
		token.RevokedAt = &value
	}
	if userAgent.Valid {
		value := userAgent.String
		token.UserAgent = &value
	}
	if ipValue.Valid {
		value := ipValue.String
		token.IP = &value
	}

	return &token, nil
}
