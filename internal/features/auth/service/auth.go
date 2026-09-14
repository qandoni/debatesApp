package auth_service

import (
	"context"
	"fmt"

	"github.com/qandoni/debatesApp/internal/core/domain"
	core_errors "github.com/qandoni/debatesApp/internal/core/errors"
	auth_contracts "github.com/qandoni/debatesApp/internal/features/auth/contracts"
)

func (s *AuthService) Login(
	ctx context.Context,
	input auth_contracts.LoginInput,
) (auth_contracts.LoginOutput, error) {
	user, err := s.usersRepo.GetUserByEmail(
		ctx,
		input.Email,
	)
	if err != nil {
		return auth_contracts.LoginOutput{}, fmt.Errorf("get user by email: %w", err)
	}
	err = s.passwordHasher.Compare(
		user.PasswordHash,
		input.Password,
	)
	if err != nil {
		return auth_contracts.LoginOutput{}, fmt.Errorf("compare password: %w", core_errors.ErrUnauthorized)
	}

	accessToken, err := s.jwtManager.GenerateAccessToken(user)
	if err != nil {
		return auth_contracts.LoginOutput{}, err
	}
	return auth_contracts.LoginOutput{
		AccessToken: accessToken,
	}, nil
}

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
