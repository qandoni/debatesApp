package auth_service

import (
	"context"
	"errors"
	"testing"

	core_auth "github.com/qandoni/debatesApp/internal/core/auth"
	"github.com/qandoni/debatesApp/internal/core/domain"
	core_errors "github.com/qandoni/debatesApp/internal/core/errors"
	core_postgres "github.com/qandoni/debatesApp/internal/core/repository/postgres"
	auth_contracts "github.com/qandoni/debatesApp/internal/features/auth/contracts"
)

type mockUsersRepository struct {
	getByEmailFn func(ctx context.Context, email string) (domain.User, error)
	createUserFn func(ctx context.Context, user domain.User) (domain.User, error)
}

func (m *mockUsersRepository) GetUserByEmail(ctx context.Context, email string) (domain.User, error) {
	return m.getByEmailFn(ctx, email)
}

func (m *mockUsersRepository) CreateUser(ctx context.Context, user domain.User) (domain.User, error) {
	return m.createUserFn(ctx, user)
}

type mockPasswordHasher struct {
	hashFn    func(password string) (string, error)
	compareFn func(hash, password string) error
}

func (m *mockPasswordHasher) Hash(password string) (string, error) {
	return m.hashFn(password)
}

func (m *mockPasswordHasher) Compare(hash, password string) error {
	return m.compareFn(hash, password)
}

type mockSha256Hasher struct {
	hashFn func(value string) string
}

func (m *mockSha256Hasher) Hash(value string) string {
	return m.hashFn(value)
}

type mockJWTManager struct {
	generateFn func(user domain.User) (string, error)
	parseFn    func(token string) (core_auth.AuthInfo, error)
}

func (m *mockJWTManager) GenerateAccessToken(user domain.User) (string, error) {
	return m.generateFn(user)
}

func (m *mockJWTManager) ParseAccessToken(token string) (core_auth.AuthInfo, error) {
	return m.parseFn(token)
}

type mockTxManager struct {
	withinFn func(ctx context.Context, fn func(ctx context.Context) error) error
}

func (m *mockTxManager) WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return m.withinFn(ctx, fn)
}

func newAuthService(
	users UsersRepository,
	hasher PasswordHasher,
	sha256 Sha256Hasher,
	jwt JWTManager,
	tx core_postgres.TransactionManager,
) *AuthService {
	return NewAuthService(users, hasher, sha256, jwt, tx)
}

func TestRegistrate_Success(t *testing.T) {
	usersRepo := &mockUsersRepository{
		createUserFn: func(ctx context.Context, user domain.User) (domain.User, error) {
			if user.Username != "john_doe" || user.Email != "john@example.com" {
				t.Fatalf("unexpected user: %+v", user)
			}
			if user.PasswordHash != "hashed_password" {
				t.Fatalf("expected hashed password, got: %s", user.PasswordHash)
			}
			return user, nil
		},
	}
	hasher := &mockPasswordHasher{
		hashFn: func(password string) (string, error) {
			if password != "secret123" {
				t.Fatalf("expected 'secret123', got: %s", password)
			}
			return "hashed_password", nil
		},
	}

	svc := newAuthService(usersRepo, hasher, &mockSha256Hasher{}, &mockJWTManager{}, &mockTxManager{})
	err := svc.Registrate(context.Background(), auth_contracts.RegistrationInput{
		UserName: "john_doe",
		Email:    "john@example.com",
		Password: "secret123",
	})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func TestRegistrate_InvalidInput(t *testing.T) {
	usersRepo := &mockUsersRepository{
		createUserFn: func(ctx context.Context, user domain.User) (domain.User, error) {
			t.Fatal("CreateUser should not be called")
			return domain.User{}, nil
		},
	}

	svc := newAuthService(usersRepo, &mockPasswordHasher{}, &mockSha256Hasher{}, &mockJWTManager{}, &mockTxManager{})
	err := svc.Registrate(context.Background(), auth_contracts.RegistrationInput{
		UserName: "ab",
		Email:    "x",
		Password: "123",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, core_errors.ErrInvalidArgument) {
		t.Fatalf("expected ErrInvalidArgument, got: %v", err)
	}
}

func TestLogin_Success(t *testing.T) {
	usersRepo := &mockUsersRepository{
		getByEmailFn: func(ctx context.Context, email string) (domain.User, error) {
			if email != "john@example.com" {
				t.Fatalf("expected email, got: %s", email)
			}
			return domain.User{ID: 1, Email: "john@example.com", PasswordHash: "hash"}, nil
		},
	}
	hasher := &mockPasswordHasher{
		compareFn: func(hash, password string) error {
			if hash != "hash" || password != "secret123" {
				t.Fatalf("unexpected compare args: %s %s", hash, password)
			}
			return nil
		},
	}
	jwt := &mockJWTManager{
		generateFn: func(user domain.User) (string, error) {
			if user.ID != 1 {
				t.Fatalf("expected user 1, got %d", user.ID)
			}
			return "access_token", nil
		},
	}

	svc := newAuthService(usersRepo, hasher, &mockSha256Hasher{}, jwt, &mockTxManager{})
	out, err := svc.Login(context.Background(), auth_contracts.LoginInput{
		Email:    "john@example.com",
		Password: "secret123",
	})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if out.AccessToken != "access_token" {
		t.Fatalf("expected 'access_token', got: %s", out.AccessToken)
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	usersRepo := &mockUsersRepository{
		getByEmailFn: func(ctx context.Context, email string) (domain.User, error) {
			return domain.User{ID: 1, Email: "john@example.com", PasswordHash: "hash"}, nil
		},
	}
	hasher := &mockPasswordHasher{
		compareFn: func(hash, password string) error {
			return errors.New("mismatch")
		},
	}
	jwt := &mockJWTManager{
		generateFn: func(user domain.User) (string, error) {
			t.Fatal("GenerateAccessToken should not be called")
			return "", nil
		},
	}

	svc := newAuthService(usersRepo, hasher, &mockSha256Hasher{}, jwt, &mockTxManager{})
	_, err := svc.Login(context.Background(), auth_contracts.LoginInput{
		Email:    "john@example.com",
		Password: "wrong",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, core_errors.ErrUnauthorized) {
		t.Fatalf("expected ErrUnauthorized, got: %v", err)
	}
}
