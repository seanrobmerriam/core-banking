package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/core-banking/services/customer-service/internal/encryption"
	"github.com/core-banking/services/customer-service/internal/models"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

// DBQuerier is an interface for database operations
type DBQuerier interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}

// pgCustomerRepository implements CustomerRepository for PostgreSQL
type pgCustomerRepository struct {
	db        DBQuerier
	encryptor *encryption.Encryptor
}

// NewCustomerRepository creates a new PostgreSQL customer repository
func NewCustomerRepository(db *sql.DB, encryptor *encryption.Encryptor) CustomerRepository {
	return &pgCustomerRepository{
		db:        db,
		encryptor: encryptor,
	}
}

// Customer operations

func (r *pgCustomerRepository) CreateCustomer(ctx context.Context, customer *models.Customer) error {
	if customer.ID == uuid.Nil {
		customer.ID = uuid.New()
	}
	customer.CreatedAt = time.Now().UTC()
	customer.UpdatedAt = time.Now().UTC()
	customer.Version = 1

	encryptedTaxID, err := r.encryptor.Encrypt(customer.TaxID)
	if err != nil {
		return fmt.Errorf("failed to encrypt tax id: %w", err)
	}

	query := `
		INSERT INTO customers (
			id, customer_number, first_name, middle_name, last_name,
			date_of_birth, tax_id, email, phone, status,
			created_at, updated_at, created_by, version
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14
		)
	`

	_, err = r.db.ExecContext(ctx, query,
		customer.ID,
		customer.CustomerNumber,
		customer.FirstName,
		customer.MiddleName,
		customer.LastName,
		customer.DateOfBirth,
		encryptedTaxID,
		customer.Email,
		customer.Phone,
		customer.Status,
		customer.CreatedAt,
		customer.UpdatedAt,
		customer.CreatedBy,
		customer.Version,
	)

	if err != nil {
		return fmt.Errorf("failed to create customer: %w", err)
	}

	return nil
}

func (r *pgCustomerRepository) GetCustomerByID(ctx context.Context, id uuid.UUID) (*models.Customer, error) {
	query := `
		SELECT id, customer_number, first_name, middle_name, last_name,
			date_of_birth, tax_id, email, phone, status,
			created_at, updated_at, created_by, updated_by, version
		FROM customers
		WHERE id = $1
	`

	customer := &models.Customer{}
	var encryptedTaxID string
	var updatedBy sql.NullString

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&customer.ID,
		&customer.CustomerNumber,
		&customer.FirstName,
		&customer.MiddleName,
		&customer.LastName,
		&customer.DateOfBirth,
		&encryptedTaxID,
		&customer.Email,
		&customer.Phone,
		&customer.Status,
		&customer.CreatedAt,
		&customer.UpdatedAt,
		&customer.CreatedBy,
		&updatedBy,
		&customer.Version,
	)

	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get customer: %w", err)
	}

	if encryptedTaxID != "" {
		decrypted, err := r.encryptor.Decrypt(encryptedTaxID)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt tax id: %w", err)
		}
		customer.TaxID = decrypted
	}

	if updatedBy.Valid {
		updatedByUUID := uuid.MustParse(updatedBy.String)
		customer.UpdatedBy = &updatedByUUID
	}

	return customer, nil
}

func (r *pgCustomerRepository) GetCustomerByNumber(ctx context.Context, customerNumber string) (*models.Customer, error) {
	query := `
		SELECT id, customer_number, first_name, middle_name, last_name,
			date_of_birth, tax_id, email, phone, status,
			created_at, updated_at, created_by, updated_by, version
		FROM customers
		WHERE customer_number = $1
	`

	customer := &models.Customer{}
	var encryptedTaxID string
	var updatedBy sql.NullString

	err := r.db.QueryRowContext(ctx, query, customerNumber).Scan(
		&customer.ID,
		&customer.CustomerNumber,
		&customer.FirstName,
		&customer.MiddleName,
		&customer.LastName,
		&customer.DateOfBirth,
		&encryptedTaxID,
		&customer.Email,
		&customer.Phone,
		&customer.Status,
		&customer.CreatedAt,
		&customer.UpdatedAt,
		&customer.CreatedBy,
		&updatedBy,
		&customer.Version,
	)

	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get customer: %w", err)
	}

	if encryptedTaxID != "" {
		decrypted, err := r.encryptor.Decrypt(encryptedTaxID)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt tax id: %w", err)
		}
		customer.TaxID = decrypted
	}

	if updatedBy.Valid {
		updatedByUUID := uuid.MustParse(updatedBy.String)
		customer.UpdatedBy = &updatedByUUID
	}

	return customer, nil
}

