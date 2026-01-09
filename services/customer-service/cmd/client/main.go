package main

import (
	"context"
	"fmt"
	"log"
	"time"

	customerpb "github.com/core-banking/services/customer-service/internal/proto/customerpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func main() {
	// Connect to the gRPC server
	conn, err := grpc.Dial("localhost:50051",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
		grpc.WithTimeout(10*time.Second),
	)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	// Create client
	client := customerpb.NewCustomerServiceClient(conn)
	ctx := context.Background()

	// Example 1: Create a customer
	fmt.Println("=== Creating Customer ===")
	customer, err := createCustomer(ctx, client)
	if err != nil {
		log.Fatalf("Failed to create customer: %v", err)
	}
	fmt.Printf("Created customer: ID=%s, Number=%s, Email=%s\n",
		customer.Id, customer.CustomerNumber, customer.Email)

	// Example 2: Get customer
	fmt.Println("\n=== Getting Customer ===")
	retrievedCustomer, err := getCustomer(ctx, client, customer.Id)
	if err != nil {
		log.Fatalf("Failed to get customer: %v", err)
	}
	fmt.Printf("Retrieved customer: %s %s\n", retrievedCustomer.FirstName, retrievedCustomer.LastName)

	// Example 3: Add address
	fmt.Println("\n=== Adding Address ===")
	address, err := addAddress(ctx, client, customer.Id)
	if err != nil {
		log.Fatalf("Failed to add address: %v", err)
	}
	fmt.Printf("Added address: %s, %s, %s %s\n",
		address.Street1, address.City, address.State, address.PostalCode)

	// Example 4: Add document
	fmt.Println("\n=== Adding Document ===")
	document, err := addDocument(ctx, client, customer.Id)
	if err != nil {
		log.Fatalf("Failed to add document: %v", err)
	}
	fmt.Printf("Added document: Type=%s, Number=%s\n",
		document.DocumentType, document.DocumentNumber)

	// Example 5: Get full profile
	fmt.Println("\n=== Getting Full Profile ===")
	profile, err := getFullProfile(ctx, client, customer.Id)
	if err != nil {
		log.Fatalf("Failed to get full profile: %v", err)
	}
	fmt.Printf("Full Profile:\n")
	fmt.Printf("  Customer: %s %s (%s)\n", profile.Customer.FirstName, profile.Customer.LastName, profile.Customer.Email)
	fmt.Printf("  Addresses: %d\n", len(profile.Addresses))
	fmt.Printf("  Documents: %d\n", len(profile.Documents))

	// Example 6: Update customer status
	fmt.Println("\n=== Updating Customer Status ===")
	updatedCustomer, err := updateCustomerStatus(ctx, client, customer.Id, "Active", "Customer verified with documents")
	if err != nil {
		log.Fatalf("Failed to update customer status: %v", err)
	}
	fmt.Printf("Updated customer status: %s -> %s\n", customer.Status, updatedCustomer.Customer.Status)

	// Example 7: Search customers
	fmt.Println("\n=== Searching Customers ===")
	customers, err := searchCustomers(ctx, client)
	if err != nil {
		log.Fatalf("Failed to search customers: %v", err)
	}
	fmt.Printf("Found %d customers\n", len(customers))

	fmt.Println("\n=== All examples completed successfully ===")
}

func createCustomer(ctx context.Context, client customerpb.CustomerServiceClient) (*customerpb.Customer, error) {
	req := &customerpb.CreateCustomerRequest{
		FirstName:   "John",
		LastName:    "Doe",
		Email:       "john.doe@example.com",
		Phone:       "+1-555-123-4567",
		DateOfBirth: timestamppb.New(time.Date(1990, 1, 15, 0, 0, 0, 0, time.UTC)),
		TaxId:       "123-45-6789",
		CreatedBy:   "system",
	}

	resp, err := client.CreateCustomer(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("CreateCustomer failed: %w", err)
	}

	return resp.Customer, nil
}

func getCustomer(ctx context.Context, client customerpb.CustomerServiceClient, id string) (*customerpb.Customer, error) {
	req := &customerpb.GetCustomerRequest{
		Id: id,
	}

	resp, err := client.GetCustomer(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("GetCustomer failed: %w", err)
	}

	return resp.Customer, nil
}

func addAddress(ctx context.Context, client customerpb.CustomerServiceClient, customerID string) (*customerpb.Address, error) {
	req := &customerpb.AddAddressRequest{
		CustomerId:  customerID,
		AddressType: "Physical",
		Street1:     "123 Main Street",
		Street2:     "Apt 4B",
		City:        "New York",
		State:       "NY",
		PostalCode:  "10001",
		Country:     "US",
		IsPrimary:   true,
		ValidFrom:   timestamppb.New(time.Now()),
	}

	resp, err := client.AddAddress(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("AddAddress failed: %w", err)
	}

	return resp.Address, nil
}

func addDocument(ctx context.Context, client customerpb.CustomerServiceClient, customerID string) (*customerpb.Document, error) {
	req := &customerpb.AddDocumentRequest{
		CustomerId:       customerID,
		DocumentType:     "Passport",
		DocumentNumber:   "AB12345678",
		IssuingAuthority: "US Department of State",
		IssuingCountry:   "US",
		IssueDate:        timestamppb.New(time.Now().AddDate(-2, 0, 0)),
		ExpiryDate:       timestamppb.New(time.Now().AddDate(5, 0, 0)),
	}

	resp, err := client.AddDocument(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("AddDocument failed: %w", err)
	}

	return resp.Document, nil
}

func getFullProfile(ctx context.Context, client customerpb.CustomerServiceClient, id string) (*customerpb.CustomerFullProfileResponse, error) {
	req := &customerpb.GetCustomerRequest{
		Id: id,
	}

	resp, err := client.GetCustomerFullProfile(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("GetCustomerFullProfile failed: %w", err)
	}

	return resp, nil
}

func updateCustomerStatus(ctx context.Context, client customerpb.CustomerServiceClient, id, newStatus, reason string) (*customerpb.UpdateCustomerStatusResponse, error) {
	req := &customerpb.UpdateCustomerStatusRequest{
		Id:        id,
		NewStatus: newStatus,
		Reason:    reason,
		ChangedBy: "admin",
	}

	resp, err := client.UpdateCustomerStatus(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("UpdateCustomerStatus failed: %w", err)
	}

	return resp, nil
}

func searchCustomers(ctx context.Context, client customerpb.CustomerServiceClient) ([]*customerpb.Customer, error) {
	req := &customerpb.SearchCustomersRequest{
		Status: "Active",
		Limit:  10,
	}

	resp, err := client.SearchCustomers(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("SearchCustomers failed: %w", err)
	}

	return resp.Customers, nil
}
