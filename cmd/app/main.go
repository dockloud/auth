package main

import (
	"log"
	"net"
	"os"

	myserver "github.com/dockloud/auth/internal/delivery/grpc"
	serviceauth "github.com/dockloud/auth/internal/delivery/grpc/auth"
	"github.com/dockloud/auth/internal/infrastructure/cache/redis"
	"github.com/dockloud/auth/internal/infrastructure/database/postgres"
	"github.com/dockloud/auth/internal/infrastructure/user"
	"github.com/dockloud/auth/internal/usecase/auth"
	authv1 "github.com/dockloud/contract/gen/auth/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
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

	listener, err := net.Listen("tcp", ":50150")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	keyring := serviceauth.Keyring{
		"gateway-key": {
			ID:      "dck-gateway-v1",
			Service: "dockloud-gateway",
			Secret:  []byte(os.Getenv("GATEWAY_SERVICE_KEY")),
		},
	}
	interceptor, err := serviceauth.NewKeyringInterceptor(keyring)
	if err != nil {
		log.Fatalf("failed to create interceptor: %v", err)
	}

	grpcServer := grpc.NewServer(grpc.UnaryInterceptor(interceptor))
	healthServer := health.NewServer()

	healthpb.RegisterHealthServer(grpcServer, healthServer)
	healthServer.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)

	authv1.RegisterAuthServiceServer(grpcServer, myserver.NewAuthServer(*service))

	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
