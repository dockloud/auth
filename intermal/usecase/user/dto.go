package user

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
	UID      string
	Username string
	Email    string
}

type UpdateInput struct {
	Username *string
	Email    *string
	Password *string
}
