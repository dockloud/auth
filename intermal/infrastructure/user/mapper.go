package user

import (
	"github.com/dockloud/auth/intermal/domain/user"
)

func DomainToCacheable(user user.User) *CacheableUser {
	return &CacheableUser{
		UID:          user.UID,
		Username:     user.Username,
		Email:        user.Email,
		PasswordHash: user.PasswordHash,
		CreatedAt:    user.CreatedAt,
	}
}
