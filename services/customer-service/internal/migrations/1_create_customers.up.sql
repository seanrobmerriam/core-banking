-- Create customers table
CREATE TYPE customer_status AS ENUM ('Pending', 'Active', 'Inactive', 'Suspended', 'Closed');

CREATE TABLE customers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    customer_number VARCHAR(50) NOT NULL UNIQUE,
    first_name VARCHAR(100) NOT NULL,
    middle_name VARCHAR(100),
    last_name VARCHAR(100) NOT NULL,
    date_of_birth DATE NOT NULL,
    tax_id TEXT NOT NULL, -- Encrypted field
    email VARCHAR(255) NOT NULL,
    phone VARCHAR(50) NOT NULL,
    status customer_status NOT NULL DEFAULT 'Pending',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    created_by UUID NOT NULL,
    updated_by UUID,
    version INTEGER NOT NULL DEFAULT 1
);

-- Create addresses table
CREATE TYPE address_type AS ENUM ('Physical', 'Mailing', 'Business');

CREATE TABLE addresses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    customer_id UUID NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
    address_type address_type NOT NULL DEFAULT 'Physical',
    street1 VARCHAR(255) NOT NULL,
    street2 VARCHAR(255),
    city VARCHAR(100) NOT NULL,
    state VARCHAR(100) NOT NULL,
    postal_code VARCHAR(20) NOT NULL,
    country VARCHAR(100) NOT NULL,
    is_primary BOOLEAN NOT NULL DEFAULT FALSE,
    valid_from TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    valid_to TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Create customer_documents table
CREATE TYPE document_type AS ENUM ('Passport', 'DriversLicense', 'NationalID', 'SSN', 'TaxID', 'UtilityBill', 'BankStatement');
CREATE TYPE verification_status AS ENUM ('Pending', 'Verified', 'Expired', 'Rejected');

CREATE TABLE customer_documents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    customer_id UUID NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
    document_type document_type NOT NULL,
    document_number TEXT NOT NULL, -- Encrypted field
    issuing_authority VARCHAR(255) NOT NULL,
    issuing_country VARCHAR(100) NOT NULL,
    issue_date DATE NOT NULL,
    expiry_date DATE NOT NULL,
    verification_status verification_status NOT NULL DEFAULT 'Pending',
    verified_at TIMESTAMP WITH TIME ZONE,
    verified_by UUID,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Create indexes for performance
CREATE INDEX idx_customers_customer_number ON customers(customer_number);
CREATE INDEX idx_customers_email ON customers(email);
CREATE INDEX idx_customers_tax_id_hash ON customers((ENCODE(SHA256(tax_id::bytea), 'hex')));
CREATE INDEX idx_customers_status ON customers(status);
CREATE INDEX idx_customers_name_gin ON customers USING GIN (to_tsvector('english', first_name || ' ' || COALESCE(middle_name, '') || ' ' || last_name));
CREATE INDEX idx_customers_email_gin ON customers USING GIN (to_tsvector('english', email));

CREATE INDEX idx_addresses_customer_id ON addresses(customer_id);
CREATE INDEX idx_addresses_is_primary ON addresses(customer_id, is_primary) WHERE is_primary = TRUE;

CREATE INDEX idx_customer_documents_customer_id ON customer_documents(customer_id);
CREATE INDEX idx_customer_documents_verification_status ON customer_documents(verification_status);

-- Create trigger for updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_customers_updated_at
    BEFORE UPDATE ON customers
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_addresses_updated_at
    BEFORE UPDATE ON addresses
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_customer_documents_updated_at
    BEFORE UPDATE ON customer_documents
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
