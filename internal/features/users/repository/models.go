package users_repository

import (
	"time"

	"github.com/qandoni/debatesApp/internal/core/domain"
)

type UserModel struct {
	ID           int
	Version      int
	Username     string
	Email        string
	PasswordHash string
	AvatarURL    *string
	Bio          *string
	CreatedAt    time.Time
	UpdatedAt    *time.Time
}

func userDomainFromModel(userModel UserModel) domain.User {
	return domain.NewUser(
		userModel.ID,
		userModel.Version,
		userModel.Username,
		userModel.Email,
		userModel.PasswordHash,
		userModel.AvatarURL,
		userModel.Bio,
		userModel.CreatedAt,
		userModel.UpdatedAt,
	)
}
