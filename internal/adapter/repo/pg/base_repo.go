package pg

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// BaseRepository provides common CRUD operations for entities
type BaseRepository struct {
	db *pgxpool.Pool
}

// NewBaseRepository creates a new base repository
func NewBaseRepository(db *pgxpool.Pool) *BaseRepository {
	return &BaseRepository{db: db}
}

// GetByIDGeneric retrieves an entity by ID using a generic query builder
func (r *BaseRepository) GetByIDGeneric(
	ctx context.Context,
	tableName string,
	columns []string,
	id uuid.UUID,
	includeDeleted bool,
	scanFunc func(...interface{}) error,
) (*interface{}, error) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE id = $1", strings.Join(columns, ", "), tableName)

	if !includeDeleted {
		query += " AND deleted_at IS NULL"
	}

	var result interface{}
	err := r.db.QueryRow(ctx, query, id).Scan(scanFunc)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get %s: %w", tableName, err)
	}

	return &result, nil
}

// SoftDeleteGeneric marks an entity as deleted by setting deleted_at
func (r *BaseRepository) SoftDeleteGeneric(ctx context.Context, tableName string, id uuid.UUID) error {
	query := fmt.Sprintf("UPDATE %s SET deleted_at = $1 WHERE id = $2", tableName)

	now := time.Now()
	_, err := r.db.Exec(ctx, query, now, id)
	if err != nil {
		return fmt.Errorf("failed to soft delete %s: %w", tableName, err)
	}

	return nil
}

// RestoreGeneric restores a soft-deleted entity by clearing deleted_at
func (r *BaseRepository) RestoreGeneric(ctx context.Context, tableName string, id uuid.UUID) error {
	query := fmt.Sprintf("UPDATE %s SET deleted_at = NULL WHERE id = $1", tableName)

	_, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to restore %s: %w", tableName, err)
	}

	return nil
}

// BuildWhereClause builds a WHERE clause from conditions
func (r *BaseRepository) BuildWhereClause(conditions []string) string {
	if len(conditions) == 0 {
		return ""
	}
	return "WHERE " + strings.Join(conditions, " AND ")
}

// BuildPaginationQuery builds a pagination query with LIMIT and OFFSET
func (r *BaseRepository) BuildPaginationQuery(baseQuery, whereClause string, limit, offset int) string {
	query := baseQuery + " " + whereClause
	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}
	if offset > 0 {
		query += fmt.Sprintf(" OFFSET %d", offset)
	}
	return query
}

// CountTotal counts total records for pagination
func (r *BaseRepository) CountTotal(ctx context.Context, baseQuery, whereClause string, args []interface{}) (int64, error) {
	countQuery := fmt.Sprintf("SELECT COUNT(*) %s %s", baseQuery, whereClause)

	var total int64
	err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("failed to count total: %w", err)
	}

	return total, nil
}
