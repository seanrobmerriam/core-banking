package models

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCustomerStatus_Value(t *testing.T) {
	tests := []struct {
		name   string
		status CustomerStatus
		want   string
	}{
		{"Pending", CustomerStatusPending, "Pending"},
		{"Active", CustomerStatusActive, "Active"},
		{"Inactive", CustomerStatusInactive, "Inactive"},
		{"Suspended", CustomerStatusSuspended, "Suspended"},
		{"Closed", CustomerStatusClosed, "Closed"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.status.Value()
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestCustomerStatus_Scan(t *testing.T) {
	tests := []struct {
		name    string
		input   interface{}
		want    CustomerStatus
		wantErr bool
	}{
		{"valid Pending", "Pending", CustomerStatusPending, false},
		{"valid Active", "Active", CustomerStatusActive, false},
		{"nil input", nil, CustomerStatusPending, false},
		{"invalid string", "Invalid", CustomerStatus(""), true},
		{"wrong type", 123, CustomerStatus(""), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got CustomerStatus
			err := got.Scan(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestAddressType_Value(t *testing.T) {
	tests := []struct {
		name     string
		addrType AddressType
		want     string
	}{
		{"Physical", AddressTypePhysical, "Physical"},
		{"Mailing", AddressTypeMailing, "Mailing"},
		{"Business", AddressTypeBusiness, "Business"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.addrType.Value()
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestAddressType_Scan(t *testing.T) {
	tests := []struct {
		name    string
		input   interface{}
		want    AddressType
		wantErr bool
	}{
		{"valid Physical", "Physical", AddressTypePhysical, false},
		{"valid Mailing", "Mailing", AddressTypeMailing, false},
		{"nil input", nil, AddressTypePhysical, false},
		{"invalid string", "Invalid", AddressType(""), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got AddressType
			err := got.Scan(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestDocumentType_Value(t *testing.T) {
	tests := []struct {
		name    string
		docType DocumentType
		want    string
	}{
		{"Passport", DocumentTypePassport, "Passport"},
		{"DriversLicense", DocumentTypeDriversLicense, "DriversLicense"},
		{"NationalID", DocumentTypeNationalID, "NationalID"},
		{"SSN", DocumentTypeSSN, "SSN"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.docType.Value()
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestDocumentType_Scan(t *testing.T) {
	tests := []struct {
		name    string
		input   interface{}
		want    DocumentType
		wantErr bool
	}{
		{"valid Passport", "Passport", DocumentTypePassport, false},
		{"nil input", nil, DocumentTypePassport, false},
		{"invalid string", "Invalid", DocumentType(""), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got DocumentType
			err := got.Scan(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestVerificationStatus_Value(t *testing.T) {
	tests := []struct {
		name   string
		status VerificationStatus
		want   string
	}{
		{"Pending", VerificationStatusPending, "Pending"},
		{"Verified", VerificationStatusVerified, "Verified"},
		{"Expired", VerificationStatusExpired, "Expired"},
		{"Rejected", VerificationStatusRejected, "Rejected"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.status.Value()
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestVerificationStatus_Scan(t *testing.T) {
	tests := []struct {
		name    string
		input   interface{}
		want    VerificationStatus
		wantErr bool
	}{
		{"valid Verified", "Verified", VerificationStatusVerified, false},
		{"nil input", nil, VerificationStatusPending, false},
		{"invalid string", "Invalid", VerificationStatus(""), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got VerificationStatus
			err := got.Scan(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestCustomer_MarshalJSON(t *testing.T) {
	customer := &Customer{
		ID:             uuid.New(),
		CustomerNumber: "CUST-001",
		FirstName:      "John",
		LastName:       "Doe",
		DateOfBirth:    time.Now().AddDate(-30, 0, 0),
		TaxID:          "123-45-6789", // Should be excluded from JSON
		Email:          "john@example.com",
		Phone:          "+1-555-1234",
		Status:         CustomerStatusActive,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		CreatedBy:      uuid.New(),
		Version:        1,
	}

	data, err := json.Marshal(customer)
	require.NoError(t, err)

	var result Customer
	err = json.Unmarshal(data, &result)
	require.NoError(t, err)

	assert.Equal(t, customer.ID, result.ID)
	assert.Equal(t, customer.CustomerNumber, result.CustomerNumber)
	assert.Equal(t, customer.FirstName, result.FirstName)
	assert.Equal(t, customer.LastName, result.LastName)
	assert.Equal(t, customer.Email, result.Email)
	assert.Equal(t, customer.Status, result.Status)
	assert.Equal(t, customer.Version, result.Version)

	// TaxID should be excluded (empty string)
	assert.Empty(t, result.TaxID)
}

func TestAddress_MarshalJSON(t *testing.T) {
	street2 := "Apt 4B"
	address := &Address{
		ID:          uuid.New(),
		CustomerID:  uuid.New(),
		AddressType: AddressTypePhysical,
		Street1:     "123 Main St",
		Street2:     &street2,
		City:        "New York",
		State:       "NY",
		PostalCode:  "10001",
		Country:     "USA",
		IsPrimary:   true,
		ValidFrom:   time.Now(),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	data, err := json.Marshal(address)
	require.NoError(t, err)

	var result Address
	err = json.Unmarshal(data, &result)
	require.NoError(t, err)

	assert.Equal(t, address.ID, result.ID)
	assert.Equal(t, address.CustomerID, result.CustomerID)
	assert.Equal(t, address.AddressType, result.AddressType)
	assert.Equal(t, address.Street1, result.Street1)
	assert.NotNil(t, result.Street2)
	assert.Equal(t, *address.Street2, *result.Street2)
	assert.Equal(t, address.City, result.City)
	assert.True(t, result.IsPrimary)
}

func TestCustomerDocument_MarshalJSON(t *testing.T) {
	doc := &CustomerDocument{
		ID:                 uuid.New(),
		CustomerID:         uuid.New(),
		DocumentType:       DocumentTypePassport,
		DocumentNumber:     "AB1234567", // Should be excluded from JSON
		IssuingAuthority:   "State Dept",
		IssuingCountry:     "USA",
		IssueDate:          time.Now().AddDate(-5, 0, 0),
		ExpiryDate:         time.Now().AddDate(5, 0, 0),
		VerificationStatus: VerificationStatusVerified,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	data, err := json.Marshal(doc)
	require.NoError(t, err)

	var result CustomerDocument
	err = json.Unmarshal(data, &result)
	require.NoError(t, err)

	assert.Equal(t, doc.ID, result.ID)
	assert.Equal(t, doc.CustomerID, result.CustomerID)
	assert.Equal(t, doc.DocumentType, result.DocumentType)
	assert.Equal(t, doc.IssuingAuthority, result.IssuingAuthority)
	assert.Equal(t, doc.IssuingCountry, result.IssuingCountry)
	assert.Equal(t, doc.VerificationStatus, result.VerificationStatus)

	// DocumentNumber should be excluded (empty string)
	assert.Empty(t, result.DocumentNumber)
}

func TestNullTime_Value(t *testing.T) {
	t.Run("valid time", func(t *testing.T) {
		now := time.Now()
		nt := NullTime{Time: now, Valid: true}

		got, err := nt.Value()
		require.NoError(t, err)
		assert.Equal(t, now, got)
	})

	t.Run("null time", func(t *testing.T) {
		nt := NullTime{Valid: false}

		got, err := nt.Value()
		require.NoError(t, err)
		assert.Nil(t, got)
	})
}

func TestNullTime_Scan(t *testing.T) {
	t.Run("scan time", func(t *testing.T) {
		now := time.Now()
		var nt NullTime
		err := nt.Scan(now)
		require.NoError(t, err)
		assert.True(t, nt.Valid)
		assert.Equal(t, now, nt.Time)
	})

	t.Run("scan nil", func(t *testing.T) {
		var nt NullTime
		err := nt.Scan(nil)
		require.NoError(t, err)
		assert.False(t, nt.Valid)
	})

	t.Run("scan invalid type", func(t *testing.T) {
		var nt NullTime
		err := nt.Scan("not a time")
		assert.Error(t, err)
	})
}
