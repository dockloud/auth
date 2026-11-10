package grpc

import (
	"context"
	"errors"
	"log"

	"github.com/dockloud/auth/internal/delivery/mapper"
	"github.com/dockloud/auth/internal/usecase/auth"
	authv1 "github.com/dockloud/contract/gen/auth/v1"
	"github.com/dockloud/user-auth/pkg/transport/grpc/metadata"
)

var ErrNoUIDPassed = errors.New("no UID passed")

type authServer struct {
	authv1.UnimplementedAuthServiceServer
	service *auth.Service
	logger  *log.Logger
}

func NewAuthServer(service auth.Service) authv1.AuthServiceServer {
	return &authServer{
		service: &service,
		logger:  log.Default(),
	}
}

func (s *authServer) Register(ctx context.Context, req *authv1.RegisterRequest) (*authv1.RegisterResponse, error) {
	s.logger.Printf("Received RegisterRequest %+v", req)

	input := mapper.ToRegisterInput(req)
	res, err := s.service.Register(ctx, input)

	if err != nil {
		s.logger.Printf("Failed to register %+v", req)
		return nil, err
	}

	return mapper.ToRegisterResponse(res), nil
}

func (s *authServer) Login(ctx context.Context, req *authv1.LoginRequest) (*authv1.LoginResponse, error) {
	s.logger.Printf("Received LoginRequest %+v", req)

	input := mapper.ToLoginInput(req)
	res, err := s.service.Login(ctx, input)

	if err != nil {
		s.logger.Printf("Failed to login %+v", req)
		return nil, err
	}

	return mapper.ToLoginResponse(res), nil
}

func (s *authServer) Update(ctx context.Context, req *authv1.UpdateUserRequest) (*authv1.UpdateUserResponse, error) {
	s.logger.Printf("Received UpdateRequest %+v", req)

	input := mapper.ToUpdateInput(req)
	uid, ok := metadata.GetUIDFromMetadata(ctx)
	if !ok {
		s.logger.Printf("Failed to update %+v", req)
		return nil, ErrNoUIDPassed
	}

	err := s.service.Update(ctx, uid, input)
	if err != nil {
		s.logger.Printf("Failed to update %+v", req)
		return nil, err
	}

	return &authv1.UpdateUserResponse{}, nil
}

func (s *authServer) Delete(ctx context.Context, req *authv1.DeleteUserRequest) (*authv1.DeleteUserResponse, error) {
	s.logger.Printf("Received DeleteRequest %+v", req)

	uid, ok := metadata.GetUIDFromMetadata(ctx)
	if !ok {
		s.logger.Printf("Failed to delete %+v", req)
		return nil, ErrNoUIDPassed
	}

	err := s.service.Delete(ctx, uid)
	if err != nil {
		s.logger.Printf("Failed to delete %+v", req)
		return nil, err
	}

	return &authv1.DeleteUserResponse{}, nil
}
