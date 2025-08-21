# Eco-Van API Backend Documentation

## Overview
The Eco-Van API is a comprehensive waste management and logistics system built with Go, featuring a layered architecture (Handler → Service → Repository), JWT authentication, and RBAC (Role-Based Access Control).

**API Response Guarantee:** All list endpoints consistently return empty arrays `[]` instead of `null` when no data is found, ensuring predictable response structures for frontend applications.

## Base URL
```
http://localhost:8080/api/v1
```

## Authentication
All protected endpoints require a valid JWT token in the Authorization header:
```
Authorization: Bearer <jwt_token>
```

## Available Endpoints

### 1. Health & Monitoring
#### GET `/healthz`
- **Description:** Application health check
- **Authentication:** None required
- **Response:** 200 OK with health status

#### GET `/metrics`
- **Description:** Application metrics (Prometheus format)
- **Authentication:** None required
- **Response:** 200 OK with metrics data

### 2. API Documentation
#### GET `/api/v1/docs`
- **Description:** OpenAPI specification in YAML format
- **Authentication:** None required
- **Response:** 200 OK with OpenAPI 3.0 specification
- **Use Case:** For API client generation, integration tools, and programmatic access

#### GET `/api/v1/docs/ui`
- **Description:** Interactive Swagger UI for API exploration
- **Authentication:** None required
- **Response:** 200 OK with HTML page containing Swagger UI
- **Use Case:** Interactive API documentation and testing interface

### 3. Authentication
#### POST `/auth/login`
- **Description:** User authentication
- **Authentication:** None required
- **Request Body:**
```json
{
  "email": "admin@example.com",
  "password": "admin123456"
}
```
- **Response:** 200 OK with access token
```json
{
  "accessToken": "eyJhbGciOiJIUzI1NiIs...",
  "refreshToken": "eyJhbGciOiJIUzI1NiIs...",
  "expiresIn": 3600
}
```

#### POST `/auth/refresh`
- **Description:** Refresh access token
- **Authentication:** Valid refresh token required
- **Request Body:**
```json
{
  "refreshToken": "eyJhbGciOiJIUzI1NiIs..."
}
```
- **Response:** 200 OK with new access token

### 4. User Management
#### GET `/users`
- **Description:** List all users
- **Authentication:** Required (Admin only)
- **Query Parameters:**
  - `page` (int): Page number (default: 1)
  - `pageSize` (int): Items per page (default: 10, max: 100)
  - `includeDeleted` (bool): Include soft-deleted users
- **Response:** 200 OK with paginated user list

### 5. Client Management
#### GET `/clients`
- **Description:** List all clients
- **Authentication:** Required (Read access)
- **Query Parameters:**
  - `page` (int): Page number
  - `pageSize` (int): Items per page
  - `includeDeleted` (bool): Include soft-deleted clients
- **Response:** 200 OK with paginated client list

#### POST `/clients`
- **Description:** Create new client
- **Authentication:** Required (Write access - Admin/Dispatcher only)
- **Request Body:**
```json
{
  "name": "Acme Corporation",
  "email": "info@acme.com",
  "phone": "+1-555-0101",
  "taxId": "TAX123456",
  "notes": "Large manufacturing company"
}
```
- **Response:** 201 Created with client details

#### GET `/clients/{id}`
- **Description:** Get client by ID
- **Authentication:** Required (Read access)
- **Response:** 200 OK with client details

#### PUT `/clients/{id}`
- **Description:** Update client
- **Authentication:** Required (Write access - Admin/Dispatcher only)
- **Request Body:** Same as POST with optional fields
- **Response:** 200 OK with updated client

#### DELETE `/clients/{id}`
- **Description:** Soft delete client
- **Authentication:** Required (Write access - Admin/Dispatcher only)
- **Response:** 204 No Content

#### POST `/clients/{id}/restore`
- **Description:** Restore soft-deleted client
- **Authentication:** Required (Write access - Admin/Dispatcher only)
- **Response:** 200 OK with restored client

### 5. Client Object Management
#### GET `/clients/{clientId}/objects`
- **Description:** List client objects
- **Authentication:** Required (Read access)
- **Response:** 200 OK with client objects list

#### POST `/clients/{clientId}/objects`
- **Description:** Create client object
- **Authentication:** Required (Write access - Admin/Dispatcher only)
- **Request Body:**
```json
{
  "name": "Main Factory",
  "address": "123 Industrial Blvd, Manufacturing District"
}
```
- **Response:** 201 Created with object details

#### GET `/clients/{clientId}/objects/{objectId}`
- **Description:** Get client object by ID
- **Authentication:** Required (Read access)
- **Response:** 200 OK with object details

