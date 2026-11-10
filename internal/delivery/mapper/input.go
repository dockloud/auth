package mapper

import (
	"github.com/dockloud/auth/internal/usecase/auth"
	authv1 "github.com/dockloud/contract/gen/auth/v1"
)

func ToRegisterInput(request *authv1.RegisterRequest) *auth.RegisterInput {
	return &auth.RegisterInput{
		Username: request.Username,
		Email:    request.Email,
		Password: request.Password,
	}
}

func ToLoginInput(request *authv1.LoginRequest) *auth.LoginInput {
	return &auth.LoginInput{
		Email:    request.Email,
		Password: request.Password,
	}
}

func ToUpdateInput(request *authv1.UpdateUserRequest) *auth.UpdateInput {
	return &auth.UpdateInput{
		Username: request.Username,
		Email:    request.Email,
		Password: request.Password,
	}
}
