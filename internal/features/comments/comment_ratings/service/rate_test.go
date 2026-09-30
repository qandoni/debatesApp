package comment_ratings_service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/qandoni/debatesApp/internal/core/domain"
	core_errors "github.com/qandoni/debatesApp/internal/core/errors"
	core_realtime "github.com/qandoni/debatesApp/internal/core/realtime"
)

type mockCommentRatingsRepository struct {
	createFn func(ctx context.Context, rating domain.CommentRating) (domain.CommentRating, error)
	getFn    func(ctx context.Context, commentID, userID int) (domain.CommentRating, error)
	updateFn func(ctx context.Context, rating domain.CommentRating) (domain.CommentRating, error)
}

func (m *mockCommentRatingsRepository) CreateCommentRating(ctx context.Context, rating domain.CommentRating) (domain.CommentRating, error) {
	return m.createFn(ctx, rating)
}

func (m *mockCommentRatingsRepository) GetByCommentAndUser(ctx context.Context, commentID, userID int) (domain.CommentRating, error) {
	return m.getFn(ctx, commentID, userID)
}

func (m *mockCommentRatingsRepository) UpdateCommentRating(ctx context.Context, rating domain.CommentRating) (domain.CommentRating, error) {
	return m.updateFn(ctx, rating)
}

type mockCommentsRepository struct {
	getByIDFn func(ctx context.Context, commentID int) (domain.Comment, error)
}

func (m *mockCommentsRepository) GetByID(ctx context.Context, commentID int) (domain.Comment, error) {
	return m.getByIDFn(ctx, commentID)
}

type mockDebatesRepository struct {
	getByPostFn func(ctx context.Context, postID int) (domain.Debate, error)
}

func (m *mockDebatesRepository) GetByPostID(ctx context.Context, postID int) (domain.Debate, error) {
	return m.getByPostFn(ctx, postID)
}

type publisherMock struct {
	topics []string
	events []core_realtime.Event
}

func (m *publisherMock) Publish(topic string, event core_realtime.Event) {
	m.topics = append(m.topics, topic)
	m.events = append(m.events, event)
}

func newRatingsService(
	ratings CommentRatingsRepository,
	comments CommentsRepository,
	debates DebatesRepository,
) *CommentRatingsService {
	return newRatingsServiceWithPublisher(ratings, comments, debates, &publisherMock{})
}

func newRatingsServiceWithPublisher(
	ratings CommentRatingsRepository,
	comments CommentsRepository,
	debates DebatesRepository,
	publisher core_realtime.Publisher,
) *CommentRatingsService {
	return NewCommentRatingsService(ratings, comments, debates, publisher)
}

func openDebate() domain.Debate {
	return domain.Debate{ID: 10, PostID: 100, Status: "OPEN"}
}

func finishedDebate() domain.Debate {
	return domain.Debate{ID: 10, PostID: 100, Status: "FINISHED"}
}

func rootArgument() domain.Comment {
	sideID := 2
	return domain.Comment{ID: 1, PostID: 100, DebateSideID: &sideID}
}

func TestRate_InvalidScore(t *testing.T) {
	svc := newRatingsService(&mockCommentRatingsRepository{}, &mockCommentsRepository{}, &mockDebatesRepository{})

	for _, score := range []int{0, 6, -1, 100} {
		_, err := svc.Rate(context.Background(), 5, 1, score)
		if err == nil {
			t.Fatalf("expected error for score %d, got nil", score)
		}
		if !errors.Is(err, core_errors.ErrInvalidArgument) {
			t.Fatalf("expected ErrInvalidArgument for score %d, got: %v", score, err)
		}
	}
}

func TestRate_NotRootComment(t *testing.T) {
	parentID := 3
	commentsRepo := &mockCommentsRepository{
		getByIDFn: func(ctx context.Context, commentID int) (domain.Comment, error) {
			return domain.Comment{ID: 1, PostID: 100, ParentCommentID: &parentID}, nil
		},
	}

	svc := newRatingsService(&mockCommentRatingsRepository{}, commentsRepo, &mockDebatesRepository{})
	_, err := svc.Rate(context.Background(), 5, 1, 4)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, core_errors.ErrInvalidArgument) {
		t.Fatalf("expected ErrInvalidArgument, got: %v", err)
	}
}

func TestRate_NotArgument(t *testing.T) {
	commentsRepo := &mockCommentsRepository{
		getByIDFn: func(ctx context.Context, commentID int) (domain.Comment, error) {
			return domain.Comment{ID: 1, PostID: 100}, nil
		},
	}

	svc := newRatingsService(&mockCommentRatingsRepository{}, commentsRepo, &mockDebatesRepository{})
	_, err := svc.Rate(context.Background(), 5, 1, 4)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, core_errors.ErrInvalidArgument) {
		t.Fatalf("expected ErrInvalidArgument, got: %v", err)
	}
}

