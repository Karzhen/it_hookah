package repository

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/karzhen/restaurant-lk/internal/domain"
)

type TagRepository interface {
	CreateTag(ctx context.Context, tag *domain.Tag) error
	UpdateTag(ctx context.Context, tag *domain.Tag) error
	DeactivateTag(ctx context.Context, tagID uuid.UUID) error
	GetTagByID(ctx context.Context, tagID uuid.UUID, adminView bool) (*domain.Tag, error)
	GetTagByCode(ctx context.Context, code string, adminView bool) (*domain.Tag, error)
	GetTagByName(ctx context.Context, name string, adminView bool) (*domain.Tag, error)
	GetTagsByIDs(ctx context.Context, tagIDs []uuid.UUID, activeOnly bool) ([]domain.Tag, error)
	ListTags(ctx context.Context, activeOnly bool, limit, offset int) ([]domain.Tag, error)

	SetProductTags(ctx context.Context, productID uuid.UUID, tagIDs []uuid.UUID) error
	ListProductTags(ctx context.Context, productID uuid.UUID, activeOnly bool) ([]domain.Tag, error)

	SetMixTags(ctx context.Context, mixID uuid.UUID, tagIDs []uuid.UUID) error
	ListMixTags(ctx context.Context, mixID uuid.UUID, activeOnly bool) ([]domain.Tag, error)
}

type PgxTagRepository struct {
	db *pgxpool.Pool
}

func NewTagRepository(db *pgxpool.Pool) *PgxTagRepository {
	return &PgxTagRepository{db: db}
}

