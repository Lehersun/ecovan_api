-- Migration: 002_equipment_transport_assignment
-- Add transport_id to equipment table and update constraints

-- Add transport_id column to equipment table
ALTER TABLE equipment ADD COLUMN IF NOT EXISTS transport_id UUID REFERENCES transport(id) ON DELETE SET NULL;

-- Drop the old constraint that only allowed client_object_id OR warehouse_id
ALTER TABLE equipment DROP CONSTRAINT IF EXISTS equipment_single_location;

-- Add new constraint that ensures exactly one of transport_id, client_object_id, or warehouse_id is set
ALTER TABLE equipment ADD CONSTRAINT equipment_single_assignment
  CHECK (
    (transport_id IS NOT NULL AND client_object_id IS NULL AND warehouse_id IS NULL) OR
    (transport_id IS NULL AND client_object_id IS NOT NULL AND warehouse_id IS NULL) OR
    (transport_id IS NULL AND client_object_id IS NULL AND warehouse_id IS NOT NULL)
  );

-- Add index for transport_id
CREATE INDEX IF NOT EXISTS idx_equipment_transport ON equipment(transport_id) WHERE deleted_at IS NULL;

-- Update the comment to reflect the new structure
COMMENT ON TABLE equipment IS 'Equipment can be assigned to exactly one of: transport, client_object, or warehouse';
