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

type clientObjectRepository struct {
	pool *pgxpool.Pool
}

// NewClientObjectRepository creates a new PostgreSQL client object repository
func NewClientObjectRepository(pool *pgxpool.Pool) port.ClientObjectRepository {
	return &clientObjectRepository{pool: pool}
}

// Create creates a new client object
func (r *clientObjectRepository) Create(ctx context.Context, clientObject *models.ClientObject) error {
	query := `
		INSERT INTO client_objects (id, client_id, name, address, geo_lat, geo_lng, notes, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	_, err := r.pool.Exec(ctx, query,
		clientObject.ID,
		clientObject.ClientID,
		clientObject.Name,
		clientObject.Address,
		clientObject.GeoLat,
		clientObject.GeoLng,
		clientObject.Notes,
		clientObject.CreatedAt,
		clientObject.UpdatedAt,
	)
	return err
}

// GetByID retrieves a client object by ID
func (r *clientObjectRepository) GetByID(ctx context.Context, id uuid.UUID, includeDeleted bool) (*models.ClientObject, error) {
	query := `
		SELECT id, client_id, name, address, geo_lat, geo_lng, notes, created_at, updated_at, deleted_at
		FROM client_objects
		WHERE id = $1
	`
	if !includeDeleted {
		query += " AND deleted_at IS NULL"
	}

	var clientObject models.ClientObject
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&clientObject.ID,
		&clientObject.ClientID,
		&clientObject.Name,
		&clientObject.Address,
		&clientObject.GeoLat,
		&clientObject.GeoLng,
		&clientObject.Notes,
		&clientObject.CreatedAt,
		&clientObject.UpdatedAt,
		&clientObject.DeletedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("client object not found")
		}
		return nil, fmt.Errorf("failed to get client object: %w", err)
	}

	return &clientObject, nil
}

// List retrieves client objects for a specific client with pagination
func (r *clientObjectRepository) List(ctx context.Context, clientID uuid.UUID, req models.ClientObjectListRequest) (*models.ClientObjectListResponse, error) {
	// Set defaults
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 {
		req.PageSize = 20
	}
	if req.PageSize > 100 {
		req.PageSize = 100
	}

	// Build base query
	baseQuery := `
		FROM client_objects
		WHERE client_id = $1
	`
	args := []interface{}{clientID}

	// Add soft-delete filter
	if !req.IncludeDeleted {
		baseQuery += " AND deleted_at IS NULL"
	}

	// Count total records
	countQuery := "SELECT COUNT(*) " + baseQuery
	var total int64
	err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, err
	}

	// Build the main query with pagination
	limitParamNum := len(args) + 1
	offsetParamNum := len(args) + 2
	mainQuery := `
		SELECT id, client_id, name, address, geo_lat, geo_lng, notes, created_at, updated_at, deleted_at
		` + baseQuery + `
		ORDER BY name
		LIMIT $` + fmt.Sprintf("%d", limitParamNum) + `
		OFFSET $` + fmt.Sprintf("%d", offsetParamNum) + `
	`

	// Calculate pagination
	offset := (req.Page - 1) * req.PageSize
	args = append(args, req.PageSize, offset)

	// Execute main query
	rows, err := r.pool.Query(ctx, mainQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []models.ClientObjectResponse
	for rows.Next() {
		var clientObject models.ClientObject
		err := rows.Scan(
			&clientObject.ID,
			&clientObject.ClientID,
			&clientObject.Name,
			&clientObject.Address,
			&clientObject.GeoLat,
			&clientObject.GeoLng,
			&clientObject.Notes,
			&clientObject.CreatedAt,
			&clientObject.UpdatedAt,
			&clientObject.DeletedAt,
		)
		if err != nil {
			return nil, err
		}
		items = append(items, clientObject.ToResponse())
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return &models.ClientObjectListResponse{
		Items:    items,
		Page:     req.Page,
		PageSize: req.PageSize,
		Total:    total,
	}, nil
}

// Update updates an existing client object
func (r *clientObjectRepository) Update(ctx context.Context, clientObject *models.ClientObject) error {
	query := `
		UPDATE client_objects
		SET name = $1, address = $2, geo_lat = $3, geo_lng = $4, notes = $5, updated_at = $6
		WHERE id = $7 AND deleted_at IS NULL
	`
	result, err := r.pool.Exec(ctx, query,
		clientObject.Name,
		clientObject.Address,
		clientObject.GeoLat,
		clientObject.GeoLng,
		clientObject.Notes,
		clientObject.UpdatedAt,
		clientObject.ID,
	)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("client object not found or already deleted")
	}

	return nil
}

// SoftDelete soft deletes a client object (guarded)
func (r *clientObjectRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	// Check for conflicts before deletion
	conflicts, err := r.GetDeleteConflicts(ctx, id)
	if err != nil {
		return err
	}

	if conflicts.HasActiveOrders || conflicts.HasActiveEquipment {
		return fmt.Errorf("cannot delete client object: %s", conflicts.Message)
	}

	query := `
		UPDATE client_objects
		SET deleted_at = now(), updated_at = now()
		WHERE id = $1 AND deleted_at IS NULL
	`
	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("client object not found or already deleted")
	}

	return nil
}

// Restore restores a soft-deleted client object
func (r *clientObjectRepository) Restore(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE client_objects
		SET deleted_at = NULL, updated_at = now()
		WHERE id = $1 AND deleted_at IS NOT NULL
	`
	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("client object not found or not deleted")
	}

	return nil
}

