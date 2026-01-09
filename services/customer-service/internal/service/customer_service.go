package service

import (
	"context"
	"fmt"
	"time"

	customerpb "github.com/core-banking/services/customer-service/internal/proto/customerpb"
	"github.com/core-banking/services/customer-service/internal/repository"
	"github.com/core-banking/services/customer-service/internal/validation"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/core-banking/services/customer-service/internal/models"
)

// CustomerService handles customer business logic
type CustomerService struct {
	customerpb.UnimplementedCustomerServiceServer
	repo      repository.CustomerRepository
	validator *validation.Validator
}

// NewCustomerService creates a new CustomerService instance
func NewCustomerService(repo repository.CustomerRepository) *CustomerService {
	return &CustomerService{
		repo:      repo,
		validator: validation.NewValidator(),
	}
}

// CreateCustomer creates a new customer with validation
func (s *CustomerService) CreateCustomer(ctx context.Context, req *customerpb.CreateCustomerRequest) (*customerpb.CreateCustomerResponse, error) {
	// Validate request
	if errs := s.validator.ValidateCustomerCreate(req); len(errs) > 0 {
		return nil, status.Errorf(codes.InvalidArgument, "%s", errs)
	}

	// Generate customer number
	customerNumber := generateCustomerNumber()

	// Parse UUID for created_by
	var createdByUUID uuid.UUID
	if req.GetCreatedBy() != "" {
		var err error
		createdByUUID, err = uuid.Parse(req.GetCreatedBy())
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid created_by UUID: %v", err)
		}
	}

	// Create customer model
	customer := &models.Customer{
		ID:             uuid.New(),
		CustomerNumber: customerNumber,
		FirstName:      req.GetFirstName(),
		MiddleName:     stringPtr(req.GetMiddleName()),
		LastName:       req.GetLastName(),
		DateOfBirth:    req.GetDateOfBirth().AsTime(),
		TaxID:          req.GetTaxId(),
		Email:          req.GetEmail(),
		Phone:          req.GetPhone(),
		Status:         models.CustomerStatusPending,
		CreatedBy:      createdByUUID,
	}

	// Save to repository
	if err := s.repo.CreateCustomer(ctx, customer); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create customer: %v", err)
	}

	// Return response
	return &customerpb.CreateCustomerResponse{
		Customer: modelToProto(customer),
	}, nil
}

// GetCustomer retrieves a customer by ID
func (s *CustomerService) GetCustomer(ctx context.Context, req *customerpb.GetCustomerRequest) (*customerpb.GetCustomerResponse, error) {
	if req.GetId() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "id is required")
	}

	customerID, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid customer id: %v", err)
	}

	customer, err := s.repo.GetCustomerByID(ctx, customerID)
	if err != nil {
		if err == repository.ErrNotFound {
			return nil, status.Errorf(codes.NotFound, "customer not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to get customer: %v", err)
	}

	return &customerpb.GetCustomerResponse{
		Customer: modelToProto(customer),
	}, nil
}

// UpdateCustomer updates an existing customer
func (s *CustomerService) UpdateCustomer(ctx context.Context, req *customerpb.UpdateCustomerRequest) (*customerpb.UpdateCustomerResponse, error) {
	// Validate request
	if errs := s.validator.ValidateCustomerUpdate(req); len(errs) > 0 {
		return nil, status.Errorf(codes.InvalidArgument, "%s", errs)
	}

	customerID, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid customer id: %v", err)
	}

	// Get existing customer
	customer, err := s.repo.GetCustomerByID(ctx, customerID)
	if err != nil {
		if err == repository.ErrNotFound {
			return nil, status.Errorf(codes.NotFound, "customer not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to get customer: %v", err)
	}

	// Update fields if provided
	if req.GetFirstName() != "" {
		customer.FirstName = req.GetFirstName()
	}
	if req.GetMiddleName() != "" {
		customer.MiddleName = stringPtr(req.GetMiddleName())
	}
	if req.GetLastName() != "" {
		customer.LastName = req.GetLastName()
	}
	if req.GetDateOfBirth() != nil {
		customer.DateOfBirth = req.GetDateOfBirth().AsTime()
	}
	if req.GetTaxId() != "" {
		customer.TaxID = req.GetTaxId()
	}
	if req.GetEmail() != "" {
		customer.Email = req.GetEmail()
	}
	if req.GetPhone() != "" {
		customer.Phone = req.GetPhone()
	}

	// Parse updated_by
	if req.GetUpdatedBy() != "" {
		updatedByUUID, err := uuid.Parse(req.GetUpdatedBy())
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid updated_by UUID: %v", err)
		}
		customer.UpdatedBy = &updatedByUUID
	}

	// Update in repository
	if err := s.repo.UpdateCustomer(ctx, customer); err != nil {
		if _, ok := err.(*repository.ErrOptimisticLock); ok {
			return nil, status.Errorf(codes.Aborted, "customer was modified by another process")
		}
		return nil, status.Errorf(codes.Internal, "failed to update customer: %v", err)
	}

	return &customerpb.UpdateCustomerResponse{
		Customer: modelToProto(customer),
	}, nil
}