func (r *pgCustomerRepository) UpdateCustomer(ctx context.Context, customer *models.Customer) error {
	customer.UpdatedAt = time.Now().UTC()
	customer.Version++

	encryptedTaxID, err := r.encryptor.Encrypt(customer.TaxID)
	if err != nil {
		return fmt.Errorf("failed to encrypt tax id: %w", err)
	}

	query := `
		UPDATE customers SET
			customer_number = $2,
			first_name = $3,
			middle_name = $4,
			last_name = $5,
			date_of_birth = $6,
			tax_id = $7,
			email = $8,
			phone = $9,
			status = $10,
			updated_at = $11,
			updated_by = $12,
			version = $13
		WHERE id = $1 AND version = $14
	`

	result, err := r.db.ExecContext(ctx, query,
		customer.ID,
		customer.CustomerNumber,
		customer.FirstName,
		customer.MiddleName,
		customer.LastName,
		customer.DateOfBirth,
		encryptedTaxID,
		customer.Email,
		customer.Phone,
		customer.Status,
		customer.UpdatedAt,
		customer.UpdatedBy,
		customer.Version,
		customer.Version-1,
	)

	if err != nil {
		return fmt.Errorf("failed to update customer: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return &ErrOptimisticLock{
			CustomerID: customer.ID,
		}
	}

	return nil
}

