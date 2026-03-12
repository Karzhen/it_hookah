package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/karzhen/restaurant-lk/internal/domain"
)

type CatalogRepository struct {
	db *pgxpool.Pool
}

func NewCatalogRepository(db *pgxpool.Pool) *CatalogRepository {
	return &CatalogRepository{db: db}
}

func (r *CatalogRepository) ListCategories(ctx context.Context, activeOnly bool) ([]domain.ProductCategory, error) {
	query := `
		SELECT id, code, name, description, is_active, created_at, updated_at
		FROM product_categories
	`
	args := make([]any, 0)
	if activeOnly {
		query += ` WHERE is_active = true`
	}
	query += ` ORDER BY name ASC`

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	categories := make([]domain.ProductCategory, 0)
	for rows.Next() {
		category, err := scanCategory(rows)
		if err != nil {
			return nil, err
		}
		categories = append(categories, *category)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return categories, nil
}

func (r *CatalogRepository) GetCategoryByID(ctx context.Context, id uuid.UUID) (*domain.ProductCategory, error) {
	query := `
		SELECT id, code, name, description, is_active, created_at, updated_at
		FROM product_categories
		WHERE id = $1
	`

	category, err := scanCategory(r.db.QueryRow(ctx, query, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return category, nil
}

func (r *CatalogRepository) CreateCategory(ctx context.Context, category *domain.ProductCategory) error {
	query := `
		INSERT INTO product_categories (id, code, name, description, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := r.db.Exec(
		ctx,
		query,
		category.ID,
		category.Code,
		category.Name,
		category.Description,
		category.IsActive,
		category.CreatedAt,
		category.UpdatedAt,
	)
	return err
}

func (r *CatalogRepository) UpdateCategory(ctx context.Context, category *domain.ProductCategory) error {
	query := `
		UPDATE product_categories
		SET code = $1, name = $2, description = $3, is_active = $4, updated_at = now()
		WHERE id = $5
	`

	result, err := r.db.Exec(
		ctx,
		query,
		category.Code,
		category.Name,
		category.Description,
		category.IsActive,
		category.ID,
	)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *CatalogRepository) DeactivateCategory(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE product_categories
		SET is_active = false, updated_at = now()
		WHERE id = $1
	`

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *CatalogRepository) ListFlavors(ctx context.Context, activeOnly bool) ([]domain.TobaccoFlavor, error) {
	query := `
		SELECT id, name, description, is_active, created_at, updated_at
		FROM tobacco_flavors
	`
	if activeOnly {
		query += ` WHERE is_active = true`
	}
	query += ` ORDER BY name ASC`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	flavors := make([]domain.TobaccoFlavor, 0)
	for rows.Next() {
		flavor, err := scanFlavor(rows)
		if err != nil {
			return nil, err
		}
		flavors = append(flavors, *flavor)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return flavors, nil
}

func (r *CatalogRepository) GetFlavorByID(ctx context.Context, id uuid.UUID) (*domain.TobaccoFlavor, error) {
	query := `
		SELECT id, name, description, is_active, created_at, updated_at
		FROM tobacco_flavors
		WHERE id = $1
	`

	flavor, err := scanFlavor(r.db.QueryRow(ctx, query, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return flavor, nil
}

func (r *CatalogRepository) GetFlavorsByIDs(ctx context.Context, ids []uuid.UUID) ([]domain.TobaccoFlavor, error) {
	if len(ids) == 0 {
		return []domain.TobaccoFlavor{}, nil
	}

	query := `
		SELECT id, name, description, is_active, created_at, updated_at
		FROM tobacco_flavors
		WHERE id = ANY($1)
	`
	rows, err := r.db.Query(ctx, query, ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]domain.TobaccoFlavor, 0, len(ids))
	for rows.Next() {
		flavor, err := scanFlavor(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, *flavor)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

func (r *CatalogRepository) CreateFlavor(ctx context.Context, flavor *domain.TobaccoFlavor) error {
	query := `
		INSERT INTO tobacco_flavors (id, name, description, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := r.db.Exec(
		ctx,
		query,
		flavor.ID,
		flavor.Name,
		flavor.Description,
		flavor.IsActive,
		flavor.CreatedAt,
		flavor.UpdatedAt,
	)
	return err
}

func (r *CatalogRepository) UpdateFlavor(ctx context.Context, flavor *domain.TobaccoFlavor) error {
	query := `
		UPDATE tobacco_flavors
		SET name = $1, description = $2, is_active = $3, updated_at = now()
		WHERE id = $4
	`

	result, err := r.db.Exec(
		ctx,
		query,
		flavor.Name,
		flavor.Description,
		flavor.IsActive,
		flavor.ID,
	)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *CatalogRepository) DeactivateFlavor(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE tobacco_flavors
		SET is_active = false, updated_at = now()
		WHERE id = $1
	`

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *CatalogRepository) ListStrengths(ctx context.Context, activeOnly bool) ([]domain.TobaccoStrength, error) {
	query := `
		SELECT id, name, level, description, is_active, created_at, updated_at
		FROM tobacco_strengths
	`
	if activeOnly {
		query += ` WHERE is_active = true`
	}
	query += ` ORDER BY level ASC, name ASC`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	strengths := make([]domain.TobaccoStrength, 0)
	for rows.Next() {
		strength, err := scanStrength(rows)
		if err != nil {
			return nil, err
		}
		strengths = append(strengths, *strength)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return strengths, nil
}

func (r *CatalogRepository) GetStrengthByID(ctx context.Context, id uuid.UUID) (*domain.TobaccoStrength, error) {
	query := `
		SELECT id, name, level, description, is_active, created_at, updated_at
		FROM tobacco_strengths
		WHERE id = $1
	`

	strength, err := scanStrength(r.db.QueryRow(ctx, query, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return strength, nil
}

func (r *CatalogRepository) CreateStrength(ctx context.Context, strength *domain.TobaccoStrength) error {
	query := `
		INSERT INTO tobacco_strengths (id, name, level, description, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := r.db.Exec(
		ctx,
		query,
		strength.ID,
		strength.Name,
		strength.Level,
		strength.Description,
		strength.IsActive,
		strength.CreatedAt,
		strength.UpdatedAt,
	)
	return err
}

func (r *CatalogRepository) UpdateStrength(ctx context.Context, strength *domain.TobaccoStrength) error {
	query := `
		UPDATE tobacco_strengths
		SET name = $1, level = $2, description = $3, is_active = $4, updated_at = now()
		WHERE id = $5
	`

	result, err := r.db.Exec(
		ctx,
		query,
		strength.Name,
		strength.Level,
		strength.Description,
		strength.IsActive,
		strength.ID,
	)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *CatalogRepository) DeactivateStrength(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE tobacco_strengths
		SET is_active = false, updated_at = now()
		WHERE id = $1
	`

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *CatalogRepository) ListProducts(ctx context.Context, filter domain.ProductFilter) ([]domain.Product, error) {
	strengthJoin := "LEFT JOIN tobacco_strengths s ON s.id = p.strength_id"
	if !filter.AdminView {
		strengthJoin = "LEFT JOIN tobacco_strengths s ON s.id = p.strength_id AND s.is_active = true"
	}

	query := `
		SELECT
			p.id, p.category_id, p.name, p.description, p.price::text, p.stock_quantity,
			p.unit, p.is_active, p.strength_id, p.created_at, p.updated_at,
			c.id, c.code, c.name, c.description, c.is_active, c.created_at, c.updated_at,
			s.id, s.name, s.level, s.description, s.is_active, s.created_at, s.updated_at
		FROM products p
		JOIN product_categories c ON c.id = p.category_id
		` + strengthJoin + `
	`

	where := []string{"1=1"}
	args := make([]any, 0)
	index := 1

	if !filter.AdminView {
		where = append(where, "p.is_active = true", "c.is_active = true")
	}
	if filter.IsActive != nil {
		where = append(where, fmt.Sprintf("p.is_active = $%d", index))
		args = append(args, *filter.IsActive)
		index++
	}
	if filter.CategoryCode != nil {
		where = append(where, fmt.Sprintf("c.code = $%d", index))
		args = append(args, *filter.CategoryCode)
		index++
	}
	if filter.Search != nil {
		where = append(where, fmt.Sprintf("(p.name ILIKE $%d OR COALESCE(p.description, '') ILIKE $%d)", index, index))
		args = append(args, "%"+*filter.Search+"%")
		index++
	}
	if filter.MinPrice != nil {
		where = append(where, fmt.Sprintf("p.price >= $%d", index))
		args = append(args, *filter.MinPrice)
		index++
	}
	if filter.MaxPrice != nil {
		where = append(where, fmt.Sprintf("p.price <= $%d", index))
		args = append(args, *filter.MaxPrice)
		index++
	}
	if filter.InStock != nil {
		if *filter.InStock {
			where = append(where, "p.stock_quantity > 0")
		} else {
			where = append(where, "p.stock_quantity = 0")
		}
	}
	if filter.StrengthID != nil {
		where = append(where, fmt.Sprintf("p.strength_id = $%d", index))
		args = append(args, *filter.StrengthID)
		index++
	}
	if filter.FlavorID != nil {
		if filter.AdminView {
			where = append(where, fmt.Sprintf("EXISTS (SELECT 1 FROM product_flavors pf WHERE pf.product_id = p.id AND pf.flavor_id = $%d)", index))
		} else {
			where = append(where, fmt.Sprintf("EXISTS (SELECT 1 FROM product_flavors pf JOIN tobacco_flavors tf ON tf.id = pf.flavor_id WHERE pf.product_id = p.id AND pf.flavor_id = $%d AND tf.is_active = true)", index))
		}
		args = append(args, *filter.FlavorID)
		index++
	}

	query += " WHERE " + strings.Join(where, " AND ")
	query += " ORDER BY p.created_at DESC"
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", index, index+1)
	args = append(args, filter.Limit, filter.Offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	products := make([]domain.Product, 0)
	productIDs := make([]uuid.UUID, 0)
	for rows.Next() {
		product, err := scanProductWithRelations(rows)
		if err != nil {
			return nil, err
		}
		products = append(products, *product)
		productIDs = append(productIDs, product.ID)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	flavorsByProductID, err := r.loadFlavorsByProductIDs(ctx, productIDs, !filter.AdminView)
	if err != nil {
		return nil, err
	}
	tagsByProductID, err := r.loadTagsByProductIDs(ctx, productIDs, !filter.AdminView)
	if err != nil {
		return nil, err
	}

	for i := range products {
		products[i].Flavors = flavorsByProductID[products[i].ID]
		products[i].Tags = tagsByProductID[products[i].ID]
	}

	return products, nil
}

func (r *CatalogRepository) GetProductByID(ctx context.Context, id uuid.UUID, adminView bool) (*domain.Product, error) {
	strengthJoin := "LEFT JOIN tobacco_strengths s ON s.id = p.strength_id"
	if !adminView {
		strengthJoin = "LEFT JOIN tobacco_strengths s ON s.id = p.strength_id AND s.is_active = true"
	}

	query := `
		SELECT
			p.id, p.category_id, p.name, p.description, p.price::text, p.stock_quantity,
			p.unit, p.is_active, p.strength_id, p.created_at, p.updated_at,
			c.id, c.code, c.name, c.description, c.is_active, c.created_at, c.updated_at,
			s.id, s.name, s.level, s.description, s.is_active, s.created_at, s.updated_at
		FROM products p
		JOIN product_categories c ON c.id = p.category_id
		` + strengthJoin + `
		WHERE p.id = $1
	`
	if !adminView {
		query += " AND p.is_active = true AND c.is_active = true"
	}

	product, err := scanProductWithRelations(r.db.QueryRow(ctx, query, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	flavorsByProductID, err := r.loadFlavorsByProductIDs(ctx, []uuid.UUID{id}, !adminView)
	if err != nil {
		return nil, err
	}
	product.Flavors = flavorsByProductID[id]
	tagsByProductID, err := r.loadTagsByProductIDs(ctx, []uuid.UUID{id}, !adminView)
	if err != nil {
		return nil, err
	}
	product.Tags = tagsByProductID[id]

	return product, nil
}

func (r *CatalogRepository) CreateProduct(ctx context.Context, input domain.ProductUpsert) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	insertProductQuery := `
		INSERT INTO products (
			id, category_id, name, description, price, stock_quantity,
			unit, is_active, strength_id, created_at, updated_at
		)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,now(),now())
	`
	_, err = tx.Exec(
		ctx,
		insertProductQuery,
		input.ID,
		input.CategoryID,
		input.Name,
		input.Description,
		input.Price,
		input.StockQuantity,
		input.Unit,
		input.IsActive,
		input.StrengthID,
	)
	if err != nil {
		return err
	}

	if err := insertProductFlavors(ctx, tx, input.ID, input.FlavorIDs); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}
	return nil
}

func (r *CatalogRepository) UpdateProduct(ctx context.Context, input domain.ProductUpsert) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	updateProductQuery := `
		UPDATE products
		SET
			category_id = $1,
			name = $2,
			description = $3,
			price = $4,
			stock_quantity = $5,
			unit = $6,
			is_active = $7,
			strength_id = $8,
			updated_at = now()
		WHERE id = $9
	`

	result, err := tx.Exec(
		ctx,
		updateProductQuery,
		input.CategoryID,
		input.Name,
		input.Description,
		input.Price,
		input.StockQuantity,
		input.Unit,
		input.IsActive,
		input.StrengthID,
		input.ID,
	)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrNotFound
	}

	_, err = tx.Exec(ctx, `DELETE FROM product_flavors WHERE product_id = $1`, input.ID)
	if err != nil {
		return err
	}

	if err := insertProductFlavors(ctx, tx, input.ID, input.FlavorIDs); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}
	return nil
}

func (r *CatalogRepository) DeactivateProduct(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE products
		SET is_active = false, updated_at = now()
		WHERE id = $1
	`

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *CatalogRepository) ApplyProductStockOperation(
	ctx context.Context,
	id uuid.UUID,
	operation domain.StockMovementOperation,
	quantity int,
	reason *string,
	createdByUserID *uuid.UUID,
) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var beforeQuantity int
	lockQuery := `
		SELECT stock_quantity
		FROM products
		WHERE id = $1
		FOR UPDATE
	`
	if err := tx.QueryRow(ctx, lockQuery, id).Scan(&beforeQuantity); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}

	afterQuantity := beforeQuantity
	switch operation {
	case domain.StockMovementOperationSet:
		afterQuantity = quantity
	case domain.StockMovementOperationIncrement:
		afterQuantity = beforeQuantity + quantity
	case domain.StockMovementOperationDecrement:
		if beforeQuantity < quantity {
			return ErrInsufficientStock
		}
		afterQuantity = beforeQuantity - quantity
	default:
		return ErrInvalidStockOperation
	}

	if afterQuantity < 0 {
		return ErrInsufficientStock
	}

	now := time.Now().UTC()
	updateQuery := `
		UPDATE products
		SET stock_quantity = $1, updated_at = $2
		WHERE id = $3
	`
	if _, err := tx.Exec(ctx, updateQuery, afterQuantity, now, id); err != nil {
		return err
	}

	movement := &domain.StockMovement{
		ID:              uuid.New(),
		ProductID:       id,
		Operation:       operation,
		Quantity:        quantity,
		BeforeQuantity:  beforeQuantity,
		AfterQuantity:   afterQuantity,
		Reason:          reason,
		CreatedByUserID: createdByUserID,
		CreatedAt:       now,
	}
	if err := createStockMovementTx(ctx, tx, movement); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}
	return nil
}

func insertProductFlavors(ctx context.Context, tx pgx.Tx, productID uuid.UUID, flavorIDs []uuid.UUID) error {
	if len(flavorIDs) == 0 {
		return nil
	}

	query := `
		INSERT INTO product_flavors (product_id, flavor_id)
		VALUES ($1, $2)
	`

	for _, flavorID := range flavorIDs {
		if _, err := tx.Exec(ctx, query, productID, flavorID); err != nil {
			return err
		}
	}

	return nil
}

func (r *CatalogRepository) loadFlavorsByProductIDs(ctx context.Context, productIDs []uuid.UUID, activeOnly bool) (map[uuid.UUID][]domain.TobaccoFlavor, error) {
	result := make(map[uuid.UUID][]domain.TobaccoFlavor, len(productIDs))
	if len(productIDs) == 0 {
		return result, nil
	}

	query := `
		SELECT pf.product_id, f.id, f.name, f.description, f.is_active, f.created_at, f.updated_at
		FROM product_flavors pf
		JOIN tobacco_flavors f ON f.id = pf.flavor_id
		WHERE pf.product_id = ANY($1)
	`
	if activeOnly {
		query += ` AND f.is_active = true`
	}
	query += ` ORDER BY f.name ASC`

	rows, err := r.db.Query(ctx, query, productIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var productID uuid.UUID
		flavor, err := scanFlavorWithProductID(rows, &productID)
		if err != nil {
			return nil, err
		}
		result[productID] = append(result[productID], *flavor)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

func (r *CatalogRepository) loadTagsByProductIDs(ctx context.Context, productIDs []uuid.UUID, activeOnly bool) (map[uuid.UUID][]domain.Tag, error) {
	result := make(map[uuid.UUID][]domain.Tag, len(productIDs))
	if len(productIDs) == 0 {
		return result, nil
	}

	query := `
		SELECT pt.product_id, t.id, t.code, t.name, t.description, t.is_active, t.created_at, t.updated_at
		FROM product_tags pt
		JOIN tags t ON t.id = pt.tag_id
		WHERE pt.product_id = ANY($1)
	`
	if activeOnly {
		query += ` AND t.is_active = true`
	}
	query += ` ORDER BY t.name ASC`

	rows, err := r.db.Query(ctx, query, productIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			productID   uuid.UUID
			tag         domain.Tag
			description sql.NullString
		)
		if err := rows.Scan(
			&productID,
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
		result[productID] = append(result[productID], tag)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

func scanCategory(row pgx.Row) (*domain.ProductCategory, error) {
	var (
		category    domain.ProductCategory
		description sql.NullString
	)

	err := row.Scan(
		&category.ID,
		&category.Code,
		&category.Name,
		&description,
		&category.IsActive,
		&category.CreatedAt,
		&category.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	if description.Valid {
		value := description.String
		category.Description = &value
	}

	return &category, nil
}

func scanFlavor(row pgx.Row) (*domain.TobaccoFlavor, error) {
	var (
		flavor      domain.TobaccoFlavor
		description sql.NullString
	)

	err := row.Scan(
		&flavor.ID,
		&flavor.Name,
		&description,
		&flavor.IsActive,
		&flavor.CreatedAt,
		&flavor.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	if description.Valid {
		value := description.String
		flavor.Description = &value
	}

	return &flavor, nil
}

func scanFlavorWithProductID(row pgx.Row, productID *uuid.UUID) (*domain.TobaccoFlavor, error) {
	var (
		flavor      domain.TobaccoFlavor
		description sql.NullString
	)

	err := row.Scan(
		productID,
		&flavor.ID,
		&flavor.Name,
		&description,
		&flavor.IsActive,
		&flavor.CreatedAt,
		&flavor.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	if description.Valid {
		value := description.String
		flavor.Description = &value
	}

	return &flavor, nil
}

func scanStrength(row pgx.Row) (*domain.TobaccoStrength, error) {
	var (
		strength    domain.TobaccoStrength
		description sql.NullString
	)

	err := row.Scan(
		&strength.ID,
		&strength.Name,
		&strength.Level,
		&description,
		&strength.IsActive,
		&strength.CreatedAt,
		&strength.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	if description.Valid {
		value := description.String
		strength.Description = &value
	}

	return &strength, nil
}

func scanProductWithRelations(row pgx.Row) (*domain.Product, error) {
	var (
		product              domain.Product
		productDescription   sql.NullString
		strengthID           uuid.NullUUID
		categoryDescription  sql.NullString
		strengthObj          domain.TobaccoStrength
		strengthDescription  sql.NullString
		strengthObjID        uuid.NullUUID
		strengthObjCreatedAt sql.NullTime
		strengthObjUpdatedAt sql.NullTime
		strengthObjIsActive  sql.NullBool
		strengthObjLevel     sql.NullInt16
		strengthObjName      sql.NullString
	)

	err := row.Scan(
		&product.ID,
		&product.CategoryID,
		&product.Name,
		&productDescription,
		&product.Price,
		&product.StockQuantity,
		&product.Unit,
		&product.IsActive,
		&strengthID,
		&product.CreatedAt,
		&product.UpdatedAt,
		&product.Category.ID,
		&product.Category.Code,
		&product.Category.Name,
		&categoryDescription,
		&product.Category.IsActive,
		&product.Category.CreatedAt,
		&product.Category.UpdatedAt,
		&strengthObjID,
		&strengthObjName,
		&strengthObjLevel,
		&strengthDescription,
		&strengthObjIsActive,
		&strengthObjCreatedAt,
		&strengthObjUpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	if productDescription.Valid {
		value := productDescription.String
		product.Description = &value
	}
	if categoryDescription.Valid {
		value := categoryDescription.String
		product.Category.Description = &value
	}
	if strengthID.Valid {
		value := strengthID.UUID
		product.StrengthID = &value
	}

	if strengthObjID.Valid {
		strengthObj.ID = strengthObjID.UUID
		if strengthObjName.Valid {
			strengthObj.Name = strengthObjName.String
		}
		if strengthObjLevel.Valid {
			strengthObj.Level = strengthObjLevel.Int16
		}
		if strengthDescription.Valid {
			value := strengthDescription.String
			strengthObj.Description = &value
		}
		if strengthObjIsActive.Valid {
			strengthObj.IsActive = strengthObjIsActive.Bool
		}
		if strengthObjCreatedAt.Valid {
			strengthObj.CreatedAt = strengthObjCreatedAt.Time
		}
		if strengthObjUpdatedAt.Valid {
			strengthObj.UpdatedAt = strengthObjUpdatedAt.Time
		}
		product.Strength = &strengthObj
	}

	return &product, nil
}
