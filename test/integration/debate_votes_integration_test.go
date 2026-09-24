package integration_test

import (
	"context"
	"sync"
	"testing"
)

func TestVote_PersistsVote(t *testing.T) {
	author := createUser(t)
	voter := createUser(t)
	post := createDebatePost(t, author.ID, 2)

	side1 := post.Debate.Sides[0].ID

	mustVote(t, voter.ID, post.Debate.ID, side1)

	gotSide, isChanged, version := getVoteRaw(t, post.Debate.ID, voter.ID)
	if gotSide != side1 {
		t.Fatalf("expected side %d in DB, got %d", side1, gotSide)
	}
	if isChanged {
		t.Fatal("expected is_changed=false for fresh vote")
	}
	if version != 1 {
		t.Fatalf("expected version=1, got %d", version)
	}
	if n := countVotes(t, post.Debate.ID); n != 1 {
		t.Fatalf("expected 1 vote in DB, got %d", n)
	}
}

func TestVote_DoubleVote_Conflict(t *testing.T) {
	author := createUser(t)
	voter := createUser(t)
	post := createDebatePost(t, author.ID, 2)

	side1 := post.Debate.Sides[0].ID
	side2 := post.Debate.Sides[1].ID

	mustVote(t, voter.ID, post.Debate.ID, side1)

	_, err := itVotesService.Vote(context.Background(), voter.ID, post.Debate.ID, side2)
	if err == nil {
		t.Fatal("expected error on double vote, got nil")
	}
	if !isConflict(err) {
		t.Fatalf("expected ErrConflict on double vote, got: %v", err)
	}
	if n := countVotes(t, post.Debate.ID); n != 1 {
		t.Fatalf("expected still 1 vote in DB, got %d", n)
	}
}

func TestVote_SideFromOtherDebate_NotFound(t *testing.T) {
	author := createUser(t)
	voter := createUser(t)
	postA := createDebatePost(t, author.ID, 2)
	postB := createDebatePost(t, author.ID, 2)

	_, err := itVotesService.Vote(context.Background(), voter.ID, postA.Debate.ID, postB.Debate.Sides[0].ID)
	if err == nil {
		t.Fatal("expected error for foreign side, got nil")
	}
	if !isNotFound(err) {
		t.Fatalf("expected ErrNotFound, got: %v", err)
	}
	if n := countVotes(t, postA.Debate.ID); n != 0 {
		t.Fatalf("expected 0 votes in DB, got %d", n)
	}
}

func TestVote_OnFinishedDebate_Fails(t *testing.T) {
	author := createUser(t)
	voter := createUser(t)
	post := createDebatePost(t, author.ID, 2)

	finishDebate(t, author.ID, post.Debate.ID)

	_, err := itVotesService.Vote(context.Background(), voter.ID, post.Debate.ID, post.Debate.Sides[0].ID)
	if err == nil {
		t.Fatal("expected error voting on finished debate, got nil")
	}
	if n := countVotes(t, post.Debate.ID); n != 0 {
		t.Fatalf("expected 0 votes in finished debate, got %d", n)
	}
}

func TestVote_AfterFinishCommit_ConcurrentMustFail(t *testing.T) {
	author := createUser(t)
	post := createDebatePost(t, author.ID, 2)

	finishDebate(t, author.ID, post.Debate.ID)

	const voters = 8
	gate := newStartGate(voters)
	var wg sync.WaitGroup
	results := make(chan error, voters)

	for i := 0; i < voters; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			u := createUser(t)
			gate.wait()
			_, err := itVotesService.Vote(context.Background(), u.ID, post.Debate.ID, post.Debate.Sides[0].ID)
			results <- err
		}()
	}
	gate.open()
	wg.Wait()
	close(results)

	for err := range results {
		if err == nil {
			t.Fatal("vote succeeded on already finished debate: race window allows votes after finish commit")
		}
	}
	if n := countVotes(t, post.Debate.ID); n != 0 {
		t.Fatalf("expected 0 votes in finished debate, got %d", n)
	}
}