#### PUT `/clients/{clientId}/objects/{objectId}`
- **Description:** Update client object
- **Authentication:** Required (Write access - Admin/Dispatcher only)
- **Response:** 200 OK with updated object

#### DELETE `/clients/{clientId}/objects/{objectId}`
- **Description:** Soft delete client object
- **Authentication:** Required (Write access - Admin/Dispatcher only)
- **Response:** 204 No Content

#### POST `/clients/{clientId}/objects/{objectId}/restore`
- **Description:** Restore soft-deleted client object
- **Authentication:** Required (Write access - Admin/Dispatcher only)
- **Response:** 200 OK with restored object

### 6. Warehouse Management
#### GET `/warehouses`
- **Description:** List all warehouses
- **Authentication:** Required (Read access)
- **Query Parameters:**
  - `page` (int): Page number
  - `pageSize` (int): Items per page
  - `includeDeleted` (bool): Include soft-deleted warehouses
- **Response:** 200 OK with paginated warehouse list

#### POST `/warehouses`
- **Description:** Create new warehouse
- **Authentication:** Required (Write access - Admin/Dispatcher only)
- **Request Body:**
```json
{
  "name": "Central Warehouse",
  "address": "100 Logistics Plaza, Central District",
  "capacity": 5000
}
```
- **Response:** 201 Created with warehouse details

#### GET `/warehouses/{id}`
- **Description:** Get warehouse by ID
- **Authentication:** Required (Read access)
- **Response:** 200 OK with warehouse details

#### PUT `/warehouses/{id}`
- **Description:** Update warehouse
- **Authentication:** Required (Write access - Admin/Dispatcher only)
- **Response:** 200 OK with updated warehouse

#### DELETE `/warehouses/{id}`
- **Description:** Soft delete warehouse
- **Authentication:** Required (Write access - Admin/Dispatcher only)
- **Response:** 204 No Content

#### POST `/warehouses/{id}/restore`
- **Description:** Restore soft-deleted warehouse
- **Authentication:** Required (Write access - Admin/Dispatcher only)
- **Response:** 200 OK with restored warehouse

### 7. Equipment Management
#### GET `/equipment`
- **Description:** List all equipment
- **Authentication:** Required (Read access)
- **Query Parameters:**
  - `page` (int): Page number
  - `pageSize` (int): Items per page
  - `warehouseId` (uuid): Filter by warehouse
  - `type` (string): Filter by equipment type
  - `includeDeleted` (bool): Include soft-deleted equipment
- **Response:** 200 OK with paginated equipment list

#### POST `/equipment`
- **Description:** Create new equipment
- **Authentication:** Required (Write access - Admin/Dispatcher only)
- **Request Body:**
```json
{
  "type": "BIN",
  "number": "BIN001",
  "volumeL": 100,
  "condition": "GOOD",
  "warehouseId": "c2488d47-53eb-491d-89f4-6998d461ee9b"
}
```
- **Equipment Types:** BIN, CONTAINER, PALLET
- **Conditions:** GOOD, DAMAGED, OUT_OF_SERVICE
- **Response:** 201 Created with equipment details

#### GET `/equipment/{id}`
- **Description:** Get equipment by ID
- **Authentication:** Required (Read access)
- **Response:** 200 OK with equipment details

#### PUT `/equipment/{id}`
- **Description:** Update equipment
- **Authentication:** Required (Write access - Admin/Dispatcher only)
- **Response:** 200 OK with updated equipment

#### DELETE `/equipment/{id}`
- **Description:** Soft delete equipment
- **Authentication:** Required (Write access - Admin/Dispatcher only)
- **Response:** 204 No Content

#### POST `/equipment/{id}/restore`
- **Description:** Restore soft-deleted equipment
- **Authentication:** Required (Write access - Admin/Dispatcher only)
- **Response:** 200 OK with restored equipment

### 8. Driver Management
#### GET `/drivers`
- **Description:** List all drivers
- **Authentication:** Required (Read access)
- **Query Parameters:**
  - `page` (int): Page number
  - `pageSize` (int): Items per page
  - `includeDeleted` (bool): Include soft-deleted drivers
- **Response:** 200 OK with paginated driver list

#### POST `/drivers`
- **Description:** Create new driver
- **Authentication:** Required (Write access - Admin/Dispatcher only)
- **Request Body:**
```json
{
  "fullName": "John Smith",
  "phone": "+1-555-1001",
  "licenseNo": "DL123456",
  "licenseClass": "B",
  "experienceYears": 5
}
```
- **License Classes:** A, B, C, D, E
- **Response:** 201 Created with driver details

