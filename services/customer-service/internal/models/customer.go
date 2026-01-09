package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

// CustomerStatus represents the status of a customer
type CustomerStatus string

const (
	CustomerStatusPending   CustomerStatus = "Pending"
	CustomerStatusActive    CustomerStatus = "Active"
	CustomerStatusInactive  CustomerStatus = "Inactive"
	CustomerStatusSuspended CustomerStatus = "Suspended"
	CustomerStatusClosed    CustomerStatus = "Closed"
)

// IsValid checks if the status is valid
func (s CustomerStatus) IsValid() bool {
	switch s {
	case CustomerStatusPending, CustomerStatusActive, CustomerStatusInactive,
		CustomerStatusSuspended, CustomerStatusClosed:
		return true
	}
	return false
}

// AddressType represents the type of address
type AddressType string

const (
	AddressTypePhysical AddressType = "Physical"
	AddressTypeMailing  AddressType = "Mailing"
	AddressTypeBusiness AddressType = "Business"
)

// IsValid checks if the address type is valid
func (t AddressType) IsValid() bool {
	switch t {
	case AddressTypePhysical, AddressTypeMailing, AddressTypeBusiness:
		return true
	}
	return false
}

// DocumentType represents the type of customer document
type DocumentType string

const (
	DocumentTypePassport       DocumentType = "Passport"
	DocumentTypeDriversLicense DocumentType = "DriversLicense"
	DocumentTypeNationalID     DocumentType = "NationalID"
	DocumentTypeSSN            DocumentType = "SSN"
	DocumentTypeTaxID          DocumentType = "TaxID"
	DocumentTypeUtilityBill    DocumentType = "UtilityBill"
	DocumentTypeBankStatement  DocumentType = "BankStatement"
)

// IsValid checks if the document type is valid
func (t DocumentType) IsValid() bool {
	switch t {
	case DocumentTypePassport, DocumentTypeDriversLicense, DocumentTypeNationalID,
		DocumentTypeSSN, DocumentTypeTaxID, DocumentTypeUtilityBill, DocumentTypeBankStatement:
		return true
	}
	return false
}

// VerificationStatus represents the verification status of a document
type VerificationStatus string

const (
	VerificationStatusPending  VerificationStatus = "Pending"
	VerificationStatusVerified VerificationStatus = "Verified"
	VerificationStatusExpired  VerificationStatus = "Expired"
	VerificationStatusRejected VerificationStatus = "Rejected"
)

// IsValid checks if the verification status is valid
func (s VerificationStatus) IsValid() bool {
	switch s {
	case VerificationStatusPending, VerificationStatusVerified,
		VerificationStatusExpired, VerificationStatusRejected:
		return true
	}
	return false
}

// Customer represents a customer entity in the core banking system
type Customer struct {
	ID             uuid.UUID      `json:"id" db:"id"`
	CustomerNumber string         `json:"customer_number" db:"customer_number"`
	FirstName      string         `json:"first_name" db:"first_name"`
	MiddleName     *string        `json:"middle_name,omitempty" db:"middle_name"`
	LastName       string         `json:"last_name" db:"last_name"`
	DateOfBirth    time.Time      `json:"date_of_birth" db:"date_of_birth"`
	TaxID          string         `json:"-" db:"tax_id"` // Encrypted field, not exposed in JSON
	Email          string         `json:"email" db:"email"`
	Phone          string         `json:"phone" db:"phone"`
	Status         CustomerStatus `json:"status" db:"status"`
	CreatedAt      time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at" db:"updated_at"`
	CreatedBy      uuid.UUID      `json:"created_by" db:"created_by"`
	UpdatedBy      *uuid.UUID     `json:"updated_by,omitempty" db:"updated_by"`
	Version        int            `json:"version" db:"version"` // Optimistic locking
}

// Address represents a customer's address
type Address struct {
	ID          uuid.UUID   `json:"id" db:"id"`
	CustomerID  uuid.UUID   `json:"customer_id" db:"customer_id"`
	AddressType AddressType `json:"address_type" db:"address_type"`
	Street1     string      `json:"street1" db:"street1"`
	Street2     *string     `json:"street2,omitempty" db:"street2"`
	City        string      `json:"city" db:"city"`
	State       string      `json:"state" db:"state"`
	PostalCode  string      `json:"postal_code" db:"postal_code"`
	Country     string      `json:"country" db:"country"`
	IsPrimary   bool        `json:"is_primary" db:"is_primary"`
	ValidFrom   time.Time   `json:"valid_from" db:"valid_from"`
	ValidTo     *time.Time  `json:"valid_to,omitempty" db:"valid_to"`
	CreatedAt   time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at" db:"updated_at"`
}

