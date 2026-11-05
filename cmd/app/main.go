package main

import (
	"log"
	"net"
	"os"

	myserver "github.com/dockloud/auth/internal/delivery/grpc"
	"github.com/dockloud/auth/internal/infrastructure/cache/redis"
	"github.com/dockloud/auth/internal/infrastructure/database/postgres"
	"github.com/dockloud/auth/internal/infrastructure/user"
	"github.com/dockloud/auth/internal/usecase/auth"
	userauthn "github.com/dockloud/auth/pkg/user_authn"
	authv1 "github.com/dockloud/contract/gen/auth/v1"
	"google.golang.org/grpc"
)

func main() {
	db, err := postgres.NewDB(os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("failed to connect to database: %s", err)
	}

	cache, err := redis.NewCache(os.Getenv("CACHE_URL"))
	if err != nil {
		log.Fatalf("failed to connect to cache: %s", err)
	}

	repo := user.NewRepository(db, cache)
	service := auth.NewService(repo)

	verifier, err := userauthn.NewVerifier(os.Getenv("JWT_AUTH_KEY"))
	if err != nil {
		log.Fatal(err)
	}

	listener, err := net.Listen("tcp", ":50150")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(
			userauthn.UnaryServerInterceptor(
				verifier,
				"/auth.v1.AuthService/Register",
				"/auth.v1.AuthService/Login",
			),
		),
	)
	authv1.RegisterAuthServiceServer(grpcServer, myserver.NewAuthServer(*service))

	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
