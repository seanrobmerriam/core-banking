package repository

import (
	"context"
	"errors"

	"github.com/core-banking/services/customer-service/internal/models"
	"github.com/google/uuid"
)

// ErrNotFound is returned when a record is not found
var ErrNotFound = errors.New("record not found")

// OptimisticLockError is returned when optimistic locking fails
type OptimisticLockError struct {
	ExpectedVersion int
	ActualVersion   int
}

func (e *OptimisticLockError) Error() string {
	return "optimistic lock error: version mismatch"
}

// CustomerRepository defines the interface for customer data operations
type CustomerRepository interface {
	// Customer operations
	CreateCustomer(ctx context.Context, customer *models.Customer) error
	GetCustomerByID(ctx context.Context, id uuid.UUID) (*models.Customer, error)
	GetCustomerByNumber(ctx context.Context, customerNumber string) (*models.Customer, error)
	UpdateCustomer(ctx context.Context, customer *models.Customer) error
	DeleteCustomer(ctx context.Context, id uuid.UUID) error
	SearchCustomers(ctx context.Context, filters models.SearchFilters) ([]*models.Customer, error)

	// Address operations
	AddAddress(ctx context.Context, address *models.Address) error
	UpdateAddress(ctx context.Context, address *models.Address) error
	GetCustomerAddresses(ctx context.Context, customerID uuid.UUID) ([]*models.Address, error)
	DeleteAddress(ctx context.Context, id uuid.UUID) error

	// Document operations
	AddDocument(ctx context.Context, doc *models.CustomerDocument) error
	UpdateDocument(ctx context.Context, doc *models.CustomerDocument) error
	GetCustomerDocuments(ctx context.Context, customerID uuid.UUID) ([]*models.CustomerDocument, error)
	DeleteDocument(ctx context.Context, id uuid.UUID) error

	// Transaction management
	BeginTx(ctx context.Context) (Tx, error)
}

// Tx represents a database transaction
type Tx interface {
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
	CustomerRepository() CustomerRepository
}

// ErrOptimisticLock is returned when a concurrent update is detected
type ErrOptimisticLock struct {
	CustomerID      uuid.UUID
	ExpectedVersion int
	ActualVersion   int
}

func (e *ErrOptimisticLock) Error() string {
	return "optimistic lock error"
}

// CustomerRepositoryEx extends CustomerRepository with additional methods for testing
type CustomerRepositoryEx interface {
	CustomerRepository
	GetDB() interface{}
}
