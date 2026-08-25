package auth_contracts

type LoginInput struct {
	Email    string
	Password string
}

type LoginOutput struct {
	AccessToken string
}
