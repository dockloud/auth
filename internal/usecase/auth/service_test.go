package auth

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/dockloud/auth/internal/domain/user"
	"github.com/dockloud/auth/internal/usecase/auth/mocks"
	"github.com/golang-jwt/jwt/v5"
	"github.com/golang/mock/gomock"
)

const testJWTKey = "test-jwt-key"

var (
	errSearchFailed = errors.New("search failed")
	errCreateFailed = errors.New("create failed")
	errAuthFailed   = errors.New("authenticate failed")
	errUpdateFailed = errors.New("update failed")
	errDeleteFailed = errors.New("delete failed")
)

func strPtr(s string) *string { return new(s) }

type nilStrPtrMatcher struct{}

func (nilStrPtrMatcher) Matches(x interface{}) bool {
	s, ok := x.(*string)
	return ok && s == nil
}

func (nilStrPtrMatcher) String() string { return "nil *string" }

func newTestService(t *testing.T, setup func(m *mocks.MockRepository)) *Service {
	t.Helper()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	repo := mocks.NewMockRepository(ctrl)
	if setup != nil {
		setup(repo)
	}
	return NewService(repo)
}

func assertError(t *testing.T, got, want error) {
	t.Helper()
	if (got == nil) != (want == nil) {
		t.Fatalf("error = %v, want %v", got, want)
	}
	if got != nil && !errors.Is(got, want) {
		t.Fatalf("error = %q, want %q", got, want)
	}
}

func validateJWT(t *testing.T, tokenString, wantUID string) {
	t.Helper()

	claims := jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(tok *jwt.Token) (any, error) {
		if _, ok := tok.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", tok.Header["alg"])
		}
		return []byte(testJWTKey), nil
	})
	if err != nil {
		t.Fatalf("failed to parse JWT: %v", err)
	}
	if !token.Valid {
		t.Fatal("JWT is invalid")
	}
	sub, err := claims.GetSubject()
	if err != nil {
		t.Fatalf("failed to read sub claim: %v", err)
	}
	if sub != wantUID {
		t.Errorf("sub = %q, want %q", sub, wantUID)
	}
	iss, err := claims.GetIssuer()
	if err != nil {
		t.Fatalf("failed to read iss claim: %v", err)
	}
	if iss != "dockloud-auth" {
		t.Errorf("iss = %q, want dockloud-auth", iss)
	}
}

