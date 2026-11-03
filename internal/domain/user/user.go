package user

import (
	"time"
)

type User struct {
	UID          string
	Username     string
	Email        string
	PasswordHash string
	CreatedAt    time.Time
}

func (u *User) UsernameValid() error {
	return ValidateUsername(u.Username)
}

func (u *User) EmailValid() error {
	return ValidateEmail(u.Email)
}

func (u *User) ChangeUsername(username string) error {
	err := ValidateUsername(username)
	if err != nil {
		return err
	}
	u.Username = username
	return nil
}

func (u *User) ChangeEmail(email string) error {
	err := ValidateEmail(email)
	if err != nil {
		return err
	}
	u.Email = email
	return nil
}
