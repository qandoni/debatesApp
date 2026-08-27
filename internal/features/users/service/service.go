package users_service

import (
	"context"

	"github.com/qandoni/debatesApp/internal/core/domain"
)

type UsersService struct {
	usersRepository UsersRepository
}
type UsersRepository interface {
	GetMyProfile(ctx context.Context, usersID int) (domain.User, error)
}

func NewUsersService(
	usersRepository UsersRepository,
) *UsersService {
	return &UsersService{
		usersRepository: usersRepository,
	}
}