#### GET `/drivers/{id}`
- **Description:** Get driver by ID
- **Authentication:** Required (Read access)
- **Response:** 200 OK with driver details

#### PUT `/drivers/{id}`
- **Description:** Update driver
- **Authentication:** Required (Write access - Admin/Dispatcher only)
- **Response:** 200 OK with updated driver

#### DELETE `/drivers/{id}`
- **Description:** Soft delete driver
- **Authentication:** Required (Write access - Admin/Dispatcher only)
- **Response:** 204 No Content

#### POST `/drivers/{id}/restore`
- **Description:** Restore soft-deleted driver
- **Authentication:** Required (Write access - Admin/Dispatcher only)
- **Response:** 200 OK with restored driver

### 9. Transport Management
#### GET `/transport`
- **Description:** List all transport vehicles
- **Authentication:** Required (Read access)
- **Query Parameters:**
  - `page` (int): Page number
  - `pageSize` (int): Items per page
  - `status` (string): Filter by status
  - `includeDeleted` (bool): Include soft-deleted transport
- **Response:** 200 OK with paginated transport list

#### POST `/transport`
- **Description:** Create new transport vehicle
- **Authentication:** Required (Write access - Admin/Dispatcher only)
- **Request Body:**
```json
{
  "plateNo": "ABC123",
  "brand": "Mercedes",
  "model": "Sprinter",
  "capacityL": 2000,
  "status": "IN_WORK"
}
```
- **Required Fields:** plateNo, brand, model, capacityL
- **Brand & Model:** Vehicle manufacturer and model (1-50 characters each)
- **Statuses:** IN_WORK, REPAIR
- **Response:** 201 Created with transport details

#### GET `/transport/{id}`
- **Description:** Get transport by ID
- **Authentication:** Required (Read access)
- **Response:** 200 OK with transport details

#### PUT `/transport/{id}`
- **Description:** Update transport
- **Authentication:** Required (Write access - Admin/Dispatcher only)
- **Response:** 200 OK with updated transport

#### DELETE `/transport/{id}`
- **Description:** Soft delete transport
- **Authentication:** Required (Write access - Admin/Dispatcher only)
- **Response:** 204 No Content

#### POST `/transport/{id}/restore`
- **Description:** Restore soft-deleted transport
- **Authentication:** Required (Write access - Admin/Dispatcher only)
- **Response:** 200 OK with restored transport

#### POST `/transport/{id}/assign-driver`
- **Description:** Assign driver to transport
- **Authentication:** Required (Write access - Admin/Dispatcher only)
- **Request Body:**
```json
{
  "driverId": "f53f66d0-d0c5-4e8e-b93c-f2c1219288be"
}
```
- **Response:** 200 OK with updated transport

#### POST `/transport/{id}/assign-equipment`
- **Description:** Assign equipment to transport
- **Authentication:** Required (Write access - Admin/Dispatcher only)
- **Request Body:**
```json
{
  "equipmentId": "406b38a8-3525-4d62-a735-b12aedcf97a3"
}
```
- **Response:** 200 OK with updated transport

### 10. Order Management
#### GET `/orders`
- **Description:** List all orders
- **Authentication:** Required (Read access)
- **Query Parameters:**
  - `page` (int): Page number
  - `pageSize` (int): Items per page
  - `status` (string): Filter by status
  - `clientId` (uuid): Filter by client
  - `includeDeleted` (bool): Include soft-deleted orders
- **Response:** 200 OK with paginated order list

#### POST `/orders`
- **Description:** Create new order
- **Authentication:** Required (Write access - Admin/Dispatcher only)
- **Request Body:**
```json
{
  "clientId": "e98745d7-7826-40e4-a3b2-22f37e16dbf0",
  "objectId": "a2516239-d5bf-4fff-90f1-b29f10076148",
  "scheduledDate": "2025-08-25T08:00:00Z",
  "scheduledWindowFrom": "08:00",
  "scheduledWindowTo": "12:00",
  "notes": "Regular waste collection from main factory"
}
```
- **Response:** 201 Created with order details

#### GET `/orders/{id}`
- **Description:** Get order by ID
- **Authentication:** Required (Read access)
- **Response:** 200 OK with order details

#### PUT `/orders/{id}`
- **Description:** Update order
- **Authentication:** Required (Write access - Admin/Dispatcher only)
- **Response:** 200 OK with updated order

#### DELETE `/orders/{id}`
- **Description:** Soft delete order
- **Authentication:** Required (Write access - Admin/Dispatcher only)
- **Business Rules:** Only DRAFT or CANCELED orders can be deleted
- **Response:** 
  - 204 No Content (if deletion successful)
  - 409 Conflict (if order cannot be deleted with reason)

