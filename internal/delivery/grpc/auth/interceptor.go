package serviceauth

import (
	"context"

	"github.com/dockloud/service-auth/core"
	"github.com/dockloud/service-auth/grpcauth"
	"google.golang.org/grpc"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

func NewKeyringInterceptor(keyring Keyring) (grpc.UnaryServerInterceptor, error) {
	verifier, err := serviceauth.NewVerifier(keyring, 0)
	if err != nil {
		return nil, err
	}

	interceptor, err := grpcauth.NewServerInterceptor(verifier, 0)
	if err != nil {
		return nil, err
	}

	return func(ctx context.Context, request any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if info.FullMethod == healthpb.Health_Check_FullMethodName {
			return handler(ctx, request)
		}
		return interceptor(ctx, request, info, handler)
	}, nil
}
