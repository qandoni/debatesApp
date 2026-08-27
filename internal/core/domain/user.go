package domain

import "time"

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
