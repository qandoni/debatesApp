package integration_test

import (
	"context"
	"fmt"
	"mime/multipart"
	"testing"
	"time"

	auth_jwt "github.com/qandoni/debatesApp/internal/features/auth/jwt"

	core_password "github.com/qandoni/debatesApp/internal/core/password"
	core_password_hash "github.com/qandoni/debatesApp/internal/core/password/hash"

	"github.com/qandoni/debatesApp/internal/core/domain"
	auth_service "github.com/qandoni/debatesApp/internal/features/auth/service"
	comments_ratings_repository "github.com/qandoni/debatesApp/internal/features/comments/comment_ratings/repository/postgres"
	comment_ratings_service "github.com/qandoni/debatesApp/internal/features/comments/comment_ratings/service"
	comments_repository "github.com/qandoni/debatesApp/internal/features/comments/repository/postgres"
	comments_service "github.com/qandoni/debatesApp/internal/features/comments/service"
	posts_contracts "github.com/qandoni/debatesApp/internal/features/posts/contracts"
	debate_sides_repository "github.com/qandoni/debatesApp/internal/features/posts/debate_sides/repository/postgres"
	debate_votes_repository "github.com/qandoni/debatesApp/internal/features/posts/debate_votes/repository/postgres"
	debate_votes_service "github.com/qandoni/debatesApp/internal/features/posts/debate_votes/service"
	debates_repository "github.com/qandoni/debatesApp/internal/features/posts/debates/repository/postgres"
	posts_repository "github.com/qandoni/debatesApp/internal/features/posts/repository"
	posts_service "github.com/qandoni/debatesApp/internal/features/posts/service"
	users_repository "github.com/qandoni/debatesApp/internal/features/users/repository"
	users_service "github.com/qandoni/debatesApp/internal/features/users/service"

	"github.com/jackc/pgx/v5"
	core_errors "github.com/qandoni/debatesApp/internal/core/errors"
)

type pgxConn struct {
	conn *pgx.Conn
}

func newPgxConn(ctx context.Context, dsn string) (*pgxConn, error) {
	connCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	conn, err := pgx.Connect(connCtx, dsn)
	if err != nil {
		return nil, err
	}
	return &pgxConn{conn: conn}, nil
}

func (c *pgxConn) Exec(ctx context.Context, sql string, args ...any) (commandTag, error) {
	tag, err := c.conn.Exec(ctx, sql, args...)
	return commandTag{rowsAffected: tag.RowsAffected()}, err
}

func (c *pgxConn) Close(ctx context.Context) error {
	return c.conn.Close(ctx)
}

type commandTag struct {
	rowsAffected int64
}

var (
	itPostsRepo    *posts_repository.PostsRepository
	itDebatesRepo  *debates_repository.DebatesRepository
	itSidesRepo    *debate_sides_repository.DebateSidesRepository
	itVotesRepo    *debate_votes_repository.DebateVotesRepository
	itCommentsRepo *comments_repository.CommentsRepository
	itRatingsRepo  *comments_ratings_repository.CommentRatingsRepository
	itUsersRepo    *users_repository.UsersRepository

	itVotesService    *debate_votes_service.DebateVotesService
	itCommentsService *comments_service.CommentsService
	itRatingsService  *comment_ratings_service.CommentRatingsService
	itPostsService    *posts_service.PostsService
	itAuthService     *auth_service.AuthService
	itUsersService    *users_service.UsersService

	itJWTManager *auth_jwt.JWTManager
	itBcrypt     *core_password.BcryptHasher
	itTestSecret = "integration-test-secret"
)