func (r *PgxTagRepository) CreateTag(ctx context.Context, tag *domain.Tag) error {
	query := `
		INSERT INTO tags (id, code, name, description, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := r.db.Exec(
		ctx,
		query,
		tag.ID,
		tag.Code,
		tag.Name,
		tag.Description,
		tag.IsActive,
		tag.CreatedAt,
		tag.UpdatedAt,
	)
	if err != nil {
		return mapTagError(err)
	}
	return nil
}

func (r *PgxTagRepository) UpdateTag(ctx context.Context, tag *domain.Tag) error {
	query := `
		UPDATE tags
		SET code = $1, name = $2, description = $3, is_active = $4, updated_at = $5
		WHERE id = $6
	`

	result, err := r.db.Exec(
		ctx,
		query,
		tag.Code,
		tag.Name,
		tag.Description,
		tag.IsActive,
		tag.UpdatedAt,
		tag.ID,
	)
	if err != nil {
		return mapTagError(err)
	}
	if result.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *PgxTagRepository) DeactivateTag(ctx context.Context, tagID uuid.UUID) error {
	query := `
		UPDATE tags
		SET is_active = false, updated_at = now()
		WHERE id = $1
	`

	result, err := r.db.Exec(ctx, query, tagID)
	if err != nil {
		return mapTagError(err)
	}
	if result.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *PgxTagRepository) GetTagByID(ctx context.Context, tagID uuid.UUID, adminView bool) (*domain.Tag, error) {
	query := `
		SELECT id, code, name, description, is_active, created_at, updated_at
		FROM tags
		WHERE id = $1
	`
	if !adminView {
		query += ` AND is_active = true`
	}

	tag, err := scanTag(r.db.QueryRow(ctx, query, tagID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return tag, nil
}

func (r *PgxTagRepository) GetTagByCode(ctx context.Context, code string, adminView bool) (*domain.Tag, error) {
	query := `
		SELECT id, code, name, description, is_active, created_at, updated_at
		FROM tags
		WHERE LOWER(code) = LOWER($1)
	`
	if !adminView {
		query += ` AND is_active = true`
	}

	tag, err := scanTag(r.db.QueryRow(ctx, query, strings.TrimSpace(code)))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return tag, nil
}

func (r *PgxTagRepository) GetTagByName(ctx context.Context, name string, adminView bool) (*domain.Tag, error) {
	query := `
		SELECT id, code, name, description, is_active, created_at, updated_at
		FROM tags
		WHERE LOWER(name) = LOWER($1)
	`
	if !adminView {
		query += ` AND is_active = true`
	}

	tag, err := scanTag(r.db.QueryRow(ctx, query, strings.TrimSpace(name)))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return tag, nil
}

func (r *PgxTagRepository) GetTagsByIDs(ctx context.Context, tagIDs []uuid.UUID, activeOnly bool) ([]domain.Tag, error) {
	if len(tagIDs) == 0 {
		return []domain.Tag{}, nil
	}

	query := `
		SELECT id, code, name, description, is_active, created_at, updated_at
		FROM tags
		WHERE id = ANY($1)
	`
	if activeOnly {
		query += ` AND is_active = true`
	}

	rows, err := r.db.Query(ctx, query, tagIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]domain.Tag, 0, len(tagIDs))
	for rows.Next() {
		tag, err := scanTag(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, *tag)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

func (r *PgxTagRepository) ListTags(ctx context.Context, activeOnly bool, limit, offset int) ([]domain.Tag, error) {
	query := `
		SELECT id, code, name, description, is_active, created_at, updated_at
		FROM tags
	`
	args := make([]any, 0, 2)
	if activeOnly {
		query += ` WHERE is_active = true`
	}
	query += ` ORDER BY name ASC LIMIT $1 OFFSET $2`
	args = append(args, limit, offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]domain.Tag, 0)
	for rows.Next() {
		tag, err := scanTag(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, *tag)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

func (r *PgxTagRepository) SetProductTags(ctx context.Context, productID uuid.UUID, tagIDs []uuid.UUID) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var existingProductID uuid.UUID
	if err := tx.QueryRow(ctx, `SELECT id FROM products WHERE id = $1 FOR UPDATE`, productID).Scan(&existingProductID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}

	if _, err := tx.Exec(ctx, `DELETE FROM product_tags WHERE product_id = $1`, productID); err != nil {
		return mapTagError(err)
	}

	if len(tagIDs) > 0 {
		now := time.Now().UTC()
		query := `INSERT INTO product_tags (product_id, tag_id, created_at) VALUES ($1, $2, $3)`
		for _, tagID := range dedupeUUIDs(tagIDs) {
			if _, err := tx.Exec(ctx, query, productID, tagID, now); err != nil {
				return mapTagError(err)
			}
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return mapTagError(err)
	}
	return nil
}

func (r *PgxTagRepository) ListProductTags(ctx context.Context, productID uuid.UUID, activeOnly bool) ([]domain.Tag, error) {
	query := `
		SELECT t.id, t.code, t.name, t.description, t.is_active, t.created_at, t.updated_at
		FROM product_tags pt
		JOIN tags t ON t.id = pt.tag_id
		WHERE pt.product_id = $1
	`
	if activeOnly {
		query += ` AND t.is_active = true`
	}
	query += ` ORDER BY t.name ASC`

	rows, err := r.db.Query(ctx, query, productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tags := make([]domain.Tag, 0)
	for rows.Next() {
		tag, err := scanTag(rows)
		if err != nil {
			return nil, err
		}
		tags = append(tags, *tag)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tags, nil
}

func (r *PgxTagRepository) SetMixTags(ctx context.Context, mixID uuid.UUID, tagIDs []uuid.UUID) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var existingMixID uuid.UUID
	if err := tx.QueryRow(ctx, `SELECT id FROM mixes WHERE id = $1 FOR UPDATE`, mixID).Scan(&existingMixID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}

	if _, err := tx.Exec(ctx, `DELETE FROM mix_tags WHERE mix_id = $1`, mixID); err != nil {
		return mapTagError(err)
	}

	if len(tagIDs) > 0 {
		now := time.Now().UTC()
		query := `INSERT INTO mix_tags (mix_id, tag_id, created_at) VALUES ($1, $2, $3)`
		for _, tagID := range dedupeUUIDs(tagIDs) {
			if _, err := tx.Exec(ctx, query, mixID, tagID, now); err != nil {
				return mapTagError(err)
			}
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return mapTagError(err)
	}
	return nil
}

func (r *PgxTagRepository) ListMixTags(ctx context.Context, mixID uuid.UUID, activeOnly bool) ([]domain.Tag, error) {
	query := `
		SELECT t.id, t.code, t.name, t.description, t.is_active, t.created_at, t.updated_at
		FROM mix_tags mt
		JOIN tags t ON t.id = mt.tag_id
		WHERE mt.mix_id = $1
	`
	if activeOnly {
		query += ` AND t.is_active = true`
	}
	query += ` ORDER BY t.name ASC`

	rows, err := r.db.Query(ctx, query, mixID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tags := make([]domain.Tag, 0)
	for rows.Next() {
		tag, err := scanTag(rows)
		if err != nil {
			return nil, err
		}
		tags = append(tags, *tag)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tags, nil
}

func scanTag(row pgx.Row) (*domain.Tag, error) {
	var (
		tag         domain.Tag
		description sql.NullString
	)

	if err := row.Scan(
		&tag.ID,
		&tag.Code,
		&tag.Name,
		&description,
		&tag.IsActive,
		&tag.CreatedAt,
		&tag.UpdatedAt,
	); err != nil {
		return nil, err
	}

	if description.Valid {
		value := description.String
		tag.Description = &value
	}

	return &tag, nil
}

func mapTagError(err error) error {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return err
	}

	if pgErr.Code == "23503" {
		return ErrNotFound
	}

	if pgErr.Code == "23505" {
		switch strings.ToLower(strings.TrimSpace(pgErr.ConstraintName)) {
		case "tags_code_key":
			return ErrDuplicateTagCode
		case "tags_name_key":
			return ErrDuplicateTagName
		}

		msg := strings.ToLower(strings.TrimSpace(pgErr.Message))
		if strings.Contains(msg, "tags_code_key") || strings.Contains(msg, "(code)") {
			return ErrDuplicateTagCode
		}
		if strings.Contains(msg, "tags_name_key") || strings.Contains(msg, "(name)") {
			return ErrDuplicateTagName
		}
	}

	return err
}

func dedupeUUIDs(values []uuid.UUID) []uuid.UUID {
	if len(values) == 0 {
		return []uuid.UUID{}
	}
	seen := make(map[uuid.UUID]struct{}, len(values))
	result := make([]uuid.UUID, 0, len(values))
	for _, value := range values {
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}