func (r *pgCustomerRepository) DeleteCustomer(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM customers WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete customer: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

func (r *pgCustomerRepository) SearchCustomers(ctx context.Context, filters models.SearchFilters) ([]*models.Customer, error) {
	var conditions []string
	var args []interface{}
	argIdx := 1

	if filters.FirstName != "" {
		conditions = append(conditions, fmt.Sprintf("first_name ILIKE $%d", argIdx))
		args = append(args, "%"+filters.FirstName+"%")
		argIdx++
	}

	if filters.LastName != "" {
		conditions = append(conditions, fmt.Sprintf("last_name ILIKE $%d", argIdx))
		args = append(args, "%"+filters.LastName+"%")
		argIdx++
	}

	if filters.Email != "" {
		conditions = append(conditions, fmt.Sprintf("email ILIKE $%d", argIdx))
		args = append(args, "%"+filters.Email+"%")
		argIdx++
	}

	if filters.Phone != "" {
		conditions = append(conditions, fmt.Sprintf("phone ILIKE $%d", argIdx))
		args = append(args, "%"+filters.Phone+"%")
		argIdx++
	}

	if filters.Status != "" {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, filters.Status)
		argIdx++
	}

	if filters.FromDate != nil {
		conditions = append(conditions, fmt.Sprintf("created_at >= $%d", argIdx))
		args = append(args, *filters.FromDate)
		argIdx++
	}

	if filters.ToDate != nil {
		conditions = append(conditions, fmt.Sprintf("created_at <= $%d", argIdx))
		args = append(args, *filters.ToDate)
		argIdx++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	limit := 50
	offset := 0
	if filters.Limit > 0 {
		limit = filters.Limit
	}
	if filters.Offset > 0 {
		offset = filters.Offset
	}

	query := fmt.Sprintf(`
		SELECT id, customer_number, first_name, middle_name, last_name,
			date_of_birth, tax_id, email, phone, status,
			created_at, updated_at, created_by, updated_by, version
		FROM customers
		%s
		ORDER BY created_at DESC
		LIMIT %d OFFSET %d
	`, whereClause, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to search customers: %w", err)
	}
	defer rows.Close()

	var customers []*models.Customer
	for rows.Next() {
		customer := &models.Customer{}
		var encryptedTaxID string
		var updatedBy sql.NullString

		err := rows.Scan(
			&customer.ID,
			&customer.CustomerNumber,
			&customer.FirstName,
			&customer.MiddleName,
			&customer.LastName,
			&customer.DateOfBirth,
			&encryptedTaxID,
			&customer.Email,
			&customer.Phone,
			&customer.Status,
			&customer.CreatedAt,
			&customer.UpdatedAt,
			&customer.CreatedBy,
			&updatedBy,
			&customer.Version,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan customer: %w", err)
		}

		if encryptedTaxID != "" {
			decrypted, err := r.encryptor.Decrypt(encryptedTaxID)
			if err != nil {
				return nil, fmt.Errorf("failed to decrypt tax id: %w", err)
			}
			customer.TaxID = decrypted
		}

		if updatedBy.Valid {
			updatedByUUID := uuid.MustParse(updatedBy.String)
			customer.UpdatedBy = &updatedByUUID
		}

		customers = append(customers, customer)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating customers: %w", err)
	}

	return customers, nil
}

// Address operations

func (r *pgCustomerRepository) AddAddress(ctx context.Context, address *models.Address) error {
	if address.ID == uuid.Nil {
		address.ID = uuid.New()
	}
	address.CreatedAt = time.Now().UTC()
	address.UpdatedAt = time.Now().UTC()

	query := `
		INSERT INTO addresses (
			id, customer_id, address_type, street1, street2,
			city, state, postal_code, country, is_primary,
			valid_from, valid_to, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14
		)
	`

	_, err := r.db.ExecContext(ctx, query,
		address.ID,
		address.CustomerID,
		address.AddressType,
		address.Street1,
		address.Street2,
		address.City,
		address.State,
		address.PostalCode,
		address.Country,
		address.IsPrimary,
		address.ValidFrom,
		address.ValidTo,
		address.CreatedAt,
		address.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to add address: %w", err)
	}

	return nil
}

func (r *pgCustomerRepository) UpdateAddress(ctx context.Context, address *models.Address) error {
	address.UpdatedAt = time.Now().UTC()

	query := `
		UPDATE addresses SET
			address_type = $2,
			street1 = $3,
			street2 = $4,
			city = $5,
			state = $6,
			postal_code = $7,
			country = $8,
			is_primary = $9,
			valid_from = $10,
			valid_to = $11,
			updated_at = $12
		WHERE id = $1
	`

	_, err := r.db.ExecContext(ctx, query,
		address.ID,
		address.AddressType,
		address.Street1,
		address.Street2,
		address.City,
		address.State,
		address.PostalCode,
		address.Country,
		address.IsPrimary,
		address.ValidFrom,
		address.ValidTo,
		address.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update address: %w", err)
	}

	return nil
}

func (r *pgCustomerRepository) GetCustomerAddresses(ctx context.Context, customerID uuid.UUID) ([]*models.Address, error) {
	query := `
		SELECT id, customer_id, address_type, street1, street2,
			city, state, postal_code, country, is_primary,
			valid_from, valid_to, created_at, updated_at
		FROM addresses
		WHERE customer_id = $1
		ORDER BY is_primary DESC, created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, customerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get addresses: %w", err)
	}
	defer rows.Close()

	var addresses []*models.Address
	for rows.Next() {
		address := &models.Address{}
		var street2 sql.NullString
		var validTo sql.NullTime

		err := rows.Scan(
			&address.ID,
			&address.CustomerID,
			&address.AddressType,
			&address.Street1,
			&street2,
			&address.City,
			&address.State,
			&address.PostalCode,
			&address.Country,
			&address.IsPrimary,
			&address.ValidFrom,
			&validTo,
			&address.CreatedAt,
			&address.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan address: %w", err)
		}

		if street2.Valid {
			street2Str := street2.String
			address.Street2 = &street2Str
		}
		if validTo.Valid {
			validToTime := validTo.Time
			address.ValidTo = &validToTime
		}

		addresses = append(addresses, address)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating addresses: %w", err)
	}

	return addresses, nil
}

func (r *pgCustomerRepository) DeleteAddress(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM addresses WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete address: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

// Document operations

func (r *pgCustomerRepository) AddDocument(ctx context.Context, doc *models.CustomerDocument) error {
	if doc.ID == uuid.Nil {
		doc.ID = uuid.New()
	}
	doc.CreatedAt = time.Now().UTC()
	doc.UpdatedAt = time.Now().UTC()

	encryptedDocNumber, err := r.encryptor.Encrypt(doc.DocumentNumber)
	if err != nil {
		return fmt.Errorf("failed to encrypt document number: %w", err)
	}

	query := `
		INSERT INTO customer_documents (
			id, customer_id, document_type, document_number,
			issuing_authority, issuing_country, issue_date, expiry_date,
			verification_status, verified_at, verified_by, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
		)
	`

	_, err = r.db.ExecContext(ctx, query,
		doc.ID,
		doc.CustomerID,
		doc.DocumentType,
		encryptedDocNumber,
		doc.IssuingAuthority,
		doc.IssuingCountry,
		doc.IssueDate,
		doc.ExpiryDate,
		doc.VerificationStatus,
		doc.VerifiedAt,
		doc.VerifiedBy,
		doc.CreatedAt,
		doc.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to add document: %w", err)
	}

	return nil
}

func (r *pgCustomerRepository) UpdateDocument(ctx context.Context, doc *models.CustomerDocument) error {
	doc.UpdatedAt = time.Now().UTC()

	encryptedDocNumber, err := r.encryptor.Encrypt(doc.DocumentNumber)
	if err != nil {
		return fmt.Errorf("failed to encrypt document number: %w", err)
	}

	query := `
		UPDATE customer_documents SET
			document_type = $2,
			document_number = $3,
			issuing_authority = $4,
			issuing_country = $5,
			issue_date = $6,
			expiry_date = $7,
			verification_status = $8,
			verified_at = $9,
			verified_by = $10,
			updated_at = $11
		WHERE id = $1
	`

	_, err = r.db.ExecContext(ctx, query,
		doc.ID,
		doc.DocumentType,
		encryptedDocNumber,
		doc.IssuingAuthority,
		doc.IssuingCountry,
		doc.IssueDate,
		doc.ExpiryDate,
		doc.VerificationStatus,
		doc.VerifiedAt,
		doc.VerifiedBy,
		doc.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update document: %w", err)
	}

	return nil
}

func (r *pgCustomerRepository) GetCustomerDocuments(ctx context.Context, customerID uuid.UUID) ([]*models.CustomerDocument, error) {
	query := `
		SELECT id, customer_id, document_type, document_number,
			issuing_authority, issuing_country, issue_date, expiry_date,
			verification_status, verified_at, verified_by, created_at, updated_at
		FROM customer_documents
		WHERE customer_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, customerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get documents: %w", err)
	}
	defer rows.Close()

	var documents []*models.CustomerDocument
	for rows.Next() {
		doc := &models.CustomerDocument{}
		var encryptedDocNumber string
		var verifiedAt sql.NullTime
		var verifiedBy sql.NullString

		err := rows.Scan(
			&doc.ID,
			&doc.CustomerID,
			&doc.DocumentType,
			&encryptedDocNumber,
			&doc.IssuingAuthority,
			&doc.IssuingCountry,
			&doc.IssueDate,
			&doc.ExpiryDate,
			&doc.VerificationStatus,
			&verifiedAt,
			&verifiedBy,
			&doc.CreatedAt,
			&doc.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan document: %w", err)
		}

		if encryptedDocNumber != "" {
			decrypted, err := r.encryptor.Decrypt(encryptedDocNumber)
			if err != nil {
				return nil, fmt.Errorf("failed to decrypt document number: %w", err)
			}
			doc.DocumentNumber = decrypted
		}

		if verifiedAt.Valid {
			verifiedAtTime := verifiedAt.Time
			doc.VerifiedAt = &verifiedAtTime
		}
		if verifiedBy.Valid {
			verifiedByUUID := uuid.MustParse(verifiedBy.String)
			doc.VerifiedBy = &verifiedByUUID
		}

		documents = append(documents, doc)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating documents: %w", err)
	}

	return documents, nil
}

