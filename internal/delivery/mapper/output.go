package mapper

import (
	"github.com/dockloud/auth/internal/usecase/auth"
	authv1 "github.com/dockloud/contract/gen/auth/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func ToLoginResponse(output *auth.LoginOutput) *authv1.LoginResponse {
	return &authv1.LoginResponse{
		Uid:       output.UID,
		JwtToken:  output.JWTToken,
		ExpiresAt: timestamppb.New(output.ExpiresAt),
	}
}

func ToRegisterResponse(output *auth.RegisterOutput) *authv1.RegisterResponse {
	return &authv1.RegisterResponse{
		Uid:      output.UID,
		Username: output.Username,
		Email:    output.Email,
	}
}
