package users_service

import (
	"context"

	"github.com/qandoni/debatesApp/internal/core/domain"
)

type UsersService struct {
	usersRepository UsersRepository
	passwordHasher  PasswordHasher
}
type UsersRepository interface {
	GetMyProfile(ctx context.Context, usersID int) (domain.User, error)
	EditProfile(ctx context.Context, userID int, user domain.User) (domain.User, error)
}

type PasswordHasher interface {
	Hash(
		password string,
	) (string, error)
	Compare(
		hash string,
		password string,
	) error
}

func NewUsersService(
	usersRepository UsersRepository,
	passwordHasher PasswordHasher,
) *UsersService {
	return &UsersService{
		usersRepository: usersRepository,
		passwordHasher:  passwordHasher,
	}
}