// ExistsByName checks if a client object with the given name exists for a client
func (r *clientObjectRepository) ExistsByName(ctx context.Context, clientID uuid.UUID, name string, excludeID *uuid.UUID) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM client_objects
			WHERE client_id = $1 AND name = $2 AND deleted_at IS NULL
		)
	`
	args := []interface{}{clientID, name}

	if excludeID != nil {
		query = `
			SELECT EXISTS(
				SELECT 1 FROM client_objects
				WHERE client_id = $1 AND name = $2 AND id != $3 AND deleted_at IS NULL
			)
		`
		args = append(args, *excludeID)
	}

	var exists bool
	err := r.pool.QueryRow(ctx, query, args...).Scan(&exists)
	return exists, err
}

// HasActiveOrders checks if there are active orders for this client object
func (r *clientObjectRepository) HasActiveOrders(ctx context.Context, clientObjectID uuid.UUID) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM orders
			WHERE object_id = $1 
			AND status IN ('DRAFT', 'SCHEDULED', 'IN_PROGRESS')
			AND deleted_at IS NULL
		)
	`
	var exists bool
	err := r.pool.QueryRow(ctx, query, clientObjectID).Scan(&exists)
	return exists, err
}

// HasActiveEquipment checks if there is active equipment placed at this client object
func (r *clientObjectRepository) HasActiveEquipment(ctx context.Context, clientObjectID uuid.UUID) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM equipment
			WHERE client_object_id = $1 AND deleted_at IS NULL
		)
	`
	var exists bool
	err := r.pool.QueryRow(ctx, query, clientObjectID).Scan(&exists)
	return exists, err
}

// GetDeleteConflicts returns detailed information about what prevents deletion
func (r *clientObjectRepository) GetDeleteConflicts(ctx context.Context, clientObjectID uuid.UUID) (*models.DeleteConflicts, error) {
	conflicts := &models.DeleteConflicts{}

	// Check for active orders
	orderQuery := `
		SELECT id FROM orders
		WHERE object_id = $1 
		AND status IN ('DRAFT', 'SCHEDULED', 'IN_PROGRESS')
		AND deleted_at IS NULL
	`
	orderRows, err := r.pool.Query(ctx, orderQuery, clientObjectID)
	if err != nil {
		return nil, err
	}
	defer orderRows.Close()

	var activeOrderIDs []uuid.UUID
	for orderRows.Next() {
		var orderID uuid.UUID
		if err := orderRows.Scan(&orderID); err != nil {
			return nil, err
		}
		activeOrderIDs = append(activeOrderIDs, orderID)
	}
	conflicts.HasActiveOrders = len(activeOrderIDs) > 0
	conflicts.ActiveOrderIDs = activeOrderIDs

	// Check for active equipment
	equipmentQuery := `
		SELECT id FROM equipment
		WHERE client_object_id = $1 AND deleted_at IS NULL
	`
	equipmentRows, err := r.pool.Query(ctx, equipmentQuery, clientObjectID)
	if err != nil {
		return nil, err
	}
	defer equipmentRows.Close()

	var activeEquipmentIDs []uuid.UUID
	for equipmentRows.Next() {
		var equipmentID uuid.UUID
		if err := equipmentRows.Scan(&equipmentID); err != nil {
			return nil, err
		}
		activeEquipmentIDs = append(activeEquipmentIDs, equipmentID)
	}
	conflicts.HasActiveEquipment = len(activeEquipmentIDs) > 0
	conflicts.ActiveEquipmentIDs = activeEquipmentIDs

	// Build conflict message
	var messages []string
	if conflicts.HasActiveOrders {
		messages = append(messages, fmt.Sprintf("has %d active orders", len(activeOrderIDs)))
	}
	if conflicts.HasActiveEquipment {
		messages = append(messages, fmt.Sprintf("has %d active equipment", len(activeEquipmentIDs)))
	}
	conflicts.Message = "Cannot delete client object: " + strings.Join(messages, ", ")

	return conflicts, nil
}