func TestService_Register(t *testing.T) {
	t.Setenv("JWT_AUTH_KEY", testJWTKey)

	created := &user.User{UID: "uid-1", Username: "alice", Email: "alice@example.com"}
	valid := &RegisterInput{Username: "alice", Email: "alice@example.com", Password: "password123"}

	tests := []struct {
		name    string
		input   *RegisterInput
		setup   func(m *mocks.MockRepository)
		want    *RegisterOutput
		wantErr error
	}{
		{
			name:  "success",
			input: valid,
			setup: func(m *mocks.MockRepository) {
				m.EXPECT().
					Search(gomock.Any(), nilStrPtrMatcher{}, gomock.Eq(strPtr("alice")), gomock.Eq(strPtr("alice@example.com"))).
					Return(&user.SearchResult{}, nil)
				m.EXPECT().
					Create(gomock.Any(), "alice", hashPassword("password123"), "alice@example.com").
					Return(created, nil)
			},
			want: &RegisterOutput{UID: "uid-1", Username: "alice", Email: "alice@example.com"},
		},
		{
			name:    "empty username",
			input:   &RegisterInput{Username: "", Email: "alice@example.com", Password: "password123"},
			wantErr: user.ErrInvalidUsername,
		},
		{
			name:    "username too short",
			input:   &RegisterInput{Username: "ab", Email: "alice@example.com", Password: "password123"},
			wantErr: user.ErrInvalidUsername,
		},
		{
			name:    "username too long",
			input:   &RegisterInput{Username: strings.Repeat("a", 65), Email: "alice@example.com", Password: "password123"},
			wantErr: user.ErrInvalidUsername,
		},
		{
			name:    "invalid email",
			input:   &RegisterInput{Username: "alice", Email: "not-an-email", Password: "password123"},
			wantErr: user.ErrInvalidEmail,
		},
		{
			name:    "password too short",
			input:   &RegisterInput{Username: "alice", Email: "alice@example.com", Password: "pass1"},
			wantErr: user.ErrInvalidPassword,
		},
		{
			name:    "password with special characters",
			input:   &RegisterInput{Username: "alice", Email: "alice@example.com", Password: "pass!w0rd1"},
			wantErr: user.ErrInvalidPassword,
		},
		{
			name:  "username already taken",
			input: valid,
			setup: func(m *mocks.MockRepository) {
				m.EXPECT().
					Search(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(&user.SearchResult{UsernameExists: true}, nil)
			},
			wantErr: ErrUsernameTaken,
		},
		{
			name:  "email already taken",
			input: valid,
			setup: func(m *mocks.MockRepository) {
				m.EXPECT().
					Search(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(&user.SearchResult{EmailExists: true}, nil)
			},
			wantErr: ErrEmailTaken,
		},
		{
			name:  "search failed",
			input: valid,
			setup: func(m *mocks.MockRepository) {
				m.EXPECT().
					Search(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, errSearchFailed)
			},
			wantErr: errSearchFailed,
		},
		{
			name:  "create failed",
			input: valid,
			setup: func(m *mocks.MockRepository) {
				m.EXPECT().
					Search(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(&user.SearchResult{}, nil)
				m.EXPECT().
					Create(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, errCreateFailed)
			},
			wantErr: errCreateFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newTestService(t, tt.setup)

			got, err := svc.Register(context.Background(), tt.input)
			assertError(t, err, tt.wantErr)
			if tt.wantErr != nil {
				if got != nil {
					t.Fatalf("got = %+v, want nil", got)
				}
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("got = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestService_Login(t *testing.T) {
	t.Setenv("JWT_AUTH_KEY", testJWTKey)

	authenticated := &user.User{UID: "uid-1", Username: "alice", Email: "alice@example.com"}
	valid := &LoginInput{Email: "alice@example.com", Password: "password123"}

	tests := []struct {
		name    string
		input   *LoginInput
		setup   func(m *mocks.MockRepository)
		wantErr error
		check   func(t *testing.T, got *LoginOutput)
	}{
		{
			name:  "success",
			input: valid,
			setup: func(m *mocks.MockRepository) {
				m.EXPECT().
					Authenticate(gomock.Any(), "alice@example.com", hashPassword("password123")).
					Return(authenticated, nil)
			},
			check: func(t *testing.T, got *LoginOutput) {
				if got.UID != "uid-1" {
					t.Errorf("UID = %q, want uid-1", got.UID)
				}
				if got.JWTToken == "" {
					t.Fatal("JWTToken is empty")
				}
				validateJWT(t, got.JWTToken, "uid-1")
				if d := time.Until(got.ExpiresAt); d < time.Hour || d > 2*time.Hour+time.Minute {
					t.Errorf("ExpiresAt = %v, want ~2h from now", got.ExpiresAt)
				}
			},
		},
		{
			name:    "invalid email",
			input:   &LoginInput{Email: "not-an-email", Password: "password123"},
			wantErr: user.ErrInvalidEmail,
		},
		{
			name:  "authenticate failed",
			input: valid,
			setup: func(m *mocks.MockRepository) {
				m.EXPECT().
					Authenticate(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, errAuthFailed)
			},
			wantErr: errAuthFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newTestService(t, tt.setup)

			got, err := svc.Login(context.Background(), tt.input)
			assertError(t, err, tt.wantErr)
			if tt.wantErr != nil {
				if got != nil {
					t.Fatalf("got = %+v, want nil", got)
				}
				return
			}
			if tt.check == nil {
				t.Fatal("check is required for success case")
			}
			tt.check(t, got)
		})
	}
}

func TestService_Update(t *testing.T) {
	t.Setenv("JWT_AUTH_KEY", testJWTKey)

	const uid = "uid-1"
	valid := &UpdateInput{
		Username: strPtr("newname"),
		Email:    strPtr("new@example.com"),
		Password: strPtr("newpassword1"),
	}
	updated := &user.User{UID: uid}

	tests := []struct {
		name    string
		input   *UpdateInput
		setup   func(m *mocks.MockRepository)
		wantErr error
	}{
		{
			name:  "success",
			input: valid,
			setup: func(m *mocks.MockRepository) {
				m.EXPECT().
					Update(gomock.Any(), uid, gomock.Eq(strPtr("newname")), gomock.Eq(strPtr("new@example.com")), gomock.Eq(strPtr("newpassword1"))).
					Return(updated, nil)
			},
		},
		{
			name: "invalid username",
			input: &UpdateInput{
				Username: strPtr("ab"),
				Email:    strPtr("new@example.com"),
				Password: strPtr("newpassword1"),
			},
			wantErr: user.ErrInvalidUsername,
		},
		{
			name: "invalid email",
			input: &UpdateInput{
				Username: strPtr("newname"),
				Email:    strPtr("bad-email"),
				Password: strPtr("newpassword1"),
			},
			wantErr: user.ErrInvalidEmail,
		},
		{
			name: "invalid password",
			input: &UpdateInput{
				Username: strPtr("newname"),
				Email:    strPtr("new@example.com"),
				Password: strPtr("short"),
			},
			wantErr: user.ErrInvalidPassword,
		},
		{
			name:  "update failed",
			input: valid,
			setup: func(m *mocks.MockRepository) {
				m.EXPECT().
					Update(gomock.Any(), uid, gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, errUpdateFailed)
			},
			wantErr: errUpdateFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newTestService(t, tt.setup)

			err := svc.Update(context.Background(), uid, tt.input)
			assertError(t, err, tt.wantErr)
		})
	}
}

func TestService_Delete(t *testing.T) {
	t.Setenv("JWT_AUTH_KEY", testJWTKey)

	const uid = "uid-1"
	existing := &user.User{UID: uid}

	tests := []struct {
		name    string
		setup   func(m *mocks.MockRepository)
		wantErr error
	}{
		{
			name: "success",
			setup: func(m *mocks.MockRepository) {
				m.EXPECT().
					Search(gomock.Any(), gomock.Eq(strPtr(uid)), nilStrPtrMatcher{}, nilStrPtrMatcher{}).
					Return(&user.SearchResult{UIDExists: true}, nil)
				m.EXPECT().
					Delete(gomock.Any(), uid).
					Return(existing, nil)
			},
		},
		{
			// Delete не должен вызываться: ожидания на него нет,
			// поэтому случайный вызов провалит тест.
			name: "user not found",
			setup: func(m *mocks.MockRepository) {
				m.EXPECT().
					Search(gomock.Any(), gomock.Eq(strPtr(uid)), nilStrPtrMatcher{}, nilStrPtrMatcher{}).
					Return(&user.SearchResult{UIDExists: false}, nil)
			},
			wantErr: ErrUserNotFound,
		},
		{
			name: "search failed",
			setup: func(m *mocks.MockRepository) {
				m.EXPECT().
					Search(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, errSearchFailed)
			},
			wantErr: errSearchFailed,
		},
		{
			name: "delete failed",
			setup: func(m *mocks.MockRepository) {
				m.EXPECT().
					Search(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(&user.SearchResult{UIDExists: true}, nil)
				m.EXPECT().
					Delete(gomock.Any(), uid).
					Return(nil, errDeleteFailed)
			},
			wantErr: errDeleteFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newTestService(t, tt.setup)

			err := svc.Delete(context.Background(), uid)
			assertError(t, err, tt.wantErr)
		})
	}
}
