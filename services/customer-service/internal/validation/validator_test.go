package validation

import (
	"testing"
	"time"

	"github.com/core-banking/services/customer-service/internal/models"
	customerpb "github.com/core-banking/services/customer-service/internal/proto/customerpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestValidateCustomerCreate(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name    string
		req     *customerpb.CreateCustomerRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid request",
			req: &customerpb.CreateCustomerRequest{
				FirstName:   "John",
				LastName:    "Doe",
				Email:       "john.doe@example.com",
				Phone:       "+1234567890",
				DateOfBirth: timestamppb.New(time.Now().AddDate(-25, 0, 0)),
				TaxId:       "123-45-6789",
				CreatedBy:   "user-uuid",
			},
			wantErr: false,
		},
		{
			name: "missing first name",
			req: &customerpb.CreateCustomerRequest{
				LastName:    "Doe",
				Email:       "john.doe@example.com",
				DateOfBirth: timestamppb.New(time.Now().AddDate(-25, 0, 0)),
			},
			wantErr: true,
			errMsg:  "first_name: is required",
		},
		{
			name: "missing last name",
			req: &customerpb.CreateCustomerRequest{
				FirstName:   "John",
				Email:       "john.doe@example.com",
				DateOfBirth: timestamppb.New(time.Now().AddDate(-25, 0, 0)),
			},
			wantErr: true,
			errMsg:  "last_name: is required",
		},
		{
			name: "invalid email format",
			req: &customerpb.CreateCustomerRequest{
				FirstName:   "John",
				LastName:    "Doe",
				Email:       "invalid-email",
				DateOfBirth: timestamppb.New(time.Now().AddDate(-25, 0, 0)),
			},
			wantErr: true,
			errMsg:  "email: is invalid format",
		},
		{
			name: "invalid phone format",
			req: &customerpb.CreateCustomerRequest{
				FirstName:   "John",
				LastName:    "Doe",
				Email:       "john.doe@example.com",
				Phone:       "invalid-phone",
				DateOfBirth: timestamppb.New(time.Now().AddDate(-25, 0, 0)),
			},
			wantErr: true,
			errMsg:  "phone: is invalid format",
		},
		{
			name: "underage customer",
			req: &customerpb.CreateCustomerRequest{
				FirstName:   "John",
				LastName:    "Doe",
				Email:       "john.doe@example.com",
				DateOfBirth: timestamppb.New(time.Now().AddDate(-10, 0, 0)),
			},
			wantErr: true,
			errMsg:  "date_of_birth: customer must be at least 18 years old",
		},
		{
			name: "first name too short",
			req: &customerpb.CreateCustomerRequest{
				FirstName:   "J",
				LastName:    "Doe",
				Email:       "john.doe@example.com",
				DateOfBirth: timestamppb.New(time.Now().AddDate(-25, 0, 0)),
			},
			wantErr: true,
			errMsg:  "first_name: must be between 2 and 100 characters",
		},
		{
			name: "valid phone - empty is allowed",
			req: &customerpb.CreateCustomerRequest{
				FirstName:   "John",
				LastName:    "Doe",
				Email:       "john.doe@example.com",
				Phone:       "",
				DateOfBirth: timestamppb.New(time.Now().AddDate(-25, 0, 0)),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := validator.ValidateCustomerCreate(tt.req)
			if tt.wantErr {
				if len(errs) == 0 {
					t.Errorf("ValidateCustomerCreate() expected error, got none")
				}
			} else {
				if len(errs) > 0 {
					t.Errorf("ValidateCustomerCreate() unexpected error: %v", errs)
				}
			}
		})
	}
}

func TestValidateAddress(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name    string
		req     *customerpb.AddAddressRequest
		wantErr bool
	}{
		{
			name: "valid address",
			req: &customerpb.AddAddressRequest{
				CustomerId:  "customer-uuid",
				Street1:     "123 Main St",
				City:        "New York",
				State:       "NY",
				PostalCode:  "10001",
				Country:     "US",
				AddressType: "Physical",
				IsPrimary:   true,
			},
			wantErr: false,
		},
		{
			name: "missing street1",
			req: &customerpb.AddAddressRequest{
				CustomerId: "customer-uuid",
				City:       "New York",
				State:      "NY",
				PostalCode: "10001",
				Country:    "US",
			},
			wantErr: true,
		},
		{
			name: "invalid country code",
			req: &customerpb.AddAddressRequest{
				CustomerId: "customer-uuid",
				Street1:    "123 Main St",
				City:       "New York",
				State:      "NY",
				PostalCode: "10001",
				Country:    "USA",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := validator.ValidateAddress(tt.req)
			if tt.wantErr && len(errs) == 0 {
				t.Errorf("ValidateAddress() expected error, got none")
			}
			if !tt.wantErr && len(errs) > 0 {
				t.Errorf("ValidateAddress() unexpected error: %v", errs)
			}
		})
	}
}

