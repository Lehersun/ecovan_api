-- Initial schema for Eco Van API (normalized, soft delete ready)
-- Migration: 001_initial_schema

-- Extensions
CREATE EXTENSION IF NOT EXISTS citext;
CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- =========================================
-- Users (Authentication & Authorization)
-- =========================================
CREATE TABLE IF NOT EXISTS users (
  id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  email         CITEXT NOT NULL,
  password_hash TEXT NOT NULL,
  role          TEXT NOT NULL CHECK (role IN ('ADMIN','DISPATCHER','DRIVER','VIEWER')),
  created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
  deleted_at    TIMESTAMPTZ
);

-- Unique email among non-deleted users
CREATE UNIQUE INDEX IF NOT EXISTS uniq_users_email_not_deleted
  ON users(email) WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_users_role ON users(role) WHERE deleted_at IS NULL;

-- =========================================
-- Clients
-- =========================================
CREATE TABLE IF NOT EXISTS clients (
  id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name       TEXT NOT NULL,
  tax_id     TEXT,
  email      CITEXT,
  phone      TEXT,
  notes      TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  deleted_at TIMESTAMPTZ
);

-- Unique among non-deleted (business name)
CREATE UNIQUE INDEX IF NOT EXISTS uniq_clients_name_not_deleted
  ON clients(name) WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_clients_email ON clients(email);
CREATE INDEX IF NOT EXISTS idx_clients_phone ON clients(phone);

