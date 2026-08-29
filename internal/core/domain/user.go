package domain

import (
	"fmt"
	"time"

	core_errors "github.com/qandoni/debatesApp/internal/core/errors"
)

type User struct {
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

func NewUser(
	iD int,
	version int,
	username string,
	email string,
	passwordHash string,
	avatarURL *string,
	bio *string,
	createdAt time.Time,
	updatedAt *time.Time,

) User {
	return User{
		iD,
		version,
		username,
		email,
		passwordHash,
		avatarURL,
		bio,
		createdAt,
		updatedAt,
	}
}

func NewUserUninitialized(
	username string,
	email string,
	passwordHash string,
	avatarURL *string,
	bio *string,
) User {
	return NewUser(
		UninitializedID,
		UninitializedVersion,
		username,
		email,
		passwordHash,
		avatarURL,
		bio,
		time.Now(),
		nil,
	)
}

type UserPatch struct {
	Username Nullable[string]
	Email    Nullable[string]
	Password Nullable[string]
	Bio      Nullable[string]
}

func NewUserPatch(
	username Nullable[string],
	email Nullable[string],
	password Nullable[string],
	bio Nullable[string],
) UserPatch {
	return UserPatch{
		Username: username,
		Email:    email,
		Password: password,
		Bio:      bio,
	}
}
func (p *UserPatch) Validate() error {
	if p.Username.Set && p.Username.Value == nil {
		return fmt.Errorf("'Username' can't be patched to NULL: %w", core_errors.ErrInvalidArgument)
	}
	if p.Email.Set && p.Email.Value == nil {
		return fmt.Errorf("`Email` can't be patched to NULL: %w", core_errors.ErrInvalidArgument)
	}

	if p.Password.Set && p.Password.Value == nil {
		return fmt.Errorf("'Password' can't be patched to NULL : %w", core_errors.ErrInvalidArgument)
	}
	return nil
}
