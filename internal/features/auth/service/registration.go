package auth_service

import (
	"context"
	"fmt"

	"github.com/qandoni/debatesApp/internal/core/domain"
	auth_contracts "github.com/qandoni/debatesApp/internal/features/auth/contracts"
)

func (s *AuthService) Registrate(
	ctx context.Context,
	input auth_contracts.RegistrationInput,
) error {
	if err := input.Validate(); err != nil {
		return fmt.Errorf("validate input data: %w", err)
	}
	passwordHash, err := s.passwordHasher.Hash(input.Password)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}
	user := domain.NewUserUninitialized(input.UserName, input.Email, passwordHash, nil, nil)
	_, err = s.usersRepo.CreateUser(ctx, user)
	if err != nil {
		return fmt.Errorf("create user in repository: %w", err)
	}
	return nil
}
