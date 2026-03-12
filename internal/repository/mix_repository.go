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

type MixRepository interface {
	CreateMix(ctx context.Context, mix *domain.Mix) error
	CreateMixItems(ctx context.Context, items []domain.MixItem) error
	GetMixByID(ctx context.Context, mixID uuid.UUID, adminView bool) (*domain.Mix, error)
	ListMixes(ctx context.Context, activeOnly bool, limit, offset int) ([]domain.Mix, error)
	ListMixItemsByMixID(ctx context.Context, mixID uuid.UUID) ([]domain.MixItem, error)
	UpdateMix(ctx context.Context, mix *domain.Mix) error
	ReplaceMixItems(ctx context.Context, mixID uuid.UUID, items []domain.MixItem) error
	DeactivateMix(ctx context.Context, mixID uuid.UUID) error
}

type PgxMixRepository struct {
	db *pgxpool.Pool
}

func NewMixRepository(db *pgxpool.Pool) *PgxMixRepository {
	return &PgxMixRepository{db: db}
}

func (r *PgxMixRepository) CreateMix(ctx context.Context, mix *domain.Mix) error {
	if mix == nil {
		return ErrInvalidMixItemProduct
	}
	if len(mix.Items) == 0 {
		return ErrInvalidMixPercentTotal
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	insertMixQuery := `
		INSERT INTO mixes (id, name, description, final_strength_label, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	if _, err := tx.Exec(
		ctx,
		insertMixQuery,
		mix.ID,
		mix.Name,
		mix.Description,
		mix.FinalStrengthLabel,
		mix.IsActive,
		mix.CreatedAt,
		mix.UpdatedAt,
	); err != nil {
		return mapMixDBError(err)
	}

	if err := insertMixItemsTx(ctx, tx, mix.Items); err != nil {
		return mapMixDBError(err)
	}

	if err := tx.Commit(ctx); err != nil {
		return mapMixDBError(err)
	}
	return nil
}

func (r *PgxMixRepository) CreateMixItems(ctx context.Context, items []domain.MixItem) error {
	if len(items) == 0 {
		return ErrInvalidMixPercentTotal
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	mixID := items[0].MixID
	if err := ensureMixExistsTx(ctx, tx, mixID); err != nil {
		return err
	}

	if err := insertMixItemsTx(ctx, tx, items); err != nil {
		return mapMixDBError(err)
	}

	if _, err := tx.Exec(ctx, `UPDATE mixes SET updated_at = $1 WHERE id = $2`, time.Now().UTC(), mixID); err != nil {
		return mapMixDBError(err)
	}

	if err := tx.Commit(ctx); err != nil {
		return mapMixDBError(err)
	}

	return nil
}

func (r *PgxMixRepository) GetMixByID(ctx context.Context, mixID uuid.UUID, adminView bool) (*domain.Mix, error) {
	query := `
		SELECT id, name, description, final_strength_label, is_active, created_at, updated_at
		FROM mixes
		WHERE id = $1
	`
	if !adminView {
		query += ` AND is_active = true`
	}

	mix, err := scanMix(r.db.QueryRow(ctx, query, mixID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	tagsByMixID, err := r.loadTagsByMixIDs(ctx, []uuid.UUID{mix.ID}, !adminView)
	if err != nil {
		return nil, err
	}
	mix.Tags = tagsByMixID[mix.ID]

	return mix, nil
}

func (r *PgxMixRepository) ListMixes(ctx context.Context, activeOnly bool, limit, offset int) ([]domain.Mix, error) {
	query := `
		SELECT id, name, description, final_strength_label, is_active, created_at, updated_at
		FROM mixes
	`
	args := make([]any, 0, 2)
	if activeOnly {
		query += ` WHERE is_active = true`
	}
	query += ` ORDER BY created_at DESC LIMIT $1 OFFSET $2`
	args = append(args, limit, offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]domain.Mix, 0)
	for rows.Next() {
		item, err := scanMix(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, *item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	mixIDs := make([]uuid.UUID, 0, len(result))
	for i := range result {
		mixIDs = append(mixIDs, result[i].ID)
	}
	tagsByMixID, err := r.loadTagsByMixIDs(ctx, mixIDs, activeOnly)
	if err != nil {
		return nil, err
	}
	for i := range result {
		result[i].Tags = tagsByMixID[result[i].ID]
	}

	return result, nil
}

func (r *PgxMixRepository) ListMixItemsByMixID(ctx context.Context, mixID uuid.UUID) ([]domain.MixItem, error) {
	query := `
		SELECT id, mix_id, product_id, percent, created_at
		FROM mix_items
		WHERE mix_id = $1
		ORDER BY created_at ASC
	`

	rows, err := r.db.Query(ctx, query, mixID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]domain.MixItem, 0)
	for rows.Next() {
		var item domain.MixItem
		if err := rows.Scan(&item.ID, &item.MixID, &item.ProductID, &item.Percent, &item.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func (r *PgxMixRepository) UpdateMix(ctx context.Context, mix *domain.Mix) error {
	if mix == nil {
		return ErrInvalidMixItemProduct
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	updateQuery := `
		UPDATE mixes
		SET name = $1, description = $2, final_strength_label = $3, is_active = $4, updated_at = $5
		WHERE id = $6
	`
	result, err := tx.Exec(
		ctx,
		updateQuery,
		mix.Name,
		mix.Description,
		mix.FinalStrengthLabel,
		mix.IsActive,
		mix.UpdatedAt,
		mix.ID,
	)
	if err != nil {
		return mapMixDBError(err)
	}
	if result.RowsAffected() == 0 {
		return ErrNotFound
	}

	if mix.Items != nil {
		if len(mix.Items) == 0 {
			return ErrInvalidMixPercentTotal
		}

		if _, err := tx.Exec(ctx, `DELETE FROM mix_items WHERE mix_id = $1`, mix.ID); err != nil {
			return mapMixDBError(err)
		}
		if err := insertMixItemsTx(ctx, tx, mix.Items); err != nil {
			return mapMixDBError(err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return mapMixDBError(err)
	}

	return nil
}

func (r *PgxMixRepository) ReplaceMixItems(ctx context.Context, mixID uuid.UUID, items []domain.MixItem) error {
	if len(items) == 0 {
		return ErrInvalidMixPercentTotal
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if err := ensureMixExistsTx(ctx, tx, mixID); err != nil {
		return err
	}

	if _, err := tx.Exec(ctx, `DELETE FROM mix_items WHERE mix_id = $1`, mixID); err != nil {
		return mapMixDBError(err)
	}
	if err := insertMixItemsTx(ctx, tx, items); err != nil {
		return mapMixDBError(err)
	}

	if _, err := tx.Exec(ctx, `UPDATE mixes SET updated_at = $1 WHERE id = $2`, time.Now().UTC(), mixID); err != nil {
		return mapMixDBError(err)
	}

	if err := tx.Commit(ctx); err != nil {
		return mapMixDBError(err)
	}
	return nil
}

func (r *PgxMixRepository) DeactivateMix(ctx context.Context, mixID uuid.UUID) error {
	query := `
		UPDATE mixes
		SET is_active = false, updated_at = now()
		WHERE id = $1
	`
	result, err := r.db.Exec(ctx, query, mixID)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func insertMixItemsTx(ctx context.Context, tx pgx.Tx, items []domain.MixItem) error {
	query := `
		INSERT INTO mix_items (id, mix_id, product_id, percent, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`

	for _, item := range items {
		if _, err := tx.Exec(ctx, query, item.ID, item.MixID, item.ProductID, item.Percent, item.CreatedAt); err != nil {
			return err
		}
	}

	return nil
}

func ensureMixExistsTx(ctx context.Context, tx pgx.Tx, mixID uuid.UUID) error {
	var existingID uuid.UUID
	if err := tx.QueryRow(ctx, `SELECT id FROM mixes WHERE id = $1 FOR UPDATE`, mixID).Scan(&existingID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}
	return nil
}

func scanMix(row pgx.Row) (*domain.Mix, error) {
	var (
		mix                domain.Mix
		description        sql.NullString
		finalStrengthLabel sql.NullString
	)

	if err := row.Scan(
		&mix.ID,
		&mix.Name,
		&description,
		&finalStrengthLabel,
		&mix.IsActive,
		&mix.CreatedAt,
		&mix.UpdatedAt,
	); err != nil {
		return nil, err
	}

	if description.Valid {
		value := description.String
		mix.Description = &value
	}
	if finalStrengthLabel.Valid {
		value := finalStrengthLabel.String
		mix.FinalStrengthLabel = &value
	}

	return &mix, nil
}

func mapMixDBError(err error) error {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return err
	}

	if pgErr.Code == "23503" {
		if strings.Contains(strings.ToLower(pgErr.ConstraintName), "product") {
			return ErrTobaccoProductNotFound
		}
	}

	if pgErr.Code == "23505" {
		if strings.EqualFold(pgErr.ConstraintName, "mix_items_mix_id_product_id_key") {
			return ErrInvalidMixItemProduct
		}
	}

	if pgErr.Code == "23514" {
		message := strings.ToLower(strings.TrimSpace(pgErr.Message))
		if strings.Contains(message, "sum of mix item percent must be 100") ||
			strings.Contains(message, "mix must contain at least 1 item") {
			return ErrInvalidMixPercentTotal
		}
		if strings.Contains(message, "mix item product must belong to hookah_tobacco category") {
			return ErrInvalidMixItemProduct
		}
	}

	return err
}

func (r *PgxMixRepository) loadTagsByMixIDs(ctx context.Context, mixIDs []uuid.UUID, activeOnly bool) (map[uuid.UUID][]domain.Tag, error) {
	result := make(map[uuid.UUID][]domain.Tag, len(mixIDs))
	if len(mixIDs) == 0 {
		return result, nil
	}

	query := `
		SELECT mt.mix_id, t.id, t.code, t.name, t.description, t.is_active, t.created_at, t.updated_at
		FROM mix_tags mt
		JOIN tags t ON t.id = mt.tag_id
		WHERE mt.mix_id = ANY($1)
	`
	if activeOnly {
		query += ` AND t.is_active = true`
	}
	query += ` ORDER BY t.name ASC`

	rows, err := r.db.Query(ctx, query, mixIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			mixID       uuid.UUID
			tag         domain.Tag
			description sql.NullString
		)
		if err := rows.Scan(
			&mixID,
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
		result[mixID] = append(result[mixID], tag)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}
