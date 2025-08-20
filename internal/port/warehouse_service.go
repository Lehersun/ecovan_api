package port

import (
	"eco-van-api/internal/models"
)

// WarehouseService defines the interface for warehouse business logic
type WarehouseService interface {
	BaseService[
		models.WarehouseResponse,
		models.CreateWarehouseRequest,
		models.UpdateWarehouseRequest,
		models.WarehouseListRequest,
		models.WarehouseListResponse,
	]
}
