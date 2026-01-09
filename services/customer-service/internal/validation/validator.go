package validation

import (
	"fmt"
	"regexp"
	"time"

	"github.com/core-banking/services/customer-service/internal/models"
	customerpb "github.com/core-banking/services/customer-service/internal/proto/customerpb"
)

// ValidationError represents a validation error with field details
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// ValidationErrors represents a collection of validation errors
type ValidationErrors []ValidationError

func (e ValidationErrors) Error() string {
	if len(e) == 0 {
		return ""
	}
	result := "validation failed: "
	for i, err := range e {
		if i > 0 {
			result += ", "
		}
		result += err.Error()
	}
	return result
}

// EmailRegex validates email format
var EmailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

// PhoneRegex validates international phone format
var PhoneRegex = regexp.MustCompile(`^\+?[1-9]\d{1,14}$`)

// TaxIDRegexes for different countries
var (
	USSSNRegex      = regexp.MustCompile(`^\d{3}-\d{2}-\d{4}$`)
	UKNINumberRegex = regexp.MustCompile(`^[A-Z]{2}\d{6}[A-Z]$`)
	EUTaxIDRegex    = regexp.MustCompile(`^[A-Z]{2}\d{8,12}$`)
)

// Validator provides validation methods for customer data
type Validator struct{}

// NewValidator creates a new Validator instance
func NewValidator() *Validator {
	return &Validator{}
}

// ValidateCustomerCreate validates customer creation data
func (v *Validator) ValidateCustomerCreate(req *customerpb.CreateCustomerRequest) ValidationErrors {
	var errs ValidationErrors

	// Required fields validation
	if req.GetFirstName() == "" {
		errs = append(errs, ValidationError{Field: "first_name", Message: "is required"})
	} else if len(req.GetFirstName()) < 2 || len(req.GetFirstName()) > 100 {
		errs = append(errs, ValidationError{Field: "first_name", Message: "must be between 2 and 100 characters"})
	}

	if req.GetLastName() == "" {
		errs = append(errs, ValidationError{Field: "last_name", Message: "is required"})
	} else if len(req.GetLastName()) < 2 || len(req.GetLastName()) > 100 {
		errs = append(errs, ValidationError{Field: "last_name", Message: "must be between 2 and 100 characters"})
	}

	if req.GetEmail() == "" {
		errs = append(errs, ValidationError{Field: "email", Message: "is required"})
	} else if !EmailRegex.MatchString(req.GetEmail()) {
		errs = append(errs, ValidationError{Field: "email", Message: "is invalid format"})
	}

	if req.GetPhone() != "" && !PhoneRegex.MatchString(req.GetPhone()) {
		errs = append(errs, ValidationError{Field: "phone", Message: "is invalid format. Use international format e.g., +1234567890"})
	}

	if req.GetDateOfBirth() == nil {
		errs = append(errs, ValidationError{Field: "date_of_birth", Message: "is required"})
	} else {
		dob := req.GetDateOfBirth().AsTime()
		if dob.After(time.Now().AddDate(-18, 0, 0)) {
			errs = append(errs, ValidationError{Field: "date_of_birth", Message: "customer must be at least 18 years old"})
		}
		if dob.Before(time.Now().AddDate(-150, 0, 0)) {
			errs = append(errs, ValidationError{Field: "date_of_birth", Message: "date of birth is too far in the past"})
		}
	}

	if req.GetTaxId() != "" {
		if err := v.ValidateTaxID(req.GetTaxId(), ""); err != nil {
			errs = append(errs, *err)
		}
	}

	return errs
}