// SearchCustomers searches for customers based on filters
func (s *CustomerService) SearchCustomers(ctx context.Context, req *customerpb.SearchCustomersRequest) (*customerpb.SearchCustomersResponse, error) {
	// Validate request
	if errs := s.validator.ValidateSearchFilters(req); len(errs) > 0 {
		return nil, status.Errorf(codes.InvalidArgument, "%s", errs)
	}

	// Build search filters
	filters := models.SearchFilters{
		FirstName: req.GetFirstName(),
		LastName:  req.GetLastName(),
		Email:     req.GetEmail(),
		Phone:     req.GetPhone(),
		Status:    models.CustomerStatus(req.GetStatus()),
		Limit:     int(req.GetLimit()),
		Offset:    int(req.GetOffset()),
	}

	if req.GetFromDate() != nil {
		fromDate := req.GetFromDate().AsTime()
		filters.FromDate = &fromDate
	}
	if req.GetToDate() != nil {
		toDate := req.GetToDate().AsTime()
		filters.ToDate = &toDate
	}

	// Search in repository
	customers, err := s.repo.SearchCustomers(ctx, filters)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to search customers: %v", err)
	}

	// Convert to proto
	protoCustomers := make([]*customerpb.Customer, len(customers))
	for i, c := range customers {
		protoCustomers[i] = modelToProto(c)
	}

	return &customerpb.SearchCustomersResponse{
		Customers: protoCustomers,
		Total:     int32(len(protoCustomers)),
	}, nil
}

// AddAddress adds a new address to a customer
func (s *CustomerService) AddAddress(ctx context.Context, req *customerpb.AddAddressRequest) (*customerpb.AddAddressResponse, error) {
	// Validate request
	if errs := s.validator.ValidateAddress(req); len(errs) > 0 {
		return nil, status.Errorf(codes.InvalidArgument, "%s", errs)
	}

	customerID, err := uuid.Parse(req.GetCustomerId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid customer id: %v", err)
	}

	// Verify customer exists
	_, err = s.repo.GetCustomerByID(ctx, customerID)
	if err != nil {
		if err == repository.ErrNotFound {
			return nil, status.Errorf(codes.NotFound, "customer not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to get customer: %v", err)
	}

	// Handle primary address logic
	addresses, err := s.repo.GetCustomerAddresses(ctx, customerID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get addresses: %v", err)
	}

	// If setting as primary, unset other primary addresses
	if req.GetIsPrimary() {
		for _, addr := range addresses {
			if addr.IsPrimary {
				addr.IsPrimary = false
				if err := s.repo.UpdateAddress(ctx, addr); err != nil {
					return nil, status.Errorf(codes.Internal, "failed to update existing primary address: %v", err)
				}
			}
		}
	} else {
		// If no primary address exists and not setting one, set this as primary
		hasPrimary := false
		for _, addr := range addresses {
			if addr.IsPrimary {
				hasPrimary = true
				break
			}
		}
		if !hasPrimary {
			// Auto-set as primary if it's the first address
			req.IsPrimary = true
		}
	}

	// Create address model
	address := &models.Address{
		ID:          uuid.New(),
		CustomerID:  customerID,
		AddressType: models.AddressType(req.GetAddressType()),
		Street1:     req.GetStreet1(),
		Street2:     stringPtr(req.GetStreet2()),
		City:        req.GetCity(),
		State:       req.GetState(),
		PostalCode:  req.GetPostalCode(),
		Country:     req.GetCountry(),
		IsPrimary:   req.GetIsPrimary(),
		ValidFrom:   time.Now().UTC(),
	}

	if req.GetValidFrom() != nil {
		address.ValidFrom = req.GetValidFrom().AsTime()
	}
	if req.GetValidTo() != nil {
		validTo := req.GetValidTo().AsTime()
		address.ValidTo = &validTo
	}

	// Save address
	if err := s.repo.AddAddress(ctx, address); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to add address: %v", err)
	}

	return &customerpb.AddAddressResponse{
		Address: addressModelToProto(address),
	}, nil
}