func TestChangeVote_SwitchesSide(t *testing.T) {
	author := createUser(t)
	voter := createUser(t)
	post := createDebatePost(t, author.ID, 2)

	side1 := post.Debate.Sides[0].ID
	side2 := post.Debate.Sides[1].ID

	mustVote(t, voter.ID, post.Debate.ID, side1)

	_, err := itVotesService.ChangeVote(context.Background(), voter.ID, post.Debate.ID, side2)
	if err != nil {
		t.Fatalf("change vote: %v", err)
	}

	gotSide, isChanged, version := getVoteRaw(t, post.Debate.ID, voter.ID)
	if gotSide != side2 {
		t.Fatalf("expected side %d after change, got %d", side2, gotSide)
	}
	if !isChanged {
		t.Fatal("expected is_changed=true after change")
	}
	if version != 2 {
		t.Fatalf("expected version=2 after change, got %d", version)
	}
}

func TestChangeVote_SecondChange_Conflict(t *testing.T) {
	author := createUser(t)
	voter := createUser(t)
	post := createDebatePost(t, author.ID, 3)

	side1 := post.Debate.Sides[0].ID
	side2 := post.Debate.Sides[1].ID
	side3 := post.Debate.Sides[2].ID

	mustVote(t, voter.ID, post.Debate.ID, side1)

	if _, err := itVotesService.ChangeVote(context.Background(), voter.ID, post.Debate.ID, side2); err != nil {
		t.Fatalf("first change vote: %v", err)
	}

	_, err := itVotesService.ChangeVote(context.Background(), voter.ID, post.Debate.ID, side3)
	if err == nil {
		t.Fatal("expected error on second change, got nil")
	}
	if !isConflict(err) {
		t.Fatalf("expected ErrConflict on second change, got: %v", err)
	}

	gotSide, _, _ := getVoteRaw(t, post.Debate.ID, voter.ID)
	if gotSide != side2 {
		t.Fatalf("expected side unchanged (%d), got %d", side2, gotSide)
	}
}

func TestChangeVote_SameSide_Conflict(t *testing.T) {
	author := createUser(t)
	voter := createUser(t)
	post := createDebatePost(t, author.ID, 2)

	side1 := post.Debate.Sides[0].ID

	mustVote(t, voter.ID, post.Debate.ID, side1)

	_, err := itVotesService.ChangeVote(context.Background(), voter.ID, post.Debate.ID, side1)
	if err == nil {
		t.Fatal("expected error changing to same side, got nil")
	}
	if !isConflict(err) {
		t.Fatalf("expected ErrConflict, got: %v", err)
	}
}

func TestChangeVote_AfterArgument_Forbidden(t *testing.T) {
	author := createUser(t)
	voter := createUser(t)
	post := createDebatePost(t, author.ID, 2)

	side1 := post.Debate.Sides[0].ID
	side2 := post.Debate.Sides[1].ID

	mustVote(t, voter.ID, post.Debate.ID, side1)
	createArgument(t, voter.ID, post.ID, side1)

	_, err := itVotesService.ChangeVote(context.Background(), voter.ID, post.Debate.ID, side2)
	if err == nil {
		t.Fatal("expected error changing vote after argument, got nil")
	}
	if !isForbidden(err) {
		t.Fatalf("expected ErrAccessForbidden, got: %v", err)
	}

	gotSide, isChanged, _ := getVoteRaw(t, post.Debate.ID, voter.ID)
	if gotSide != side1 || isChanged {
		t.Fatalf("vote must be untouched after forbidden change: side=%d changed=%v", gotSide, isChanged)
	}
}

func TestChangeVote_RaceWithCreateArgument_Invariant(t *testing.T) {
	for iteration := 0; iteration < 10; iteration++ {
		author := createUser(t)
		voter := createUser(t)
		post := createDebatePost(t, author.ID, 2)

		side1 := post.Debate.Sides[0].ID
		side2 := post.Debate.Sides[1].ID

		mustVote(t, voter.ID, post.Debate.ID, side1)

		gate := newStartGate(2)
		var wg sync.WaitGroup
		var changeErr, argErr error

		wg.Add(2)
		go func() {
			defer wg.Done()
			gate.wait()
			_, changeErr = itVotesService.ChangeVote(context.Background(), voter.ID, post.Debate.ID, side2)
		}()
		go func() {
			defer wg.Done()
			gate.wait()
			_, argErr = itCommentsService.CreateArgument(context.Background(), voter.ID, post.ID, side1, "racy argument")
		}()

		gate.open()
		wg.Wait()

		_, isChanged, _ := getVoteRaw(t, post.Debate.ID, voter.ID)
		hasArgument := argErr == nil

		if hasArgument && isChanged {
			t.Fatalf(
				"RACE BUG (iteration %d): argument created AND vote changed concurrently. changeErr=%v argErr=%v",
				iteration, changeErr, argErr,
			)
		}
	}
}

