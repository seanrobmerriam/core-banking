package service

import (
	"context"
	"testing"
	"time"

	"github.com/core-banking/services/customer-service/internal/models"
	customerpb "github.com/core-banking/services/customer-service/internal/proto/customerpb"
	"github.com/core-banking/services/customer-service/internal/repository"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// MockRepository is a mock implementation of CustomerRepository for testing
type MockRepository struct {
	customers map[uuid.UUID]*models.Customer
	addresses map[uuid.UUID][]*models.Address
	documents map[uuid.UUID][]*models.CustomerDocument
	nextErr   error
}

func NewMockRepository() *MockRepository {
	return &MockRepository{
		customers: make(map[uuid.UUID]*models.Customer),
		addresses: make(map[uuid.UUID][]*models.Address),
		documents: make(map[uuid.UUID][]*models.CustomerDocument),
	}
}

func (m *MockRepository) CreateCustomer(ctx context.Context, customer *models.Customer) error {
	if m.nextErr != nil {
		return m.nextErr
	}
	customer.CreatedAt = time.Now().UTC()
	customer.UpdatedAt = time.Now().UTC()
	customer.Version = 1
	m.customers[customer.ID] = customer
	return nil
}

func (m *MockRepository) GetCustomerByID(ctx context.Context, id uuid.UUID) (*models.Customer, error) {
	if m.nextErr != nil {
		return nil, m.nextErr
	}
	customer, exists := m.customers[id]
	if !exists {
		return nil, repository.ErrNotFound
	}
	return customer, nil
}

func (m *MockRepository) GetCustomerByNumber(ctx context.Context, customerNumber string) (*models.Customer, error) {
	for _, customer := range m.customers {
		if customer.CustomerNumber == customerNumber {
			return customer, nil
		}
	}
	return nil, repository.ErrNotFound
}

func (m *MockRepository) UpdateCustomer(ctx context.Context, customer *models.Customer) error {
	if m.nextErr != nil {
		return m.nextErr
	}
	if _, exists := m.customers[customer.ID]; !exists {
		return repository.ErrNotFound
	}
	customer.UpdatedAt = time.Now().UTC()
	customer.Version++
	m.customers[customer.ID] = customer
	return nil
}

func (m *MockRepository) DeleteCustomer(ctx context.Context, id uuid.UUID) error {
	if _, exists := m.customers[id]; !exists {
		return repository.ErrNotFound
	}
	delete(m.customers, id)
	return nil
}

func (m *MockRepository) SearchCustomers(ctx context.Context, filters models.SearchFilters) ([]*models.Customer, error) {
	if m.nextErr != nil {
		return nil, m.nextErr
	}
	var results []*models.Customer
	for _, customer := range m.customers {
		if filters.FirstName != "" && customer.FirstName != filters.FirstName {
			continue
		}
		if filters.LastName != "" && customer.LastName != filters.LastName {
			continue
		}
		if filters.Email != "" && customer.Email != filters.Email {
			continue
		}
		if filters.Status != "" && customer.Status != filters.Status {
			continue
		}
		results = append(results, customer)
	}
	return results, nil
}

func (m *MockRepository) AddAddress(ctx context.Context, address *models.Address) error {
	if m.nextErr != nil {
		return m.nextErr
	}
	address.CreatedAt = time.Now().UTC()
	address.UpdatedAt = time.Now().UTC()
	m.addresses[address.CustomerID] = append(m.addresses[address.CustomerID], address)
	return nil
}

func (m *MockRepository) UpdateAddress(ctx context.Context, address *models.Address) error {
	if m.nextErr != nil {
		return m.nextErr
	}
	address.UpdatedAt = time.Now().UTC()
	return nil
}

func (m *MockRepository) GetCustomerAddresses(ctx context.Context, customerID uuid.UUID) ([]*models.Address, error) {
	if m.nextErr != nil {
		return nil, m.nextErr
	}
	return m.addresses[customerID], nil
}

func (m *MockRepository) DeleteAddress(ctx context.Context, id uuid.UUID) error {
	return nil
}

func (m *MockRepository) AddDocument(ctx context.Context, doc *models.CustomerDocument) error {
	if m.nextErr != nil {
		return m.nextErr
	}
	doc.CreatedAt = time.Now().UTC()
	doc.UpdatedAt = time.Now().UTC()
	m.documents[doc.CustomerID] = append(m.documents[doc.CustomerID], doc)
	return nil
}

func (m *MockRepository) UpdateDocument(ctx context.Context, doc *models.CustomerDocument) error {
	if m.nextErr != nil {
		return m.nextErr
	}
	doc.UpdatedAt = time.Now().UTC()
	return nil
}

func (m *MockRepository) GetCustomerDocuments(ctx context.Context, customerID uuid.UUID) ([]*models.CustomerDocument, error) {
	if m.nextErr != nil {
		return nil, m.nextErr
	}
	return m.documents[customerID], nil
}

func (m *MockRepository) DeleteDocument(ctx context.Context, id uuid.UUID) error {
	return nil
}

func (m *MockRepository) BeginTx(ctx context.Context) (repository.Tx, error) {
	return nil, nil
}

// Integration tests for CustomerService

func TestCustomerService_CreateCustomer(t *testing.T) {
	repo := NewMockRepository()
	svc := NewCustomerService(repo)
	ctx := context.Background()

	tests := []struct {
		name    string
		req     *customerpb.CreateCustomerRequest
		wantErr codes.Code
	}{
		{
			name: "valid customer creation",
			req: &customerpb.CreateCustomerRequest{
				FirstName:   "John",
				LastName:    "Doe",
				Email:       "john.doe@example.com",
				Phone:       "+1234567890",
				DateOfBirth: timestamppb.New(time.Now().AddDate(-25, 0, 0)),
				TaxId:       "123-45-6789",
				CreatedBy:   uuid.New().String(),
			},
			wantErr: codes.OK,
		},
		{
			name: "invalid email format",
			req: &customerpb.CreateCustomerRequest{
				FirstName:   "John",
				LastName:    "Doe",
				Email:       "invalid-email",
				DateOfBirth: timestamppb.New(time.Now().AddDate(-25, 0, 0)),
			},
			wantErr: codes.InvalidArgument,
		},
		{
			name: "missing required fields",
			req: &customerpb.CreateCustomerRequest{
				FirstName: "John",
			},
			wantErr: codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := svc.CreateCustomer(ctx, tt.req)
			if tt.wantErr == codes.OK {
				if err != nil {
					t.Errorf("CreateCustomer() unexpected error: %v", err)
				}
				if resp == nil || resp.Customer == nil {
					t.Error("CreateCustomer() expected customer in response")
				}
			} else {
				if err == nil {
					t.Error("CreateCustomer() expected error, got nil")
					return
				}
				st, ok := status.FromError(err)
				if !ok {
					t.Errorf("CreateCustomer() error is not a gRPC status error")
					return
				}
				if st.Code() != tt.wantErr {
					t.Errorf("CreateCustomer() got code %v, want %v", st.Code(), tt.wantErr)
				}
			}
		})
	}
}