#### POST `/orders/{id}/restore`
- **Description:** Restore soft-deleted order
- **Authentication:** Required (Write access - Admin/Dispatcher only)
- **Response:** 200 OK with restored order

#### PUT `/orders/{id}/status`
- **Description:** Update order status
- **Authentication:** Required (Write access - Admin/Dispatcher only)
- **Request Body:**
```json
{
  "status": "SCHEDULED"
}
```
- **Order Statuses:** DRAFT, SCHEDULED, IN_PROGRESS, COMPLETED, CANCELED
- **Response:** 200 OK with updated order

## HTTP Status Codes

### Success Responses
- **200 OK:** Request successful, data returned
- **201 Created:** Resource created successfully
- **204 No Content:** Request successful, no content to return

### Client Error Responses
- **400 Bad Request:** Invalid request format or parameters
- **401 Unauthorized:** Authentication required or failed
- **403 Forbidden:** Insufficient permissions
- **404 Not Found:** Resource not found
- **409 Conflict:** Business logic violation (e.g., cannot delete order in certain status)
- **422 Unprocessable Entity:** Validation error

### Server Error Responses
- **500 Internal Server Error:** Unexpected server error

## Error Response Format
```json
{
  "type": "/errors/validation-error",
  "title": "Validation Error",
  "status": 422,
  "detail": "Validation failed"
}
```

## Business Logic Rules

### Order Management
- **Deletion Rules:** Only DRAFT or CANCELED orders can be deleted
- **Status Transitions:** Orders follow a specific workflow (DRAFT → SCHEDULED → IN_PROGRESS → COMPLETED)
- **Scheduling:** Orders require valid client and object references

### Equipment Management
- **Condition Tracking:** Equipment can be GOOD, DAMAGED, or OUT_OF_SERVICE
- **Warehouse Assignment:** Equipment must be assigned to a valid warehouse

### Transport Management
- **Vehicle Information:** Transport includes brand and model for identification
- **Driver Assignment:** Transport can be assigned to available drivers (nullable)
- **Equipment Assignment:** Transport can carry equipment for delivery (nullable)
- **Capacity:** Transport has specified capacity in liters

## Rate Limiting
Currently no rate limiting implemented.

## Pagination
All list endpoints support pagination with the following parameters:
- `page`: Page number (1-based)
- `pageSize`: Items per page (1-100)

Response includes:
```json
{
  "items": [...],
  "page": 1,
  "pageSize": 10,
  "total": 150
}
```

**Important:** The `items` field will always be an array, even when empty. It will never be `null` - if no data is found, it will return an empty array `[]`.

**Example of empty list response:**
```json
{
  "items": [],
  "page": 1,
  "pageSize": 10,
  "total": 0
}
```

## Soft Delete
All entities support soft delete functionality:
- DELETE operations mark records as deleted but don't remove them
- RESTORE endpoints can recover soft-deleted records
- `includeDeleted` parameter controls visibility in list operations

## Response Consistency
All API responses follow consistent patterns:
- **List responses:** Always return `items` as an array, never `null`
- **Empty results:** Return empty array `[]` when no data is found
- **Single item responses:** Return the item directly or `null` if not found
- **Error responses:** Always follow the standard error format

## Data Validation
All input data is validated using:
- Required field validation
- Data type validation
- Business rule validation
- Format validation (email, phone, etc.)

## Security Features
- JWT-based authentication
- Role-based access control (RBAC)
- Input validation and sanitization
- SQL injection protection via parameterized queries
- CORS configuration for web clients

## Development Notes
- Database: PostgreSQL with migrations
- ORM: SQLC for type-safe database operations
- Testing: Unit tests, integration tests, and HTTP tests
- Logging: Structured logging with different levels
- Configuration: Environment-based configuration management

## OpenAPI Specification
The API provides a complete OpenAPI 3.0 specification that can be accessed at:
- **Raw Specification:** `GET /api/v1/docs` - Returns YAML format
- **Interactive UI:** `GET /api/v1/docs/ui` - Returns Swagger UI

### Benefits of OpenAPI Specification
- **API Client Generation:** Generate client libraries in various languages
- **Integration Tools:** Use with tools like Postman, Insomnia, or custom integrations
- **Documentation:** Always up-to-date API documentation
- **Testing:** Interactive testing interface via Swagger UI
- **Contract Testing:** Validate API contracts in CI/CD pipelines

### OpenAPI Features
- Complete endpoint definitions with request/response schemas
- Authentication requirements and security schemes
- Request/response examples
- Error response definitions
- Data model schemas for all entities
- Query parameter documentation
- Path parameter validation rules
