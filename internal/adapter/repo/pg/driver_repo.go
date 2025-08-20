package pg

import (
	"context"
	"eco-van-api/internal/models"
	"eco-van-api/internal/port"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type driverRepository struct {
	*BaseRepository
	pool *pgxpool.Pool
}

// NewDriverRepository creates a new driver repository
func NewDriverRepository(pool *pgxpool.Pool) port.DriverRepository {
	return &driverRepository{
		BaseRepository: NewBaseRepository(pool),
		pool:           pool,
	}
}

// Create creates a new driver
func (r *driverRepository) Create(ctx context.Context, driver *models.Driver) error {
	query := `
		INSERT INTO drivers (id, full_name, phone, license_no, license_class, photo, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	now := time.Now()
	driver.ID = uuid.New()
	driver.CreatedAt = now
	driver.UpdatedAt = now

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
		query += DeletedAtFilter
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
	whereClause := r.BuildWhereClause(whereClauses)

	// Count total
	total, err := r.CountTotal(ctx, baseQuery, whereClause, args)
	if err != nil {
		return nil, fmt.Errorf("failed to count drivers: %w", err)
	}

	// Build pagination query
	query := fmt.Sprintf(`
		SELECT id, full_name, phone, license_no, license_class, photo, created_at, updated_at, deleted_at
		%s
		%s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, baseQuery, whereClause, argIndex, argIndex+1)

	// Calculate pagination
	limit := req.PageSize
	offset := (req.Page - 1) * req.PageSize
	args = append(args, limit, offset)

	rows, err := r.pool.Query(ctx, query, args...)
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

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over drivers: %w", err)
	}

	return &models.DriverListResponse{
		Items:    items,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

// Update updates an existing driver
func (r *driverRepository) Update(ctx context.Context, driver *models.Driver) error {
	query := `
		UPDATE drivers
		SET full_name = $1, phone = $2, license_no = $3, license_class = $4, photo = $5, updated_at = $6
		WHERE id = $7
	`

	driver.UpdatedAt = time.Now()
	_, err := r.pool.Exec(ctx, query,
		driver.FullName, driver.Phone, driver.LicenseNo,
		driver.LicenseClass, driver.Photo, driver.UpdatedAt, driver.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update driver: %w", err)
	}
	return nil
}

// SoftDelete marks driver as deleted by setting deleted_at
func (r *driverRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	return r.SoftDeleteGeneric(ctx, "drivers", id)
}

// Restore restores a soft-deleted driver by clearing deleted_at
func (r *driverRepository) Restore(ctx context.Context, id uuid.UUID) error {
	return r.RestoreGeneric(ctx, "drivers", id)
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
func (r *driverRepository) ListAvailable(ctx context.Context) ([]models.Driver, error) {
	query := `
		SELECT id, full_name, phone, license_no, license_class, photo, created_at, updated_at, deleted_at
		FROM drivers
		WHERE deleted_at IS NULL
		ORDER BY full_name
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list available drivers: %w", err)
	}
	defer rows.Close()

	var drivers []models.Driver
	for rows.Next() {
		var driver models.Driver
		err := rows.Scan(
			&driver.ID, &driver.FullName, &driver.Phone, &driver.LicenseNo,
			&driver.LicenseClass, &driver.Photo, &driver.CreatedAt, &driver.UpdatedAt, &driver.DeletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan driver: %w", err)
		}
		drivers = append(drivers, driver)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over available drivers: %w", err)
	}

	return drivers, nil
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
