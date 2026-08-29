package users_service

import (
	"context"
	"fmt"

	"github.com/qandoni/debatesApp/internal/core/domain"
)

func (s *UsersService) EditProfile(ctx context.Context, userID int, profilePatch domain.UserPatch) (domain.User, error) {
	user, err := s.usersRepository.GetMyProfile(ctx, userID)
	if err != nil {
		return domain.User{}, fmt.Errorf("get user: %w", err)
	}
	if user, err = s.ApplyPatch(user, profilePatch); err != nil {
		return domain.User{}, fmt.Errorf("apply user patch: %w", err)
	}
	patchedUser, err := s.usersRepository.EditProfile(ctx, userID, user)
	if err != nil {
		return domain.User{}, fmt.Errorf("edit profile: %w", err)
	}
	return patchedUser, nil
}

func (s *UsersService) ApplyPatch(
	user domain.User,
	patch domain.UserPatch,
) (domain.User, error) {
	tmp := user
	if patch.Username.Set {
		tmp.Username = *patch.Username.Value
	}
	if patch.Email.Set {
		tmp.Email = *patch.Email.Value
	}
	if patch.Password.Set {
		if patch.Password.Value == nil {
			return domain.User{}, fmt.Errorf("password can't be NULL")
		}
		hashed, err := s.passwordHasher.Hash(*patch.Password.Value)
		if err != nil {
			return domain.User{}, fmt.Errorf("hash password: %w", err)
		}
		tmp.PasswordHash = hashed
	}
	if patch.Bio.Set {
		tmp.Bio = patch.Bio.Value
	}

	return tmp, nil
}
