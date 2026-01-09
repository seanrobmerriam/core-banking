package repository

import (
	"database/sql"
	"testing"
	"time"

	"github.com/core-banking/services/customer-service/internal/encryption"
	"github.com/core-banking/services/customer-service/internal/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockDB is a mock implementation of DBQuerier for testing
type MockDB struct {
	rows       *sql.Rows
	nextError  error
	execResult sql.Result
}

func setupTestDB(t *testing.T) (*sql.DB, func()) {
	// Skip if not running integration tests
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// For unit tests, we'll use a mock approach
	// In a real scenario, you would connect to a test database
	t.Skip("Integration tests require a running PostgreSQL database")
	return nil, func() {}
}

func setupTestEncryptor(t *testing.T) *encryption.Encryptor {
	key := "12345678901234567890123456789012"
	encryptor, err := encryption.NewEncryptor(key)
	require.NoError(t, err)
	return encryptor
}

func TestCustomerStatus_IsValid(t *testing.T) {
	tests := []struct {
		name   string
		status models.CustomerStatus
		valid  bool
	}{
		{"Pending is valid", models.CustomerStatusPending, true},
		{"Active is valid", models.CustomerStatusActive, true},
		{"Inactive is valid", models.CustomerStatusInactive, true},
		{"Suspended is valid", models.CustomerStatusSuspended, true},
		{"Closed is valid", models.CustomerStatusClosed, true},
		{"Empty is invalid", models.CustomerStatus(""), false},
		{"Unknown is invalid", models.CustomerStatus("Unknown"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.valid, tt.status.IsValid())
		})
	}
}

func TestAddressType_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		addrType models.AddressType
		valid    bool
	}{
		{"Physical is valid", models.AddressTypePhysical, true},
		{"Mailing is valid", models.AddressTypeMailing, true},
		{"Business is valid", models.AddressTypeBusiness, true},
		{"Empty is invalid", models.AddressType(""), false},
		{"Unknown is invalid", models.AddressType("Unknown"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.valid, tt.addrType.IsValid())
		})
	}
}

func TestDocumentType_IsValid(t *testing.T) {
	tests := []struct {
		name    string
		docType models.DocumentType
		valid   bool
	}{
		{"Passport is valid", models.DocumentTypePassport, true},
		{"DriversLicense is valid", models.DocumentTypeDriversLicense, true},
		{"NationalID is valid", models.DocumentTypeNationalID, true},
		{"SSN is valid", models.DocumentTypeSSN, true},
		{"TaxID is valid", models.DocumentTypeTaxID, true},
		{"UtilityBill is valid", models.DocumentTypeUtilityBill, true},
		{"BankStatement is valid", models.DocumentTypeBankStatement, true},
		{"Empty is invalid", models.DocumentType(""), false},
		{"Unknown is invalid", models.DocumentType("Unknown"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.valid, tt.docType.IsValid())
		})
	}
}

func TestVerificationStatus_IsValid(t *testing.T) {
	tests := []struct {
		name   string
		status models.VerificationStatus
		valid  bool
	}{
		{"Pending is valid", models.VerificationStatusPending, true},
		{"Verified is valid", models.VerificationStatusVerified, true},
		{"Expired is valid", models.VerificationStatusExpired, true},
		{"Rejected is valid", models.VerificationStatusRejected, true},
		{"Empty is invalid", models.VerificationStatus(""), false},
		{"Unknown is invalid", models.VerificationStatus("Unknown"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.valid, tt.status.IsValid())
		})
	}
}

func TestCustomer_Model(t *testing.T) {
	t.Run("create customer model", func(t *testing.T) {
		now := time.Now().UTC()
		createdBy := uuid.New()
		customer := &models.Customer{
			ID:             uuid.New(),
			CustomerNumber: "CUST-001",
			FirstName:      "John",
			LastName:       "Doe",
			DateOfBirth:    now.AddDate(-30, 0, 0),
			TaxID:          "123-45-6789",
			Email:          "john.doe@example.com",
			Phone:          "+1-555-123-4567",
			Status:         models.CustomerStatusActive,
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedBy:      createdBy,
			Version:        1,
		}

		assert.NotEmpty(t, customer.ID)
		assert.Equal(t, "CUST-001", customer.CustomerNumber)
		assert.Equal(t, "John", customer.FirstName)
		assert.Equal(t, "Doe", customer.LastName)
		assert.Equal(t, models.CustomerStatusActive, customer.Status)
		assert.Equal(t, 1, customer.Version)
	})

	t.Run("customer with middle name", func(t *testing.T) {
		middleName := "Michael"
		customer := &models.Customer{
			ID:             uuid.New(),
			CustomerNumber: "CUST-002",
			FirstName:      "John",
			MiddleName:     &middleName,
			LastName:       "Doe",
			DateOfBirth:    time.Now().AddDate(-25, 0, 0),
			TaxID:          "987-65-4321",
			Email:          "john.michael.doe@example.com",
			Phone:          "+1-555-987-6543",
			Status:         models.CustomerStatusPending,
			CreatedBy:      uuid.New(),
			Version:        1,
		}

		assert.NotNil(t, customer.MiddleName)
		assert.Equal(t, "Michael", *customer.MiddleName)
	})
}

