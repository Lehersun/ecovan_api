package http

// Common error messages used across handlers
const (
	// Entity not found messages
	ErrEquipmentNotFound    = "equipment not found"
	ErrClientNotFound       = "client not found"
	ErrClientObjectNotFound = "client object not found"
	ErrWarehouseNotFound    = "warehouse not found"
	ErrDriverNotFound       = "driver not found"
	ErrUserNotFound         = "user not found"

	// Query parameter values
	QueryParamIncludeDeleted = "true"

	// Common prefixes
	ErrClientObjectNamePrefix   = "client object with name"
	ErrCannotDeleteClientObject = "cannot delete client object"

	// Magic numbers
	MaxPageSize = 100
)