// AddDocument adds a new document to a customer
func (s *CustomerService) AddDocument(ctx context.Context, req *customerpb.AddDocumentRequest) (*customerpb.AddDocumentResponse, error) {
	// Validate request
	if errs := s.validator.ValidateDocument(req); len(errs) > 0 {
		return nil, status.Errorf(codes.InvalidArgument, "%s", errs)
	}

	customerID, err := uuid.Parse(req.GetCustomerId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid customer id: %v", err)
	}

	// Verify customer exists
	customer, err := s.repo.GetCustomerByID(ctx, customerID)
	if err != nil {
		if err == repository.ErrNotFound {
			return nil, status.Errorf(codes.NotFound, "customer not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to get customer: %v", err)
	}

	// Create document model
	doc := &models.CustomerDocument{
		ID:                 uuid.New(),
		CustomerID:         customerID,
		DocumentType:       models.DocumentType(req.GetDocumentType()),
		DocumentNumber:     req.GetDocumentNumber(),
		IssuingAuthority:   req.GetIssuingAuthority(),
		IssuingCountry:     req.GetIssuingCountry(),
		IssueDate:          req.GetIssueDate().AsTime(),
		ExpiryDate:         req.GetExpiryDate().AsTime(),
		VerificationStatus: models.VerificationStatusPending,
	}

	// Save document
	if err := s.repo.AddDocument(ctx, doc); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to add document: %v", err)
	}

	// Check if customer can be activated (has at least one verified identity document)
	if customer.Status == models.CustomerStatusPending {
		docs, err := s.repo.GetCustomerDocuments(ctx, customerID)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to get documents: %v", err)
		}

		hasVerifiedIdentity := false
		for _, d := range docs {
			if isIdentityDocument(d.DocumentType) && d.VerificationStatus == models.VerificationStatusVerified {
				hasVerifiedIdentity = true
				break
			}
		}

		if hasVerifiedIdentity {
			// Auto-activate customer
			customer.Status = models.CustomerStatusActive
			if err := s.repo.UpdateCustomer(ctx, customer); err != nil {
				return nil, status.Errorf(codes.Internal, "failed to update customer status: %v", err)
			}
		}
	}

	return &customerpb.AddDocumentResponse{
		Document: documentModelToProto(doc),
	}, nil
}

// UpdateCustomerStatus updates the status of a customer
func (s *CustomerService) UpdateCustomerStatus(ctx context.Context, req *customerpb.UpdateCustomerStatusRequest) (*customerpb.UpdateCustomerStatusResponse, error) {
	if req.GetId() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "id is required")
	}

	if req.GetNewStatus() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "new_status is required")
	}

	if req.GetReason() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "reason is required")
	}

	customerID, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid customer id: %v", err)
	}

	// Get existing customer
	customer, err := s.repo.GetCustomerByID(ctx, customerID)
	if err != nil {
		if err == repository.ErrNotFound {
			return nil, status.Errorf(codes.NotFound, "customer not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to get customer: %v", err)
	}

	// Validate status transition
	currentStatus := customer.Status
	newStatus := models.CustomerStatus(req.GetNewStatus())

	if err := s.validator.ValidateStatusTransition(currentStatus, newStatus); err != nil {
		return nil, status.Errorf(codes.FailedPrecondition, "%v", err)
	}

	// Update status
	previousStatus := customer.Status
	customer.Status = newStatus

	var changedByUUID uuid.UUID
	if req.GetChangedBy() != "" {
		changedByUUID, err = uuid.Parse(req.GetChangedBy())
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid changed_by UUID: %v", err)
		}
		customer.UpdatedBy = &changedByUUID
	}

	// Update customer
	if err := s.repo.UpdateCustomer(ctx, customer); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update customer status: %v", err)
	}

	// Create status change record
	statusChange := &models.StatusChange{
		ID:             uuid.New(),
		CustomerID:     customerID,
		PreviousStatus: previousStatus,
		NewStatus:      newStatus,
		Reason:         req.GetReason(),
		ChangedBy:      changedByUUID,
		ChangedAt:      time.Now().UTC(),
	}

	// Log status change (in a real system, save to database)
	fmt.Printf("Status change: %s -> %s, Reason: %s\n", previousStatus, newStatus, req.GetReason())

	return &customerpb.UpdateCustomerStatusResponse{
		Customer:     modelToProto(customer),
		StatusChange: statusChangeModelToProto(statusChange),
	}, nil
}