func (r *pgCustomerRepository) DeleteDocument(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM customer_documents WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete document: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

// Transaction management

func (r *pgCustomerRepository) BeginTx(ctx context.Context) (Tx, error) {
	tx, err := r.db.(*sql.DB).BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	return &pgTx{tx: tx, encryptor: r.encryptor}, nil
}

// pgTx implements Tx for PostgreSQL
type pgTx struct {
	tx        *sql.Tx
	encryptor *encryption.Encryptor
}

func (t *pgTx) Commit(ctx context.Context) error {
	if err := t.tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil
}

func (t *pgTx) Rollback(ctx context.Context) error {
	if err := t.tx.Rollback(); err != nil {
		return fmt.Errorf("failed to rollback transaction: %w", err)
	}
	return nil
}

func (t *pgTx) CustomerRepository() CustomerRepository {
	return &txCustomerRepository{
		tx:        t.tx,
		encryptor: t.encryptor,
	}
}

// txCustomerRepository wraps a transaction for CustomerRepository
type txCustomerRepository struct {
	tx        *sql.Tx
	encryptor *encryption.Encryptor
}

func (r *txCustomerRepository) CreateCustomer(ctx context.Context, customer *models.Customer) error {
	if customer.ID == uuid.Nil {
		customer.ID = uuid.New()
	}
	customer.CreatedAt = time.Now().UTC()
	customer.UpdatedAt = time.Now().UTC()
	customer.Version = 1

	encryptedTaxID, err := r.encryptor.Encrypt(customer.TaxID)
	if err != nil {
		return fmt.Errorf("failed to encrypt tax id: %w", err)
	}

	query := `
		INSERT INTO customers (
			id, customer_number, first_name, middle_name, last_name,
			date_of_birth, tax_id, email, phone, status,
			created_at, updated_at, created_by, version
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14
		)
	`

	_, err = r.tx.ExecContext(ctx, query,
		customer.ID,
		customer.CustomerNumber,
		customer.FirstName,
		customer.MiddleName,
		customer.LastName,
		customer.DateOfBirth,
		encryptedTaxID,
		customer.Email,
		customer.Phone,
		customer.Status,
		customer.CreatedAt,
		customer.UpdatedAt,
		customer.CreatedBy,
		customer.Version,
	)

	if err != nil {
		return fmt.Errorf("failed to create customer: %w", err)
	}

	return nil
}

func (r *txCustomerRepository) GetCustomerByID(ctx context.Context, id uuid.UUID) (*models.Customer, error) {
	query := `
		SELECT id, customer_number, first_name, middle_name, last_name,
			date_of_birth, tax_id, email, phone, status,
			created_at, updated_at, created_by, updated_by, version
		FROM customers
		WHERE id = $1
	`

	customer := &models.Customer{}
	var encryptedTaxID string
	var updatedBy sql.NullString

	err := r.tx.QueryRowContext(ctx, query, id).Scan(
		&customer.ID,
		&customer.CustomerNumber,
		&customer.FirstName,
		&customer.MiddleName,
		&customer.LastName,
		&customer.DateOfBirth,
		&encryptedTaxID,
		&customer.Email,
		&customer.Phone,
		&customer.Status,
		&customer.CreatedAt,
		&customer.UpdatedAt,
		&customer.CreatedBy,
		&updatedBy,
		&customer.Version,
	)

	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get customer: %w", err)
	}

	if encryptedTaxID != "" {
		decrypted, err := r.encryptor.Decrypt(encryptedTaxID)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt tax id: %w", err)
		}
		customer.TaxID = decrypted
	}

	if updatedBy.Valid {
		updatedByUUID := uuid.MustParse(updatedBy.String)
		customer.UpdatedBy = &updatedByUUID
	}

	return customer, nil
}

