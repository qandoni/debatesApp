package integration_test

import (
	"context"
	"testing"

	core_errors "github.com/qandoni/debatesApp/internal/core/errors"
)

func TestCreateComment_RegularPost_Persists(t *testing.T) {
	author := createUser(t)
	commenter := createUser(t)
	post := createRegularPost(t, author.ID)

	comment, err := itCommentsService.CreateComment(context.Background(), commenter.ID, post.ID, nil, "hello")
	if err != nil {
		t.Fatalf("create comment: %v", err)
	}

	stored, err := itCommentsRepo.GetByID(context.Background(), comment.ID)
	if err != nil {
		t.Fatalf("get comment: %v", err)
	}
	if stored.Content != "hello" || stored.AuthorID != commenter.ID || stored.PostID != post.ID {
		t.Fatalf("unexpected stored comment: %+v", stored)
	}
	if stored.DebateSideID != nil {
		t.Fatalf("regular comment must have NULL debate_side_id, got %d", *stored.DebateSideID)
	}
}

func TestCreateComment_OnDebatePost_WithoutParent_Conflict(t *testing.T) {
	author := createUser(t)
	commenter := createUser(t)
	post := createDebatePost(t, author.ID, 2)

	_, err := itCommentsService.CreateComment(context.Background(), commenter.ID, post.ID, nil, "no parent")
	if err == nil {
		t.Fatal("expected error creating plain comment on debate post, got nil")
	}
	if !isConflict(err) {
		t.Fatalf("expected ErrConflict, got: %v", err)
	}
}

func TestCreateComment_Reply_InheritsDebateSide(t *testing.T) {
	author := createUser(t)
	voter := createUser(t)
	replier := createUser(t)
	post := createDebatePost(t, author.ID, 2)

	side1 := post.Debate.Sides[0].ID
	mustVote(t, voter.ID, post.Debate.ID, side1)
	arg := createArgument(t, voter.ID, post.ID, side1)

	reply := createReply(t, replier.ID, arg)

	stored, err := itCommentsRepo.GetByID(context.Background(), reply.ID)
	if err != nil {
		t.Fatalf("get reply: %v", err)
	}
	if stored.ParentCommentID == nil || *stored.ParentCommentID != arg.ID {
		t.Fatalf("expected parent_comment_id=%d, got %v", arg.ID, stored.ParentCommentID)
	}
	if stored.DebateSideID == nil || *stored.DebateSideID != side1 {
		t.Fatalf("reply must inherit debate_side_id=%d, got %v", side1, stored.DebateSideID)
	}
}

func TestCreateComment_ParentFromOtherPost_Conflict(t *testing.T) {
	author := createUser(t)
	voter := createUser(t)
	postA := createDebatePost(t, author.ID, 2)
	postB := createDebatePost(t, author.ID, 2)

	sideA := postA.Debate.Sides[0].ID
	mustVote(t, voter.ID, postA.Debate.ID, sideA)
	argA := createArgument(t, voter.ID, postA.ID, sideA)

	parentID := argA.ID
	_, err := itCommentsService.CreateComment(context.Background(), voter.ID, postB.ID, &parentID, "cross-post reply")
	if err == nil {
		t.Fatal("expected error for cross-post parent, got nil")
	}
	if !isConflict(err) {
		t.Fatalf("expected ErrConflict, got: %v", err)
	}
}

func TestCreateArgument_WithoutVote_Fails(t *testing.T) {
	author := createUser(t)
	outsider := createUser(t)
	post := createDebatePost(t, author.ID, 2)

	side1 := post.Debate.Sides[0].ID

	_, err := itCommentsService.CreateArgument(context.Background(), outsider.ID, post.ID, side1, "no vote")
	if err == nil {
		t.Fatal("expected error creating argument without vote, got nil")
	}
	if !isNotFound(err) {
		t.Fatalf("expected ErrNotFound (no vote), got: %v", err)
	}
}

func TestCreateArgument_WrongSide_Fails(t *testing.T) {
	author := createUser(t)
	voter := createUser(t)
	post := createDebatePost(t, author.ID, 2)

	side1 := post.Debate.Sides[0].ID
	side2 := post.Debate.Sides[1].ID

	mustVote(t, voter.ID, post.Debate.ID, side1)

	_, err := itCommentsService.CreateArgument(context.Background(), voter.ID, post.ID, side2, "wrong side")
	if err == nil {
		t.Fatal("expected error creating argument for other side, got nil")
	}
	if !isConflict(err) {
		t.Fatalf("expected ErrConflict, got: %v", err)
	}
}

func TestCreateArgument_Success_Persists(t *testing.T) {
	author := createUser(t)
	voter := createUser(t)
	post := createDebatePost(t, author.ID, 2)

	side1 := post.Debate.Sides[0].ID
	mustVote(t, voter.ID, post.Debate.ID, side1)

	arg := createArgument(t, voter.ID, post.ID, side1)

	stored, err := itCommentsRepo.GetByID(context.Background(), arg.ID)
	if err != nil {
		t.Fatalf("get argument: %v", err)
	}
	if stored.DebateSideID == nil || *stored.DebateSideID != side1 {
		t.Fatalf("expected debate_side_id=%d, got %v", side1, stored.DebateSideID)
	}
	if stored.ParentCommentID != nil {
		t.Fatalf("argument must be root, got parent %d", *stored.ParentCommentID)
	}
}

