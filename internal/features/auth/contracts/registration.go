package auth_contracts

import (
	"fmt"

	core_errors "github.com/qandoni/debatesApp/internal/core/errors"
)

type RegistrationInput struct {
	UserName string
	Email    string
	Password string
}

func (i *RegistrationInput) Validate() error {
	userNameLen := len([]rune(i.UserName))
	if userNameLen < 5 || userNameLen > 40 {
		return fmt.Errorf(
			"invalid 'UserName' len: %d: %w",
			userNameLen,
			core_errors.ErrInvalidArgument,
		)
	}
	emailLen := len([]rune(i.Email))
	if emailLen < 3 || emailLen > 100 {
		return fmt.Errorf(
			"invalid 'Email' len: %d: %w",
			emailLen,
			core_errors.ErrInvalidArgument,
		)
	}
	passwordLen := len([]rune(i.Password))
	if passwordLen < 6 || passwordLen > 50 {
		return fmt.Errorf(
			"invalid 'Password' len: %d: %w",
			passwordLen,
			core_errors.ErrInvalidArgument,
		)
	}
	return nil
}