func TestRate_DebateFinished(t *testing.T) {
	commentsRepo := &mockCommentsRepository{
		getByIDFn: func(ctx context.Context, commentID int) (domain.Comment, error) {
			return rootArgument(), nil
		},
	}
	debatesRepo := &mockDebatesRepository{
		getByPostFn: func(ctx context.Context, postID int) (domain.Debate, error) {
			return finishedDebate(), nil
		},
	}

	svc := newRatingsService(&mockCommentRatingsRepository{}, commentsRepo, debatesRepo)
	_, err := svc.Rate(context.Background(), 5, 1, 4)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, core_errors.ErrAccessForbidden) {
		t.Fatalf("expected ErrAccessForbidden, got: %v", err)
	}
}

func TestRate_FirstRatingCreates(t *testing.T) {
	commentsRepo := &mockCommentsRepository{
		getByIDFn: func(ctx context.Context, commentID int) (domain.Comment, error) {
			return rootArgument(), nil
		},
	}
	debatesRepo := &mockDebatesRepository{
		getByPostFn: func(ctx context.Context, postID int) (domain.Debate, error) {
			return openDebate(), nil
		},
	}
	ratingsRepo := &mockCommentRatingsRepository{
		getFn: func(ctx context.Context, commentID, userID int) (domain.CommentRating, error) {
			return domain.CommentRating{}, core_errors.ErrNotFound
		},
		createFn: func(ctx context.Context, rating domain.CommentRating) (domain.CommentRating, error) {
			if rating.Rating != 4 || rating.CommentID != 1 || rating.UserID != 5 {
				t.Fatalf("unexpected rating: %+v", rating)
			}
			rating.ID = 1
			return rating, nil
		},
	}

	svc := newRatingsService(ratingsRepo, commentsRepo, debatesRepo)
	created, err := svc.Rate(context.Background(), 5, 1, 4)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if created.ID != 1 {
		t.Fatalf("expected ID 1, got %d", created.ID)
	}
}

func TestRate_ChangeRatingUpdates(t *testing.T) {
	commentsRepo := &mockCommentsRepository{
		getByIDFn: func(ctx context.Context, commentID int) (domain.Comment, error) {
			return rootArgument(), nil
		},
	}
	debatesRepo := &mockDebatesRepository{
		getByPostFn: func(ctx context.Context, postID int) (domain.Debate, error) {
			return openDebate(), nil
		},
	}
	ratingsRepo := &mockCommentRatingsRepository{
		getFn: func(ctx context.Context, commentID, userID int) (domain.CommentRating, error) {
			return domain.CommentRating{ID: 1, CommentID: 1, UserID: 5, Rating: 2}, nil
		},
		updateFn: func(ctx context.Context, rating domain.CommentRating) (domain.CommentRating, error) {
			if rating.Rating != 5 {
				t.Fatalf("expected rating 5, got %d", rating.Rating)
			}
			if rating.UpdatedAt == nil {
				t.Fatal("expected UpdatedAt to be set")
			}
			return rating, nil
		},
	}

	svc := newRatingsService(ratingsRepo, commentsRepo, debatesRepo)
	updated, err := svc.Rate(context.Background(), 5, 1, 5)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if updated.Rating != 5 {
		t.Fatalf("expected rating 5, got %d", updated.Rating)
	}
}

func TestRate_SameRatingReturnsExisting(t *testing.T) {
	commentsRepo := &mockCommentsRepository{
		getByIDFn: func(ctx context.Context, commentID int) (domain.Comment, error) {
			return rootArgument(), nil
		},
	}
	debatesRepo := &mockDebatesRepository{
		getByPostFn: func(ctx context.Context, postID int) (domain.Debate, error) {
			return openDebate(), nil
		},
	}
	ratingsRepo := &mockCommentRatingsRepository{
		getFn: func(ctx context.Context, commentID, userID int) (domain.CommentRating, error) {
			return domain.CommentRating{ID: 1, CommentID: 1, UserID: 5, Rating: 3}, nil
		},
		updateFn: func(ctx context.Context, rating domain.CommentRating) (domain.CommentRating, error) {
			t.Fatal("Update should not be called for same rating")
			return domain.CommentRating{}, nil
		},
	}

	svc := newRatingsService(ratingsRepo, commentsRepo, debatesRepo)
	existing, err := svc.Rate(context.Background(), 5, 1, 3)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if existing.Rating != 3 {
		t.Fatalf("expected rating 3, got %d", existing.Rating)
	}
}

// assertSingleEvent проверяет, что сервис отправил ровно одно событие в нужный топик.
func assertSingleEvent(
	t *testing.T,
	publisher *publisherMock,
	topic string,
	eventType string,
) core_realtime.Event {
	t.Helper()

	if len(publisher.topics) != 1 || len(publisher.events) != 1 {
		t.Fatalf(
			"expected exactly one event, got topics=%v events=%v",
			publisher.topics,
			publisher.events,
		)
	}
	if publisher.topics[0] != topic {
		t.Fatalf("expected topic %q, got %q", topic, publisher.topics[0])
	}
	if publisher.events[0].Type != eventType {
		t.Fatalf(
			"expected event type %q, got %q",
			eventType,
			publisher.events[0].Type,
		)
	}

	return publisher.events[0]
}

