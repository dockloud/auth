package userauthn

import (
	"context"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type TokenVerifier interface {
	Verify(rawToken string) (string, error)
}

func UnaryServerInterceptor(
	verifier TokenVerifier,
	publicMethods ...string,
) grpc.UnaryServerInterceptor {
	public := make(map[string]struct{}, len(publicMethods))

	for _, method := range publicMethods {
		public[method] = struct{}{}
	}

	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {

		if _, ok := public[info.FullMethod]; ok {
			return handler(ctx, req)
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(
				codes.Unauthenticated,
				"missing authorization metadata",
			)
		}

		values := md.Get("authorization")

		if len(values) != 1 {
			return nil, status.Error(
				codes.Unauthenticated,
				"missing authorization token",
			)
		}

		parts := strings.Fields(values[0])

		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			return nil, status.Error(
				codes.Unauthenticated,
				"invalid authorization format",
			)
		}

		uid, err := verifier.Verify(parts[1])
		if err != nil {
			return nil, status.Error(
				codes.Unauthenticated,
				"invalid or expired token",
			)
		}

		ctx = contextWithUserID(ctx, uid)

		return handler(ctx, req)
	}
}
