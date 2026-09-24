package integration_test

import (
	"context"
	"strings"
	"sync"
	"testing"

	auth_contracts "github.com/qandoni/debatesApp/internal/features/auth/contracts"

	"github.com/qandoni/debatesApp/internal/core/domain"
	core_errors "github.com/qandoni/debatesApp/internal/core/errors"
)

func TestRegistration_Login_FullFlow(t *testing.T) {
	email := uniqueEmail()
	password := "S3cure-Pass!"

	err := itAuthService.Registrate(context.Background(), auth_contracts.RegistrationInput{
		UserName: "integration_user",
		Email:    email,
		Password: password,
	})
	if err != nil {
		t.Fatalf("registrate: %v", err)
	}

	user, err := itUsersRepo.GetUserByEmail(context.Background(), email)
	if err != nil {
		t.Fatalf("get user by email: %v", err)
	}
	if user.PasswordHash == password {
		t.Fatal("password stored in plaintext!")
	}
	if strings.HasPrefix(user.PasswordHash, "$2") == false {
		t.Fatalf("expected bcrypt hash, got: %s", user.PasswordHash)
	}

	output, err := itAuthService.Login(context.Background(), auth_contracts.LoginInput{
		Email:    email,
		Password: password,
	})
	if err != nil {
		t.Fatalf("login: %v", err)
	}

	info, err := itJWTManager.ParseAccessToken(output.AccessToken)
	if err != nil {
		t.Fatalf("parse token: %v", err)
	}
	if info.UserID != user.ID {
		t.Fatalf("expected userID %d in token, got %d", user.ID, info.UserID)
	}
}

func TestLogin_WrongPassword_Unauthorized(t *testing.T) {
	email := uniqueEmail()

	if err := itAuthService.Registrate(context.Background(), auth_contracts.RegistrationInput{
		UserName: "wrong_pass_user",
		Email:    email,
		Password: "correct-password",
	}); err != nil {
		t.Fatalf("registrate: %v", err)
	}

	_, err := itAuthService.Login(context.Background(), auth_contracts.LoginInput{
		Email:    email,
		Password: "wrong-password",
	})
	if err == nil {
		t.Fatal("expected error for wrong password, got nil")
	}
	if !is(err, core_errors.ErrUnauthorized) {
		t.Fatalf("expected ErrUnauthorized, got: %v", err)
	}
}

func TestLogin_UnknownEmail_NotFound(t *testing.T) {
	_, err := itAuthService.Login(context.Background(), auth_contracts.LoginInput{
		Email:    "no-such-user@test.it",
		Password: "whatever",
	})
	if err == nil {
		t.Fatal("expected error for unknown email, got nil")
	}
	if !isNotFound(err) {
		t.Fatalf("expected ErrNotFound, got: %v", err)
	}
}

func TestRegistration_DuplicateEmail_Sequential(t *testing.T) {
	email := uniqueEmail()

	input := auth_contracts.RegistrationInput{
		UserName: "duplicate_email_user",
		Email:    email,
		Password: "password-1",
	}
	if err := itAuthService.Registrate(context.Background(), input); err != nil {
		t.Fatalf("first registration: %v", err)
	}

	err := itAuthService.Registrate(context.Background(), input)
	if err == nil {
		t.Fatal("expected error on duplicate email registration, got nil")
	}
	if !isConflict(err) {
		t.Fatalf("raw DB error leaked from duplicate registration (unique violation not mapped to ErrConflict): %v", err)
	}
	if n := countUsersByEmail(t, email); n != 1 {
		t.Fatalf("expected exactly 1 user with email, got %d", n)
	}
}

func TestRegistration_DuplicateEmail_Concurrent(t *testing.T) {
	email := uniqueEmail()

	const attempts = 8
	gate := newStartGate(attempts)
	var wg sync.WaitGroup
	results := make(chan error, attempts)

	for i := 0; i < attempts; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			gate.wait()
			err := itAuthService.Registrate(context.Background(), auth_contracts.RegistrationInput{
				UserName: "concurrent_dup_user",
				Email:    email,
				Password: "password-1",
			})
			results <- err
		}()
	}
	gate.open()
	wg.Wait()
	close(results)

	successes := 0
	for err := range results {
		if err == nil {
			successes++
			continue
		}
		if !isConflict(err) {
			t.Fatalf("raw DB error leaked from concurrent registration: %v", err)
		}
	}
	if successes != 1 {
		t.Fatalf("expected exactly 1 successful registration, got %d", successes)
	}
	if n := countUsersByEmail(t, email); n != 1 {
		t.Fatalf("expected exactly 1 user with email, got %d", n)
	}
}

func TestEditProfile_OptimisticLocking_Concurrent(t *testing.T) {
	user := createUser(t)

	newName1 := "concurrent_name_1"
	newName2 := "concurrent_name_2"

	gate := newStartGate(2)
	var wg sync.WaitGroup
	results := make(chan error, 2)

	patch := func(name string) func(ctx context.Context) {
		return func(ctx context.Context) {
			_, err := itUsersService.EditProfile(ctx, user.ID, domain.NewUserPatch(
				domain.Nullable[string]{Value: &name, Set: true},
				domain.Nullable[string]{},
				domain.Nullable[string]{},
				domain.Nullable[string]{},
			))
			results <- err
		}
	}

	wg.Add(2)
	go func() { defer wg.Done(); f := patch(newName1); gate.wait(); f(context.Background()) }()
	go func() { defer wg.Done(); f := patch(newName2); gate.wait(); f(context.Background()) }()

	gate.open()
	wg.Wait()
	close(results)

	successes := 0
	for err := range results {
		if err == nil {
			successes++
			continue
		}
		if !isConflict(err) {
			t.Fatalf("unexpected error type on concurrent profile edit: %v", err)
		}
	}
	if successes != 1 {
		t.Fatalf("expected exactly 1 successful concurrent edit, got %d", successes)
	}

	updated, err := itUsersRepo.GetMyProfile(context.Background(), user.ID)
	if err != nil {
		t.Fatalf("get profile: %v", err)
	}
	if updated.Version != user.Version+1 {
		t.Fatalf("expected version %d after single successful edit, got %d", user.Version+1, updated.Version)
	}
	if updated.Username != newName1 && updated.Username != newName2 {
		t.Fatalf("unexpected username after concurrent edit: %s", updated.Username)
	}
}

func TestEditProfile_DuplicateEmail_RawError(t *testing.T) {
	userA := createUser(t)
	userB := createUser(t)

	_, err := itUsersService.EditProfile(context.Background(), userA.ID, domain.NewUserPatch(
		domain.Nullable[string]{}, // username не меняем
		domain.Nullable[string]{Value: &userB.Email, Set: true},
		domain.Nullable[string]{},
		domain.Nullable[string]{},
	))
	if err == nil {
		t.Fatal("expected error on duplicate email edit, got nil")
	}
	if !isConflict(err) {
		t.Fatalf("raw DB error leaked from duplicate email edit: %v", err)
	}
}
