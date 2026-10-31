package user

import "time"

type User struct {
	UID          string
	Username     string
	Email        string
	PasswordHash string
	CreatedAt    time.Time
}
