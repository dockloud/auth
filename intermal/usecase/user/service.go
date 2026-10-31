package user

import (
	"context"

	"github.com/dockloud/auth/intermal/domain/user"
)

type Repository interface {
	Create(ctx context.Context, username string, passwordHash string, email string) (*user.User, error)
	Get(ctx context.Context, uid string) (*user.User, error)
	Update(ctx context.Context, uid string, username *string, email *string, passwordHash *string) (*user.User, error)
	Delete(ctx context.Context, uid string) (*user.User, error)
	Search(ctx context.Context, username *string, email *string) (*user.SearchResult, error)
}

type Service struct {
	repository Repository
}

func NewService(repository Repository) *Service {
	return &Service{repository: repository}
}
