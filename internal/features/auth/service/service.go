package auth_service

import (
	"context"

	core_auth "github.com/qandoni/debatesApp/internal/core/auth"
	"github.com/qandoni/debatesApp/internal/core/domain"
	core_postgres "github.com/qandoni/debatesApp/internal/core/repository/postgres"
)

type AuthService struct {
	usersRepo      UsersRepository
	passwordHasher PasswordHasher
	sha256Hasher   Sha256Hasher
	jwtManager     JWTManager
	txManager      core_postgres.TransactionManager
}

type UsersRepository interface {
	GetUserByEmail(
		ctx context.Context,
		email string,
	) (domain.User, error)
	CreateUser(
		ctx context.Context,
		input domain.User,
	) (domain.User, error)
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

type Sha256Hasher interface {
	Hash(value string) string
}

type JWTManager interface {
	GenerateAccessToken(
		user domain.User,
	) (string, error)
	ParseAccessToken(
		token string,
	) (core_auth.AuthInfo, error)
}

func NewAuthService(
	usersRepository UsersRepository,
	passwordHasher PasswordHasher,
	sha256Hasher Sha256Hasher,
	jwtManager JWTManager,
	txManager core_postgres.TransactionManager,
) *AuthService {
	return &AuthService{
		usersRepo:      usersRepository,
		passwordHasher: passwordHasher,
		sha256Hasher:   sha256Hasher,
		jwtManager:     jwtManager,
		txManager:      txManager,
	}
}