// GetCustomerFullProfile retrieves the complete customer profile
func (s *CustomerService) GetCustomerFullProfile(ctx context.Context, req *customerpb.GetCustomerRequest) (*customerpb.CustomerFullProfileResponse, error) {
	if req.GetId() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "id is required")
	}

	customerID, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid customer id: %v", err)
	}

	// Get customer
	customer, err := s.repo.GetCustomerByID(ctx, customerID)
	if err != nil {
		if err == repository.ErrNotFound {
			return nil, status.Errorf(codes.NotFound, "customer not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to get customer: %v", err)
	}

	// Get addresses
	addresses, err := s.repo.GetCustomerAddresses(ctx, customerID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get addresses: %v", err)
	}

	// Get documents
	documents, err := s.repo.GetCustomerDocuments(ctx, customerID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get documents: %v", err)
	}

	// Convert to proto
	protoAddresses := make([]*customerpb.Address, len(addresses))
	for i, a := range addresses {
		protoAddresses[i] = addressModelToProto(a)
	}

	protoDocuments := make([]*customerpb.Document, len(documents))
	for i, d := range documents {
		protoDocuments[i] = documentModelToProto(d)
	}

	return &customerpb.CustomerFullProfileResponse{
		Customer:      modelToProto(customer),
		Addresses:     protoAddresses,
		Documents:     protoDocuments,
		StatusHistory: []*customerpb.StatusChange{},
	}, nil
}

// Helper functions

func stringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func modelToProto(c *models.Customer) *customerpb.Customer {
	customer := &customerpb.Customer{
		Id:             c.ID.String(),
		CustomerNumber: c.CustomerNumber,
		FirstName:      c.FirstName,
		LastName:       c.LastName,
		Email:          c.Email,
		Phone:          c.Phone,
		Status:         string(c.Status),
		Version:        int32(c.Version),
	}

	if c.MiddleName != nil {
		customer.MiddleName = *c.MiddleName
	}

	if !c.DateOfBirth.IsZero() {
		customer.DateOfBirth = timestamppb.New(c.DateOfBirth)
	}

	customer.CreatedAt = timestamppb.New(c.CreatedAt)
	customer.UpdatedAt = timestamppb.New(c.UpdatedAt)
	customer.CreatedBy = c.CreatedBy.String()

	if c.UpdatedBy != nil {
		customer.UpdatedBy = c.UpdatedBy.String()
	}

	return customer
}

func addressModelToProto(a *models.Address) *customerpb.Address {
	addr := &customerpb.Address{
		Id:          a.ID.String(),
		CustomerId:  a.CustomerID.String(),
		AddressType: string(a.AddressType),
		Street1:     a.Street1,
		City:        a.City,
		State:       a.State,
		PostalCode:  a.PostalCode,
		Country:     a.Country,
		IsPrimary:   a.IsPrimary,
	}

	if a.Street2 != nil {
		addr.Street2 = *a.Street2
	}

	addr.ValidFrom = timestamppb.New(a.ValidFrom)
	addr.CreatedAt = timestamppb.New(a.CreatedAt)
	addr.UpdatedAt = timestamppb.New(a.UpdatedAt)

	if a.ValidTo != nil {
		addr.ValidTo = timestamppb.New(*a.ValidTo)
	}

	return addr
}

func documentModelToProto(d *models.CustomerDocument) *customerpb.Document {
	doc := &customerpb.Document{
		Id:                 d.ID.String(),
		CustomerId:         d.CustomerID.String(),
		DocumentType:       string(d.DocumentType),
		DocumentNumber:     d.DocumentNumber,
		IssuingAuthority:   d.IssuingAuthority,
		IssuingCountry:     d.IssuingCountry,
		VerificationStatus: string(d.VerificationStatus),
	}

	doc.IssueDate = timestamppb.New(d.IssueDate)
	doc.ExpiryDate = timestamppb.New(d.ExpiryDate)
	doc.CreatedAt = timestamppb.New(d.CreatedAt)
	doc.UpdatedAt = timestamppb.New(d.UpdatedAt)

	if d.VerifiedAt != nil {
		doc.VerifiedAt = timestamppb.New(*d.VerifiedAt)
	}
	if d.VerifiedBy != nil {
		doc.VerifiedBy = d.VerifiedBy.String()
	}

	return doc
}

func statusChangeModelToProto(sc *models.StatusChange) *customerpb.StatusChange {
	return &customerpb.StatusChange{
		Id:             sc.ID.String(),
		CustomerId:     sc.CustomerID.String(),
		PreviousStatus: string(sc.PreviousStatus),
		NewStatus:      string(sc.NewStatus),
		Reason:         sc.Reason,
		ChangedBy:      sc.ChangedBy.String(),
		ChangedAt:      timestamppb.New(sc.ChangedAt),
	}
}

func generateCustomerNumber() string {
	return fmt.Sprintf("CUST-%d", time.Now().UnixNano())
}

func isIdentityDocument(docType models.DocumentType) bool {
	switch docType {
	case models.DocumentTypePassport, models.DocumentTypeDriversLicense,
		models.DocumentTypeNationalID, models.DocumentTypeSSN:
		return true
	default:
		return false
	}
}
