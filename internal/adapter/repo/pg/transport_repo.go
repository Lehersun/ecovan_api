package pg

import (
	"context"
	"fmt"
	"strings"
	"time"

	"eco-van-api/internal/models"
	"eco-van-api/internal/port"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// transportRepository implements port.TransportRepository
type transportRepository struct {
	*BaseRepository
	pool *pgxpool.Pool
}

// NewTransportRepository creates a new transport repository
func NewTransportRepository(pool *pgxpool.Pool) port.TransportRepository {
	return &transportRepository{
		BaseRepository: NewBaseRepository(pool),
		pool:           pool,
	}
}

// Create creates a new transport
func (r *transportRepository) Create(ctx context.Context, transport *models.Transport) error {
	query := `
		INSERT INTO transport (id, plate_no, brand, model, capacity_l, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	now := time.Now()
	transport.ID = uuid.New()
	transport.CreatedAt = now
	transport.UpdatedAt = now

	_, err := r.pool.Exec(ctx, query,
		transport.ID,
		transport.PlateNo,
		transport.Brand,
		transport.Model,
		transport.CapacityL,
		transport.Status,
		transport.CreatedAt,
		transport.UpdatedAt,
	)

	return err
}

// GetByID retrieves a transport by ID, optionally including soft-deleted
func (r *transportRepository) GetByID(ctx context.Context, id uuid.UUID, includeDeleted bool) (*models.Transport, error) {
	query := `
		SELECT id, plate_no, brand, model, capacity_l, current_driver_id, current_equipment_id, status, created_at, updated_at, deleted_at
		FROM transport
		WHERE id = $1
	`

	if !includeDeleted {
		query += " AND deleted_at IS NULL"
	}

	var transport models.Transport
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&transport.ID,
		&transport.PlateNo,
		&transport.Brand,
		&transport.Model,
		&transport.CapacityL,
		&transport.CurrentDriverID,
		&transport.CurrentEquipmentID,
		&transport.Status,
		&transport.CreatedAt,
		&transport.UpdatedAt,
		&transport.DeletedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get transport: %w", err)
	}

	return &transport, nil
}

// Update updates an existing transport
func (r *transportRepository) Update(ctx context.Context, transport *models.Transport) error {
	query := `
		UPDATE transport 
		SET plate_no = $1, brand = $2, model = $3, capacity_l = $4, status = $5, updated_at = $6
		WHERE id = $7 AND deleted_at IS NULL
	`

	transport.UpdatedAt = time.Now()
	result, err := r.pool.Exec(ctx, query,
		transport.PlateNo,
		transport.Brand,
		transport.Model,
		transport.CapacityL,
		transport.Status,
		transport.UpdatedAt,
		transport.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update transport: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("transport not found or already deleted")
	}

	return nil
}

// SoftDelete soft-deletes a transport
func (r *transportRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	return r.SoftDeleteGeneric(ctx, "transport", id)
}

// Restore restores a soft-deleted transport
func (r *transportRepository) Restore(ctx context.Context, id uuid.UUID) error {
	return r.RestoreGeneric(ctx, "transport", id)
}

// List implements the list method with filtering
func (r *transportRepository) List(ctx context.Context, req models.TransportListRequest) (*models.TransportListResponse, error) {
	// Build the base query
	baseQuery := "FROM transport WHERE 1=1"
	var conditions []string
	var args []interface{}
	argIndex := 1

	// Add status filter if provided
	if req.Status != nil && *req.Status != "" {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argIndex))
		args = append(args, *req.Status)
		argIndex++
	}

	// Add soft-delete filter
	if !req.IncludeDeleted {
		conditions = append(conditions, "deleted_at IS NULL")
	}

	// Build WHERE clause
	if len(conditions) > 0 {
		baseQuery += " AND " + strings.Join(conditions, " AND ")
	}

	// Count total items
	countQuery := "SELECT COUNT(*) " + baseQuery
	var total int64
	err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("failed to count transport items: %w", err)
	}

	// Get items with pagination
	query := `
		SELECT id, plate_no, brand, model, capacity_l, current_driver_id, current_equipment_id, status, created_at, updated_at, deleted_at
		` + baseQuery + `
		ORDER BY created_at DESC
		LIMIT $` + fmt.Sprintf("%d", argIndex) + ` OFFSET $` + fmt.Sprintf("%d", argIndex+1)

	offset := (req.Page - 1) * req.PageSize
	args = append(args, req.PageSize, offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query transport items: %w", err)
	}
	defer rows.Close()

	transports := make([]models.Transport, 0)
	for rows.Next() {
		var transport models.Transport
		err := rows.Scan(
			&transport.ID,
			&transport.PlateNo,
			&transport.Brand,
			&transport.Model,
			&transport.CapacityL,
			&transport.CurrentDriverID,
			&transport.CurrentEquipmentID,
			&transport.Status,
			&transport.CreatedAt,
			&transport.UpdatedAt,
			&transport.DeletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan transport: %w", err)
		}
		transports = append(transports, transport)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over transport rows: %w", err)
	}

	// Convert to response format
	responses := make([]models.TransportResponse, len(transports))
	for i := range transports {
		responses[i] = transports[i].ToResponse()
	}

	return &models.TransportListResponse{
		Items:    responses,
		Page:     req.Page,
		PageSize: req.PageSize,
		Total:    total,
	}, nil
}

// ExistsByPlateNo checks if a transport exists with the given plate number (excluding soft-deleted)
func (r *transportRepository) ExistsByPlateNo(ctx context.Context, plateNo string, excludeID *uuid.UUID) (bool, error) {
	query := "SELECT EXISTS(SELECT 1 FROM transport WHERE plate_no = $1" + DeletedAtFilter
	args := []interface{}{plateNo}
	argIndex := 2

	if excludeID != nil {
		query += fmt.Sprintf(" AND id != $%d", argIndex)
		args = append(args, *excludeID)
	}
	query += ")"

	var exists bool
	err := r.pool.QueryRow(ctx, query, args...).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check transport existence by plate number: %w", err)
	}

	return exists, nil
}

// HasActiveDriver checks if transport has an assigned driver
func (r *transportRepository) HasActiveDriver(ctx context.Context, transportID uuid.UUID) (bool, error) {
	query := "SELECT EXISTS(SELECT 1 FROM transport WHERE id = $1 AND current_driver_id IS NOT NULL" + DeletedAtFilter + ")"

	var exists bool
	err := r.pool.QueryRow(ctx, query, transportID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check if transport has active driver: %w", err)
	}

	return exists, nil
}

// HasActiveEquipment checks if transport has assigned equipment
func (r *transportRepository) HasActiveEquipment(ctx context.Context, transportID uuid.UUID) (bool, error) {
	query := "SELECT EXISTS(SELECT 1 FROM transport WHERE id = $1 AND current_equipment_id IS NOT NULL" + DeletedAtFilter + ")"

	var exists bool
	err := r.pool.QueryRow(ctx, query, transportID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check if transport has active equipment: %w", err)
	}

	return exists, nil
}

// HasActiveOrders checks if transport has active orders (DRAFT, SCHEDULED, IN_PROGRESS)
func (r *transportRepository) HasActiveOrders(ctx context.Context, transportID uuid.UUID) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM orders 
			WHERE transport_id = $1 
			AND status IN ('DRAFT', 'SCHEDULED', 'IN_PROGRESS') 
			AND deleted_at IS NULL
		)`

	var exists bool
	err := r.pool.QueryRow(ctx, query, transportID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check if transport has active orders: %w", err)
	}

	return exists, nil
}

// AssignDriver assigns a driver to transport
func (r *transportRepository) AssignDriver(ctx context.Context, transportID, driverID uuid.UUID) error {
	query := "UPDATE transport SET current_driver_id = $1, updated_at = NOW() WHERE id = $2 AND deleted_at IS NULL"

	result, err := r.pool.Exec(ctx, query, driverID, transportID)
	if err != nil {
		return fmt.Errorf("failed to assign driver to transport: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("transport not found or already deleted")
	}

	return nil
}

// UnassignDriver removes driver assignment from transport
func (r *transportRepository) UnassignDriver(ctx context.Context, transportID uuid.UUID) error {
	query := "UPDATE transport SET current_driver_id = NULL, updated_at = NOW() WHERE id = $1 AND deleted_at IS NULL"

	result, err := r.pool.Exec(ctx, query, transportID)
	if err != nil {
		return fmt.Errorf("failed to unassign driver from transport: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("transport not found or already deleted")
	}

	return nil
}

// AssignEquipment assigns equipment to transport
func (r *transportRepository) AssignEquipment(ctx context.Context, transportID, equipmentID uuid.UUID) error {
	// Start a transaction to update both tables
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil {
			// Log the rollback error but don't return it since we're already returning an error
			_ = err // explicitly ignore the error
		}
	}()

	// Update transport table
	transportQuery := "UPDATE transport SET current_equipment_id = $1, updated_at = NOW() WHERE id = $2 AND deleted_at IS NULL"
	result, err := tx.Exec(ctx, transportQuery, equipmentID, transportID)
	if err != nil {
		return fmt.Errorf("failed to update transport: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("transport not found or already deleted")
	}

	// Update equipment table to set transport_id
	equipmentQuery := "UPDATE equipment SET transport_id = $1, client_object_id = NULL, " +
		"warehouse_id = NULL, updated_at = NOW() WHERE id = $2 AND deleted_at IS NULL"
	result, err = tx.Exec(ctx, equipmentQuery, transportID, equipmentID)
	if err != nil {
		return fmt.Errorf("failed to update equipment: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("equipment not found or already deleted")
	}

	// Commit transaction
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// UnassignEquipment removes equipment assignment from transport
func (r *transportRepository) UnassignEquipment(ctx context.Context, transportID uuid.UUID) error {
	// Start a transaction to update both tables
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil {
			// Log the rollback error but don't return it since we're already returning an error
			_ = err // explicitly ignore the error
		}
	}()

	// Get the equipment ID that was assigned to this transport
	var equipmentID *uuid.UUID
	equipmentQuery := "SELECT current_equipment_id FROM transport WHERE id = $1 AND deleted_at IS NULL"
	err = tx.QueryRow(ctx, equipmentQuery, transportID).Scan(&equipmentID)
	if err != nil {
		return fmt.Errorf("failed to get equipment ID from transport: %w", err)
	}

	if equipmentID == nil {
		return fmt.Errorf("no equipment assigned to transport")
	}

	// Update transport table
	transportQuery := "UPDATE transport SET current_equipment_id = NULL, updated_at = NOW() WHERE id = $1 AND deleted_at IS NULL"
	result, err := tx.Exec(ctx, transportQuery, transportID)
	if err != nil {
		return fmt.Errorf("failed to update transport: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("transport not found or already deleted")
	}

	// Update equipment table to clear transport_id
	equipmentUpdateQuery := "UPDATE equipment SET transport_id = NULL, updated_at = NOW() WHERE id = $1 AND deleted_at IS NULL"
	result, err = tx.Exec(ctx, equipmentUpdateQuery, *equipmentID)
	if err != nil {
		return fmt.Errorf("failed to update equipment: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("equipment not found or already deleted")
	}

	// Commit transaction
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetAvailable returns transport with status IN_WORK and no soft-delete
func (r *transportRepository) GetAvailable(ctx context.Context, req models.TransportListRequest) (*models.TransportListResponse, error) {
	// Build base query for available transport
	baseQuery := "FROM transport WHERE status = 'IN_WORK' AND deleted_at IS NULL"
	args := []interface{}{}

	// Count total available transport
	countQuery := "SELECT COUNT(*) " + baseQuery
	var total int64
	err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("failed to count available transport: %w", err)
	}

	// Get available transport with pagination
	query := `
		SELECT id, plate_no, brand, model, capacity_l, current_driver_id, current_equipment_id, status, created_at, updated_at, deleted_at
		` + baseQuery + `
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2`

	offset := (req.Page - 1) * req.PageSize
	args = append(args, req.PageSize, offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query available transport: %w", err)
	}
	defer rows.Close()

	transports := make([]models.Transport, 0)
	for rows.Next() {
		var transport models.Transport
		err := rows.Scan(
			&transport.ID,
			&transport.PlateNo,
			&transport.Brand,
			&transport.Model,
			&transport.CapacityL,
			&transport.CurrentDriverID,
			&transport.CurrentEquipmentID,
			&transport.Status,
			&transport.CreatedAt,
			&transport.UpdatedAt,
			&transport.DeletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan transport: %w", err)
		}
		transports = append(transports, transport)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over transport rows: %w", err)
	}

	// Convert to response format
	responses := make([]models.TransportResponse, len(transports))
	for i := range transports {
		responses[i] = transports[i].ToResponse()
	}

	return &models.TransportListResponse{
		Items:    responses,
		Page:     req.Page,
		PageSize: req.PageSize,
		Total:    total,
	}, nil
}

// IsDriverAssignedToOtherTransport checks if driver is assigned to another non-deleted transport
func (r *transportRepository) IsDriverAssignedToOtherTransport(ctx context.Context, driverID, excludeTransportID uuid.UUID) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM transport 
			WHERE current_driver_id = $1 
			AND id != $2 
			AND deleted_at IS NULL
		)`

	var exists bool
	err := r.pool.QueryRow(ctx, query, driverID, excludeTransportID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check if driver is assigned to other transport: %w", err)
	}

	return exists, nil
}

// IsEquipmentAssignedToOtherTransport checks if equipment is assigned to another non-deleted transport
func (r *transportRepository) IsEquipmentAssignedToOtherTransport(ctx context.Context,
	equipmentID, excludeTransportID uuid.UUID) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM transport 
			WHERE current_equipment_id = $1 
			AND id != $2 
			AND deleted_at IS NULL
		)`

	var exists bool
	err := r.pool.QueryRow(ctx, query, equipmentID, excludeTransportID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check if equipment is assigned to other transport: %w", err)
	}

	return exists, nil
}

// IsEquipmentAvailableForAssignment checks if equipment can be assigned (no client_object_id, warehouse_id, or transport_id)
func (r *transportRepository) IsEquipmentAvailableForAssignment(ctx context.Context, equipmentID uuid.UUID) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM equipment 
			WHERE id = $1 
			AND client_object_id IS NULL 
			AND warehouse_id IS NULL 
			AND transport_id IS NULL
			AND deleted_at IS NULL
		)`

	var exists bool
	err := r.pool.QueryRow(ctx, query, equipmentID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check if equipment is available for assignment: %w", err)
	}

	return exists, nil
}
