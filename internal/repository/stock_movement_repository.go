package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/karzhen/restaurant-lk/internal/domain"
)

type StockMovementRepository interface {
	CreateStockMovement(ctx context.Context, movement *domain.StockMovement) error
	ListStockMovements(ctx context.Context, filter domain.StockMovementFilter) ([]domain.StockMovement, error)
	ListStockMovementsByProductID(
		ctx context.Context,
		productID uuid.UUID,
		operation *domain.StockMovementOperation,
		limit,
		offset int,
	) ([]domain.StockMovement, error)
}

type PgxStockMovementRepository struct {
	db *pgxpool.Pool
}

func NewStockMovementRepository(db *pgxpool.Pool) *PgxStockMovementRepository {
	return &PgxStockMovementRepository{db: db}
}

func (r *PgxStockMovementRepository) CreateStockMovement(ctx context.Context, movement *domain.StockMovement) error {
	return insertStockMovement(ctx, r.db, movement)
}

func (r *PgxStockMovementRepository) ListStockMovements(ctx context.Context, filter domain.StockMovementFilter) ([]domain.StockMovement, error) {
	query := `
		SELECT
			sm.id,
			sm.product_id,
			p.name AS product_name,
			sm.operation,
			sm.quantity,
			sm.before_quantity,
			sm.after_quantity,
			sm.reason,
			sm.created_by_user_id,
			sm.created_at
		FROM stock_movements sm
		LEFT JOIN products p ON p.id = sm.product_id
	`

	where := make([]string, 0, 2)
	args := make([]any, 0, 4)
	argIndex := 1

	if filter.ProductID != nil {
		where = append(where, fmt.Sprintf("sm.product_id = $%d", argIndex))
		args = append(args, *filter.ProductID)
		argIndex++
	}
	if filter.Operation != nil {
		where = append(where, fmt.Sprintf("sm.operation = $%d", argIndex))
		args = append(args, *filter.Operation)
		argIndex++
	}
	if len(where) > 0 {
		query += " WHERE " + strings.Join(where, " AND ")
	}

	query += fmt.Sprintf(" ORDER BY sm.created_at DESC LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, filter.Limit, filter.Offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	movements := make([]domain.StockMovement, 0)
	for rows.Next() {
		movement, err := scanStockMovementWithProductName(rows)
		if err != nil {
			return nil, err
		}
		movements = append(movements, *movement)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return movements, nil
}

func (r *PgxStockMovementRepository) ListStockMovementsByProductID(
	ctx context.Context,
	productID uuid.UUID,
	operation *domain.StockMovementOperation,
	limit,
	offset int,
) ([]domain.StockMovement, error) {
	query := `
		SELECT
			sm.id,
			sm.product_id,
			p.name AS product_name,
			sm.operation,
			sm.quantity,
			sm.before_quantity,
			sm.after_quantity,
			sm.reason,
			sm.created_by_user_id,
			sm.created_at
		FROM stock_movements sm
		LEFT JOIN products p ON p.id = sm.product_id
		WHERE sm.product_id = $1
	`

	args := []any{productID}
	argIndex := 2

	if operation != nil {
		query += fmt.Sprintf(" AND sm.operation = $%d", argIndex)
		args = append(args, *operation)
		argIndex++
	}

	query += fmt.Sprintf(" ORDER BY sm.created_at DESC LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, limit, offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	movements := make([]domain.StockMovement, 0)
	for rows.Next() {
		movement, err := scanStockMovementWithProductName(rows)
		if err != nil {
			return nil, err
		}
		movements = append(movements, *movement)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return movements, nil
}

type stockMovementScanner interface {
	Scan(dest ...any) error
}

func scanStockMovementWithProductName(scanner stockMovementScanner) (*domain.StockMovement, error) {
	var (
		movement    domain.StockMovement
		productName sql.NullString
	)
	if err := scanner.Scan(
		&movement.ID,
		&movement.ProductID,
		&productName,
		&movement.Operation,
		&movement.Quantity,
		&movement.BeforeQuantity,
		&movement.AfterQuantity,
		&movement.Reason,
		&movement.CreatedByUserID,
		&movement.CreatedAt,
	); err != nil {
		return nil, err
	}
	if productName.Valid {
		value := productName.String
		movement.ProductName = &value
	}

	return &movement, nil
}

type stockMovementExecutor interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
}

func createStockMovementTx(ctx context.Context, tx pgx.Tx, movement *domain.StockMovement) error {
	return insertStockMovement(ctx, tx, movement)
}

func insertStockMovement(ctx context.Context, exec stockMovementExecutor, movement *domain.StockMovement) error {
	if movement.CreatedAt.IsZero() {
		movement.CreatedAt = time.Now().UTC()
	}

	query := `
		INSERT INTO stock_movements (
			id,
			product_id,
			operation,
			quantity,
			before_quantity,
			after_quantity,
			reason,
			created_by_user_id,
			created_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err := exec.Exec(
		ctx,
		query,
		movement.ID,
		movement.ProductID,
		movement.Operation,
		movement.Quantity,
		movement.BeforeQuantity,
		movement.AfterQuantity,
		movement.Reason,
		movement.CreatedByUserID,
		movement.CreatedAt,
	)
	return err
}
