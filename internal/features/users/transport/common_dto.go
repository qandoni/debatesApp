package users_http_transport

import (
	"fmt"
	"net/mail"
	"time"

	"github.com/qandoni/debatesApp/internal/core/domain"
	core_http_types "github.com/qandoni/debatesApp/internal/core/transport/http/types"
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

type PatchUserRequest struct {
	Username core_http_types.Nullable[string] `json:"username"`
	Email    core_http_types.Nullable[string] `json:"email"`
	Password core_http_types.Nullable[string] `json:"password"`
	Bio      core_http_types.Nullable[string] `json:"bio"`
}

func (r *PatchUserRequest) Validate() error {
	if r.Username.Set {
		if r.Username.Value == nil {
			return fmt.Errorf("'Username' can't be NULL")
		}
		usernameLen := len([]rune(*r.Username.Value))
		if usernameLen < 5 || usernameLen > 25 {
			return fmt.Errorf("'Username' must be between 5 and 25 symbols")
		}
	}
	if r.Email.Set {
		if r.Email.Value == nil {
			return fmt.Errorf("'Email' can't be NULL")
		}
		_, err := mail.ParseAddress(*r.Email.Value)
		if err != nil {
			return fmt.Errorf("wrong format of the email")
		}
		emailLen := len([]rune(*r.Email.Value))
		if emailLen < 5 || emailLen > 40 {
			return fmt.Errorf("'Email' must be between 5 and 40 symbols")
		}
	}
	if r.Password.Set {
		if r.Password.Value == nil {
			return fmt.Errorf("'Password' can't be NULL")
		}
		passwordLen := len([]rune(*r.Password.Value))
		if passwordLen < 6 || passwordLen > 50 {
			return fmt.Errorf("'Password' must be between 6 and 50 symbols")
		}
	}

	return nil
}
