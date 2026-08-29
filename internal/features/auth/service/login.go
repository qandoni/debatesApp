package auth_service

import (
	"context"
	"fmt"

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
