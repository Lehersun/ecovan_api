-- Remove priority indexes
DROP INDEX IF EXISTS idx_orders_priority_status;
DROP INDEX IF EXISTS idx_orders_priority;

-- Remove priority column
ALTER TABLE orders DROP COLUMN IF EXISTS priority;

-- Remove comment
COMMENT ON TABLE orders IS NULL;