func TestAddress_Model(t *testing.T) {
	t.Run("create address model", func(t *testing.T) {
		customerID := uuid.New()
		street2 := "Apt 4B"
		validTo := time.Now().AddDate(1, 0, 0)

		address := &models.Address{
			ID:          uuid.New(),
			CustomerID:  customerID,
			AddressType: models.AddressTypePhysical,
			Street1:     "123 Main Street",
			Street2:     &street2,
			City:        "New York",
			State:       "NY",
			PostalCode:  "10001",
			Country:     "USA",
			IsPrimary:   true,
			ValidFrom:   time.Now().UTC(),
			ValidTo:     &validTo,
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
		}

		assert.NotEmpty(t, address.ID)
		assert.Equal(t, customerID, address.CustomerID)
		assert.Equal(t, models.AddressTypePhysical, address.AddressType)
		assert.Equal(t, "123 Main Street", address.Street1)
		assert.NotNil(t, address.Street2)
		assert.Equal(t, "Apt 4B", *address.Street2)
		assert.True(t, address.IsPrimary)
	})
}

func TestCustomerDocument_Model(t *testing.T) {
	t.Run("create document model", func(t *testing.T) {
		customerID := uuid.New()
		verifiedBy := uuid.New()
		verifiedAt := time.Now().UTC()

		doc := &models.CustomerDocument{
			ID:                 uuid.New(),
			CustomerID:         customerID,
			DocumentType:       models.DocumentTypePassport,
			DocumentNumber:     "AB1234567",
			IssuingAuthority:   "US Department of State",
			IssuingCountry:     "USA",
			IssueDate:          time.Now().AddDate(-5, 0, 0),
			ExpiryDate:         time.Now().AddDate(5, 0, 0),
			VerificationStatus: models.VerificationStatusVerified,
			VerifiedAt:         &verifiedAt,
			VerifiedBy:         &verifiedBy,
			CreatedAt:          time.Now().UTC(),
			UpdatedAt:          time.Now().UTC(),
		}

		assert.NotEmpty(t, doc.ID)
		assert.Equal(t, customerID, doc.CustomerID)
		assert.Equal(t, models.DocumentTypePassport, doc.DocumentType)
		assert.Equal(t, "AB1234567", doc.DocumentNumber)
		assert.Equal(t, models.VerificationStatusVerified, doc.VerificationStatus)
		assert.NotNil(t, doc.VerifiedAt)
		assert.NotNil(t, doc.VerifiedBy)
	})
}

func TestSearchFilters_Model(t *testing.T) {
	t.Run("create search filters", func(t *testing.T) {
		fromDate := time.Now().AddDate(-1, 0, 0)
		toDate := time.Now()

		filters := models.SearchFilters{
			FirstName: "John",
			LastName:  "Doe",
			Email:     "john.doe@example.com",
			Phone:     "+1-555",
			Status:    models.CustomerStatusActive,
			FromDate:  &fromDate,
			ToDate:    &toDate,
			Limit:     10,
			Offset:    0,
		}

		assert.Equal(t, "John", filters.FirstName)
		assert.Equal(t, "Doe", filters.LastName)
		assert.Equal(t, models.CustomerStatusActive, filters.Status)
		assert.Equal(t, 10, filters.Limit)
		assert.Equal(t, 0, filters.Offset)
	})
}

func TestErrOptimisticLock(t *testing.T) {
	t.Run("error message", func(t *testing.T) {
		customerID := uuid.New()
		err := &ErrOptimisticLock{
			CustomerID:      customerID,
			ExpectedVersion: 1,
			ActualVersion:   2,
		}

		assert.Contains(t, err.Error(), "optimistic lock error")
		assert.Equal(t, customerID, err.CustomerID)
	})
}

func TestDBQuerier_Interface(t *testing.T) {
	t.Run("DBQuerier interface compliance", func(t *testing.T) {
		// This test ensures that *sql.DB implements DBQuerier
		var _ DBQuerier = (*sql.DB)(nil)
	})

	t.Run("DBQuerier interface compliance for Tx", func(t *testing.T) {
		// This test ensures that *sql.Tx implements DBQuerier
		var _ DBQuerier = (*sql.Tx)(nil)
	})
}

func TestNullTime(t *testing.T) {
	t.Run("null time with valid value", func(t *testing.T) {
		now := time.Now()
		nt := models.NullTime{
			Time:  now,
			Valid: true,
		}

		value, err := nt.Value()
		require.NoError(t, err)
		assert.Equal(t, now, value)
	})

	t.Run("null time with nil value", func(t *testing.T) {
		nt := models.NullTime{
			Time:  time.Time{},
			Valid: false,
		}

		value, err := nt.Value()
		require.NoError(t, err)
		assert.Nil(t, value)
	})

	t.Run("scan null time", func(t *testing.T) {
		nt := &models.NullTime{}
		err := nt.Scan(nil)
		require.NoError(t, err)
		assert.False(t, nt.Valid)
	})

	t.Run("scan valid time", func(t *testing.T) {
		nt := &models.NullTime{}
		now := time.Now()
		err := nt.Scan(now)
		require.NoError(t, err)
		assert.True(t, nt.Valid)
		assert.Equal(t, now, nt.Time)
	})
}