func (r *txCustomerRepository) GetCustomerByNumber(ctx context.Context, customerNumber string) (*models.Customer, error) {
	query := `
		SELECT id, customer_number, first_name, middle_name, last_name,
			date_of_birth, tax_id, email, phone, status,
			created_at, updated_at, created_by, updated_by, version
		FROM customers
		WHERE customer_number = $1
	`

	customer := &models.Customer{}
	var encryptedTaxID string
	var updatedBy sql.NullString

	err := r.tx.QueryRowContext(ctx, query, customerNumber).Scan(
		&customer.ID,
		&customer.CustomerNumber,
		&customer.FirstName,
		&customer.MiddleName,
		&customer.LastName,
		&customer.DateOfBirth,
		&encryptedTaxID,
		&customer.Email,
		&customer.Phone,
		&customer.Status,
		&customer.CreatedAt,
		&customer.UpdatedAt,
		&customer.CreatedBy,
		&updatedBy,
		&customer.Version,
	)

	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get customer: %w", err)
	}

	if encryptedTaxID != "" {
		decrypted, err := r.encryptor.Decrypt(encryptedTaxID)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt tax id: %w", err)
		}
		customer.TaxID = decrypted
	}

	if updatedBy.Valid {
		updatedByUUID := uuid.MustParse(updatedBy.String)
		customer.UpdatedBy = &updatedByUUID
	}

	return customer, nil
}

