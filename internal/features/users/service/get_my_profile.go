package users_service

import (
	"context"
	"fmt"

	"github.com/qandoni/debatesApp/internal/core/domain"
)

func (s *UsersService) GetMyProfile(ctx context.Context, userID int) (domain.User, error) {
	user, err := s.usersRepository.GetMyProfile(ctx, userID)
	if err != nil {
		return domain.User{}, fmt.Errorf("get user from repository: %w", err)
	}
	return user, nil
}
