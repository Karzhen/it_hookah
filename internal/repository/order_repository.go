package repository

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/karzhen/restaurant-lk/internal/domain"
)

type OrderRepository interface {
	CreateOrder(ctx context.Context, order *domain.Order) error
	CreateOrderItems(ctx context.Context, items []domain.OrderItem) error
	CheckoutFromCart(ctx context.Context, userID uuid.UUID) (*domain.Order, error)
	GetOrderByID(ctx context.Context, orderID uuid.UUID) (*domain.Order, error)
	ListOrdersByUserID(ctx context.Context, userID uuid.UUID) ([]domain.Order, error)
	ListAllOrders(ctx context.Context, limit, offset int) ([]domain.Order, error)
	ListOrderItemsByOrderID(ctx context.Context, orderID uuid.UUID) ([]domain.OrderItem, error)
	UpdateOrderStatus(ctx context.Context, orderID uuid.UUID, status domain.OrderStatus) error
	UpdateOrderTotalAmount(ctx context.Context, orderID uuid.UUID, totalAmount string) error
}

type PgxOrderRepository struct {
	db *pgxpool.Pool
}

func NewOrderRepository(db *pgxpool.Pool) *PgxOrderRepository {
	return &PgxOrderRepository{db: db}
}

func (r *PgxOrderRepository) CreateOrder(ctx context.Context, order *domain.Order) error {
	query := `
		INSERT INTO orders (id, user_id, status, total_amount, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := r.db.Exec(
		ctx,
		query,
		order.ID,
		order.UserID,
		order.Status,
		order.TotalAmount,
		order.CreatedAt,
		order.UpdatedAt,
	)
	return err
}

func (r *PgxOrderRepository) CreateOrderItems(ctx context.Context, items []domain.OrderItem) error {
	if len(items) == 0 {
		return nil
	}

	query := `
		INSERT INTO order_items (
			id, order_id, product_id, product_name_snapshot,
			unit_price_snapshot, quantity, subtotal, created_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	for _, item := range items {
		if _, err := r.db.Exec(
			ctx,
			query,
			item.ID,
			item.OrderID,
			item.ProductID,
			item.ProductNameSnap,
			item.UnitPriceSnap,
			item.Quantity,
			item.Subtotal,
			item.CreatedAt,
		); err != nil {
			return err
		}
	}

	return nil
}

func (r *PgxOrderRepository) CheckoutFromCart(ctx context.Context, userID uuid.UUID) (*domain.Order, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	cartID, err := findCartIDForCheckoutTx(ctx, tx, userID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, ErrEmptyCart
		}
		return nil, err
	}

	cartItems, err := lockCheckoutItemsTx(ctx, tx, cartID)
	if err != nil {
		return nil, err
	}
	if len(cartItems) == 0 {
		return nil, ErrEmptyCart
	}

	now := time.Now().UTC()
	orderID := uuid.New()
	orderItems := make([]domain.OrderItem, 0, len(cartItems))
	totalCents := int64(0)

	for _, item := range cartItems {
		if !item.ProductIsActive {
			return nil, ErrForbidden
		}
		if item.Quantity > item.StockQuantity {
			return nil, ErrInsufficientStock
		}

		unitPriceCents, err := moneyToCents(item.UnitPrice)
		if err != nil {
			return nil, err
		}
		subtotalCents := unitPriceCents * int64(item.Quantity)
		totalCents += subtotalCents

		orderItems = append(orderItems, domain.OrderItem{
			ID:              uuid.New(),
			OrderID:         orderID,
			ProductID:       item.ProductID,
			ProductNameSnap: item.ProductName,
			UnitPriceSnap:   centsToMoney(unitPriceCents),
			Quantity:        item.Quantity,
			Subtotal:        centsToMoney(subtotalCents),
			CreatedAt:       now,
		})
	}

	order := &domain.Order{
		ID:          orderID,
		UserID:      userID,
		Status:      domain.OrderStatusPending,
		TotalAmount: centsToMoney(totalCents),
		CreatedAt:   now,
		UpdatedAt:   now,
		Items:       orderItems,
	}

	if err := createOrderTx(ctx, tx, order); err != nil {
		return nil, err
	}
	if err := createOrderItemsTx(ctx, tx, orderItems); err != nil {
		return nil, err
	}
	if err := decrementProductStockTx(ctx, tx, cartItems, userID, now); err != nil {
		return nil, err
	}
	if err := clearCartItemsTx(ctx, tx, cartID, now); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return order, nil
}