-- =========================================
-- Warehouses (1:N with equipment)
-- =========================================
CREATE TABLE IF NOT EXISTS warehouses (
  id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name       TEXT NOT NULL,
  address    TEXT,
  notes      TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  deleted_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX IF NOT EXISTS uniq_warehouse_name_not_deleted
  ON warehouses(name) WHERE deleted_at IS NULL;

-- =========================================
-- Client Objects (nested under clients)
-- =========================================
CREATE TABLE IF NOT EXISTS client_objects (
  id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  client_id  UUID NOT NULL REFERENCES clients(id),
  name       TEXT NOT NULL,
  address    TEXT NOT NULL,
  geo_lat    NUMERIC(9,6),
  geo_lng    NUMERIC(9,6),
  notes      TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  deleted_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_client_objects_client
  ON client_objects(client_id) WHERE deleted_at IS NULL;

-- NOTE: NO direct equipment_id here. Equipment references object, not vice versa.

-- =========================================
-- Equipment (1:N placement: client_object OR warehouse)
-- =========================================
CREATE TABLE IF NOT EXISTS equipment (
  id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  number           TEXT, -- optional inventory number
  type             TEXT NOT NULL CHECK (type IN ('BIN','CONTAINER')),
  volume_l         INTEGER NOT NULL CHECK (volume_l > 0),
  condition        TEXT NOT NULL CHECK (condition IN ('GOOD','DAMAGED','OUT_OF_SERVICE')),
  photo            TEXT,
  -- placement (mutually exclusive) when NOT on transport
  client_object_id UUID REFERENCES client_objects(id) ON DELETE SET NULL,
  warehouse_id     UUID REFERENCES warehouses(id)     ON DELETE SET NULL,
  created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
  deleted_at       TIMESTAMPTZ
);

-- Exclusivity: equipment is at client_object OR warehouse (or nowhere)
DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_constraint WHERE conname = 'equipment_single_location'
  ) THEN
    ALTER TABLE equipment
      ADD CONSTRAINT equipment_single_location
      CHECK (client_object_id IS NULL OR warehouse_id IS NULL);
  END IF;
END$$;

CREATE INDEX IF NOT EXISTS idx_equipment_type ON equipment(type) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_equipment_client_object ON equipment(client_object_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_equipment_warehouse     ON equipment(warehouse_id)     WHERE deleted_at IS NULL;

-- =========================================
-- Drivers (no is_available column; availability computed by NOT EXISTS transport link)
-- =========================================
CREATE TABLE IF NOT EXISTS drivers (
  id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  full_name     TEXT NOT NULL,
  phone         TEXT,
  license_no    TEXT NOT NULL,
  license_class TEXT NOT NULL,
  photo         TEXT,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
  deleted_at    TIMESTAMPTZ
);

-- License unique among non-deleted
CREATE UNIQUE INDEX IF NOT EXISTS uniq_drivers_license_not_deleted
  ON drivers(license_no) WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_drivers_name    ON drivers(full_name);
CREATE INDEX IF NOT EXISTS idx_drivers_phone   ON drivers(phone);
CREATE INDEX IF NOT EXISTS idx_drivers_license ON drivers(license_no);

-- =========================================
-- Transport (1:1 links to driver and equipment; availability via status)
-- =========================================
CREATE TABLE IF NOT EXISTS transport (
  id                    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  plate_no              TEXT NOT NULL,
  capacity_l            INTEGER NOT NULL CHECK (capacity_l > 0),
  current_driver_id     UUID REFERENCES drivers(id),
  current_equipment_id  UUID REFERENCES equipment(id),
  status                TEXT NOT NULL DEFAULT 'IN_WORK' CHECK (status IN ('IN_WORK','REPAIR')),
  created_at            TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at            TIMESTAMPTZ NOT NULL DEFAULT now(),
  deleted_at            TIMESTAMPTZ
);

-- Unique plate among non-deleted transports
CREATE UNIQUE INDEX IF NOT EXISTS uniq_transport_plate_not_deleted
  ON transport(plate_no) WHERE deleted_at IS NULL;

-- Enforce 1:1 for links among non-deleted transports
CREATE UNIQUE INDEX IF NOT EXISTS uniq_transport_current_driver_not_deleted
  ON transport(current_driver_id)
  WHERE deleted_at IS NULL AND current_driver_id IS NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS uniq_transport_current_equipment_not_deleted
  ON transport(current_equipment_id)
  WHERE deleted_at IS NULL AND current_equipment_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_transport_status ON transport(status) WHERE deleted_at IS NULL;

-- =========================================
-- Orders (state machine fields; guard-friendly indexes)
-- =========================================
CREATE TABLE IF NOT EXISTS orders (
  id                    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  client_id             UUID NOT NULL REFERENCES clients(id),
  object_id             UUID NOT NULL REFERENCES client_objects(id),
  scheduled_date        DATE NOT NULL,
  scheduled_window_from TIME,
  scheduled_window_to   TIME,
  status                TEXT NOT NULL CHECK (status IN ('DRAFT','SCHEDULED','IN_PROGRESS','COMPLETED','CANCELED')),
  transport_id          UUID REFERENCES transport(id) ON DELETE SET NULL,
  notes                 TEXT,
  created_by            UUID,
  created_at            TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at            TIMESTAMPTZ NOT NULL DEFAULT now(),
  deleted_at            TIMESTAMPTZ
);

-- Filters and guards
CREATE INDEX IF NOT EXISTS idx_orders_status_date
  ON orders(status, scheduled_date) WHERE deleted_at IS NULL;

-- Active orders (unfinished) for guarded deletes
CREATE INDEX IF NOT EXISTS idx_orders_active_transport
  ON orders(transport_id)
  WHERE deleted_at IS NULL AND status IN ('DRAFT','SCHEDULED','IN_PROGRESS');

CREATE INDEX IF NOT EXISTS idx_orders_active_object
  ON orders(object_id)
  WHERE deleted_at IS NULL AND status IN ('DRAFT','SCHEDULED','IN_PROGRESS');

-- =========================================
-- Photos (file references)
-- =========================================
CREATE TABLE IF NOT EXISTS photos (
  id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  entity_type   TEXT NOT NULL CHECK (entity_type IN ('clients','client_objects','equipment','transport','drivers','orders')),
  entity_id     UUID NOT NULL,
  filename      TEXT NOT NULL,
  original_name TEXT NOT NULL,
  mime_type     TEXT NOT NULL,
  size          INTEGER NOT NULL CHECK (size >= 0),
  created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_photos_entity ON photos(entity_type, entity_id);

-- =========================================
-- updated_at triggers
-- =========================================
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = now();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'trg_users_updated_at') THEN
    CREATE TRIGGER trg_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
  END IF;

  IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'trg_clients_updated_at') THEN
    CREATE TRIGGER trg_clients_updated_at BEFORE UPDATE ON clients
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
  END IF;

  IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'trg_warehouses_updated_at') THEN
    CREATE TRIGGER trg_warehouses_updated_at BEFORE UPDATE ON warehouses
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
  END IF;

  IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'trg_client_objects_updated_at') THEN
    CREATE TRIGGER trg_client_objects_updated_at BEFORE UPDATE ON client_objects
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
  END IF;

  IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'trg_equipment_updated_at') THEN
    CREATE TRIGGER trg_equipment_updated_at BEFORE UPDATE ON equipment
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
  END IF;

  IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'trg_drivers_updated_at') THEN
    CREATE TRIGGER trg_drivers_updated_at BEFORE UPDATE ON drivers
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
  END IF;

  IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'trg_transport_updated_at') THEN
    CREATE TRIGGER trg_transport_updated_at BEFORE UPDATE ON transport
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
  END IF;

  IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'trg_orders_updated_at') THEN
    CREATE TRIGGER trg_orders_updated_at BEFORE UPDATE ON orders
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
  END IF;
END$$;
