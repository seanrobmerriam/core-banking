-- Drop triggers
DROP TRIGGER IF EXISTS update_customer_documents_updated_at ON customer_documents;
DROP TRIGGER IF EXISTS update_addresses_updated_at ON addresses;
DROP TRIGGER IF EXISTS update_customers_updated_at ON customers;

-- Drop functions
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop tables
DROP TABLE IF EXISTS customer_documents;
DROP TABLE IF EXISTS addresses;
DROP TABLE IF EXISTS customers;

-- Drop types
DROP TYPE IF EXISTS verification_status;
DROP TYPE IF EXISTS document_type;
DROP TYPE IF EXISTS address_type;
DROP TYPE IF EXISTS customer_status;
