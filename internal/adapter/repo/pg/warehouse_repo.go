package pg

import (
	"context"
	"fmt"
	"strings"

	"eco-van-api/internal/models"
	"eco-van-api/internal/port"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// warehouseRepository implements port.WarehouseRepository
type warehouseRepository struct {
	pool *pgxpool.Pool
}

// NewWarehouseRepository creates a new warehouse repository
func NewWarehouseRepository(pool *pgxpool.Pool) port.WarehouseRepository {
	return &warehouseRepository{pool: pool}
}

// Create creates a new warehouse
func (r *warehouseRepository) Create(ctx context.Context, warehouse *models.Warehouse) error {
	query := `
		INSERT INTO warehouses (id, name, address, notes, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := r.pool.Exec(ctx, query,
		warehouse.ID,
		warehouse.Name,
		warehouse.Address,
		warehouse.Notes,
		warehouse.CreatedAt,
		warehouse.UpdatedAt,
	)

	return err
}

// GetByID retrieves a warehouse by ID, optionally including soft-deleted
func (r *warehouseRepository) GetByID(ctx context.Context, id uuid.UUID, includeDeleted bool) (*models.Warehouse, error) {
	query := `
		SELECT id, name, address, notes, created_at, updated_at, deleted_at
		FROM warehouses
		WHERE id = $1
	`

	if !includeDeleted {
		query += " AND deleted_at IS NULL"
	}

	var warehouse models.Warehouse
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&warehouse.ID,
		&warehouse.Name,
		&warehouse.Address,
		&warehouse.Notes,
		&warehouse.CreatedAt,
		&warehouse.UpdatedAt,
		&warehouse.DeletedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get warehouse: %w", err)
	}

	return &warehouse, nil
}

// List retrieves warehouses with pagination and filtering
func (r *warehouseRepository) List(ctx context.Context, req models.WarehouseListRequest) (*models.WarehouseListResponse, error) {
	// Build base query
	baseQuery := "FROM warehouses"
	whereClauses := []string{}
	args := []interface{}{}

	// Add soft-delete filter
	if !req.IncludeDeleted {
		whereClauses = append(whereClauses, "deleted_at IS NULL")
	}

	// Build WHERE clause
	whereClause := ""
	if len(whereClauses) > 0 {
		whereClause = "WHERE " + strings.Join(whereClauses, " AND ")
	}

	// Count total
	countQuery := fmt.Sprintf("SELECT COUNT(*) %s %s", baseQuery, whereClause)
	var total int64
	err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, err
	}

	// Get paginated results
	offset := (req.Page - 1) * req.PageSize
	query := fmt.Sprintf(`
		SELECT id, name, address, notes, created_at, updated_at, deleted_at
		%s %s
		ORDER BY name
		LIMIT $%d OFFSET $%d
	`, baseQuery, whereClause, len(args)+1, len(args)+2)

	args = append(args, req.PageSize, offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var warehouses []models.Warehouse
	for rows.Next() {
		var warehouse models.Warehouse
		err := rows.Scan(
			&warehouse.ID,
			&warehouse.Name,
			&warehouse.Address,
			&warehouse.Notes,
			&warehouse.CreatedAt,
			&warehouse.UpdatedAt,
			&warehouse.DeletedAt,
		)
		if err != nil {
			return nil, err
		}
		warehouses = append(warehouses, warehouse)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return &models.WarehouseListResponse{
		Items:    warehouses,
		Page:     req.Page,
		PageSize: req.PageSize,
		Total:    total,
	}, nil
}

// Update updates an existing warehouse
func (r *warehouseRepository) Update(ctx context.Context, warehouse *models.Warehouse) error {
	query := `
			UPDATE warehouses
			SET name = $1, address = $2, notes = $3, updated_at = $4
			WHERE id = $5 AND deleted_at IS NULL
		`

	result, err := r.pool.Exec(ctx, query,
		warehouse.Name,
		warehouse.Address,
		warehouse.Notes,
		warehouse.UpdatedAt,
		warehouse.ID,
	)

	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("warehouse not found")
	}

	return nil
}

// SoftDelete marks a warehouse as deleted by setting deleted_at
func (r *warehouseRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	query := `
			UPDATE warehouses
			SET deleted_at = now()
			WHERE id = $1 AND deleted_at IS NULL
		`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("warehouse not found")
	}

	return nil
}

// Restore restores a soft-deleted warehouse by clearing deleted_at
func (r *warehouseRepository) Restore(ctx context.Context, id uuid.UUID) error {
	query := `
			UPDATE warehouses
			SET deleted_at = NULL
			WHERE id = $1 AND deleted_at IS NOT NULL
		`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("warehouse not found")
	}

	return nil
}

// ExistsByName checks if a warehouse exists with the given name (excluding soft-deleted)
func (r *warehouseRepository) ExistsByName(ctx context.Context, name string, excludeID *uuid.UUID) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM warehouses
			WHERE name = $1 AND deleted_at IS NULL
		)
	`

	args := []interface{}{name}
	if excludeID != nil {
		query = `
			SELECT EXISTS(
				SELECT 1 FROM warehouses
				WHERE name = $1 AND id != $2 AND deleted_at IS NULL
			)
		`
		args = append(args, *excludeID)
	}

	var exists bool
	err := r.pool.QueryRow(ctx, query, args...).Scan(&exists)
	return exists, err
}

// HasActiveEquipment checks if a warehouse has any non-deleted equipment
func (r *warehouseRepository) HasActiveEquipment(ctx context.Context, warehouseID uuid.UUID) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM equipment
			WHERE warehouse_id = $1 AND deleted_at IS NULL
		)
	`

	var hasEquipment bool
	err := r.pool.QueryRow(ctx, query, warehouseID).Scan(&hasEquipment)
	return hasEquipment, err
}