func (r *txCustomerRepository) UpdateCustomer(ctx context.Context, customer *models.Customer) error {
	customer.UpdatedAt = time.Now().UTC()
	customer.Version++

	encryptedTaxID, err := r.encryptor.Encrypt(customer.TaxID)
	if err != nil {
		return fmt.Errorf("failed to encrypt tax id: %w", err)
	}

	query := `
		UPDATE customers SET
			customer_number = $2,
			first_name = $3,
			middle_name = $4,
			last_name = $5,
			date_of_birth = $6,
			tax_id = $7,
			email = $8,
			phone = $9,
			status = $10,
			updated_at = $11,
			updated_by = $12,
			version = $13
		WHERE id = $1 AND version = $14
	`

	result, err := r.tx.ExecContext(ctx, query,
		customer.ID,
		customer.CustomerNumber,
		customer.FirstName,
		customer.MiddleName,
		customer.LastName,
		customer.DateOfBirth,
		encryptedTaxID,
		customer.Email,
		customer.Phone,
		customer.Status,
		customer.UpdatedAt,
		customer.UpdatedBy,
		customer.Version,
		customer.Version-1,
	)

	if err != nil {
		return fmt.Errorf("failed to update customer: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return &ErrOptimisticLock{
			CustomerID: customer.ID,
		}
	}

	return nil
}

