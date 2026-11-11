package serviceauth

import (
	"context"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"
)

func TestNewKeyringInterceptorBypassesHealthCheck(t *testing.T) {
	interceptor, err := NewKeyringInterceptor(Keyring{})
	if err != nil {
		t.Fatalf("NewKeyringInterceptor() error = %v", err)
	}

	handlerCalled := false
	response, err := interceptor(
		context.Background(),
		&healthpb.HealthCheckRequest{},
		&grpc.UnaryServerInfo{FullMethod: healthpb.Health_Check_FullMethodName},
		func(context.Context, any) (any, error) {
			handlerCalled = true
			return &healthpb.HealthCheckResponse{Status: healthpb.HealthCheckResponse_SERVING}, nil
		},
	)
	if err != nil {
		t.Fatalf("health check returned error: %v", err)
	}
	if !handlerCalled {
		t.Fatal("health check handler was not called")
	}
	if got := response.(*healthpb.HealthCheckResponse).GetStatus(); got != healthpb.HealthCheckResponse_SERVING {
		t.Errorf("health check status = %v, want SERVING", got)
	}
}

func TestNewKeyringInterceptorAuthenticatesOtherMethods(t *testing.T) {
	interceptor, err := NewKeyringInterceptor(Keyring{})
	if err != nil {
		t.Fatalf("NewKeyringInterceptor() error = %v", err)
	}

	handlerCalled := false
	_, err = interceptor(
		context.Background(),
		&healthpb.HealthCheckRequest{},
		&grpc.UnaryServerInfo{FullMethod: "/auth.v1.AuthService/Check"},
		func(context.Context, any) (any, error) {
			handlerCalled = true
			return nil, nil
		},
	)
	if got := status.Code(err); got != codes.Unauthenticated {
		t.Errorf("method error code = %v, want %v", got, codes.Unauthenticated)
	}
	if handlerCalled {
		t.Error("unauthenticated method handler was called")
	}
}
