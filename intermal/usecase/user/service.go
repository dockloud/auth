package user

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"
	"time"

	"github.com/dockloud/auth/intermal/domain/user"
	"github.com/golang-jwt/jwt/v5"
)

type Repository interface {
	Create(ctx context.Context, username string, passwordHash string, email string) (*user.User, error)
	Get(ctx context.Context, uid string) (*user.User, error)
	Update(ctx context.Context, uid string, username *string, email *string, passwordHash *string) (*user.User, error)
	Delete(ctx context.Context, uid string) (*user.User, error)
	Search(ctx context.Context, uid *string, username *string, email *string) (*user.SearchResult, error)
	Authenticate(ctx context.Context, email string, passwordHash string) (*user.User, error)
}

type Service struct {
	repository Repository
}

var (
	ErrUsernameTaken = errors.New("username already taken")
	ErrEmailTaken    = errors.New("email already taken")
	ErrUserNotFound  = errors.New("user not found")
)

func NewService(repository Repository) *Service {
	return &Service{repository: repository}
}

func hashPassword(password string) string {
	secret := []byte(os.Getenv("JWT_AUTH_KEY"))
	passwordBytes := []byte(password)

	h := hmac.New(sha256.New, secret)
	h.Write(passwordBytes)
	signatureBytes := h.Sum(nil)
	signatureHex := hex.EncodeToString(signatureBytes)

	return signatureHex
}

func generateJWTToken(uid string) (*LoginOutput, error) {
	secret := []byte(os.Getenv("JWT_AUTH_KEY"))
	created := time.Now()
	expires := created.Add(time.Hour * 2)
	issue := "dockloud-auth"

	claims := jwt.MapClaims{
		"sub": uid,            // идентификатор
		"iss": issue,          // кем выдан
		"iat": created.Unix(), // когда выдан
		"exp": expires.Unix(), // когда истекает
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(secret)
	if err != nil {
		return nil, err
	}

	return &LoginOutput{
		UID:       uid,
		JWTToken:  tokenString,
		ExpiresAt: expires,
	}, nil
}

func (s *Service) Register(ctx context.Context, input RegisterInput) (*RegisterOutput, error) {
	if err := user.ValidateUsername(input.Username); err != nil {
		return nil, err
	}
	if err := user.ValidateEmail(input.Email); err != nil {
		return nil, err
	}
	if err := user.ValidatePassword(input.Password); err != nil {
		return nil, err
	}

	searchResult, err := s.repository.Search(ctx, nil, &input.Username, &input.Email)
	if err != nil {
		return nil, err
	}

	switch {
	case searchResult.UsernameExists:
		return nil, ErrUsernameTaken
	case searchResult.EmailExists:
		return nil, ErrEmailTaken
	}

	u, err := s.repository.Create(ctx, input.Username, hashPassword(input.Password), input.Email)
	if err != nil {
		return nil, err
	}

	return &RegisterOutput{
		UID:      u.UID,
		Username: u.Username,
		Email:    u.Email,
	}, nil
}

func (s *Service) Login(ctx context.Context, input LoginInput) (*LoginOutput, error) {
	if err := user.ValidateEmail(input.Email); err != nil {
		return nil, err
	}

	u, err := s.repository.Authenticate(ctx, input.Email, hashPassword(input.Password))
	if err != nil {
		return nil, err
	}

	output, err := generateJWTToken(u.UID)
	if err != nil {
		return nil, err
	}

	return output, nil
}

func (s *Service) Update(ctx context.Context, uid string, input UpdateInput) error {
	if err := user.ValidateUsername(*input.Username); err != nil {
		return err
	}
	if err := user.ValidateEmail(*input.Email); err != nil {
		return err
	}
	if err := user.ValidatePassword(*input.Password); err != nil {
		return err
	}

	_, err := s.repository.Update(ctx, uid, input.Username, input.Email, input.Password)
	if err != nil {
		return err
	}
	return nil
}

func (s *Service) Delete(ctx context.Context, uid string) error {
	searchResult, err := s.repository.Search(ctx, &uid, nil, nil)
	if err != nil {
		return err
	}

	if !searchResult.UIDExists {
		return ErrUserNotFound
	}

	_, err = s.repository.Delete(ctx, uid)
	if err != nil {
		return err
	}

	return nil
}