func TestCustomerService_GetCustomer(t *testing.T) {
	repo := NewMockRepository()
	svc := NewCustomerService(repo)
	ctx := context.Background()

	// Create a customer first
	customerID := uuid.New()
	repo.customers[customerID] = &models.Customer{
		ID:             customerID,
		CustomerNumber: "CUST-123",
		FirstName:      "John",
		LastName:       "Doe",
		Email:          "john@example.com",
		Status:         models.CustomerStatusActive,
		CreatedAt:      time.Now(),
		Version:        1,
	}

	tests := []struct {
		name    string
		id      string
		wantErr codes.Code
	}{
		{
			name:    "existing customer",
			id:      customerID.String(),
			wantErr: codes.OK,
		},
		{
			name:    "non-existing customer",
			id:      uuid.New().String(),
			wantErr: codes.NotFound,
		},
		{
			name:    "invalid UUID",
			id:      "invalid-uuid",
			wantErr: codes.InvalidArgument,
		},
		{
			name:    "empty ID",
			id:      "",
			wantErr: codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := svc.GetCustomer(ctx, &customerpb.GetCustomerRequest{Id: tt.id})
			if tt.wantErr == codes.OK {
				if err != nil {
					t.Errorf("GetCustomer() unexpected error: %v", err)
				}
				if resp == nil || resp.Customer == nil {
					t.Error("GetCustomer() expected customer in response")
				}
			} else {
				if err == nil {
					t.Error("GetCustomer() expected error, got nil")
					return
				}
				st, ok := status.FromError(err)
				if !ok {
					t.Errorf("GetCustomer() error is not a gRPC status error")
					return
				}
				if st.Code() != tt.wantErr {
					t.Errorf("GetCustomer() got code %v, want %v", st.Code(), tt.wantErr)
				}
			}
		})
	}
}

