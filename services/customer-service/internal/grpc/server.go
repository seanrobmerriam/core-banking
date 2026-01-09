package grpc

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"time"

	customerpb "github.com/core-banking/services/customer-service/internal/proto/customerpb"
	"github.com/core-banking/services/customer-service/internal/repository"
	"github.com/core-banking/services/customer-service/internal/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// Server represents the gRPC server for customer service
type Server struct {
	customerpb.UnimplementedCustomerServiceServer
	customerService *service.CustomerService
	grpcServer      *grpc.Server
	listener        net.Listener
}

// Config holds the server configuration
type Config struct {
	Port        int
	MaxRecvSize int
	MaxSendSize int
	Timeout     time.Duration
	EnableAuth  bool
}

// NewServer creates a new gRPC server
func NewServer(repo repository.CustomerRepository, cfg Config) *Server {
	// Create customer service
	customerService := service.NewCustomerService(repo)

	// Create unary interceptors
	unaryInterceptors := []grpc.UnaryServerInterceptor{
		loggingUnaryInterceptor,
		recoveryUnaryInterceptor,
		timeoutUnaryInterceptor(cfg.Timeout),
		metadataUnaryInterceptor,
	}

	// Create stream interceptors
	streamInterceptors := []grpc.StreamServerInterceptor{
		loggingStreamInterceptor,
		recoveryStreamInterceptor,
	}

	// Create gRPC server with options
	grpcOpts := []grpc.ServerOption{
		grpc.MaxRecvMsgSize(cfg.MaxRecvSize * 1024 * 1024),
		grpc.MaxSendMsgSize(cfg.MaxSendSize * 1024 * 1024),
		grpc.ChainUnaryInterceptor(unaryInterceptors...),
		grpc.ChainStreamInterceptor(streamInterceptors...),
	}

	grpcServer := grpc.NewServer(grpcOpts...)

	// Register customer service
	customerpb.RegisterCustomerServiceServer(grpcServer, customerService)

	// Create listener
	addr := fmt.Sprintf(":%d", cfg.Port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("Failed to listen on %s: %v", addr, err)
	}

	return &Server{
		customerService: customerService,
		grpcServer:      grpcServer,
		listener:        listener,
	}
}

// Start starts the gRPC server
func (s *Server) Start() error {
	log.Printf("Starting gRPC server on %s", s.listener.Addr().String())
	return s.grpcServer.Serve(s.listener)
}

// Stop gracefully stops the gRPC server
func (s *Server) Stop() {
	log.Println("Stopping gRPC server...")
	s.grpcServer.GracefulStop()
}

// CreateClient creates a gRPC client for testing
func CreateClient(addr string) (customerpb.CustomerServiceClient, *grpc.ClientConn, error) {
	conn, err := grpc.Dial(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
		grpc.WithTimeout(5*time.Second),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect: %w", err)
	}
	return customerpb.NewCustomerServiceClient(conn), conn, nil
}

// Interceptor functions

func loggingUnaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	start := time.Now()
	log.Printf("gRPC unary request: %s", info.FullMethod)
	resp, err := handler(ctx, req)
	duration := time.Since(start)
	code := codes.OK
	if err != nil {
		if st, ok := status.FromError(err); ok {
			code = st.Code()
		}
	}
	log.Printf("gRPC unary response: %s, code=%s, duration=%v", info.FullMethod, code, duration)
	return resp, err
}

func loggingStreamInterceptor(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	start := time.Now()
	log.Printf("gRPC stream request: %s", info.FullMethod)
	err := handler(srv, ss)
	duration := time.Since(start)
	code := codes.OK
	if err != nil {
		if st, ok := status.FromError(err); ok {
			code = st.Code()
		}
	}
	log.Printf("gRPC stream response: %s, code=%s, duration=%v", info.FullMethod, code, duration)
	return err
}

func recoveryUnaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Panic in unary handler: %v", r)
		}
	}()
	return handler(ctx, req)
}

func recoveryStreamInterceptor(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Panic in stream handler: %v", r)
		}
	}()
	return handler(srv, ss)
}

func timeoutUnaryInterceptor(timeout time.Duration) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		ctx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()
		select {
		case <-ctx.Done():
			return nil, status.Errorf(codes.DeadlineExceeded, "request timeout after %v", timeout)
		default:
			return handler(ctx, req)
		}
	}
}

func metadataUnaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		if requestID := md.Get("x-request-id"); len(requestID) > 0 {
			ctx = context.WithValue(ctx, "request_id", requestID[0])
		}
		if userID := md.Get("x-user-id"); len(userID) > 0 {
			ctx = context.WithValue(ctx, "user_id", userID[0])
		}
	}
	return handler(ctx, req)
}

// StreamWrapper wraps a ServerStream to add logging
type StreamWrapper struct {
	grpc.ServerStream
	info  *grpc.StreamServerInfo
	start time.Time
}

func (s *StreamWrapper) RecvMsg(m interface{}) error {
	err := s.ServerStream.RecvMsg(m)
	if err == io.EOF {
		return err
	}
	if err != nil {
		log.Printf("Error receiving message: %v", err)
		return err
	}
	return nil
}

func (s *StreamWrapper) SendMsg(m interface{}) error {
	err := s.ServerStream.SendMsg(m)
	if err != nil && err != io.EOF {
		log.Printf("Error sending message: %v", err)
	}
	return err
}
