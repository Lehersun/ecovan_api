package pg

import (
	"context"
	"eco-van-api/internal/models"
	"eco-van-api/internal/port"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type driverRepository struct {
	pool *pgxpool.Pool
}

// NewDriverRepository creates a new driver repository
func NewDriverRepository(pool *pgxpool.Pool) port.DriverRepository {
	return &driverRepository{pool: pool}
}

// Create creates a new driver
func (r *driverRepository) Create(ctx context.Context, driver *models.Driver) error {
	query := `
		INSERT INTO drivers (id, full_name, phone, license_no, license_class, photo, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := r.pool.Exec(ctx, query,
		driver.ID, driver.FullName, driver.Phone, driver.LicenseNo,
		driver.LicenseClass, driver.Photo, driver.CreatedAt, driver.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create driver: %w", err)
	}
	return nil
}

// GetByID retrieves driver by ID, optionally including soft-deleted
func (r *driverRepository) GetByID(ctx context.Context, id uuid.UUID, includeDeleted bool) (*models.Driver, error) {
	query := `
		SELECT id, full_name, phone, license_no, license_class, photo, created_at, updated_at, deleted_at
		FROM drivers WHERE id = $1
	`
	if !includeDeleted {
		query += " AND deleted_at IS NULL"
	}

	var driver models.Driver
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&driver.ID, &driver.FullName, &driver.Phone, &driver.LicenseNo,
		&driver.LicenseClass, &driver.Photo, &driver.CreatedAt, &driver.UpdatedAt, &driver.DeletedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get driver: %w", err)
	}
	return &driver, nil
}

// List retrieves drivers with pagination and filtering
func (r *driverRepository) List(ctx context.Context, req models.DriverListRequest) (*models.DriverListResponse, error) {
	// Build base query
	baseQuery := "FROM drivers"
	whereClauses := []string{}
	args := []interface{}{}
	argIndex := 1

	// Add soft-delete filter
	if !req.IncludeDeleted {
		whereClauses = append(whereClauses, "deleted_at IS NULL")
	}

	// Add license class filter
	if req.LicenseClass != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("license_class = $%d", argIndex))
		args = append(args, *req.LicenseClass)
		argIndex++
	}

	// Add search query filter
	if req.Q != nil && *req.Q != "" {
		searchQuery := fmt.Sprintf("(full_name ILIKE $%d OR license_no ILIKE $%d)", argIndex, argIndex)
		whereClauses = append(whereClauses, searchQuery)
		args = append(args, "%"+*req.Q+"%")
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
		return nil, fmt.Errorf("failed to count drivers: %w", err)
	}

	// Build SELECT query with pagination
	selectQuery := fmt.Sprintf(`
		SELECT id, full_name, phone, license_no, license_class, photo, created_at, updated_at, deleted_at
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
		return nil, fmt.Errorf("failed to list drivers: %w", err)
	}
	defer rows.Close()

	var items []models.DriverResponse
	for rows.Next() {
		var driver models.Driver
		err := rows.Scan(
			&driver.ID, &driver.FullName, &driver.Phone, &driver.LicenseNo,
			&driver.LicenseClass, &driver.Photo, &driver.CreatedAt, &driver.UpdatedAt, &driver.DeletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan driver: %w", err)
		}
		items = append(items, driver.ToResponse())
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating driver rows: %w", err)
	}

	return &models.DriverListResponse{
		Items:    items,
		Page:     req.Page,
		PageSize: req.PageSize,
		Total:    total,
	}, nil
}

// Update updates an existing driver
func (r *driverRepository) Update(ctx context.Context, driver *models.Driver) error {
	query := `
		UPDATE drivers 
		SET full_name = $1, phone = $2, license_no = $3, license_class = $4, photo = $5, updated_at = $6
		WHERE id = $7
	`
	_, err := r.pool.Exec(ctx, query,
		driver.FullName, driver.Phone, driver.LicenseNo, driver.LicenseClass,
		driver.Photo, driver.UpdatedAt, driver.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update driver: %w", err)
	}
	return nil
}

// SoftDelete marks driver as deleted by setting deleted_at
func (r *driverRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE drivers SET deleted_at = now() WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to soft delete driver: %w", err)
	}
	return nil
}