func TestUpdateComment_NotAuthor_Forbidden(t *testing.T) {
	author := createUser(t)
	commenter := createUser(t)
	outsider := createUser(t)
	post := createRegularPost(t, author.ID)

	comment, err := itCommentsService.CreateComment(context.Background(), commenter.ID, post.ID, nil, "original")
	if err != nil {
		t.Fatalf("create comment fixture: %v", err)
	}

	_, err = itCommentsService.UpdateComment(context.Background(), outsider.ID, comment.ID, "hacked")
	if err == nil {
		t.Fatal("expected error updating foreign comment, got nil")
	}
	if !isForbidden(err) {
		t.Fatalf("expected ErrAccessForbidden, got: %v", err)
	}

	stored, err := itCommentsRepo.GetByID(context.Background(), comment.ID)
	if err != nil {
		t.Fatalf("get comment: %v", err)
	}
	if stored.Content != "original" {
		t.Fatalf("comment content must be untouched, got %q", stored.Content)
	}
}

func TestUpdateComment_Author_Success(t *testing.T) {
	author := createUser(t)
	commenter := createUser(t)
	post := createRegularPost(t, author.ID)

	comment, err := itCommentsService.CreateComment(context.Background(), commenter.ID, post.ID, nil, "original")
	if err != nil {
		t.Fatalf("create comment fixture: %v", err)
	}

	updated, err := itCommentsService.UpdateComment(context.Background(), commenter.ID, comment.ID, "edited")
	if err != nil {
		t.Fatalf("update comment: %v", err)
	}
	if updated.Content != "edited" {
		t.Fatalf("expected edited content, got %q", updated.Content)
	}
	if updated.Version != comment.Version+1 {
		t.Fatalf("expected version %d, got %d", comment.Version+1, updated.Version)
	}
}

func TestUpdateComment_ArgumentAfterFinish_Conflict(t *testing.T) {
	author := createUser(t)
	voter := createUser(t)
	post := createDebatePost(t, author.ID, 2)

	side1 := post.Debate.Sides[0].ID
	mustVote(t, voter.ID, post.Debate.ID, side1)
	arg := createArgument(t, voter.ID, post.ID, side1)

	finishDebate(t, author.ID, post.Debate.ID)

	_, err := itCommentsService.UpdateComment(context.Background(), voter.ID, arg.ID, "late edit")
	if err == nil {
		t.Fatal("expected error editing argument after debate finished, got nil")
	}
	if !isConflict(err) {
		t.Fatalf("expected ErrConflict, got: %v", err)
	}
}

func TestSetAuthorLike_OnlyDebateAuthor(t *testing.T) {
	author := createUser(t)
	voter := createUser(t)
	outsider := createUser(t)
	post := createDebatePost(t, author.ID, 2)

	side1 := post.Debate.Sides[0].ID
	mustVote(t, voter.ID, post.Debate.ID, side1)
	arg := createArgument(t, voter.ID, post.ID, side1)

	_, err := itCommentsService.SetAuthorLike(context.Background(), outsider.ID, arg.ID, true)
	if err == nil {
		t.Fatal("expected error for non-author like, got nil")
	}
	if !isForbidden(err) {
		t.Fatalf("expected ErrAccessForbidden, got: %v", err)
	}

	updated, err := itCommentsService.SetAuthorLike(context.Background(), author.ID, arg.ID, true)
	if err != nil {
		t.Fatalf("author like: %v", err)
	}
	if !updated.AuthorLiked {
		t.Fatal("expected author_liked=true")
	}

	stored, err := itCommentsRepo.GetByID(context.Background(), arg.ID)
	if err != nil {
		t.Fatalf("get argument: %v", err)
	}
	if !stored.AuthorLiked {
		t.Fatal("author_liked not persisted in DB")
	}
}

func TestSetAuthorLike_ReplyComment_Forbidden(t *testing.T) {
	author := createUser(t)
	voter := createUser(t)
	replier := createUser(t)
	post := createDebatePost(t, author.ID, 2)

	side1 := post.Debate.Sides[0].ID
	mustVote(t, voter.ID, post.Debate.ID, side1)
	arg := createArgument(t, voter.ID, post.ID, side1)
	reply := createReply(t, replier.ID, arg)

	_, err := itCommentsService.SetAuthorLike(context.Background(), author.ID, reply.ID, true)
	if err == nil {
		t.Fatal("expected error liking a reply, got nil")
	}
	if !isInvalidArgument(err) {
		t.Fatalf("expected ErrInvalidArgument, got: %v", err)
	}
}

func TestSetAuthorLike_AfterFinish_Conflict(t *testing.T) {
	author := createUser(t)
	voter := createUser(t)
	post := createDebatePost(t, author.ID, 2)

	side1 := post.Debate.Sides[0].ID
	mustVote(t, voter.ID, post.Debate.ID, side1)
	arg := createArgument(t, voter.ID, post.ID, side1)

	finishDebate(t, author.ID, post.Debate.ID)

	_, err := itCommentsService.SetAuthorLike(context.Background(), author.ID, arg.ID, true)
	if err == nil {
		t.Fatal("expected error liking after debate finished, got nil")
	}
	if !is(err, core_errors.ErrConflict) {
		t.Fatalf("expected ErrConflict, got: %v", err)
	}
}