// CustomerDocument represents a customer's identification document
type CustomerDocument struct {
	ID                 uuid.UUID          `json:"id" db:"id"`
	CustomerID         uuid.UUID          `json:"customer_id" db:"customer_id"`
	DocumentType       DocumentType       `json:"document_type" db:"document_type"`
	DocumentNumber     string             `json:"-" db:"document_number"` // Encrypted field
	IssuingAuthority   string             `json:"issuing_authority" db:"issuing_authority"`
	IssuingCountry     string             `json:"issuing_country" db:"issuing_country"`
	IssueDate          time.Time          `json:"issue_date" db:"issue_date"`
	ExpiryDate         time.Time          `json:"expiry_date" db:"expiry_date"`
	VerificationStatus VerificationStatus `json:"verification_status" db:"verification_status"`
	VerifiedAt         *time.Time         `json:"verified_at,omitempty" db:"verified_at"`
	VerifiedBy         *uuid.UUID         `json:"verified_by,omitempty" db:"verified_by"`
	CreatedAt          time.Time          `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time          `json:"updated_at" db:"updated_at"`
}

// SearchFilters represents the filters for customer search
type SearchFilters struct {
	FirstName string         `json:"first_name,omitempty"`
	LastName  string         `json:"last_name,omitempty"`
	Email     string         `json:"email,omitempty"`
	Phone     string         `json:"phone,omitempty"`
	Status    CustomerStatus `json:"status,omitempty"`
	FromDate  *time.Time     `json:"from_date,omitempty"`
	ToDate    *time.Time     `json:"to_date,omitempty"`
	Limit     int            `json:"limit,omitempty"`
	Offset    int            `json:"offset,omitempty"`
}

// StatusChange represents a customer status change record
type StatusChange struct {
	ID             uuid.UUID      `json:"id" db:"id"`
	CustomerID     uuid.UUID      `json:"customer_id" db:"customer_id"`
	PreviousStatus CustomerStatus `json:"previous_status" db:"previous_status"`
	NewStatus      CustomerStatus `json:"new_status" db:"new_status"`
	Reason         string         `json:"reason" db:"reason"`
	ChangedBy      uuid.UUID      `json:"changed_by" db:"changed_by"`
	ChangedAt      time.Time      `json:"changed_at" db:"changed_at"`
}

// Value implements driver.Valuer for CustomerStatus
func (s CustomerStatus) Value() (driver.Value, error) {
	return string(s), nil
}

// Scan implements sql.Scanner for CustomerStatus
func (s *CustomerStatus) Scan(value interface{}) error {
	if value == nil {
		*s = CustomerStatusPending
		return nil
	}
	str, ok := value.(string)
	if !ok {
		return errors.New("failed to scan CustomerStatus")
	}
	*s = CustomerStatus(str)
	if !s.IsValid() {
		return errors.New("invalid CustomerStatus value")
	}
	return nil
}

// Value implements driver.Valuer for AddressType
func (t AddressType) Value() (driver.Value, error) {
	return string(t), nil
}

// Scan implements sql.Scanner for AddressType
func (t *AddressType) Scan(value interface{}) error {
	if value == nil {
		*t = AddressTypePhysical
		return nil
	}
	str, ok := value.(string)
	if !ok {
		return errors.New("failed to scan AddressType")
	}
	*t = AddressType(str)
	if !t.IsValid() {
		return errors.New("invalid AddressType value")
	}
	return nil
}

// Value implements driver.Valuer for DocumentType
func (t DocumentType) Value() (driver.Value, error) {
	return string(t), nil
}

// Scan implements sql.Scanner for DocumentType
func (t *DocumentType) Scan(value interface{}) error {
	if value == nil {
		*t = DocumentTypePassport
		return nil
	}
	str, ok := value.(string)
	if !ok {
		return errors.New("failed to scan DocumentType")
	}
	*t = DocumentType(str)
	if !t.IsValid() {
		return errors.New("invalid DocumentType value")
	}
	return nil
}

// Value implements driver.Valuer for VerificationStatus
func (s VerificationStatus) Value() (driver.Value, error) {
	return string(s), nil
}

// Scan implements sql.Scanner for VerificationStatus
func (s *VerificationStatus) Scan(value interface{}) error {
	if value == nil {
		*s = VerificationStatusPending
		return nil
	}
	str, ok := value.(string)
	if !ok {
		return errors.New("failed to scan VerificationStatus")
	}
	*s = VerificationStatus(str)
	if !s.IsValid() {
		return errors.New("invalid VerificationStatus value")
	}
	return nil
}

// NullTime is a wrapper around time.Time for SQL null handling
type NullTime struct {
	Time  time.Time
	Valid bool
}

// Value implements driver.Valuer for NullTime
func (nt NullTime) Value() (driver.Value, error) {
	if !nt.Valid {
		return nil, nil
	}
	return nt.Time, nil
}

// Scan implements sql.Scanner for NullTime
func (nt *NullTime) Scan(value interface{}) error {
	if value == nil {
		nt.Time, nt.Valid = time.Time{}, false
		return nil
	}
	nt.Valid = true
	switch v := value.(type) {
	case time.Time:
		nt.Time = v
	default:
		return errors.New("failed to scan NullTime")
	}
	return nil
}

// MarshalJSON implements json.Marshaler for Customer
func (c Customer) MarshalJSON() ([]byte, error) {
	type Alias Customer
	return json.Marshal(&struct {
		Alias
	}{
		Alias: Alias(c),
	})
}

// MarshalJSON implements json.Marshaler for Address
func (a Address) MarshalJSON() ([]byte, error) {
	type Alias Address
	return json.Marshal(&struct {
		Alias
	}{
		Alias: Alias(a),
	})
}

// MarshalJSON implements json.Marshaler for CustomerDocument
func (d CustomerDocument) MarshalJSON() ([]byte, error) {
	type Alias CustomerDocument
	return json.Marshal(&struct {
		Alias
	}{
		Alias: Alias(d),
	})
}