// Restore restores a soft-deleted driver by clearing deleted_at
func (r *driverRepository) Restore(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE drivers SET deleted_at = NULL WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to restore driver: %w", err)
	}
	return nil
}

// IsAssignedToTransport checks if driver is currently assigned to a transport
func (r *driverRepository) IsAssignedToTransport(ctx context.Context, driverID uuid.UUID) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM transport 
			WHERE current_driver_id = $1 AND deleted_at IS NULL
		)
	`
	var exists bool
	err := r.pool.QueryRow(ctx, query, driverID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check driver transport assignment: %w", err)
	}
	return exists, nil
}

// ExistsByLicenseNo checks if driver exists with the given license number (excluding soft-deleted)
func (r *driverRepository) ExistsByLicenseNo(ctx context.Context, licenseNo string, excludeID *uuid.UUID) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM drivers WHERE license_no = $1 AND deleted_at IS NULL`
	args := []interface{}{licenseNo}
	argIndex := 2

	if excludeID != nil {
		query += fmt.Sprintf(" AND id != $%d", argIndex)
		args = append(args, *excludeID)
	}
	query += ")"

	var exists bool
	err := r.pool.QueryRow(ctx, query, args...).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check driver license existence: %w", err)
	}
	return exists, nil
}

// ListAvailable retrieves available drivers (not assigned to any transport)
func (r *driverRepository) ListAvailable(ctx context.Context, req models.DriverListRequest) (*models.DriverListResponse, error) {
	// Build base query for available drivers
	baseQuery := `
		FROM drivers d
		WHERE NOT EXISTS (
			SELECT 1 FROM transport t 
			WHERE t.current_driver_id = d.id AND t.deleted_at IS NULL
		)
	`
	whereClauses := []string{}
	args := []interface{}{}
	argIndex := 1

	// Add soft-delete filter
	if !req.IncludeDeleted {
		whereClauses = append(whereClauses, "d.deleted_at IS NULL")
	}

	// Add license class filter
	if req.LicenseClass != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("d.license_class = $%d", argIndex))
		args = append(args, *req.LicenseClass)
		argIndex++
	}

	// Add search query filter
	if req.Q != nil && *req.Q != "" {
		searchQuery := fmt.Sprintf("(d.full_name ILIKE $%d OR d.license_no ILIKE $%d)", argIndex, argIndex)
		whereClauses = append(whereClauses, searchQuery)
		args = append(args, "%"+*req.Q+"%")
		argIndex++
	}

	// Build WHERE clause
	whereClause := ""
	if len(whereClauses) > 0 {
		whereClause = "AND " + strings.Join(whereClauses, " AND ")
	}

	// Count total
	countQuery := fmt.Sprintf("SELECT COUNT(*) %s %s", baseQuery, whereClause)
	var total int64
	err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("failed to count available drivers: %w", err)
	}

	// Build SELECT query with pagination
	selectQuery := fmt.Sprintf(`
		SELECT d.id, d.full_name, d.phone, d.license_no, d.license_class, d.photo, d.created_at, d.updated_at, d.deleted_at
		%s %s
		ORDER BY d.created_at DESC
		LIMIT $%d OFFSET $%d
	`, baseQuery, whereClause, argIndex, argIndex+1)

	// Add pagination args
	offset := (req.Page - 1) * req.PageSize
	args = append(args, req.PageSize, offset)

	// Execute query
	rows, err := r.pool.Query(ctx, selectQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list available drivers: %w", err)
	}
	defer rows.Close()

	var items []models.DriverResponse
	for rows.Next() {
		var driver models.Driver
		err := rows.Scan(
			&driver.ID, &driver.FullName, &driver.Phone, &driver.LicenseNo,
			&driver.LicenseClass, &driver.Photo, &driver.CreatedAt, &driver.UpdatedAt, &driver.DeletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan driver: %w", err)
		}
		items = append(items, driver.ToResponse())
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating driver rows: %w", err)
	}

	return &models.DriverListResponse{
		Items:    items,
		Page:     req.Page,
		PageSize: req.PageSize,
		Total:    total,
	}, nil
}
