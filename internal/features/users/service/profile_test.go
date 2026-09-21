package users_service

import (
	"context"
	"testing"

	"github.com/qandoni/debatesApp/internal/core/domain"
)

type mockUsersRepository struct {
	getMyProfileFn func(ctx context.Context, userID int) (domain.User, error)
	editProfileFn  func(ctx context.Context, userID int, user domain.User) (domain.User, error)
}

func (m *mockUsersRepository) GetMyProfile(ctx context.Context, userID int) (domain.User, error) {
	return m.getMyProfileFn(ctx, userID)
}

func (m *mockUsersRepository) EditProfile(ctx context.Context, userID int, user domain.User) (domain.User, error) {
	return m.editProfileFn(ctx, userID, user)
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

func newUsersService(
	users UsersRepository,
	hasher PasswordHasher,
) *UsersService {
	return NewUsersService(users, hasher)
}

func baseUser() domain.User {
	return domain.User{
		ID:           1,
		Username:     "john",
		Email:        "john@example.com",
		PasswordHash: "old_hash",
		Bio:          nil,
	}
}

func TestApplyPatch_UpdateUsernameAndEmail(t *testing.T) {
	svc := newUsersService(&mockUsersRepository{}, &mockPasswordHasher{})

	newUsername := "new_john"
	newEmail := "new@example.com"

	patched, err := svc.ApplyPatch(baseUser(), domain.UserPatch{
		Username: domain.Nullable[string]{Value: &newUsername, Set: true},
		Email:    domain.Nullable[string]{Value: &newEmail, Set: true},
	})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if patched.Username != "new_john" {
		t.Fatalf("expected 'new_john', got: %s", patched.Username)
	}
	if patched.Email != "new@example.com" {
		t.Fatalf("expected 'new@example.com', got: %s", patched.Email)
	}
	if patched.PasswordHash != "old_hash" {
		t.Fatalf("password hash should not change, got: %s", patched.PasswordHash)
	}
}

func TestApplyPatch_UpdatePasswordHashes(t *testing.T) {
	hasher := &mockPasswordHasher{
		hashFn: func(password string) (string, error) {
			if password != "new_secret" {
				t.Fatalf("expected 'new_secret', got: %s", password)
			}
			return "new_hashed", nil
		},
	}
	svc := newUsersService(&mockUsersRepository{}, hasher)

	newPassword := "new_secret"

	patched, err := svc.ApplyPatch(baseUser(), domain.UserPatch{
		Password: domain.Nullable[string]{Value: &newPassword, Set: true},
	})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if patched.PasswordHash != "new_hashed" {
		t.Fatalf("expected 'new_hashed', got: %s", patched.PasswordHash)
	}
}

func TestApplyPatch_PasswordNull(t *testing.T) {
	svc := newUsersService(&mockUsersRepository{}, &mockPasswordHasher{})

	_, err := svc.ApplyPatch(baseUser(), domain.UserPatch{
		Password: domain.Nullable[string]{Value: nil, Set: true},
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestEditProfile_Success(t *testing.T) {
	usersRepo := &mockUsersRepository{
		getMyProfileFn: func(ctx context.Context, userID int) (domain.User, error) {
			if userID != 1 {
				t.Fatalf("expected user 1, got %d", userID)
			}
			return baseUser(), nil
		},
		editProfileFn: func(ctx context.Context, userID int, user domain.User) (domain.User, error) {
			if userID != 1 {
				t.Fatalf("expected user 1, got %d", userID)
			}
			if user.Username != "new_john" {
				t.Fatalf("expected 'new_john', got: %s", user.Username)
			}
			return user, nil
		},
	}
	svc := newUsersService(usersRepo, &mockPasswordHasher{})

	newUsername := "new_john"

	patched, err := svc.EditProfile(context.Background(), 1, domain.UserPatch{
		Username: domain.Nullable[string]{Value: &newUsername, Set: true},
	})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if patched.Username != "new_john" {
		t.Fatalf("expected 'new_john', got: %s", patched.Username)
	}
}