func wireDependencies() {
	itPostsRepo = posts_repository.NewPostsRepository(itPool, itTimeout)
	itDebatesRepo = debates_repository.NewDebatesRepository(itPool, itTimeout)
	itSidesRepo = debate_sides_repository.NewDebateSidesRepository(itPool, itTimeout)
	itVotesRepo = debate_votes_repository.NewDebateVotesRepository(itPool, itTimeout)
	itCommentsRepo = comments_repository.NewCommentsRepository(itPool, itTimeout)
	itRatingsRepo = comments_ratings_repository.NewCommentRatingsRepository(itPool, itTimeout)
	itUsersRepo = users_repository.NewUsersRepository(itPool, itTimeout)

	itJWTManager = auth_jwt.NewJWTManager(itTestSecret)
	itBcrypt = core_password.NewBcryptHasher()

	itVotesService = debate_votes_service.NewDebateVotesService(
		itVotesRepo, itDebatesRepo, itSidesRepo, itCommentsRepo,
	)
	itCommentsService = comments_service.NewCommentsService(
		itCommentsRepo, itPostsRepo, itDebatesRepo, itSidesRepo, itVotesRepo, itTxManager,
	)
	itRatingsService = comment_ratings_service.NewCommentRatingsService(
		itRatingsRepo, itCommentsRepo, itDebatesRepo,
	)
	itPostsService = posts_service.NewPostsService(
		itPostsRepo, nopImagesService{}, itDebatesRepo, itSidesRepo, itTxManager,
	)
	itAuthService = auth_service.NewAuthService(
		itUsersRepo, itBcrypt, core_password_hash.NewSHA256Hasher(), itJWTManager, itTxManager,
	)
	itUsersService = users_service.NewUsersService(itUsersRepo, itBcrypt)
}

type nopImagesService struct{}

func (nopImagesService) GetByPostID(ctx context.Context, postID int) ([]domain.PostImage, error) {
	return nil, nil
}

func (nopImagesService) CreatePostImages(ctx context.Context, userID, postID int, files []*multipart.FileHeader) ([]domain.PostImage, error) {
	return nil, nil
}

func (nopImagesService) DeleteRecordsByPostID(ctx context.Context, postID int) error { return nil }

func (nopImagesService) GetByPostIDs(ctx context.Context, postIDs []int) (map[int][]domain.PostImage, error) {
	return map[int][]domain.PostImage{}, nil
}

func (nopImagesService) DeleteFromStorage(ctx context.Context, images []domain.PostImage) error {
	return nil
}

var fixtureSeq int64

func uniqueEmail() string {
	fixtureSeq++
	return fmt.Sprintf("user_%d_%d@test.it", time.Now().UnixNano(), fixtureSeq)
}

func createUser(t *testing.T) domain.User {
	t.Helper()

	hash, err := itBcrypt.Hash("P@ssw0rd!")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	user, err := itUsersRepo.CreateUser(context.Background(), domain.NewUserUninitialized(
		uniqueEmail(), uniqueEmail(), hash, nil, nil,
	))
	if err != nil {
		t.Fatalf("create user fixture: %v", err)
	}
	return user
}

func createRegularPost(t *testing.T, authorID int) domain.Post {
	t.Helper()

	post, err := itPostsService.CreatePost(context.Background(), posts_contracts.CreatePostInput{
		AuthorID: authorID,
		Content:  "regular post content",
		IsDebate: false,
	})
	if err != nil {
		t.Fatalf("create regular post fixture: %v", err)
	}
	return post
}

func createDebatePost(t *testing.T, authorID int, sidesCount int) domain.Post {
	t.Helper()

	sides := make([]posts_contracts.CreateDebateSideInput, 0, sidesCount)
	for i := 0; i < sidesCount; i++ {
		sides = append(sides, posts_contracts.CreateDebateSideInput{
			Name: fmt.Sprintf("side-%d", i+1),
		})
	}

	post, err := itPostsService.CreatePost(context.Background(), posts_contracts.CreatePostInput{
		AuthorID: authorID,
		Content:  "debate post content",
		IsDebate: true,
		Debate: &posts_contracts.CreateDebateInput{
			Sides: sides,
		},
	})
	if err != nil {
		t.Fatalf("create debate post fixture: %v", err)
	}
	if post.Debate == nil || len(post.Debate.Sides) != sidesCount {
		t.Fatalf("debate fixture is broken: sides=%v", post.Debate)
	}
	return post
}

func mustVote(t *testing.T, userID, debateID, sideID int) domain.DebateVote {
	t.Helper()

	vote, err := itVotesService.Vote(context.Background(), userID, debateID, sideID)
	if err != nil {
		t.Fatalf("vote fixture (user=%d debate=%d side=%d): %v", userID, debateID, sideID, err)
	}
	return vote
}

