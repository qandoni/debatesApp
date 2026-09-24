package integration_test

import (
	"context"
	"sync"
	"testing"

	core_errors "github.com/qandoni/debatesApp/internal/core/errors"
)

func TestRate_CreatesRating(t *testing.T) {
	author := createUser(t)
	rater := createUser(t)
	post := createDebatePost(t, author.ID, 2)

	side1 := post.Debate.Sides[0].ID
	mustVote(t, rater.ID, post.Debate.ID, side1)
	arg := createArgument(t, rater.ID, post.ID, side1)

	rating, err := itRatingsService.Rate(context.Background(), author.ID, arg.ID, 5)
	if err != nil {
		t.Fatalf("rate: %v", err)
	}
	if rating.Rating != 5 {
		t.Fatalf("expected score 5, got %d", rating.Rating)
	}
	if n := countCommentRatings(t, arg.ID); n != 1 {
		t.Fatalf("expected 1 rating row in DB, got %d", n)
	}
}

func TestRate_UpdatesExisting(t *testing.T) {
	author := createUser(t)
	rater := createUser(t)
	post := createDebatePost(t, author.ID, 2)

	side1 := post.Debate.Sides[0].ID
	mustVote(t, rater.ID, post.Debate.ID, side1)
	arg := createArgument(t, rater.ID, post.ID, side1)

	if _, err := itRatingsService.Rate(context.Background(), author.ID, arg.ID, 3); err != nil {
		t.Fatalf("rate 3: %v", err)
	}
	updated, err := itRatingsService.Rate(context.Background(), author.ID, arg.ID, 5)
	if err != nil {
		t.Fatalf("rate 5: %v", err)
	}
	if updated.Rating != 5 {
		t.Fatalf("expected updated score 5, got %d", updated.Rating)
	}
	if updated.Version != 2 {
		t.Fatalf("expected version=2 after update, got %d", updated.Version)
	}
	if n := countCommentRatings(t, arg.ID); n != 1 {
		t.Fatalf("expected still 1 rating row, got %d", n)
	}
}

func TestRate_SameScore_Idempotent(t *testing.T) {
	author := createUser(t)
	rater := createUser(t)
	post := createDebatePost(t, author.ID, 2)

	side1 := post.Debate.Sides[0].ID
	mustVote(t, rater.ID, post.Debate.ID, side1)
	arg := createArgument(t, rater.ID, post.ID, side1)

	first, err := itRatingsService.Rate(context.Background(), author.ID, arg.ID, 4)
	if err != nil {
		t.Fatalf("first rate: %v", err)
	}
	second, err := itRatingsService.Rate(context.Background(), author.ID, arg.ID, 4)
	if err != nil {
		t.Fatalf("second rate: %v", err)
	}
	if second.ID != first.ID || second.Version != first.Version {
		t.Fatalf("same-score rate must be idempotent: first=%+v second=%+v", first, second)
	}
}

func TestRate_InvalidScore(t *testing.T) {
	author := createUser(t)
	rater := createUser(t)
	post := createDebatePost(t, author.ID, 2)

	side1 := post.Debate.Sides[0].ID
	mustVote(t, rater.ID, post.Debate.ID, side1)
	arg := createArgument(t, rater.ID, post.ID, side1)

	for _, score := range []int{0, 6, -1} {
		_, err := itRatingsService.Rate(context.Background(), author.ID, arg.ID, score)
		if err == nil {
			t.Fatalf("expected error for score %d, got nil", score)
		}
		if !isInvalidArgument(err) {
			t.Fatalf("expected ErrInvalidArgument for score %d, got: %v", score, err)
		}
	}
	if n := countCommentRatings(t, arg.ID); n != 0 {
		t.Fatalf("expected 0 ratings, got %d", n)
	}
}

func TestRate_ReplyComment_Forbidden(t *testing.T) {
	author := createUser(t)
	rater := createUser(t)
	replier := createUser(t)
	post := createDebatePost(t, author.ID, 2)

	side1 := post.Debate.Sides[0].ID
	mustVote(t, rater.ID, post.Debate.ID, side1)
	arg := createArgument(t, rater.ID, post.ID, side1)
	reply := createReply(t, replier.ID, arg)

	_, err := itRatingsService.Rate(context.Background(), author.ID, reply.ID, 5)
	if err == nil {
		t.Fatal("expected error rating a reply, got nil")
	}
	if !isInvalidArgument(err) {
		t.Fatalf("expected ErrInvalidArgument, got: %v", err)
	}
}

func TestRate_NonArgumentComment_Forbidden(t *testing.T) {
	author := createUser(t)
	commenter := createUser(t)
	post := createRegularPost(t, author.ID)

	comment, err := itCommentsService.CreateComment(context.Background(), commenter.ID, post.ID, nil, "just a comment")
	if err != nil {
		t.Fatalf("create comment fixture: %v", err)
	}

	_, err = itRatingsService.Rate(context.Background(), author.ID, comment.ID, 5)
	if err == nil {
		t.Fatal("expected error rating non-argument comment, got nil")
	}
	if !isInvalidArgument(err) {
		t.Fatalf("expected ErrInvalidArgument, got: %v", err)
	}
}

func TestRate_AfterDebateFinished_Forbidden(t *testing.T) {
	author := createUser(t)
	rater := createUser(t)
	post := createDebatePost(t, author.ID, 2)

	side1 := post.Debate.Sides[0].ID
	mustVote(t, rater.ID, post.Debate.ID, side1)
	arg := createArgument(t, rater.ID, post.ID, side1)

	finishDebate(t, author.ID, post.Debate.ID)

	_, err := itRatingsService.Rate(context.Background(), author.ID, arg.ID, 5)
	if err == nil {
		t.Fatal("expected error rating after debate finished, got nil")
	}
	if !is(err, core_errors.ErrAccessForbidden) {
		t.Fatalf("expected ErrAccessForbidden, got: %v", err)
	}
}

func TestRate_ConcurrentFirstRate_NoRawDBError(t *testing.T) {
	author := createUser(t)
	rater := createUser(t)
	post := createDebatePost(t, author.ID, 2)

	side1 := post.Debate.Sides[0].ID
	mustVote(t, rater.ID, post.Debate.ID, side1)
	arg := createArgument(t, rater.ID, post.ID, side1)

	const raters = 12
	gate := newStartGate(raters)
	var wg sync.WaitGroup
	results := make(chan error, raters)

	for i := 0; i < raters; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			gate.wait()
			_, err := itRatingsService.Rate(context.Background(), author.ID, arg.ID, 5)
			results <- err
		}()
	}
	gate.open()
	wg.Wait()
	close(results)

	for err := range results {
		if err == nil {
			continue
		}
		if !isConflict(err) {
			t.Fatalf("raw DB error leaked from concurrent Rate (unique violation not mapped to ErrConflict): %v", err)
		}
	}
	if n := countCommentRatings(t, arg.ID); n != 1 {
		t.Fatalf("expected exactly 1 rating row, got %d", n)
	}
}
