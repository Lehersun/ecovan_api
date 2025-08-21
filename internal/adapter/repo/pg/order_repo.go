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

// orderRepository implements port.OrderRepository for PostgreSQL
type orderRepository struct {
	db *pgxpool.Pool
}

// NewOrderRepository creates a new order repository
func NewOrderRepository(db *pgxpool.Pool) port.OrderRepository {
	return &orderRepository{db: db}
}

// Create creates a new order
func (r *orderRepository) Create(ctx context.Context, order *models.Order) error {
	query := `
		INSERT INTO orders (
			client_id, object_id, scheduled_date, scheduled_window_from, 
			scheduled_window_to, status, transport_id, notes, created_by, 
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id, created_at, updated_at
	`

	now := time.Now()
	order.CreatedAt = now
	order.UpdatedAt = now

	err := r.db.QueryRow(ctx, query,
		order.ClientID,
		order.ObjectID,
		order.ScheduledDate,
		order.ScheduledWindowFrom,
		order.ScheduledWindowTo,
		order.Status,
		order.TransportID,
		order.Notes,
		order.CreatedBy,
		order.CreatedAt,
		order.UpdatedAt,
	).Scan(&order.ID, &order.CreatedAt, &order.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create order: %w", err)
	}

	return nil
}

// GetByID retrieves an order by ID
func (r *orderRepository) GetByID(ctx context.Context, id uuid.UUID, includeDeleted bool) (*models.Order, error) {
	query := `
		SELECT id, client_id, object_id, scheduled_date, scheduled_window_from,
		       scheduled_window_to, status, transport_id, notes, created_by,
		       created_at, updated_at, deleted_at
		FROM orders
		WHERE id = $1
	`

	if !includeDeleted {
		query += " AND deleted_at IS NULL"
	}

	var order models.Order
	err := r.db.QueryRow(ctx, query, id).Scan(
		&order.ID,
		&order.ClientID,
		&order.ObjectID,
		&order.ScheduledDate,
		&order.ScheduledWindowFrom,
		&order.ScheduledWindowTo,
		&order.Status,
		&order.TransportID,
		&order.Notes,
		&order.CreatedBy,
		&order.CreatedAt,
		&order.UpdatedAt,
		&order.DeletedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	return &order, nil
}

// Update updates an existing order
func (r *orderRepository) Update(ctx context.Context, order *models.Order) error {
	query := `
		UPDATE orders
		SET client_id = $1, object_id = $2, scheduled_date = $3, 
		    scheduled_window_from = $4, scheduled_window_to = $5, 
		    status = $6, transport_id = $7, notes = $8, updated_at = $9
		WHERE id = $10
	`

	order.UpdatedAt = time.Now()
	result, err := r.db.Exec(ctx, query,
		order.ClientID,
		order.ObjectID,
		order.ScheduledDate,
		order.ScheduledWindowFrom,
		order.ScheduledWindowTo,
		order.Status,
		order.TransportID,
		order.Notes,
		order.UpdatedAt,
		order.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update order: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("order not found")
	}

	return nil
}

// SoftDelete marks an order as deleted by setting deleted_at
func (r *orderRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	query := "UPDATE orders SET deleted_at = $1 WHERE id = $2"

	now := time.Now()
	result, err := r.db.Exec(ctx, query, now, id)
	if err != nil {
		return fmt.Errorf("failed to soft delete order: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("order not found")
	}

	return nil
}

// Restore restores a soft-deleted order by clearing deleted_at
func (r *orderRepository) Restore(ctx context.Context, id uuid.UUID) error {
	query := "UPDATE orders SET deleted_at = NULL WHERE id = $1"

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to restore order: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("order not found")
	}

	return nil
}

// List retrieves orders with pagination and filtering
func (r *orderRepository) List(ctx context.Context, req models.OrderListRequest) (*models.OrderListResponse, error) {
	// Build the WHERE clause
	whereClauses := []string{}
	args := []interface{}{}

	// Add soft-delete filter
	if !req.IncludeDeleted {
		whereClauses = append(whereClauses, "deleted_at IS NULL")
	}

	// Add status filter
	if req.Status != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("status = $%d", len(args)+1))
		args = append(args, string(*req.Status))
	}

	// Add date filter
	if req.Date != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("scheduled_date = $%d", len(args)+1))
		args = append(args, *req.Date)
	}

	// Add client filter
	if req.ClientID != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("client_id = $%d", len(args)+1))
		args = append(args, *req.ClientID)
	}

	// Add object filter
	if req.ObjectID != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("object_id = $%d", len(args)+1))
		args = append(args, *req.ObjectID)
	}

	// Build WHERE clause
	whereClause := ""
	if len(whereClauses) > 0 {
		whereClause = "WHERE " + strings.Join(whereClauses, " AND ")
	}

	// Count total
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM orders %s", whereClause)
	var total int64
	err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("failed to count orders: %w", err)
	}

	// Calculate pagination
	limit := req.PageSize
	offset := (req.Page - 1) * req.PageSize

	// Build the main query with pagination
	limitParamNum := len(args) + 1
	const offsetIncrement = 2
	offsetParamNum := len(args) + offsetIncrement
	mainQuery := fmt.Sprintf(`
		SELECT id, client_id, object_id, scheduled_date, scheduled_window_from,
		       scheduled_window_to, status, transport_id, notes, created_by,
		       created_at, updated_at, deleted_at
		FROM orders
		%s
		ORDER BY scheduled_date DESC, created_at DESC
		LIMIT $%d
		OFFSET $%d
	`, whereClause, limitParamNum, offsetParamNum)

	args = append(args, limit, offset)

	rows, err := r.db.Query(ctx, mainQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list orders: %w", err)
	}
	defer rows.Close()

	orders := make([]models.Order, 0)
	for rows.Next() {
		var order models.Order
		err := rows.Scan(
			&order.ID,
			&order.ClientID,
			&order.ObjectID,
			&order.ScheduledDate,
			&order.ScheduledWindowFrom,
			&order.ScheduledWindowTo,
			&order.Status,
			&order.TransportID,
			&order.Notes,
			&order.CreatedBy,
			&order.CreatedAt,
			&order.UpdatedAt,
			&order.DeletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan order: %w", err)
		}
		orders = append(orders, order)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over orders: %w", err)
	}

	// Convert to responses
	responses := make([]models.OrderResponse, len(orders))
	for i := range orders {
		responses[i] = orders[i].ToResponse()
	}

	return &models.OrderListResponse{
		Items:    responses,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

// ExistsByClientAndObject checks if an order exists for the given client and object
func (r *orderRepository) ExistsByClientAndObject(ctx context.Context, clientID, objectID uuid.UUID, excludeID *uuid.UUID) (bool, error) {
	query := "SELECT EXISTS(SELECT 1 FROM orders WHERE client_id = $1 AND object_id = $2 AND deleted_at IS NULL"
	args := []interface{}{clientID, objectID}

	if excludeID != nil {
		query += " AND id != $3"
		args = append(args, *excludeID)
	}
	query += ")"

	var exists bool
	err := r.db.QueryRow(ctx, query, args...).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check order existence: %w", err)
	}

	return exists, nil
}

