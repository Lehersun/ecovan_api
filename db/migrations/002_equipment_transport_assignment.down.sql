-- Reverse migration: Remove transport_id from equipment table and revert to old constraint

-- Drop the new constraint
ALTER TABLE equipment DROP CONSTRAINT IF EXISTS equipment_single_assignment;

-- Drop the index
DROP INDEX IF EXISTS idx_equipment_transport;

-- Drop the transport_id column (this will also drop the foreign key constraint)
ALTER TABLE equipment DROP COLUMN IF EXISTS transport_id;

-- Add back the old constraint that only allowed client_object_id OR warehouse_id
ALTER TABLE equipment ADD CONSTRAINT equipment_single_location
  CHECK (
    (client_object_id IS NOT NULL AND warehouse_id IS NULL) OR
    (client_object_id IS NULL AND warehouse_id IS NOT NULL) OR
    (client_object_id IS NULL AND warehouse_id IS NULL)
  );