func (r *PgxOrderRepository) GetOrderByID(ctx context.Context, orderID uuid.UUID) (*domain.Order, error) {
	query := `
		SELECT id, user_id, status, total_amount::text, created_at, updated_at
		FROM orders
		WHERE id = $1
	`

	order, err := scanOrder(r.db.QueryRow(ctx, query, orderID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return order, nil
}

func (r *PgxOrderRepository) ListOrdersByUserID(ctx context.Context, userID uuid.UUID) ([]domain.Order, error) {
	query := `
		SELECT id, user_id, status, total_amount::text, created_at, updated_at
		FROM orders
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	orders := make([]domain.Order, 0)
	for rows.Next() {
		order, err := scanOrder(rows)
		if err != nil {
			return nil, err
		}
		orders = append(orders, *order)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return orders, nil
}

func (r *PgxOrderRepository) ListAllOrders(ctx context.Context, limit, offset int) ([]domain.Order, error) {
	query := `
		SELECT id, user_id, status, total_amount::text, created_at, updated_at
		FROM orders
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	orders := make([]domain.Order, 0)
	for rows.Next() {
		order, err := scanOrder(rows)
		if err != nil {
			return nil, err
		}
		orders = append(orders, *order)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return orders, nil
}

func (r *PgxOrderRepository) ListOrderItemsByOrderID(ctx context.Context, orderID uuid.UUID) ([]domain.OrderItem, error) {
	query := `
		SELECT
			id, order_id, product_id, product_name_snapshot,
			unit_price_snapshot::text, quantity, subtotal::text, created_at
		FROM order_items
		WHERE order_id = $1
		ORDER BY created_at ASC
	`

	rows, err := r.db.Query(ctx, query, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]domain.OrderItem, 0)
	for rows.Next() {
		item, err := scanOrderItem(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, *item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func (r *PgxOrderRepository) UpdateOrderStatus(ctx context.Context, orderID uuid.UUID, status domain.OrderStatus) error {
	query := `
		UPDATE orders
		SET status = $1, updated_at = now()
		WHERE id = $2
	`

	result, err := r.db.Exec(ctx, query, status, orderID)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

func (r *PgxOrderRepository) UpdateOrderTotalAmount(ctx context.Context, orderID uuid.UUID, totalAmount string) error {
	query := `
		UPDATE orders
		SET total_amount = $1, updated_at = now()
		WHERE id = $2
	`

	result, err := r.db.Exec(ctx, query, totalAmount, orderID)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

type checkoutCartItem struct {
	ProductID       uuid.UUID
	Quantity        int
	ProductName     string
	UnitPrice       string
	StockQuantity   int
	ProductIsActive bool
}

func findCartIDForCheckoutTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID) (uuid.UUID, error) {
	query := `
		SELECT id
		FROM carts
		WHERE user_id = $1
		FOR UPDATE
	`

	var cartID uuid.UUID
	if err := tx.QueryRow(ctx, query, userID).Scan(&cartID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return uuid.Nil, ErrNotFound
		}
		return uuid.Nil, err
	}
	return cartID, nil
}

func lockCheckoutItemsTx(ctx context.Context, tx pgx.Tx, cartID uuid.UUID) ([]checkoutCartItem, error) {
	query := `
		SELECT
			ci.product_id,
			ci.quantity,
			p.name,
			p.price::text,
			p.stock_quantity,
			p.is_active
		FROM cart_items ci
		JOIN products p ON p.id = ci.product_id
		WHERE ci.cart_id = $1
		ORDER BY ci.created_at ASC
		FOR UPDATE OF ci, p
	`

	rows, err := tx.Query(ctx, query, cartID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]checkoutCartItem, 0)
	for rows.Next() {
		var item checkoutCartItem
		if err := rows.Scan(
			&item.ProductID,
			&item.Quantity,
			&item.ProductName,
			&item.UnitPrice,
			&item.StockQuantity,
			&item.ProductIsActive,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func createOrderTx(ctx context.Context, tx pgx.Tx, order *domain.Order) error {
	query := `
		INSERT INTO orders (id, user_id, status, total_amount, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := tx.Exec(
		ctx,
		query,
		order.ID,
		order.UserID,
		order.Status,
		order.TotalAmount,
		order.CreatedAt,
		order.UpdatedAt,
	)
	return err
}

func createOrderItemsTx(ctx context.Context, tx pgx.Tx, items []domain.OrderItem) error {
	if len(items) == 0 {
		return nil
	}

	query := `
		INSERT INTO order_items (
			id, order_id, product_id, product_name_snapshot,
			unit_price_snapshot, quantity, subtotal, created_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	for _, item := range items {
		if _, err := tx.Exec(
			ctx,
			query,
			item.ID,
			item.OrderID,
			item.ProductID,
			item.ProductNameSnap,
			item.UnitPriceSnap,
			item.Quantity,
			item.Subtotal,
			item.CreatedAt,
		); err != nil {
			return err
		}
	}

	return nil
}

func decrementProductStockTx(ctx context.Context, tx pgx.Tx, items []checkoutCartItem, userID uuid.UUID, now time.Time) error {
	query := `
		UPDATE products
		SET stock_quantity = stock_quantity - $1, updated_at = $2
		WHERE id = $3 AND is_active = true AND stock_quantity >= $1
	`

	reason := "order checkout"
	createdByUserID := userID

	for _, item := range items {
		result, err := tx.Exec(ctx, query, item.Quantity, now, item.ProductID)
		if err != nil {
			return err
		}
		if result.RowsAffected() == 0 {
			if !item.ProductIsActive {
				return ErrForbidden
			}
			return ErrInsufficientStock
		}

		movement := &domain.StockMovement{
			ID:              uuid.New(),
			ProductID:       item.ProductID,
			Operation:       domain.StockMovementOperationCheckoutDecrement,
			Quantity:        item.Quantity,
			BeforeQuantity:  item.StockQuantity,
			AfterQuantity:   item.StockQuantity - item.Quantity,
			Reason:          &reason,
			CreatedByUserID: &createdByUserID,
			CreatedAt:       now,
		}
		if err := createStockMovementTx(ctx, tx, movement); err != nil {
			return err
		}
	}

	return nil
}

func clearCartItemsTx(ctx context.Context, tx pgx.Tx, cartID uuid.UUID, now time.Time) error {
	if _, err := tx.Exec(ctx, `DELETE FROM cart_items WHERE cart_id = $1`, cartID); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `UPDATE carts SET updated_at = $1 WHERE id = $2`, now, cartID); err != nil {
		return err
	}
	return nil
}

func scanOrder(row pgx.Row) (*domain.Order, error) {
	var (
		order  domain.Order
		status string
	)

	if err := row.Scan(
		&order.ID,
		&order.UserID,
		&status,
		&order.TotalAmount,
		&order.CreatedAt,
		&order.UpdatedAt,
	); err != nil {
		return nil, err
	}

	order.Status = domain.OrderStatus(status)
	return &order, nil
}

func scanOrderItem(row pgx.Row) (*domain.OrderItem, error) {
	var item domain.OrderItem
	if err := row.Scan(
		&item.ID,
		&item.OrderID,
		&item.ProductID,
		&item.ProductNameSnap,
		&item.UnitPriceSnap,
		&item.Quantity,
		&item.Subtotal,
		&item.CreatedAt,
	); err != nil {
		return nil, err
	}

	return &item, nil
}

func moneyToCents(value string) (int64, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return 0, fmt.Errorf("money value is empty")
	}

	sign := int64(1)
	if strings.HasPrefix(trimmed, "-") {
		sign = -1
		trimmed = strings.TrimPrefix(trimmed, "-")
	}

	parts := strings.Split(trimmed, ".")
	if len(parts) > 2 {
		return 0, fmt.Errorf("invalid money value: %s", value)
	}

	wholePart := parts[0]
	if wholePart == "" {
		wholePart = "0"
	}

	fracPart := "00"
	if len(parts) == 2 {
		fracPart = parts[1]
	}
	if len(fracPart) == 0 {
		fracPart = "00"
	}
	if len(fracPart) == 1 {
		fracPart += "0"
	}
	if len(fracPart) > 2 {
		return 0, fmt.Errorf("invalid money precision: %s", value)
	}

	whole, err := strconv.ParseInt(wholePart, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid money value: %s", value)
	}
	frac, err := strconv.ParseInt(fracPart, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid money value: %s", value)
	}

	return sign * (whole*100 + frac), nil
}

func centsToMoney(cents int64) string {
	sign := ""
	if cents < 0 {
		sign = "-"
		cents = -cents
	}

	whole := cents / 100
	frac := cents % 100
	return fmt.Sprintf("%s%d.%02d", sign, whole, frac)
}