func TestFinishDebate_WinnerSide_Majority(t *testing.T) {
	author := createUser(t)
	post := createDebatePost(t, author.ID, 2)

	side1 := post.Debate.Sides[0].ID
	side2 := post.Debate.Sides[1].ID

	for i := 0; i < 3; i++ {
		mustVote(t, createUser(t).ID, post.Debate.ID, side1)
	}
	mustVote(t, createUser(t).ID, post.Debate.ID, side2)

	finishDebate(t, author.ID, post.Debate.ID)

	status, winner := getDebateRaw(t, post.Debate.ID)
	if status != "FINISHED" {
		t.Fatalf("expected FINISHED, got %s", status)
	}
	if winner == nil || *winner != side1 {
		t.Fatalf("expected winner side %d, got %v", side1, winner)
	}
}

func TestFinishDebate_WinnerSide_Tie_Null(t *testing.T) {
	author := createUser(t)
	post := createDebatePost(t, author.ID, 2)

	side1 := post.Debate.Sides[0].ID
	side2 := post.Debate.Sides[1].ID

	mustVote(t, createUser(t).ID, post.Debate.ID, side1)
	mustVote(t, createUser(t).ID, post.Debate.ID, side2)

	finishDebate(t, author.ID, post.Debate.ID)

	status, winner := getDebateRaw(t, post.Debate.ID)
	if status != "FINISHED" {
		t.Fatalf("expected FINISHED, got %s", status)
	}
	if winner != nil {
		t.Fatalf("expected NULL winner on tie, got %d", *winner)
	}
}

func TestFinishDebate_Concurrent_DoubleFinish(t *testing.T) {
	author := createUser(t)
	post := createDebatePost(t, author.ID, 2)

	const finishers = 4
	gate := newStartGate(finishers)
	var wg sync.WaitGroup
	results := make(chan error, finishers)

	for i := 0; i < finishers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			gate.wait()
			results <- itVotesService.FinishDebate(context.Background(), author.ID, post.Debate.ID)
		}()
	}
	gate.open()
	wg.Wait()
	close(results)

	successes := 0
	for err := range results {
		if err == nil {
			successes++
		} else if !isConflict(err) {
			t.Fatalf("unexpected error type on concurrent finish: %v", err)
		}
	}
	if successes != 1 {
		t.Fatalf("expected exactly 1 successful finish, got %d", successes)
	}

	status, _ := getDebateRaw(t, post.Debate.ID)
	if status != "FINISHED" {
		t.Fatalf("expected FINISHED, got %s", status)
	}
}

func TestFinishDebate_NotAuthor_Forbidden(t *testing.T) {
	author := createUser(t)
	outsider := createUser(t)
	post := createDebatePost(t, author.ID, 2)

	err := itVotesService.FinishDebate(context.Background(), outsider.ID, post.Debate.ID)
	if err == nil {
		t.Fatal("expected error for non-author finish, got nil")
	}
	if !isForbidden(err) {
		t.Fatalf("expected ErrAccessForbidden, got: %v", err)
	}

	status, _ := getDebateRaw(t, post.Debate.ID)
	if status != "OPEN" {
		t.Fatalf("debate must stay OPEN, got %s", status)
	}
}

func TestFinishDebate_UnknownDebate_NotFound(t *testing.T) {
	author := createUser(t)

	err := itVotesService.FinishDebate(context.Background(), author.ID, 999_999_999)
	if err == nil {
		t.Fatal("expected error for unknown debate, got nil")
	}
	if !isNotFound(err) {
		t.Fatalf("expected ErrNotFound, got: %v", err)
	}
}
