package pg

import (
	"context"
	"fmt"
	"time"

	"eco-van-api/internal/models"
	"eco-van-api/internal/port"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// warehouseRepository implements port.WarehouseRepository
type warehouseRepository struct {
	*BaseRepository
	pool *pgxpool.Pool
}

// NewWarehouseRepository creates a new warehouse repository
func NewWarehouseRepository(pool *pgxpool.Pool) port.WarehouseRepository {
	return &warehouseRepository{
		BaseRepository: NewBaseRepository(pool),
		pool:           pool,
	}
}

// Create creates a new warehouse
func (r *warehouseRepository) Create(ctx context.Context, warehouse *models.Warehouse) error {
	query := `
		INSERT INTO warehouses (id, name, address, notes, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	now := time.Now()
	warehouse.ID = uuid.New()
	warehouse.CreatedAt = now
	warehouse.UpdatedAt = now

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
	whereClause := r.BuildWhereClause(whereClauses)

	// Count total
	total, err := r.CountTotal(ctx, baseQuery, whereClause, args)
	if err != nil {
		return nil, fmt.Errorf("failed to count warehouses: %w", err)
	}

	// Build pagination query
	// Build the main query with pagination
	limitParamNum := len(args) + 1
	const offsetIncrement = 2
	offsetParamNum := len(args) + offsetIncrement
	mainQuery := fmt.Sprintf(`
		SELECT id, name, address, geo_lat, geo_lng, notes, created_at, updated_at, deleted_at
		%s
		ORDER BY name
		LIMIT $%d
		OFFSET $%d
	`, baseQuery, limitParamNum, offsetParamNum)

	// Calculate pagination
	limit := req.PageSize
	offset := (req.Page - 1) * req.PageSize
	args = append(args, limit, offset)

	rows, err := r.pool.Query(ctx, mainQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list warehouses: %w", err)
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
			return nil, fmt.Errorf("failed to scan warehouse: %w", err)
		}
		warehouses = append(warehouses, warehouse)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over warehouses: %w", err)
	}

	return &models.WarehouseListResponse{
		Items:    warehouses,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

// Update updates an existing warehouse
func (r *warehouseRepository) Update(ctx context.Context, warehouse *models.Warehouse) error {
	query := `
		UPDATE warehouses
		SET name = $1, address = $2, notes = $3, updated_at = $4
		WHERE id = $5
	`

	warehouse.UpdatedAt = time.Now()
	_, err := r.pool.Exec(ctx, query,
		warehouse.Name,
		warehouse.Address,
		warehouse.Notes,
		warehouse.UpdatedAt,
		warehouse.ID,
	)

	return err
}

// SoftDelete marks a warehouse as deleted by setting deleted_at
func (r *warehouseRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	return r.SoftDeleteGeneric(ctx, "warehouses", id)
}

// Restore restores a soft-deleted warehouse by clearing deleted_at
func (r *warehouseRepository) Restore(ctx context.Context, id uuid.UUID) error {
	return r.RestoreGeneric(ctx, "warehouses", id)
}

// ExistsByName checks if a warehouse exists with the given name (excluding soft-deleted)
//
//nolint:dupl // Similar pattern across repositories but with different field names
func (r *warehouseRepository) ExistsByName(ctx context.Context, name string, excludeID *uuid.UUID) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM warehouses WHERE name = $1 AND deleted_at IS NULL`
	args := []interface{}{name}
	argIndex := 2

	if excludeID != nil {
		query += fmt.Sprintf(" AND id != $%d", argIndex)
		args = append(args, *excludeID)
	}
	query += ")"

	var exists bool
	err := r.pool.QueryRow(ctx, query, args...).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check warehouse name existence: %w", err)
	}

	return exists, nil
}

// HasActiveEquipment checks if a warehouse has any non-deleted equipment
func (r *warehouseRepository) HasActiveEquipment(ctx context.Context, warehouseID uuid.UUID) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM equipment 
			WHERE warehouse_id = $1 AND deleted_at IS NULL
		)
	`
	var exists bool
	err := r.pool.QueryRow(ctx, query, warehouseID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check warehouse equipment: %w", err)
	}
	return exists, nil
}
