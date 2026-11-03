package user

import "time"

type CacheableUser struct {
	UID          string    `redis:"uid"`
	Username     string    `redis:"username"`
	Email        string    `redis:"email"`
	PasswordHash string    `redis:"password_hash"`
	CreatedAt    time.Time `redis:"created_at"`
}
