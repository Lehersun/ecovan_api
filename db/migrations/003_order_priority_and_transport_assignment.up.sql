-- Add priority field to orders table
ALTER TABLE orders ADD COLUMN IF NOT EXISTS priority TEXT NOT NULL DEFAULT 'MEDIUM' CHECK (priority IN ('LOW', 'MEDIUM', 'HIGH'));

-- Update status transitions to allow IN_PROGRESS -> COMPLETED, CANCELED
-- This is already handled in the application logic, but we can add a comment
COMMENT ON TABLE orders IS 'Orders with priority levels and transport assignment support. Status transitions: DRAFT->SCHEDULED/CANCELED, SCHEDULED->IN_PROGRESS/CANCELED, IN_PROGRESS->COMPLETED/CANCELED, COMPLETED/CANCELED->no further transitions';

-- Add index for priority filtering
CREATE INDEX IF NOT EXISTS idx_orders_priority ON orders(priority) WHERE deleted_at IS NULL;

-- Add index for priority + status filtering
CREATE INDEX IF NOT EXISTS idx_orders_priority_status ON orders(priority, status) WHERE deleted_at IS NULL;