// ValidateCustomerUpdate validates customer update data
func (v *Validator) ValidateCustomerUpdate(req *customerpb.UpdateCustomerRequest) ValidationErrors {
	var errs ValidationErrors

	if req.GetId() == "" {
		errs = append(errs, ValidationError{Field: "id", Message: "is required"})
	}

	if req.GetVersion() == 0 {
		errs = append(errs, ValidationError{Field: "version", Message: "is required for optimistic locking"})
	}

	if req.GetFirstName() != "" && (len(req.GetFirstName()) < 2 || len(req.GetFirstName()) > 100) {
		errs = append(errs, ValidationError{Field: "first_name", Message: "must be between 2 and 100 characters"})
	}

	if req.GetLastName() != "" && (len(req.GetLastName()) < 2 || len(req.GetLastName()) > 100) {
		errs = append(errs, ValidationError{Field: "last_name", Message: "must be between 2 and 100 characters"})
	}

	if req.GetEmail() != "" && !EmailRegex.MatchString(req.GetEmail()) {
		errs = append(errs, ValidationError{Field: "email", Message: "is invalid format"})
	}

	if req.GetPhone() != "" && !PhoneRegex.MatchString(req.GetPhone()) {
		errs = append(errs, ValidationError{Field: "phone", Message: "is invalid format. Use international format e.g., +1234567890"})
	}

	if req.GetDateOfBirth() != nil {
		dob := req.GetDateOfBirth().AsTime()
		if dob.After(time.Now().AddDate(-18, 0, 0)) {
			errs = append(errs, ValidationError{Field: "date_of_birth", Message: "customer must be at least 18 years old"})
		}
	}

	if req.GetTaxId() != "" {
		if err := v.ValidateTaxID(req.GetTaxId(), ""); err != nil {
			errs = append(errs, *err)
		}
	}

	return errs
}

// ValidateTaxID validates tax ID format based on country
func (v *Validator) ValidateTaxID(taxID, country string) *ValidationError {
	// Basic format validation - alphanumeric with possible hyphens
	if len(taxID) < 5 || len(taxID) > 50 {
		return &ValidationError{Field: "tax_id", Message: "must be between 5 and 50 characters"}
	}

	// Country-specific validation
	switch country {
	case "US":
		if !USSSNRegex.MatchString(taxID) {
			return &ValidationError{Field: "tax_id", Message: "invalid US SSN format (XXX-XX-XXXX)"}
		}
	case "UK":
		if !UKNINumberRegex.MatchString(taxID) {
			return &ValidationError{Field: "tax_id", Message: "invalid UK NINO format"}
		}
	case "DE", "FR", "IT", "ES": // EU countries
		if !EUTaxIDRegex.MatchString(taxID) {
			return &ValidationError{Field: "tax_id", Message: "invalid EU tax ID format"}
		}
	default:
		// Generic validation for other countries
		matched, _ := regexp.MatchString(`^[A-Z0-9]{5,20}$`, taxID)
		if !matched {
			return &ValidationError{Field: "tax_id", Message: "invalid tax ID format"}
		}
	}

	return nil
}

// ValidateAddress validates address data
func (v *Validator) ValidateAddress(req *customerpb.AddAddressRequest) ValidationErrors {
	var errs ValidationErrors

	if req.GetCustomerId() == "" {
		errs = append(errs, ValidationError{Field: "customer_id", Message: "is required"})
	}

	if req.GetStreet1() == "" {
		errs = append(errs, ValidationError{Field: "street1", Message: "is required"})
	} else if len(req.GetStreet1()) > 200 {
		errs = append(errs, ValidationError{Field: "street1", Message: "must not exceed 200 characters"})
	}

	if req.GetCity() == "" {
		errs = append(errs, ValidationError{Field: "city", Message: "is required"})
	} else if len(req.GetCity()) > 100 {
		errs = append(errs, ValidationError{Field: "city", Message: "must not exceed 100 characters"})
	}

	if req.GetState() == "" {
		errs = append(errs, ValidationError{Field: "state", Message: "is required"})
	} else if len(req.GetState()) > 100 {
		errs = append(errs, ValidationError{Field: "state", Message: "must not exceed 100 characters"})
	}

	if req.GetPostalCode() == "" {
		errs = append(errs, ValidationError{Field: "postal_code", Message: "is required"})
	} else if len(req.GetPostalCode()) > 20 {
		errs = append(errs, ValidationError{Field: "postal_code", Message: "must not exceed 20 characters"})
	}

	if req.GetCountry() == "" {
		errs = append(errs, ValidationError{Field: "country", Message: "is required"})
	} else if len(req.GetCountry()) != 2 {
		errs = append(errs, ValidationError{Field: "country", Message: "must be a 2-letter ISO country code"})
	}

	if req.GetAddressType() != "" {
		addrType := models.AddressType(req.GetAddressType())
		if !addrType.IsValid() {
			errs = append(errs, ValidationError{Field: "address_type", Message: "must be Physical, Mailing, or Business"})
		}
	}

	return errs
}

