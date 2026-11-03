package user

import "time"

type RegisterInput struct {
	Username string
	Email    string
	Password string
}

type RegisterOutput struct {
	UID      string
	Username string
	Email    string
}

type LoginInput struct {
	Email    string
	Password string
}

type LoginOutput struct {
	UID       string
	JWTToken  string
	ExpiresAt time.Time
}

type UpdateInput struct {
	Username *string
	Email    *string
	Password *string
}
