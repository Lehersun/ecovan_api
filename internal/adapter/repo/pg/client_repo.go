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

// clientRepository implements port.ClientRepository
type clientRepository struct {
	*BaseRepository
	db *pgxpool.Pool
}

// NewClientRepository creates a new client repository
func NewClientRepository(db *pgxpool.Pool) port.ClientRepository {
	return &clientRepository{
		BaseRepository: NewBaseRepository(db),
		db:             db,
	}
}

// Create creates a new client
func (r *clientRepository) Create(ctx context.Context, client *models.Client) error {
	query := `
		INSERT INTO clients (id, name, tax_id, email, phone, notes, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	now := time.Now()
	client.ID = uuid.New()
	client.CreatedAt = now
	client.UpdatedAt = now

	_, err := r.db.Exec(ctx, query,
		client.ID,
		client.Name,
		client.TaxID,
		client.Email,
		client.Phone,
		client.Notes,
		client.CreatedAt,
		client.UpdatedAt,
	)

	return err
}

// GetByID retrieves a client by ID, optionally including soft-deleted
//
//nolint:dupl // Similar pattern across repositories but with different models and fields
func (r *clientRepository) GetByID(ctx context.Context, id uuid.UUID, includeDeleted bool) (*models.Client, error) {
	query := `
		SELECT id, name, tax_id, email, phone, notes, created_at, updated_at, deleted_at
		FROM clients
		WHERE id = $1
	`

	if !includeDeleted {
		query += DeletedAtFilter
	}

	var client models.Client
	err := r.db.QueryRow(ctx, query, id).Scan(
		&client.ID,
		&client.Name,
		&client.TaxID,
		&client.Email,
		&client.Phone,
		&client.Notes,
		&client.CreatedAt,
		&client.UpdatedAt,
		&client.DeletedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get client: %w", err)
	}

	return &client, nil
}

// List retrieves clients with pagination and filtering
func (r *clientRepository) List(ctx context.Context, req models.ClientListRequest) (*models.ClientListResponse, error) {
	// Build the WHERE clause
	whereClause := "WHERE 1=1"
	args := []interface{}{}

	// Add search filter
	if req.Query != "" {
		searchPattern := "%" + req.Query + "%"
		whereClause += `
			AND (
				name ILIKE $1 OR
				email ILIKE $1 OR
				tax_id ILIKE $1
			)`
		args = append(args, searchPattern)
	}

	// Add soft-delete filter
	if !req.IncludeDeleted {
		whereClause += DeletedAtFilter
	}

	// Count total
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM clients %s", whereClause)
	var total int64
	err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("failed to count clients: %w", err)
	}

	// Calculate pagination
	limit := req.PageSize
	offset := (req.Page - 1) * req.PageSize

	// Build the main query with pagination
	limitParamNum := len(args) + 1
	const offsetIncrement = 2
	offsetParamNum := len(args) + offsetIncrement
	mainQuery := fmt.Sprintf(`
		SELECT id, name, tax_id, email, phone, notes, created_at, updated_at, deleted_at
		FROM clients
		%s
		ORDER BY name
		LIMIT $%d
		OFFSET $%d
	`, whereClause, limitParamNum, offsetParamNum)

	args = append(args, limit, offset)

	rows, err := r.db.Query(ctx, mainQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list clients: %w", err)
	}
	defer rows.Close()

	var clients []models.Client
	for rows.Next() {
		var client models.Client
		err := rows.Scan(
			&client.ID,
			&client.Name,
			&client.TaxID,
			&client.Email,
			&client.Phone,
			&client.Notes,
			&client.CreatedAt,
			&client.UpdatedAt,
			&client.DeletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan client: %w", err)
		}
		clients = append(clients, client)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over clients: %w", err)
	}

	return &models.ClientListResponse{
		Items:    clients,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

// Update updates an existing client
func (r *clientRepository) Update(ctx context.Context, client *models.Client) error {
	query := `
		UPDATE clients
		SET name = $1, tax_id = $2, email = $3, phone = $4, notes = $5, updated_at = $6
		WHERE id = $7
	`

	client.UpdatedAt = time.Now()
	_, err := r.db.Exec(ctx, query,
		client.Name,
		client.TaxID,
		client.Email,
		client.Phone,
		client.Notes,
		client.UpdatedAt,
		client.ID,
	)

	return err
}

// SoftDelete marks a client as deleted by setting deleted_at
func (r *clientRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	return r.SoftDeleteGeneric(ctx, "clients", id)
}

// Restore restores a soft-deleted client by clearing deleted_at
func (r *clientRepository) Restore(ctx context.Context, id uuid.UUID) error {
	return r.RestoreGeneric(ctx, "clients", id)
}

// ExistsByName checks if a client exists with the given name (excluding soft-deleted)
func (r *clientRepository) ExistsByName(ctx context.Context, name string, excludeID *uuid.UUID) (bool, error) {
	query := "SELECT EXISTS(SELECT 1 FROM clients WHERE name = $1 AND deleted_at IS NULL"
	args := []interface{}{name}

	if excludeID != nil {
		query += " AND id != $2"
		args = append(args, *excludeID)
	}
	query += ")"

	var exists bool
	err := r.db.QueryRow(ctx, query, args...).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check client existence: %w", err)
	}

	return exists, nil
}