// HasActiveOrders checks if a client object has any active orders
func (r *orderRepository) HasActiveOrders(ctx context.Context, objectID uuid.UUID) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM orders 
			WHERE object_id = $1 
			AND deleted_at IS NULL
			AND status IN ('DRAFT', 'SCHEDULED', 'IN_PROGRESS')
		)
	`

	var exists bool
	err := r.db.QueryRow(ctx, query, objectID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check active orders: %w", err)
	}

	return exists, nil
}

// GetActiveOrdersByObject returns active orders for a specific object
func (r *orderRepository) GetActiveOrdersByObject(ctx context.Context, objectID uuid.UUID) ([]models.Order, error) {
	query := `
		SELECT id, client_id, object_id, scheduled_date, scheduled_window_from,
		       scheduled_window_to, status, transport_id, notes, created_by,
		       created_at, updated_at, deleted_at
		FROM orders
		WHERE object_id = $1 
		AND deleted_at IS NULL
		AND status IN ('DRAFT', 'SCHEDULED', 'IN_PROGRESS')
		ORDER BY scheduled_date ASC, created_at ASC
	`

	rows, err := r.db.Query(ctx, query, objectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get active orders: %w", err)
	}
	defer rows.Close()

	orders := make([]models.Order, 0)
	for rows.Next() {
		var order models.Order
		err := rows.Scan(
			&order.ID,
			&order.ClientID,
			&order.ObjectID,
			&order.ScheduledDate,
			&order.ScheduledWindowFrom,
			&order.ScheduledWindowTo,
			&order.Status,
			&order.TransportID,
			&order.Notes,
			&order.CreatedBy,
			&order.CreatedAt,
			&order.UpdatedAt,
			&order.DeletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan order: %w", err)
		}
		orders = append(orders, order)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over orders: %w", err)
	}

	return orders, nil
}
