# Database Fixtures

This directory contains sample data fixtures for the Eco Van API development and testing environment.

## Files

- `001_sample_data.sql` - Comprehensive sample data for all entities

## Sample Data Overview

### Users (Authentication)
- **Admin**: `admin@example.com` / `admin123456`
- **Dispatcher**: `dispatcher@example.com` / `dispatcher123456`
- **Driver**: `driver@example.com` / `driver123456`
- **Viewer**: `viewer@example.com` / `viewer123456`

### Clients
- ООО "ЭкоСервис" - Large client with regular orders
- ИП Иванов А.С. - Private entrepreneur
- ООО "Зеленый Мир" - Environmental company
- АО "Чистый Город" - Municipal enterprise

### Warehouses
- Центральный склад - Main company warehouse
- Северный склад - Northern districts warehouse
- Южный склад - Southern districts warehouse

### Client Objects
- Multiple locations for each client (offices, warehouses, stores)
- Includes geographic coordinates for mapping

### Drivers
- 4 drivers with different license classes
- Realistic Russian names and license numbers

### Equipment
- **11 pieces** of equipment in various states:
  - 5 assigned to client objects
  - 4 assigned to warehouses
  - 2 assigned to transport
- Mix of containers and bins
- Different conditions (GOOD, DAMAGED, OUT_OF_SERVICE)

### Transport
- **4 vehicles** with different capacities:
  - ГАЗ ГАЗель Next (1500L)
  - Ford Transit (1200L)
  - Mercedes-Benz Sprinter (2000L)
  - ГАЗ Соболь (800L)
- Different statuses (IN_WORK, REPAIR)

### Orders
- **4 sample orders** in various states:
  - SCHEDULED, DRAFT, IN_PROGRESS, COMPLETED
- Realistic scheduling with time windows

## Usage

### Development Environment
The fixtures are automatically loaded when running:
```bash
make db          # Development database
make test-db     # Test database
```

### Manual Loading
To manually load fixtures into an existing database:
```bash
psql -h localhost -U app -d eco_van_db -f db/fixtures/001_sample_data.sql
```

### Test Environment
For testing, fixtures are loaded automatically when running:
```bash
make test-integration
```

## Data Relationships

The fixtures create a realistic ecosystem:
- Clients have multiple objects
- Equipment is distributed across clients, warehouses, and transport
- Transport has drivers and equipment assigned
- Orders link clients, objects, and transport
- All entities maintain referential integrity

## Notes

- All UUIDs are fixed for consistent testing
- Russian company names and addresses for realism
- Equipment placement follows the "exactly one" rule
- Transport assignments respect capacity constraints
- Orders demonstrate the complete workflow

## Customization

To modify fixtures:
1. Edit `001_sample_data.sql`
2. Update this README if needed
3. Test with `make test-integration`
4. Reload development database with `make dev-reset`