func (r *txCustomerRepository) DeleteCustomer(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM customers WHERE id = $1`

	result, err := r.tx.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete customer: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

func (r *txCustomerRepository) SearchCustomers(ctx context.Context, filters models.SearchFilters) ([]*models.Customer, error) {
	var conditions []string
	var args []interface{}
	argIdx := 1

	if filters.FirstName != "" {
		conditions = append(conditions, fmt.Sprintf("first_name ILIKE $%d", argIdx))
		args = append(args, "%"+filters.FirstName+"%")
		argIdx++
	}

	if filters.LastName != "" {
		conditions = append(conditions, fmt.Sprintf("last_name ILIKE $%d", argIdx))
		args = append(args, "%"+filters.LastName+"%")
		argIdx++
	}

	if filters.Email != "" {
		conditions = append(conditions, fmt.Sprintf("email ILIKE $%d", argIdx))
		args = append(args, "%"+filters.Email+"%")
		argIdx++
	}

	if filters.Phone != "" {
		conditions = append(conditions, fmt.Sprintf("phone ILIKE $%d", argIdx))
		args = append(args, "%"+filters.Phone+"%")
		argIdx++
	}

	if filters.Status != "" {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, filters.Status)
		argIdx++
	}

	if filters.FromDate != nil {
		conditions = append(conditions, fmt.Sprintf("created_at >= $%d", argIdx))
		args = append(args, *filters.FromDate)
		argIdx++
	}

	if filters.ToDate != nil {
		conditions = append(conditions, fmt.Sprintf("created_at <= $%d", argIdx))
		args = append(args, *filters.ToDate)
		argIdx++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	limit := 50
	offset := 0
	if filters.Limit > 0 {
		limit = filters.Limit
	}
	if filters.Offset > 0 {
		offset = filters.Offset
	}

	query := fmt.Sprintf(`
		SELECT id, customer_number, first_name, middle_name, last_name,
			date_of_birth, tax_id, email, phone, status,
			created_at, updated_at, created_by, updated_by, version
		FROM customers
		%s
		ORDER BY created_at DESC
		LIMIT %d OFFSET %d
	`, whereClause, limit, offset)

	rows, err := r.tx.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to search customers: %w", err)
	}
	defer rows.Close()

	var customers []*models.Customer
	for rows.Next() {
		customer := &models.Customer{}
		var encryptedTaxID string
		var updatedBy sql.NullString

		err := rows.Scan(
			&customer.ID,
			&customer.CustomerNumber,
			&customer.FirstName,
			&customer.MiddleName,
			&customer.LastName,
			&customer.DateOfBirth,
			&encryptedTaxID,
			&customer.Email,
			&customer.Phone,
			&customer.Status,
			&customer.CreatedAt,
			&customer.UpdatedAt,
			&customer.CreatedBy,
			&updatedBy,
			&customer.Version,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan customer: %w", err)
		}

		if encryptedTaxID != "" {
			decrypted, err := r.encryptor.Decrypt(encryptedTaxID)
			if err != nil {
				return nil, fmt.Errorf("failed to decrypt tax id: %w", err)
			}
			customer.TaxID = decrypted
		}

		if updatedBy.Valid {
			updatedByUUID := uuid.MustParse(updatedBy.String)
			customer.UpdatedBy = &updatedByUUID
		}

		customers = append(customers, customer)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating customers: %w", err)
	}

	return customers, nil
}

func (r *txCustomerRepository) AddAddress(ctx context.Context, address *models.Address) error {
	if address.ID == uuid.Nil {
		address.ID = uuid.New()
	}
	address.CreatedAt = time.Now().UTC()
	address.UpdatedAt = time.Now().UTC()

	query := `
		INSERT INTO addresses (
			id, customer_id, address_type, street1, street2,
			city, state, postal_code, country, is_primary,
			valid_from, valid_to, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14
		)
	`

	_, err := r.tx.ExecContext(ctx, query,
		address.ID,
		address.CustomerID,
		address.AddressType,
		address.Street1,
		address.Street2,
		address.City,
		address.State,
		address.PostalCode,
		address.Country,
		address.IsPrimary,
		address.ValidFrom,
		address.ValidTo,
		address.CreatedAt,
		address.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to add address: %w", err)
	}

	return nil
}

func (r *txCustomerRepository) UpdateAddress(ctx context.Context, address *models.Address) error {
	address.UpdatedAt = time.Now().UTC()

	query := `
		UPDATE addresses SET
			address_type = $2,
			street1 = $3,
			street2 = $4,
			city = $5,
			state = $6,
			postal_code = $7,
			country = $8,
			is_primary = $9,
			valid_from = $10,
			valid_to = $11,
			updated_at = $12
		WHERE id = $1
	`

	_, err := r.tx.ExecContext(ctx, query,
		address.ID,
		address.AddressType,
		address.Street1,
		address.Street2,
		address.City,
		address.State,
		address.PostalCode,
		address.Country,
		address.IsPrimary,
		address.ValidFrom,
		address.ValidTo,
		address.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update address: %w", err)
	}

	return nil
}

func (r *txCustomerRepository) GetCustomerAddresses(ctx context.Context, customerID uuid.UUID) ([]*models.Address, error) {
	query := `
		SELECT id, customer_id, address_type, street1, street2,
			city, state, postal_code, country, is_primary,
			valid_from, valid_to, created_at, updated_at
		FROM addresses
		WHERE customer_id = $1
		ORDER BY is_primary DESC, created_at DESC
	`

	rows, err := r.tx.QueryContext(ctx, query, customerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get addresses: %w", err)
	}
	defer rows.Close()

	var addresses []*models.Address
	for rows.Next() {
		address := &models.Address{}
		var street2 sql.NullString
		var validTo sql.NullTime

		err := rows.Scan(
			&address.ID,
			&address.CustomerID,
			&address.AddressType,
			&address.Street1,
			&street2,
			&address.City,
			&address.State,
			&address.PostalCode,
			&address.Country,
			&address.IsPrimary,
			&address.ValidFrom,
			&validTo,
			&address.CreatedAt,
			&address.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan address: %w", err)
		}

		if street2.Valid {
			street2Str := street2.String
			address.Street2 = &street2Str
		}
		if validTo.Valid {
			validToTime := validTo.Time
			address.ValidTo = &validToTime
		}

		addresses = append(addresses, address)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating addresses: %w", err)
	}

	return addresses, nil
}

func (r *txCustomerRepository) DeleteAddress(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM addresses WHERE id = $1`

	result, err := r.tx.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete address: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

func (r *txCustomerRepository) AddDocument(ctx context.Context, doc *models.CustomerDocument) error {
	if doc.ID == uuid.Nil {
		doc.ID = uuid.New()
	}
	doc.CreatedAt = time.Now().UTC()
	doc.UpdatedAt = time.Now().UTC()

	encryptedDocNumber, err := r.encryptor.Encrypt(doc.DocumentNumber)
	if err != nil {
		return fmt.Errorf("failed to encrypt document number: %w", err)
	}

	query := `
		INSERT INTO customer_documents (
			id, customer_id, document_type, document_number,
			issuing_authority, issuing_country, issue_date, expiry_date,
			verification_status, verified_at, verified_by, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
		)
	`

	_, err = r.tx.ExecContext(ctx, query,
		doc.ID,
		doc.CustomerID,
		doc.DocumentType,
		encryptedDocNumber,
		doc.IssuingAuthority,
		doc.IssuingCountry,
		doc.IssueDate,
		doc.ExpiryDate,
		doc.VerificationStatus,
		doc.VerifiedAt,
		doc.VerifiedBy,
		doc.CreatedAt,
		doc.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to add document: %w", err)
	}

	return nil
}

func (r *txCustomerRepository) UpdateDocument(ctx context.Context, doc *models.CustomerDocument) error {
	doc.UpdatedAt = time.Now().UTC()

	encryptedDocNumber, err := r.encryptor.Encrypt(doc.DocumentNumber)
	if err != nil {
		return fmt.Errorf("failed to encrypt document number: %w", err)
	}

	query := `
		UPDATE customer_documents SET
			document_type = $2,
			document_number = $3,
			issuing_authority = $4,
			issuing_country = $5,
			issue_date = $6,
			expiry_date = $7,
			verification_status = $8,
			verified_at = $9,
			verified_by = $10,
			updated_at = $11
		WHERE id = $1
	`

	_, err = r.tx.ExecContext(ctx, query,
		doc.ID,
		doc.DocumentType,
		encryptedDocNumber,
		doc.IssuingAuthority,
		doc.IssuingCountry,
		doc.IssueDate,
		doc.ExpiryDate,
		doc.VerificationStatus,
		doc.VerifiedAt,
		doc.VerifiedBy,
		doc.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update document: %w", err)
	}

	return nil
}

func (r *txCustomerRepository) GetCustomerDocuments(ctx context.Context, customerID uuid.UUID) ([]*models.CustomerDocument, error) {
	query := `
		SELECT id, customer_id, document_type, document_number,
			issuing_authority, issuing_country, issue_date, expiry_date,
			verification_status, verified_at, verified_by, created_at, updated_at
		FROM customer_documents
		WHERE customer_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.tx.QueryContext(ctx, query, customerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get documents: %w", err)
	}
	defer rows.Close()

	var documents []*models.CustomerDocument
	for rows.Next() {
		doc := &models.CustomerDocument{}
		var encryptedDocNumber string
		var verifiedAt sql.NullTime
		var verifiedBy sql.NullString

		err := rows.Scan(
			&doc.ID,
			&doc.CustomerID,
			&doc.DocumentType,
			&encryptedDocNumber,
			&doc.IssuingAuthority,
			&doc.IssuingCountry,
			&doc.IssueDate,
			&doc.ExpiryDate,
			&doc.VerificationStatus,
			&verifiedAt,
			&verifiedBy,
			&doc.CreatedAt,
			&doc.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan document: %w", err)
		}

		if encryptedDocNumber != "" {
			decrypted, err := r.encryptor.Decrypt(encryptedDocNumber)
			if err != nil {
				return nil, fmt.Errorf("failed to decrypt document number: %w", err)
			}
			doc.DocumentNumber = decrypted
		}

		if verifiedAt.Valid {
			verifiedAtTime := verifiedAt.Time
			doc.VerifiedAt = &verifiedAtTime
		}
		if verifiedBy.Valid {
			verifiedByUUID := uuid.MustParse(verifiedBy.String)
			doc.VerifiedBy = &verifiedByUUID
		}

		documents = append(documents, doc)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating documents: %w", err)
	}

	return documents, nil
}

func (r *txCustomerRepository) DeleteDocument(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM customer_documents WHERE id = $1`

	result, err := r.tx.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete document: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

func (r *txCustomerRepository) BeginTx(ctx context.Context) (Tx, error) {
	return nil, fmt.Errorf("nested transactions not supported")
}
