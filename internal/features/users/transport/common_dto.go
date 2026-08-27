package users_http_transport

import (
	"time"

	"github.com/qandoni/debatesApp/internal/core/domain"
)

type UserDTOResponse struct {
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

func UserDTOFromDomain(user domain.User) UserDTOResponse {
	return UserDTOResponse{
		ID:           user.ID,
		Version:      user.Version,
		Username:     user.Username,
		Email:        user.Email,
		PasswordHash: user.PasswordHash,
		AvatarURL:    user.AvatarURL,
		Bio:          user.Bio,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
	}
}
