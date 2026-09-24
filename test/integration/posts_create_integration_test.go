package integration_test

import (
	"context"
	"errors"
	"testing"

	"github.com/qandoni/debatesApp/internal/core/domain"
	posts_contracts "github.com/qandoni/debatesApp/internal/features/posts/contracts"
)

func TestCreatePost_RegularPost_Persists(t *testing.T) {
	author := createUser(t)

	post := createRegularPost(t, author.ID)

	if !postExists(t, post.ID) {
		t.Fatalf("post %d not found in DB", post.ID)
	}
	if post.IsDebate {
		t.Fatal("expected is_debate=false")
	}
}

func TestCreatePost_DebatePost_PersistsAllParts(t *testing.T) {
	author := createUser(t)

	post := createDebatePost(t, author.ID, 3)

	if !postExists(t, post.ID) {
		t.Fatalf("post %d not found in DB", post.ID)
	}

	status, _ := getDebateRaw(t, post.Debate.ID)
	if status != "OPEN" {
		t.Fatalf("expected debate OPEN, got %s", status)
	}

	for i, side := range post.Debate.Sides {
		var (
			debateID     int
			name         string
			displayOrder int
		)
		err := itPool.QueryRow(context.Background(),
			`SELECT debate_id, name, display_order FROM debatesApp.debate_sides WHERE id = $1`,
			side.ID).Scan(&debateID, &name, &displayOrder)
		if err != nil {
			t.Fatalf("side %d not found in DB: %v", side.ID, err)
		}
		if debateID != post.Debate.ID {
			t.Fatalf("side %d belongs to debate %d, expected %d", side.ID, debateID, post.Debate.ID)
		}
		if displayOrder != i+1 {
			t.Fatalf("side %d: expected display_order %d, got %d", side.ID, i+1, displayOrder)
		}
	}
}

func TestTxManager_Rollback_OnFailure(t *testing.T) {
	author := createUser(t)

	sentinel := errors.New("boom")

	err := itTxManager.WithinTransaction(context.Background(), func(txCtx context.Context) error {
		_, err := itPostsRepo.CreatePost(txCtx, domain.NewPostUninitialized(
			author.ID, "must be rolled back", false,
		))
		if err != nil {
			t.Fatalf("create post inside tx: %v", err)
		}
		return sentinel
	})
	if !errors.Is(err, sentinel) {
		t.Fatalf("expected sentinel error, got: %v", err)
	}

	// Пост, созданный внутри откаченной транзакции, не должен существовать.
	var count int
	err = itPool.QueryRow(context.Background(),
		`SELECT COUNT(*) FROM debatesApp.posts WHERE author_id = $1 AND content = $2`,
		author.ID, "must be rolled back").Scan(&count)
	if err != nil {
		t.Fatalf("count posts: %v", err)
	}
	if count != 0 {
		t.Fatalf("ROLLBACK FAILED: %d post(s) survived the aborted transaction", count)
	}
}

func TestTxManager_Commit_OnSuccess(t *testing.T) {
	author := createUser(t)

	var postID int
	err := itTxManager.WithinTransaction(context.Background(), func(txCtx context.Context) error {
		post, err := itPostsRepo.CreatePost(txCtx, domain.NewPostUninitialized(
			author.ID, "committed post", false,
		))
		postID = post.ID
		return err
	})
	if err != nil {
		t.Fatalf("within transaction: %v", err)
	}
	if !postExists(t, postID) {
		t.Fatalf("post %d was not committed", postID)
	}
}

func TestCreatePost_UnknownAuthor_NoOrphans(t *testing.T) {
	const unknownAuthorID = 999_999_999

	_, err := itPostsService.CreatePost(context.Background(), posts_contracts.CreatePostInput{
		AuthorID: unknownAuthorID,
		Content:  "orphan attempt",
		IsDebate: true,
		Debate: &posts_contracts.CreateDebateInput{
			Sides: []posts_contracts.CreateDebateSideInput{
				{Name: "a"}, {Name: "b"},
			},
		},
	})
	if err == nil {
		t.Fatal("expected error for unknown author, got nil")
	}

	var count int
	if err := itPool.QueryRow(context.Background(),
		`SELECT COUNT(*) FROM debatesApp.posts WHERE author_id = $1`, unknownAuthorID).Scan(&count); err != nil {
		t.Fatalf("count posts: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected 0 posts for unknown author, got %d", count)
	}
}