func TestRate_NewRatingEmitsRatingUpdated(t *testing.T) {
	commentsRepo := &mockCommentsRepository{
		getByIDFn: func(ctx context.Context, commentID int) (domain.Comment, error) {
			return rootArgument(), nil
		},
	}
	debatesRepo := &mockDebatesRepository{
		getByPostFn: func(ctx context.Context, postID int) (domain.Debate, error) {
			return openDebate(), nil
		},
	}
	ratingsRepo := &mockCommentRatingsRepository{
		getFn: func(ctx context.Context, commentID, userID int) (domain.CommentRating, error) {
			return domain.CommentRating{}, core_errors.ErrNotFound
		},
		createFn: func(ctx context.Context, rating domain.CommentRating) (domain.CommentRating, error) {
			rating.ID = 1
			return rating, nil
		},
	}
	publisher := &publisherMock{}

	svc := newRatingsServiceWithPublisher(ratingsRepo, commentsRepo, debatesRepo, publisher)

	_, err := svc.Rate(context.Background(), 5, 1, 4)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	event := assertSingleEvent(
		t,
		publisher,
		core_realtime.PostTopic(100),
		core_realtime.EventArgumentRatingUpdated,
	)

	data, ok := event.Data.(core_realtime.RatingData)
	if !ok {
		t.Fatalf("expected RatingData, got %T", event.Data)
	}
	if data.CommentID != 1 || data.PostID != 100 || data.UserID != 5 || data.Score != 4 {
		t.Fatalf("unexpected payload: %+v", data)
	}
	if !data.IsNew {
		t.Fatalf("expected is_new=true for created rating, got: %+v", data)
	}
}

func TestRate_ChangedRatingEmitsRatingUpdated(t *testing.T) {
	commentsRepo := &mockCommentsRepository{
		getByIDFn: func(ctx context.Context, commentID int) (domain.Comment, error) {
			return rootArgument(), nil
		},
	}
	debatesRepo := &mockDebatesRepository{
		getByPostFn: func(ctx context.Context, postID int) (domain.Debate, error) {
			return openDebate(), nil
		},
	}
	ratingsRepo := &mockCommentRatingsRepository{
		getFn: func(ctx context.Context, commentID, userID int) (domain.CommentRating, error) {
			return domain.CommentRating{ID: 1, CommentID: 1, UserID: 5, Rating: 2}, nil
		},
		updateFn: func(ctx context.Context, rating domain.CommentRating) (domain.CommentRating, error) {
			return rating, nil
		},
	}
	publisher := &publisherMock{}

	svc := newRatingsServiceWithPublisher(ratingsRepo, commentsRepo, debatesRepo, publisher)

	_, err := svc.Rate(context.Background(), 5, 1, 5)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	event := assertSingleEvent(
		t,
		publisher,
		core_realtime.PostTopic(100),
		core_realtime.EventArgumentRatingUpdated,
	)

	data, ok := event.Data.(core_realtime.RatingData)
	if !ok {
		t.Fatalf("expected RatingData, got %T", event.Data)
	}
	if data.CommentID != 1 || data.PostID != 100 || data.UserID != 5 || data.Score != 5 {
		t.Fatalf("unexpected payload: %+v", data)
	}
	if data.IsNew {
		t.Fatalf("expected is_new=false for updated rating, got: %+v", data)
	}
}

func TestRate_SameRatingEmitsNothing(t *testing.T) {
	commentsRepo := &mockCommentsRepository{
		getByIDFn: func(ctx context.Context, commentID int) (domain.Comment, error) {
			return rootArgument(), nil
		},
	}
	debatesRepo := &mockDebatesRepository{
		getByPostFn: func(ctx context.Context, postID int) (domain.Debate, error) {
			return openDebate(), nil
		},
	}
	ratingsRepo := &mockCommentRatingsRepository{
		getFn: func(ctx context.Context, commentID, userID int) (domain.CommentRating, error) {
			return domain.CommentRating{ID: 1, CommentID: 1, UserID: 5, Rating: 3}, nil
		},
		updateFn: func(ctx context.Context, rating domain.CommentRating) (domain.CommentRating, error) {
			t.Fatal("Update should not be called for same rating")
			return domain.CommentRating{}, nil
		},
	}
	publisher := &publisherMock{}

	svc := newRatingsServiceWithPublisher(ratingsRepo, commentsRepo, debatesRepo, publisher)

	_, err := svc.Rate(context.Background(), 5, 1, 3)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if len(publisher.topics) != 0 || len(publisher.events) != 0 {
		t.Fatalf("expected no events, got topics=%v events=%v", publisher.topics, publisher.events)
	}
}

var _ = time.Now