func TestValidateDocument(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name    string
		req     *customerpb.AddDocumentRequest
		wantErr bool
	}{
		{
			name: "valid document",
			req: &customerpb.AddDocumentRequest{
				CustomerId:       "customer-uuid",
				DocumentType:     "Passport",
				DocumentNumber:   "AB1234567",
				IssuingAuthority: "US Department of State",
				IssuingCountry:   "US",
				IssueDate:        timestamppb.New(time.Now().AddDate(-2, 0, 0)),
				ExpiryDate:       timestamppb.New(time.Now().AddDate(5, 0, 0)),
			},
			wantErr: false,
		},
		{
			name: "expired document",
			req: &customerpb.AddDocumentRequest{
				CustomerId:     "customer-uuid",
				DocumentType:   "Passport",
				DocumentNumber: "AB1234567",
				IssuingCountry: "US",
				IssueDate:      timestamppb.New(time.Now().AddDate(-5, 0, 0)),
				ExpiryDate:     timestamppb.New(time.Now().AddDate(-1, 0, 0)),
			},
			wantErr: true,
		},
		{
			name: "expiry before issue",
			req: &customerpb.AddDocumentRequest{
				CustomerId:     "customer-uuid",
				DocumentType:   "Passport",
				DocumentNumber: "AB1234567",
				IssuingCountry: "US",
				IssueDate:      timestamppb.New(time.Now()),
				ExpiryDate:     timestamppb.New(time.Now().AddDate(-1, 0, 0)),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := validator.ValidateDocument(tt.req)
			if tt.wantErr && len(errs) == 0 {
				t.Errorf("ValidateDocument() expected error, got none")
			}
			if !tt.wantErr && len(errs) > 0 {
				t.Errorf("ValidateDocument() unexpected error: %v", errs)
			}
		})
	}
}

func TestValidateStatusTransition(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name          string
		currentStatus models.CustomerStatus
		newStatus     models.CustomerStatus
		wantErr       bool
	}{
		{
			name:          "pending to active",
			currentStatus: models.CustomerStatusPending,
			newStatus:     models.CustomerStatusActive,
			wantErr:       false,
		},
		{
			name:          "pending to closed",
			currentStatus: models.CustomerStatusPending,
			newStatus:     models.CustomerStatusClosed,
			wantErr:       false,
		},
		{
			name:          "active to inactive",
			currentStatus: models.CustomerStatusActive,
			newStatus:     models.CustomerStatusInactive,
			wantErr:       false,
		},
		{
			name:          "active to suspended",
			currentStatus: models.CustomerStatusActive,
			newStatus:     models.CustomerStatusSuspended,
			wantErr:       false,
		},
		{
			name:          "active to closed",
			currentStatus: models.CustomerStatusActive,
			newStatus:     models.CustomerStatusClosed,
			wantErr:       false,
		},
		{
			name:          "inactive to active",
			currentStatus: models.CustomerStatusInactive,
			newStatus:     models.CustomerStatusActive,
			wantErr:       false,
		},
		{
			name:          "suspended to active",
			currentStatus: models.CustomerStatusSuspended,
			newStatus:     models.CustomerStatusActive,
			wantErr:       false,
		},
		{
			name:          "closed to active - invalid",
			currentStatus: models.CustomerStatusClosed,
			newStatus:     models.CustomerStatusActive,
			wantErr:       true,
		},
		{
			name:          "pending to inactive - invalid",
			currentStatus: models.CustomerStatusPending,
			newStatus:     models.CustomerStatusInactive,
			wantErr:       true,
		},
		{
			name:          "active to pending - invalid",
			currentStatus: models.CustomerStatusActive,
			newStatus:     models.CustomerStatusPending,
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateStatusTransition(tt.currentStatus, tt.newStatus)
			if tt.wantErr && err == nil {
				t.Errorf("ValidateStatusTransition() expected error, got none")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("ValidateStatusTransition() unexpected error: %v", err)
			}
		})
	}
}

func TestValidateSearchFilters(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name    string
		req     *customerpb.SearchCustomersRequest
		wantErr bool
	}{
		{
			name: "valid filters",
			req: &customerpb.SearchCustomersRequest{
				FirstName: "John",
				LastName:  "Doe",
				Email:     "john@example.com",
				Status:    "Active",
				Limit:     10,
				Offset:    0,
			},
			wantErr: false,
		},
		{
			name: "limit too high",
			req: &customerpb.SearchCustomersRequest{
				Limit: 200,
			},
			wantErr: true,
		},
		{
			name: "negative offset",
			req: &customerpb.SearchCustomersRequest{
				Offset: -1,
			},
			wantErr: true,
		},
		{
			name: "invalid status",
			req: &customerpb.SearchCustomersRequest{
				Status: "InvalidStatus",
			},
			wantErr: true,
		},
		{
			name: "invalid email",
			req: &customerpb.SearchCustomersRequest{
				Email: "invalid-email",
			},
			wantErr: true,
		},
		{
			name:    "empty filters - valid",
			req:     &customerpb.SearchCustomersRequest{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := validator.ValidateSearchFilters(tt.req)
			if tt.wantErr && len(errs) == 0 {
				t.Errorf("ValidateSearchFilters() expected error, got none")
			}
			if !tt.wantErr && len(errs) > 0 {
				t.Errorf("ValidateSearchFilters() unexpected error: %v", errs)
			}
		})
	}
}