func TestCustomerService_UpdateCustomerStatus(t *testing.T) {
	repo := NewMockRepository()
	svc := NewCustomerService(repo)
	ctx := context.Background()

	// Create a customer first
	customerID := uuid.New()
	repo.customers[customerID] = &models.Customer{
		ID:             customerID,
		CustomerNumber: "CUST-123",
		FirstName:      "John",
		LastName:       "Doe",
		Email:          "john@example.com",
		Status:         models.CustomerStatusPending,
		CreatedAt:      time.Now(),
		Version:        1,
	}

	tests := []struct {
		name      string
		id        string
		newStatus string
		reason    string
		changedBy string
		wantErr   codes.Code
	}{
		{
			name:      "pending to active",
			id:        customerID.String(),
			newStatus: "Active",
			reason:    "Customer verified",
			changedBy: uuid.New().String(),
			wantErr:   codes.OK,
		},
		{
			name:      "invalid transition",
			id:        customerID.String(),
			newStatus: "Suspended",
			reason:    "Invalid reason",
			changedBy: uuid.New().String(),
			wantErr:   codes.FailedPrecondition,
		},
		{
			name:      "missing reason",
			id:        customerID.String(),
			newStatus: "Active",
			reason:    "",
			changedBy: uuid.New().String(),
			wantErr:   codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := svc.UpdateCustomerStatus(ctx, &customerpb.UpdateCustomerStatusRequest{
				Id:        tt.id,
				NewStatus: tt.newStatus,
				Reason:    tt.reason,
				ChangedBy: tt.changedBy,
			})
			if tt.wantErr == codes.OK {
				if err != nil {
					t.Errorf("UpdateCustomerStatus() unexpected error: %v", err)
				}
				if resp == nil || resp.Customer == nil {
					t.Error("UpdateCustomerStatus() expected customer in response")
				}
			} else {
				if err == nil {
					t.Error("UpdateCustomerStatus() expected error, got nil")
					return
				}
				st, ok := status.FromError(err)
				if !ok {
					t.Errorf("UpdateCustomerStatus() error is not a gRPC status error")
					return
				}
				if st.Code() != tt.wantErr {
					t.Errorf("UpdateCustomerStatus() got code %v, want %v", st.Code(), tt.wantErr)
				}
			}
		})
	}
}

