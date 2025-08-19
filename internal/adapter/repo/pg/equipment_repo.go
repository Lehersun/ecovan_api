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

// equipmentRepository implements port.EquipmentRepository
type equipmentRepository struct {
	pool *pgxpool.Pool
}

// NewEquipmentRepository creates a new equipment repository
func NewEquipmentRepository(pool *pgxpool.Pool) port.EquipmentRepository {
	return &equipmentRepository{pool: pool}
}

// Create creates a new equipment
func (r *equipmentRepository) Create(ctx context.Context, equipment *models.Equipment) error {
	query := `
		INSERT INTO equipment (id, number, type, volume_l, condition, photo, client_object_id, warehouse_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	_, err := r.pool.Exec(ctx, query,
		equipment.ID,
		equipment.Number,
		equipment.Type,
		equipment.VolumeL,
		equipment.Condition,
		equipment.Photo,
		equipment.ClientObjectID,
		equipment.WarehouseID,
		equipment.CreatedAt,
		equipment.UpdatedAt,
	)

	return err
}

// GetByID retrieves equipment by ID, optionally including soft-deleted
func (r *equipmentRepository) GetByID(ctx context.Context, id uuid.UUID, includeDeleted bool) (*models.Equipment, error) {
	query := `
		SELECT id, number, type, volume_l, condition, photo, client_object_id, warehouse_id, created_at, updated_at, deleted_at
		FROM equipment
		WHERE id = $1
	`

	if !includeDeleted {
		query += " AND deleted_at IS NULL"
	}

	var equipment models.Equipment
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&equipment.ID,
		&equipment.Number,
		&equipment.Type,
		&equipment.VolumeL,
		&equipment.Condition,
		&equipment.Photo,
		&equipment.ClientObjectID,
		&equipment.WarehouseID,
		&equipment.CreatedAt,
		&equipment.UpdatedAt,
		&equipment.DeletedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get equipment: %w", err)
	}

	return &equipment, nil
}

// List retrieves equipment with pagination and filtering
func (r *equipmentRepository) List(ctx context.Context, req models.EquipmentListRequest) (*models.EquipmentListResponse, error) {
	// Build base query
	baseQuery := "FROM equipment"
	whereClauses := []string{}
	args := []interface{}{}
	argIndex := 1

	// Add soft-delete filter
	if !req.IncludeDeleted {
		whereClauses = append(whereClauses, "deleted_at IS NULL")
	}

	// Add type filter
	if req.Type != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("type = $%d", argIndex))
		args = append(args, string(*req.Type))
		argIndex++
	}

	// Add client object filter
	if req.ClientObjectID != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("client_object_id = $%d", argIndex))
		args = append(args, *req.ClientObjectID)
		argIndex++
	}

	// Add warehouse filter
	if req.WarehouseID != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("warehouse_id = $%d", argIndex))
		args = append(args, *req.WarehouseID)
		argIndex++
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
		return nil, fmt.Errorf("failed to count equipment: %w", err)
	}

	// Build SELECT query with pagination
	selectQuery := fmt.Sprintf(`
		SELECT id, number, type, volume_l, condition, photo, client_object_id, warehouse_id, created_at, updated_at, deleted_at
		%s %s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, baseQuery, whereClause, argIndex, argIndex+1)

	// Add pagination args
	offset := (req.Page - 1) * req.PageSize
	args = append(args, req.PageSize, offset)

	// Execute query
	rows, err := r.pool.Query(ctx, selectQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list equipment: %w", err)
	}
	defer rows.Close()

	var items []models.EquipmentResponse
	for rows.Next() {
		var equipment models.Equipment
		err := rows.Scan(
			&equipment.ID,
			&equipment.Number,
			&equipment.Type,
			&equipment.VolumeL,
			&equipment.Condition,
			&equipment.Photo,
			&equipment.ClientObjectID,
			&equipment.WarehouseID,
			&equipment.CreatedAt,
			&equipment.UpdatedAt,
			&equipment.DeletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan equipment: %w", err)
		}
		items = append(items, equipment.ToResponse())
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating equipment rows: %w", err)
	}

	return &models.EquipmentListResponse{
		Items:    items,
		Page:     req.Page,
		PageSize: req.PageSize,
		Total:    total,
	}, nil
}

// Update updates an existing equipment
func (r *equipmentRepository) Update(ctx context.Context, equipment *models.Equipment) error {
	query := `
		UPDATE equipment
		SET number = $2, type = $3, volume_l = $4, condition = $5, photo = $6, 
		    client_object_id = $7, warehouse_id = $8, updated_at = $9
		WHERE id = $1
	`

	result, err := r.pool.Exec(ctx, query,
		equipment.ID,
		equipment.Number,
		equipment.Type,
		equipment.VolumeL,
		equipment.Condition,
		equipment.Photo,
		equipment.ClientObjectID,
		equipment.WarehouseID,
		equipment.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update equipment: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("equipment not found")
	}

	return nil
}

// SoftDelete marks equipment as deleted by setting deleted_at
func (r *equipmentRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE equipment
		SET deleted_at = now(), updated_at = now()
		WHERE id = $1 AND deleted_at IS NULL
	`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete equipment: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("equipment not found or already deleted")
	}

	return nil
}

// Restore restores a soft-deleted equipment by clearing deleted_at
func (r *equipmentRepository) Restore(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE equipment
		SET deleted_at = NULL, updated_at = now()
		WHERE id = $1 AND deleted_at IS NOT NULL
	`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to restore equipment: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("equipment not found or not deleted")
	}

	return nil
}

// IsAttachedToTransport checks if equipment is currently attached to a transport
func (r *equipmentRepository) IsAttachedToTransport(ctx context.Context, equipmentID uuid.UUID) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM transport 
			WHERE current_equipment_id = $1 AND deleted_at IS NULL
		)
	`

	var exists bool
	err := r.pool.QueryRow(ctx, query, equipmentID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check transport attachment: %w", err)
	}

	return exists, nil
}

// ExistsByNumber checks if equipment exists with the given number (excluding soft-deleted)
func (r *equipmentRepository) ExistsByNumber(ctx context.Context, number string, excludeID *uuid.UUID) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM equipment 
			WHERE number = $1 AND deleted_at IS NULL
		)
	`

	if excludeID != nil {
		query = `
			SELECT EXISTS(
				SELECT 1 FROM equipment 
				WHERE number = $1 AND id != $2 AND deleted_at IS NULL
			)
		`
	}

	var exists bool
	var err error
	if excludeID != nil {
		err = r.pool.QueryRow(ctx, query, number, *excludeID).Scan(&exists)
	} else {
		err = r.pool.QueryRow(ctx, query, number).Scan(&exists)
	}

	if err != nil {
		return false, fmt.Errorf("failed to check equipment number existence: %w", err)
	}

	return exists, nil
}