// ValidateDocument validates document data
func (v *Validator) ValidateDocument(req *customerpb.AddDocumentRequest) ValidationErrors {
	var errs ValidationErrors

	if req.GetCustomerId() == "" {
		errs = append(errs, ValidationError{Field: "customer_id", Message: "is required"})
	}

	if req.GetDocumentType() == "" {
		errs = append(errs, ValidationError{Field: "document_type", Message: "is required"})
	} else {
		docType := models.DocumentType(req.GetDocumentType())
		if !docType.IsValid() {
			errs = append(errs, ValidationError{Field: "document_type", Message: "is invalid document type"})
		}
	}

	if req.GetDocumentNumber() == "" {
		errs = append(errs, ValidationError{Field: "document_number", Message: "is required"})
	} else if len(req.GetDocumentNumber()) > 50 {
		errs = append(errs, ValidationError{Field: "document_number", Message: "must not exceed 50 characters"})
	}

	if req.GetIssuingCountry() == "" {
		errs = append(errs, ValidationError{Field: "issuing_country", Message: "is required"})
	} else if len(req.GetIssuingCountry()) != 2 {
		errs = append(errs, ValidationError{Field: "issuing_country", Message: "must be a 2-letter ISO country code"})
	}

	if req.GetIssuingAuthority() == "" {
		errs = append(errs, ValidationError{Field: "issuing_authority", Message: "is required"})
	}

	if req.GetIssueDate() == nil {
		errs = append(errs, ValidationError{Field: "issue_date", Message: "is required"})
	}

	if req.GetExpiryDate() == nil {
		errs = append(errs, ValidationError{Field: "expiry_date", Message: "is required"})
	} else {
		expiryDate := req.GetExpiryDate().AsTime()
		if expiryDate.Before(time.Now()) {
			errs = append(errs, ValidationError{Field: "expiry_date", Message: "must be in the future"})
		}
		if req.GetIssueDate() != nil {
			issueDate := req.GetIssueDate().AsTime()
			if expiryDate.Before(issueDate) {
				errs = append(errs, ValidationError{Field: "expiry_date", Message: "must be after issue date"})
			}
		}
	}

	return errs
}

// ValidateStatusTransition validates status transition rules
func (v *Validator) ValidateStatusTransition(currentStatus, newStatus models.CustomerStatus) error {
	// Define valid status transitions
	validTransitions := map[models.CustomerStatus][]models.CustomerStatus{
		models.CustomerStatusPending:   {models.CustomerStatusActive, models.CustomerStatusClosed},
		models.CustomerStatusActive:    {models.CustomerStatusInactive, models.CustomerStatusSuspended, models.CustomerStatusClosed},
		models.CustomerStatusInactive:  {models.CustomerStatusActive, models.CustomerStatusSuspended, models.CustomerStatusClosed},
		models.CustomerStatusSuspended: {models.CustomerStatusActive, models.CustomerStatusInactive, models.CustomerStatusClosed},
		models.CustomerStatusClosed:    {}, // No transitions from Closed
	}

	allowedTransitions, exists := validTransitions[currentStatus]
	if !exists {
		return fmt.Errorf("invalid current status: %s", currentStatus)
	}

	for _, allowed := range allowedTransitions {
		if allowed == newStatus {
			return nil
		}
	}

	return fmt.Errorf("invalid status transition from %s to %s", currentStatus, newStatus)
}

// ValidateSearchFilters validates search filter parameters
func (v *Validator) ValidateSearchFilters(req *customerpb.SearchCustomersRequest) ValidationErrors {
	var errs ValidationErrors

	if req.GetLimit() < 0 {
		errs = append(errs, ValidationError{Field: "limit", Message: "must be non-negative"})
	}
	if req.GetLimit() > 100 {
		errs = append(errs, ValidationError{Field: "limit", Message: "must not exceed 100"})
	}

	if req.GetOffset() < 0 {
		errs = append(errs, ValidationError{Field: "offset", Message: "must be non-negative"})
	}

	if req.GetStatus() != "" {
		status := models.CustomerStatus(req.GetStatus())
		if !status.IsValid() {
			errs = append(errs, ValidationError{Field: "status", Message: "is invalid status"})
		}
	}

	if req.GetEmail() != "" && !EmailRegex.MatchString(req.GetEmail()) {
		errs = append(errs, ValidationError{Field: "email", Message: "is invalid format"})
	}

	if req.GetPhone() != "" && !PhoneRegex.MatchString(req.GetPhone()) {
		errs = append(errs, ValidationError{Field: "phone", Message: "is invalid format"})
	}

	return errs
}