func TestCustomerService_AddAddress(t *testing.T) {
	repo := NewMockRepository()
	svc := NewCustomerService(repo)
	ctx := context.Background()

	// Create a customer first
	customerID := uuid.New()
	repo.customers[customerID] = &models.Customer{
		ID:             customerID,
		CustomerNumber: "CUST-123",
		FirstName:      "John",
		LastName:       "Doe",
		Email:          "john@example.com",
		Status:         models.CustomerStatusActive,
		CreatedAt:      time.Now(),
		Version:        1,
	}

	tests := []struct {
		name    string
		req     *customerpb.AddAddressRequest
		wantErr codes.Code
	}{
		{
			name: "valid address",
			req: &customerpb.AddAddressRequest{
				CustomerId:  customerID.String(),
				Street1:     "123 Main St",
				City:        "New York",
				State:       "NY",
				PostalCode:  "10001",
				Country:     "US",
				AddressType: "Physical",
				IsPrimary:   true,
			},
			wantErr: codes.OK,
		},
		{
			name: "non-existing customer",
			req: &customerpb.AddAddressRequest{
				CustomerId: uuid.New().String(),
				Street1:    "123 Main St",
				City:       "New York",
				State:      "NY",
				PostalCode: "10001",
				Country:    "US",
			},
			wantErr: codes.NotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := svc.AddAddress(ctx, tt.req)
			if tt.wantErr == codes.OK {
				if err != nil {
					t.Errorf("AddAddress() unexpected error: %v", err)
				}
				if resp == nil || resp.Address == nil {
					t.Error("AddAddress() expected address in response")
				}
			} else {
				if err == nil {
					t.Error("AddAddress() expected error, got nil")
					return
				}
				st, ok := status.FromError(err)
				if !ok {
					t.Errorf("AddAddress() error is not a gRPC status error")
					return
				}
				if st.Code() != tt.wantErr {
					t.Errorf("AddAddress() got code %v, want %v", st.Code(), tt.wantErr)
				}
			}
		})
	}
}

func TestCustomerService_AddDocument(t *testing.T) {
	repo := NewMockRepository()
	svc := NewCustomerService(repo)
	ctx := context.Background()

	// Create a customer first
	customerID := uuid.New()
	repo.customers[customerID] = &models.Customer{
		ID:             customerID,
		CustomerNumber: "CUST-123",
		FirstName:      "John",
		LastName:       "Doe",
		Email:          "john@example.com",
		Status:         models.CustomerStatusPending,
		CreatedAt:      time.Now(),
		Version:        1,
	}

	tests := []struct {
		name    string
		req     *customerpb.AddDocumentRequest
		wantErr codes.Code
	}{
		{
			name: "valid document",
			req: &customerpb.AddDocumentRequest{
				CustomerId:       customerID.String(),
				DocumentType:     "Passport",
				DocumentNumber:   "AB1234567",
				IssuingAuthority: "US State Dept",
				IssuingCountry:   "US",
				IssueDate:        timestamppb.New(time.Now().AddDate(-2, 0, 0)),
				ExpiryDate:       timestamppb.New(time.Now().AddDate(5, 0, 0)),
			},
			wantErr: codes.OK,
		},
		{
			name: "expired document",
			req: &customerpb.AddDocumentRequest{
				CustomerId:     customerID.String(),
				DocumentType:   "Passport",
				DocumentNumber: "AB1234567",
				IssuingCountry: "US",
				IssueDate:      timestamppb.New(time.Now().AddDate(-5, 0, 0)),
				ExpiryDate:     timestamppb.New(time.Now().AddDate(-1, 0, 0)),
			},
			wantErr: codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := svc.AddDocument(ctx, tt.req)
			if tt.wantErr == codes.OK {
				if err != nil {
					t.Errorf("AddDocument() unexpected error: %v", err)
				}
				if resp == nil || resp.Document == nil {
					t.Error("AddDocument() expected document in response")
				}
			} else {
				if err == nil {
					t.Error("AddDocument() expected error, got nil")
					return
				}
				st, ok := status.FromError(err)
				if !ok {
					t.Errorf("AddDocument() error is not a gRPC status error")
					return
				}
				if st.Code() != tt.wantErr {
					t.Errorf("AddDocument() got code %v, want %v", st.Code(), tt.wantErr)
				}
			}
		})
	}
}