func createArgument(t *testing.T, userID, postID, sideID int) domain.Comment {
	t.Helper()

	comment, err := itCommentsService.CreateArgument(context.Background(), userID, postID, sideID, "argument content")
	if err != nil {
		t.Fatalf("create argument fixture (user=%d post=%d side=%d): %v", userID, postID, sideID, err)
	}
	return comment
}

func createReply(t *testing.T, userID int, parent domain.Comment) domain.Comment {
	t.Helper()

	parentID := parent.ID
	comment, err := itCommentsService.CreateComment(context.Background(), userID, parent.PostID, &parentID, "reply content")
	if err != nil {
		t.Fatalf("create reply fixture: %v", err)
	}
	return comment
}

func finishDebate(t *testing.T, authorID, debateID int) {
	t.Helper()

	if err := itVotesService.FinishDebate(context.Background(), authorID, debateID); err != nil {
		t.Fatalf("finish debate fixture: %v", err)
	}
}

type startGate struct {
	ch chan struct{}
}

func newStartGate(n int) *startGate {
	return &startGate{ch: make(chan struct{}, n)}
}

func (g *startGate) wait() { <-g.ch }

func (g *startGate) open() { close(g.ch) }

func isConflict(err error) bool {
	return err != nil && is(err, core_errors.ErrConflict)
}

func isNotFound(err error) bool {
	return err != nil && is(err, core_errors.ErrNotFound)
}

func isForbidden(err error) bool {
	return err != nil && is(err, core_errors.ErrAccessForbidden)
}

func isInvalidArgument(err error) bool {
	return err != nil && is(err, core_errors.ErrInvalidArgument)
}

func is(err, target error) bool {
	for e := err; e != nil; {
		if e == target {
			return true
		}
		type unwrapper interface{ Unwrap() error }
		u, ok := e.(unwrapper)
		if !ok {
			return false
		}
		e = u.Unwrap()
	}
	return false
}

func countVotes(t *testing.T, debateID int) int {
	t.Helper()
	var n int
	err := itPool.QueryRow(context.Background(),
		`SELECT COUNT(*) FROM debatesApp.debate_votes WHERE debate_id = $1`, debateID).Scan(&n)
	if err != nil {
		t.Fatalf("count votes: %v", err)
	}
	return n
}

func countCommentRatings(t *testing.T, commentID int) int {
	t.Helper()
	var n int
	err := itPool.QueryRow(context.Background(),
		`SELECT COUNT(*) FROM debatesApp.comment_ratings WHERE comment_id = $1`, commentID).Scan(&n)
	if err != nil {
		t.Fatalf("count comment ratings: %v", err)
	}
	return n
}

func countUsersByEmail(t *testing.T, email string) int {
	t.Helper()
	var n int
	err := itPool.QueryRow(context.Background(),
		`SELECT COUNT(*) FROM debatesApp.users WHERE email = $1`, email).Scan(&n)
	if err != nil {
		t.Fatalf("count users by email: %v", err)
	}
	return n
}

func getVoteRaw(t *testing.T, debateID, userID int) (sideID int, isChanged bool, version int) {
	t.Helper()
	err := itPool.QueryRow(context.Background(),
		`SELECT debate_side_id, is_changed, version FROM debatesApp.debate_votes WHERE debate_id = $1 AND user_id = $2`,
		debateID, userID).Scan(&sideID, &isChanged, &version)
	if err != nil {
		t.Fatalf("get vote raw: %v", err)
	}
	return sideID, isChanged, version
}

func getDebateRaw(t *testing.T, debateID int) (status string, winnerSideID *int) {
	t.Helper()
	err := itPool.QueryRow(context.Background(),
		`SELECT status, winner_side_id FROM debatesApp.debates WHERE id = $1`, debateID).
		Scan(&status, &winnerSideID)
	if err != nil {
		t.Fatalf("get debate raw: %v", err)
	}
	return status, winnerSideID
}

func postExists(t *testing.T, postID int) bool {
	t.Helper()
	var exists bool
	err := itPool.QueryRow(context.Background(),
		`SELECT EXISTS (SELECT 1 FROM debatesApp.posts WHERE id = $1)`, postID).Scan(&exists)
	if err != nil {
		t.Fatalf("post exists: %v", err)
	}
	return exists
}
