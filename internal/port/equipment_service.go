package port

import (
	"eco-van-api/internal/models"
)

// EquipmentService defines the interface for equipment business logic
type EquipmentService interface {
	BaseService[
		models.EquipmentResponse,
		models.CreateEquipmentRequest,
		models.UpdateEquipmentRequest,
		models.EquipmentListRequest,
		models.EquipmentListResponse,
	]
}