func TestCustomerService_SearchCustomers(t *testing.T) {
	repo := NewMockRepository()
	svc := NewCustomerService(repo)
	ctx := context.Background()

	// Create some test customers
	customers := []*models.Customer{
		{
			ID:             uuid.New(),
			CustomerNumber: "CUST-001",
			FirstName:      "John",
			LastName:       "Doe",
			Email:          "john@example.com",
			Status:         models.CustomerStatusActive,
			CreatedAt:      time.Now(),
		},
		{
			ID:             uuid.New(),
			CustomerNumber: "CUST-002",
			FirstName:      "Jane",
			LastName:       "Smith",
			Email:          "jane@example.com",
			Status:         models.CustomerStatusActive,
			CreatedAt:      time.Now(),
		},
	}
	for _, c := range customers {
		repo.customers[c.ID] = c
	}

	tests := []struct {
		name      string
		req       *customerpb.SearchCustomersRequest
		wantCount int
		wantErr   bool
	}{
		{
			name: "search by first name",
			req: &customerpb.SearchCustomersRequest{
				FirstName: "John",
			},
			wantCount: 1,
			wantErr:   false,
		},
		{
			name: "search all active customers",
			req: &customerpb.SearchCustomersRequest{
				Status: "Active",
			},
			wantCount: 2,
			wantErr:   false,
		},
		{
			name: "search with limit",
			req: &customerpb.SearchCustomersRequest{
				Limit: 1,
			},
			wantCount: 1,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := svc.SearchCustomers(ctx, tt.req)
			if tt.wantErr {
				if err == nil {
					t.Error("SearchCustomers() expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("SearchCustomers() unexpected error: %v", err)
				}
				if resp == nil {
					t.Error("SearchCustomers() expected response")
				}
				if len(resp.Customers) != tt.wantCount {
					t.Errorf("SearchCustomers() got %d customers, want %d", len(resp.Customers), tt.wantCount)
				}
			}
		})
	}
}

func TestCustomerService_GetCustomerFullProfile(t *testing.T) {
	repo := NewMockRepository()
	svc := NewCustomerService(repo)
	ctx := context.Background()

	// Create a customer with addresses and documents
	customerID := uuid.New()
	customer := &models.Customer{
		ID:             customerID,
		CustomerNumber: "CUST-123",
		FirstName:      "John",
		LastName:       "Doe",
		Email:          "john@example.com",
		Status:         models.CustomerStatusActive,
		CreatedAt:      time.Now(),
	}
	repo.customers[customerID] = customer

	// Add address
	addressID := uuid.New()
	repo.addresses[customerID] = []*models.Address{
		{
			ID:         addressID,
			CustomerID: customerID,
			Street1:    "123 Main St",
			City:       "New York",
			State:      "NY",
			Country:    "US",
			IsPrimary:  true,
			ValidFrom:  time.Now(),
		},
	}

	// Add document
	docID := uuid.New()
	repo.documents[customerID] = []*models.CustomerDocument{
		{
			ID:                 docID,
			CustomerID:         customerID,
			DocumentType:       models.DocumentTypePassport,
			DocumentNumber:     "AB1234567",
			IssuingCountry:     "US",
			IssueDate:          time.Now().AddDate(-2, 0, 0),
			ExpiryDate:         time.Now().AddDate(5, 0, 0),
			VerificationStatus: models.VerificationStatusPending,
		},
	}

	resp, err := svc.GetCustomerFullProfile(ctx, &customerpb.GetCustomerRequest{Id: customerID.String()})
	if err != nil {
		t.Fatalf("GetCustomerFullProfile() error: %v", err)
	}

	if resp.Customer == nil {
		t.Error("GetCustomerFullProfile() expected customer in response")
	}
	if len(resp.Addresses) != 1 {
		t.Errorf("GetCustomerFullProfile() got %d addresses, want 1", len(resp.Addresses))
	}
	if len(resp.Documents) != 1 {
		t.Errorf("GetCustomerFullProfile() got %d documents, want 1", len(resp.Documents))
	}
}
