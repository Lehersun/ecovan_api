package port

import (
	"eco-van-api/internal/models"
)

// ClientService defines the interface for client business logic
type ClientService interface {
	BaseService[
		models.ClientResponse,
		models.CreateClientRequest,
		models.UpdateClientRequest,
		models.ClientListRequest,
		models.ClientListResponse,
	]
}
