package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/karzhen/restaurant-lk/internal/domain"
)

type CartRepository struct {
	db *pgxpool.Pool
}

func NewCartRepository(db *pgxpool.Pool) *CartRepository {
	return &CartRepository{db: db}
}

func (r *CartRepository) GetCartByUserID(ctx context.Context, userID uuid.UUID) (*domain.Cart, error) {
	query := `
		SELECT id, user_id, created_at, updated_at
		FROM carts
		WHERE user_id = $1
	`

	var cart domain.Cart
	err := r.db.QueryRow(ctx, query, userID).Scan(
		&cart.ID,
		&cart.UserID,
		&cart.CreatedAt,
		&cart.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &cart, nil
}

func (r *CartRepository) GetOrCreateCartByUserID(ctx context.Context, userID uuid.UUID) (*domain.Cart, error) {
	query := `
		INSERT INTO carts (id, user_id, created_at, updated_at)
		VALUES ($1, $2, now(), now())
		ON CONFLICT (user_id) DO UPDATE SET updated_at = carts.updated_at
		RETURNING id, user_id, created_at, updated_at
	`

	var cart domain.Cart
	err := r.db.QueryRow(ctx, query, uuid.New(), userID).Scan(
		&cart.ID,
		&cart.UserID,
		&cart.CreatedAt,
		&cart.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &cart, nil
}

func (r *CartRepository) ListCartItems(ctx context.Context, cartID uuid.UUID) ([]domain.CartItem, error) {
	query := `
		SELECT id, cart_id, product_id, quantity, created_at, updated_at
		FROM cart_items
		WHERE cart_id = $1
		ORDER BY created_at ASC
	`

	rows, err := r.db.Query(ctx, query, cartID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]domain.CartItem, 0)
	for rows.Next() {
		var item domain.CartItem
		if err := rows.Scan(&item.ID, &item.CartID, &item.ProductID, &item.Quantity, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func (r *CartRepository) AddOrIncrementItem(ctx context.Context, userID uuid.UUID, productID uuid.UUID, quantity int, maxStock int) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	cartID, err := getOrCreateCartIDTx(ctx, tx, userID)
	if err != nil {
		return err
	}

	currentQty, err := lockCartItemQuantityTx(ctx, tx, cartID, productID)
	if err != nil {
		return err
	}

	newQty := currentQty + quantity
	if newQty > maxStock {
		return ErrInsufficientStock
	}

	now := time.Now().UTC()
	if currentQty == 0 {
		query := `
			INSERT INTO cart_items (id, cart_id, product_id, quantity, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6)
		`
		if _, err := tx.Exec(ctx, query, uuid.New(), cartID, productID, newQty, now, now); err != nil {
			return err
		}
	} else {
		query := `
			UPDATE cart_items
			SET quantity = $1, updated_at = $2
			WHERE cart_id = $3 AND product_id = $4
		`
		if _, err := tx.Exec(ctx, query, newQty, now, cartID, productID); err != nil {
			return err
		}
	}

	if _, err := tx.Exec(ctx, `UPDATE carts SET updated_at = $1 WHERE id = $2`, now, cartID); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	return nil
}

func (r *CartRepository) SetItemQuantity(ctx context.Context, userID uuid.UUID, productID uuid.UUID, quantity int, maxStock int) error {
	if quantity > maxStock {
		return ErrInsufficientStock
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	cartID, err := findCartIDByUserTx(ctx, tx, userID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return ErrNotFound
		}
		return err
	}

	query := `
		UPDATE cart_items
		SET quantity = $1, updated_at = $2
		WHERE cart_id = $3 AND product_id = $4
	`
	now := time.Now().UTC()
	result, err := tx.Exec(ctx, query, quantity, now, cartID, productID)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrNotFound
	}

	if _, err := tx.Exec(ctx, `UPDATE carts SET updated_at = $1 WHERE id = $2`, now, cartID); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	return nil
}

func (r *CartRepository) RemoveItem(ctx context.Context, userID uuid.UUID, productID uuid.UUID) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	cartID, err := findCartIDByUserTx(ctx, tx, userID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return ErrNotFound
		}
		return err
	}

	query := `
		DELETE FROM cart_items
		WHERE cart_id = $1 AND product_id = $2
	`
	result, err := tx.Exec(ctx, query, cartID, productID)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrNotFound
	}

	now := time.Now().UTC()
	if _, err := tx.Exec(ctx, `UPDATE carts SET updated_at = $1 WHERE id = $2`, now, cartID); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	return nil
}

func (r *CartRepository) ClearCart(ctx context.Context, userID uuid.UUID) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	cartID, err := findCartIDByUserTx(ctx, tx, userID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil
		}
		return err
	}

	if _, err := tx.Exec(ctx, `DELETE FROM cart_items WHERE cart_id = $1`, cartID); err != nil {
		return err
	}

	now := time.Now().UTC()
	if _, err := tx.Exec(ctx, `UPDATE carts SET updated_at = $1 WHERE id = $2`, now, cartID); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	return nil
}

func getOrCreateCartIDTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID) (uuid.UUID, error) {
	query := `
		INSERT INTO carts (id, user_id, created_at, updated_at)
		VALUES ($1, $2, now(), now())
		ON CONFLICT (user_id) DO UPDATE SET updated_at = carts.updated_at
		RETURNING id
	`

	var cartID uuid.UUID
	if err := tx.QueryRow(ctx, query, uuid.New(), userID).Scan(&cartID); err != nil {
		return uuid.Nil, err
	}

	return cartID, nil
}

func findCartIDByUserTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID) (uuid.UUID, error) {
	query := `SELECT id FROM carts WHERE user_id = $1`

	var cartID uuid.UUID
	if err := tx.QueryRow(ctx, query, userID).Scan(&cartID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return uuid.Nil, ErrNotFound
		}
		return uuid.Nil, err
	}

	return cartID, nil
}

func lockCartItemQuantityTx(ctx context.Context, tx pgx.Tx, cartID, productID uuid.UUID) (int, error) {
	query := `
		SELECT quantity
		FROM cart_items
		WHERE cart_id = $1 AND product_id = $2
		FOR UPDATE
	`

	var quantity int
	err := tx.QueryRow(ctx, query, cartID, productID).Scan(&quantity)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, nil
		}
		return 0, err
	}

	return quantity, nil
}
