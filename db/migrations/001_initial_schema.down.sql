-- Reverse migration: Drop all tables and schemas created in 001_initial_schema.up.sql

-- Drop tables in reverse order of creation (to handle foreign key dependencies)
DROP TABLE IF EXISTS transport CASCADE;
DROP TABLE IF EXISTS orders CASCADE;
DROP TABLE IF EXISTS equipment CASCADE;
DROP TABLE IF EXISTS drivers CASCADE;
DROP TABLE IF EXISTS client_objects CASCADE;
DROP TABLE IF EXISTS clients CASCADE;
DROP TABLE IF EXISTS warehouses CASCADE;
DROP TABLE IF EXISTS users CASCADE;

-- Drop custom types
DROP TYPE IF EXISTS equipment_type CASCADE;
DROP TYPE IF EXISTS equipment_condition CASCADE;
DROP TYPE IF EXISTS order_status CASCADE;
DROP TYPE IF EXISTS driver_status CASCADE;
DROP TYPE IF EXISTS transport_status CASCADE;
DROP TYPE IF EXISTS user_role CASCADE;

-- Drop extensions (if safe to do so)
-- Note: We don't drop uuid-ossp as it might be used by other schemas
