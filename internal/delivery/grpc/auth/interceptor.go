package serviceauth

import (
	"github.com/dockloud/service-auth/core"
	"github.com/dockloud/service-auth/grpcauth"
	"google.golang.org/grpc"
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

	return interceptor, nil
}
