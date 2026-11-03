package grpc

import (
	"context"
	"log"

	"github.com/dockloud/auth/internal/delivery/mapper"
	"github.com/dockloud/auth/internal/usecase/auth"
	authv1 "github.com/dockloud/contract/gen/auth/v1"
)

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

func (s *authServer) Update(ctx context.Context, req *authv1.UpdateRequest) (*authv1.UpdateResponse, error) {
	s.logger.Printf("Received UpdateRequest %+v", req)

	input, uid := mapper.ToUpdateInput(req)
	err := s.service.Update(ctx, string(uid), input)

	if err != nil {
		s.logger.Printf("Failed to update %+v", req)
		return nil, err
	}

	return &authv1.UpdateResponse{}, nil
}

func (s *authServer) Delete(ctx context.Context, req *authv1.DeleteRequest) (*authv1.DeleteResponse, error) {
	s.logger.Printf("Received DeleteRequest %+v", req)

	input := mapper.ToDeleteInput(req)
	err := s.service.Delete(ctx, string(input))

	if err != nil {
		s.logger.Printf("Failed to delete %+v", req)
		return nil, err
	}

	return &authv1.DeleteResponse{}, nil
}
