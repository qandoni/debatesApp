package debate_votes_service

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/qandoni/debatesApp/internal/core/domain"
	core_enum "github.com/qandoni/debatesApp/internal/core/enum"
	core_errors "github.com/qandoni/debatesApp/internal/core/errors"
	core_realtime "github.com/qandoni/debatesApp/internal/core/realtime"
)

type mockDebateVotesRepository struct {
	createFn          func(ctx context.Context, vote domain.DebateVote) (domain.DebateVote, error)
	updateFn          func(ctx context.Context, debateID, userID, debateSideID int, updatedAt time.Time) (domain.DebateVote, error)
	getByDebateUserFn func(ctx context.Context, debateID, userID int) (domain.DebateVote, error)
}

func (m *mockDebateVotesRepository) Create(ctx context.Context, vote domain.DebateVote) (domain.DebateVote, error) {
	return m.createFn(ctx, vote)
}

func (m *mockDebateVotesRepository) Update(ctx context.Context, debateID, userID, debateSideID int, updatedAt time.Time) (domain.DebateVote, error) {
	return m.updateFn(ctx, debateID, userID, debateSideID, updatedAt)
}

func (m *mockDebateVotesRepository) GetByDebateAndUser(ctx context.Context, debateID, userID int) (domain.DebateVote, error) {
	return m.getByDebateUserFn(ctx, debateID, userID)
}

type mockDebatesRepository struct {
	getByIDFn          func(ctx context.Context, debateID int) (domain.Debate, error)
	finishFn           func(ctx context.Context, debateID int) error
	getAuthorFn        func(ctx context.Context, debateID int) (int, error)
	getByPostFn        func(ctx context.Context, postID int) (domain.Debate, error)
	getByIDForUpdateFn func(ctx context.Context, debateID int) (domain.Debate, error)
	createDebateFn     func(ctx context.Context, debate domain.Debate) (domain.Debate, error)
}

// GetByIDForUpdate по умолчанию отдаёт тот же стаб, что и GetByID;
// отдельный getByIDForUpdateFn позволяет тесту проверить, что вызван именно
// блокирующий вариант (FOR UPDATE).
func (m *mockDebatesRepository) GetByIDForUpdate(ctx context.Context, debateID int) (domain.Debate, error) {
	if m.getByIDForUpdateFn != nil {
		return m.getByIDForUpdateFn(ctx, debateID)
	}
	return m.getByIDFn(ctx, debateID)
}

type mockTxManager struct {
	withinFn func(ctx context.Context, fn func(ctx context.Context) error) error
}

func (m *mockTxManager) WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	if m.withinFn == nil {
		return fn(ctx)
	}
	return m.withinFn(ctx, fn)
}

func (m *mockDebatesRepository) GetByID(ctx context.Context, debateID int) (domain.Debate, error) {
	return m.getByIDFn(ctx, debateID)
}

func (m *mockDebatesRepository) FinishDebate(ctx context.Context, debateID int) error {
	return m.finishFn(ctx, debateID)
}

func (m *mockDebatesRepository) GetAuthorID(ctx context.Context, debateID int) (int, error) {
	return m.getAuthorFn(ctx, debateID)
}

func (m *mockDebatesRepository) GetByPostID(ctx context.Context, postID int) (domain.Debate, error) {
	return m.getByPostFn(ctx, postID)
}

func (m *mockDebatesRepository) CreateDebate(ctx context.Context, debate domain.Debate) (domain.Debate, error) {
	return m.createDebateFn(ctx, debate)
}

type mockDebateSidesRepository struct {
	getByDebateIDFn func(ctx context.Context, debateID int) ([]domain.DebateSide, error)
	createSideFn    func(ctx context.Context, side domain.DebateSide) (domain.DebateSide, error)
}

func (m *mockDebateSidesRepository) GetByDebateID(ctx context.Context, debateID int) ([]domain.DebateSide, error) {
	return m.getByDebateIDFn(ctx, debateID)
}

func (m *mockDebateSidesRepository) CreateDebateSide(ctx context.Context, side domain.DebateSide) (domain.DebateSide, error) {
	return m.createSideFn(ctx, side)
}

type mockCommentsRepository struct {
	hasArgumentFn func(ctx context.Context, userID, postID int) (bool, error)
}

func (m *mockCommentsRepository) HasUserArgumentInPost(ctx context.Context, userID, postID int) (bool, error) {
	return m.hasArgumentFn(ctx, userID, postID)
}

func openDebate(debateID, postID int) domain.Debate {
	return domain.NewDebate(
		debateID,
		postID,
		core_enum.DebateStatusOpen,
		nil,
		time.Now(),
		nil,
		nil,
	)
}

func finishedDebate(debateID, postID int) domain.Debate {
	return domain.NewDebate(
		debateID,
		postID,
		core_enum.DebateStatusFinished,
		nil,
		time.Now(),
		nil,
		nil,
	)
}

type publisherMock struct {
	topics []string
	events []core_realtime.Event
	err    error
}

func (m *publisherMock) Publish(topic string, event core_realtime.Event) error {
	m.topics = append(m.topics, topic)
	m.events = append(m.events, event)
	return m.err
}

func newTestService(
	debateVotes DebateVotesRepository,
	debates DebatesRepository,
	sides DebateSidesRepository,
	comments CommentsRepository,
) *DebateVotesService {
	return newTestServiceWithPublisher(
		debateVotes,
		debates,
		sides,
		comments,
		&publisherMock{},
	)
}

func newTestServiceWithPublisher(
	debateVotes DebateVotesRepository,
	debates DebatesRepository,
	sides DebateSidesRepository,
	comments CommentsRepository,
	publisher core_realtime.Publisher,
) *DebateVotesService {
	return NewDebateVotesService(debateVotes, debates, sides, comments, &mockTxManager{}, publisher)
}

func TestVote_Success(t *testing.T) {
	voteRepo := &mockDebateVotesRepository{
		createFn: func(ctx context.Context, vote domain.DebateVote) (domain.DebateVote, error) {
			if vote.DebateID != 10 || vote.UserID != 5 || vote.DebateSideID != 2 {
				t.Fatalf("unexpected vote: %+v", vote)
			}
			return vote, nil
		},
	}
	debatesRepo := &mockDebatesRepository{
		getByIDFn: func(ctx context.Context, debateID int) (domain.Debate, error) {
			return openDebate(10, 100), nil
		},
	}
	sidesRepo := &mockDebateSidesRepository{
		getByDebateIDFn: func(ctx context.Context, debateID int) ([]domain.DebateSide, error) {
			return []domain.DebateSide{{ID: 1}, {ID: 2}, {ID: 3}}, nil
		},
	}

	svc := newTestService(voteRepo, debatesRepo, sidesRepo, &mockCommentsRepository{})
	vote, err := svc.Vote(context.Background(), 5, 10, 2)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if vote.DebateSideID != 2 {
		t.Fatalf("expected side 2, got %d", vote.DebateSideID)
	}
}

func TestVote_DebateNotFound(t *testing.T) {
	voteRepo := &mockDebateVotesRepository{
		createFn: func(ctx context.Context, vote domain.DebateVote) (domain.DebateVote, error) {
			t.Fatal("Create should not be called")
			return domain.DebateVote{}, nil
		},
	}
	debatesRepo := &mockDebatesRepository{
		getByIDFn: func(ctx context.Context, debateID int) (domain.Debate, error) {
			return domain.Debate{}, core_errors.ErrNotFound
		},
	}

	svc := newTestService(voteRepo, debatesRepo, &mockDebateSidesRepository{}, &mockCommentsRepository{})
	_, err := svc.Vote(context.Background(), 5, 10, 2)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestVote_DebateClosed(t *testing.T) {
	voteRepo := &mockDebateVotesRepository{
		createFn: func(ctx context.Context, vote domain.DebateVote) (domain.DebateVote, error) {
			t.Fatal("Create should not be called")
			return domain.DebateVote{}, nil
		},
	}
	debatesRepo := &mockDebatesRepository{
		getByIDFn: func(ctx context.Context, debateID int) (domain.Debate, error) {
			return finishedDebate(10, 100), nil
		},
	}

	svc := newTestService(voteRepo, debatesRepo, &mockDebateSidesRepository{}, &mockCommentsRepository{})
	_, err := svc.Vote(context.Background(), 5, 10, 2)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestVote_SideNotBelongToDebate(t *testing.T) {
	voteRepo := &mockDebateVotesRepository{
		createFn: func(ctx context.Context, vote domain.DebateVote) (domain.DebateVote, error) {
			t.Fatal("Create should not be called")
			return domain.DebateVote{}, nil
		},
	}
	debatesRepo := &mockDebatesRepository{
		getByIDFn: func(ctx context.Context, debateID int) (domain.Debate, error) {
			return openDebate(10, 100), nil
		},
	}
	sidesRepo := &mockDebateSidesRepository{
		getByDebateIDFn: func(ctx context.Context, debateID int) ([]domain.DebateSide, error) {
			return []domain.DebateSide{{ID: 1}, {ID: 2}}, nil
		},
	}

	svc := newTestService(voteRepo, debatesRepo, sidesRepo, &mockCommentsRepository{})
	_, err := svc.Vote(context.Background(), 5, 10, 99)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, core_errors.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got: %v", err)
	}
}

func TestChangeVote_Success(t *testing.T) {
	voteRepo := &mockDebateVotesRepository{
		updateFn: func(ctx context.Context, debateID, userID, debateSideID int, updatedAt time.Time) (domain.DebateVote, error) {
			if debateID != 10 || userID != 5 || debateSideID != 3 {
				t.Fatalf("unexpected update args: %d %d %d", debateID, userID, debateSideID)
			}
			return domain.NewDebateVote(1, 2, 10, 5, 3, time.Now(), &updatedAt, true), nil
		},
	}
	debatesRepo := &mockDebatesRepository{
		getByIDFn: func(ctx context.Context, debateID int) (domain.Debate, error) {
			return openDebate(10, 100), nil
		},
	}
	sidesRepo := &mockDebateSidesRepository{
		getByDebateIDFn: func(ctx context.Context, debateID int) ([]domain.DebateSide, error) {
			return []domain.DebateSide{{ID: 1}, {ID: 3}}, nil
		},
	}
	commentsRepo := &mockCommentsRepository{
		hasArgumentFn: func(ctx context.Context, userID, postID int) (bool, error) {
			return false, nil
		},
	}

	svc := newTestService(voteRepo, debatesRepo, sidesRepo, commentsRepo)
	vote, err := svc.ChangeVote(context.Background(), 5, 10, 3)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if vote.DebateSideID != 3 {
		t.Fatalf("expected side 3, got %d", vote.DebateSideID)
	}
}

func TestChangeVote_ForbiddenAfterArgument(t *testing.T) {
	voteRepo := &mockDebateVotesRepository{
		updateFn: func(ctx context.Context, debateID, userID, debateSideID int, updatedAt time.Time) (domain.DebateVote, error) {
			t.Fatal("Update should not be called")
			return domain.DebateVote{}, nil
		},
	}
	debatesRepo := &mockDebatesRepository{
		getByIDFn: func(ctx context.Context, debateID int) (domain.Debate, error) {
			return openDebate(10, 100), nil
		},
	}
	sidesRepo := &mockDebateSidesRepository{
		getByDebateIDFn: func(ctx context.Context, debateID int) ([]domain.DebateSide, error) {
			return []domain.DebateSide{{ID: 1}, {ID: 3}}, nil
		},
	}
	commentsRepo := &mockCommentsRepository{
		hasArgumentFn: func(ctx context.Context, userID, postID int) (bool, error) {
			return true, nil
		},
	}

	svc := newTestService(voteRepo, debatesRepo, sidesRepo, commentsRepo)
	_, err := svc.ChangeVote(context.Background(), 5, 10, 3)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, core_errors.ErrAccessForbidden) {
		t.Fatalf("expected ErrAccessForbidden, got: %v", err)
	}
}

func TestChangeVote_DebateClosed(t *testing.T) {
	voteRepo := &mockDebateVotesRepository{
		updateFn: func(ctx context.Context, debateID, userID, debateSideID int, updatedAt time.Time) (domain.DebateVote, error) {
			t.Fatal("Update should not be called")
			return domain.DebateVote{}, nil
		},
	}
	debatesRepo := &mockDebatesRepository{
		getByIDFn: func(ctx context.Context, debateID int) (domain.Debate, error) {
			return finishedDebate(10, 100), nil
		},
	}

	svc := newTestService(voteRepo, debatesRepo, &mockDebateSidesRepository{}, &mockCommentsRepository{})
	_, err := svc.ChangeVote(context.Background(), 5, 10, 3)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestFinishDebate_Success(t *testing.T) {
	debatesRepo := &mockDebatesRepository{
		getByIDFn: func(ctx context.Context, debateID int) (domain.Debate, error) {
			return openDebate(10, 100), nil
		},
		getAuthorFn: func(ctx context.Context, debateID int) (int, error) {
			return 5, nil
		},
		finishFn: func(ctx context.Context, debateID int) error {
			if debateID != 10 {
				t.Fatalf("expected debate 10, got %d", debateID)
			}
			return nil
		},
	}

	svc := newTestService(&mockDebateVotesRepository{}, debatesRepo, &mockDebateSidesRepository{}, &mockCommentsRepository{})
	err := svc.FinishDebate(context.Background(), 5, 10)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func TestFinishDebate_NotAuthor(t *testing.T) {
	debatesRepo := &mockDebatesRepository{
		getByIDFn: func(ctx context.Context, debateID int) (domain.Debate, error) {
			return openDebate(10, 100), nil
		},
		getAuthorFn: func(ctx context.Context, debateID int) (int, error) {
			return 5, nil
		},
		finishFn: func(ctx context.Context, debateID int) error {
			t.Fatal("FinishDebate should not be called")
			return nil
		},
	}

	svc := newTestService(&mockDebateVotesRepository{}, debatesRepo, &mockDebateSidesRepository{}, &mockCommentsRepository{})
	err := svc.FinishDebate(context.Background(), 99, 10)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, core_errors.ErrAccessForbidden) {
		t.Fatalf("expected ErrAccessForbidden, got: %v", err)
	}
}

func TestFinishDebate_AlreadyFinished(t *testing.T) {
	debatesRepo := &mockDebatesRepository{
		getByIDFn: func(ctx context.Context, debateID int) (domain.Debate, error) {
			return finishedDebate(10, 100), nil
		},
		getAuthorFn: func(ctx context.Context, debateID int) (int, error) {
			return 5, nil
		},
		finishFn: func(ctx context.Context, debateID int) error {
			t.Fatal("FinishDebate should not be called")
			return nil
		},
	}

	svc := newTestService(&mockDebateVotesRepository{}, debatesRepo, &mockDebateSidesRepository{}, &mockCommentsRepository{})
	err := svc.FinishDebate(context.Background(), 5, 10)
	if err == nil {
		t.Fatal("expected error, got nil")
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

func TestVote_EmitsVoteCreated(t *testing.T) {
	voteRepo := &mockDebateVotesRepository{
		createFn: func(ctx context.Context, vote domain.DebateVote) (domain.DebateVote, error) {
			vote.ID = 1
			return vote, nil
		},
	}
	debatesRepo := &mockDebatesRepository{
		getByIDFn: func(ctx context.Context, debateID int) (domain.Debate, error) {
			return openDebate(10, 100), nil
		},
	}
	sidesRepo := &mockDebateSidesRepository{
		getByDebateIDFn: func(ctx context.Context, debateID int) ([]domain.DebateSide, error) {
			return []domain.DebateSide{{ID: 1}, {ID: 2}, {ID: 3}}, nil
		},
	}
	publisher := &publisherMock{}

	svc := newTestServiceWithPublisher(
		voteRepo,
		debatesRepo,
		sidesRepo,
		&mockCommentsRepository{},
		publisher,
	)

	_, err := svc.Vote(context.Background(), 5, 10, 2)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	event := assertSingleEvent(
		t,
		publisher,
		core_realtime.PostTopic(100),
		core_realtime.EventDebateVoteCreated,
	)

	data, ok := event.Data.(core_realtime.VoteData)
	if !ok {
		t.Fatalf("expected VoteData, got %T", event.Data)
	}
	if data.DebateID != 10 || data.PostID != 100 || data.UserID != 5 || data.DebateSideID != 2 {
		t.Fatalf("unexpected payload: %+v", data)
	}
	if data.IsChanged {
		t.Fatalf("expected is_changed=false for new vote, got: %+v", data)
	}
}

func TestChangeVote_EmitsVoteChanged(t *testing.T) {
	voteRepo := &mockDebateVotesRepository{
		updateFn: func(ctx context.Context, debateID, userID, debateSideID int, updatedAt time.Time) (domain.DebateVote, error) {
			return domain.NewDebateVote(1, 2, debateID, userID, debateSideID, time.Now(), &updatedAt, true), nil
		},
	}
	debatesRepo := &mockDebatesRepository{
		getByIDFn: func(ctx context.Context, debateID int) (domain.Debate, error) {
			return openDebate(10, 100), nil
		},
	}
	sidesRepo := &mockDebateSidesRepository{
		getByDebateIDFn: func(ctx context.Context, debateID int) ([]domain.DebateSide, error) {
			return []domain.DebateSide{{ID: 1}, {ID: 3}}, nil
		},
	}
	commentsRepo := &mockCommentsRepository{
		hasArgumentFn: func(ctx context.Context, userID, postID int) (bool, error) {
			return false, nil
		},
	}
	publisher := &publisherMock{}

	svc := newTestServiceWithPublisher(voteRepo, debatesRepo, sidesRepo, commentsRepo, publisher)

	_, err := svc.ChangeVote(context.Background(), 5, 10, 3)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	event := assertSingleEvent(
		t,
		publisher,
		core_realtime.PostTopic(100),
		core_realtime.EventDebateVoteChanged,
	)

	data, ok := event.Data.(core_realtime.VoteData)
	if !ok {
		t.Fatalf("expected VoteData, got %T", event.Data)
	}
	if data.DebateID != 10 || data.PostID != 100 || data.UserID != 5 || data.DebateSideID != 3 {
		t.Fatalf("unexpected payload: %+v", data)
	}
	if !data.IsChanged {
		t.Fatalf("expected is_changed=true for changed vote, got: %+v", data)
	}
}

func TestFinishDebate_EmitsDebateFinished(t *testing.T) {
	winnerSideID := 2
	finishedAt := time.Now()
	var getByIDCalls int

	debatesRepo := &mockDebatesRepository{
		getByIDFn: func(ctx context.Context, debateID int) (domain.Debate, error) {
			getByIDCalls++
			if getByIDCalls == 1 {
				return openDebate(10, 100), nil
			}

			return domain.NewDebate(
				10,
				100,
				core_enum.DebateStatusFinished,
				nil,
				time.Now(),
				&finishedAt,
				&winnerSideID,
			), nil
		},
		getAuthorFn: func(ctx context.Context, debateID int) (int, error) {
			return 5, nil
		},
		finishFn: func(ctx context.Context, debateID int) error {
			return nil
		},
	}
	publisher := &publisherMock{}

	svc := newTestServiceWithPublisher(
		&mockDebateVotesRepository{},
		debatesRepo,
		&mockDebateSidesRepository{},
		&mockCommentsRepository{},
		publisher,
	)

	err := svc.FinishDebate(context.Background(), 5, 10)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if getByIDCalls != 2 {
		t.Fatalf("expected debate to be re-read after finishing, got %d calls", getByIDCalls)
	}

	event := assertSingleEvent(
		t,
		publisher,
		core_realtime.PostTopic(100),
		core_realtime.EventDebateFinished,
	)

	data, ok := event.Data.(core_realtime.DebateFinishedData)
	if !ok {
		t.Fatalf("expected DebateFinishedData, got %T", event.Data)
	}
	if data.DebateID != 10 || data.PostID != 100 || data.FinishedByUserID != 5 {
		t.Fatalf("unexpected payload: %+v", data)
	}
	if data.WinnerSideID == nil || *data.WinnerSideID != winnerSideID {
		t.Fatalf("expected winner side %d, got: %+v", winnerSideID, data.WinnerSideID)
	}
}

func TestFinishDebate_ErrorEmitsNothing(t *testing.T) {
	debatesRepo := &mockDebatesRepository{
		getByIDFn: func(ctx context.Context, debateID int) (domain.Debate, error) {
			return openDebate(10, 100), nil
		},
		getAuthorFn: func(ctx context.Context, debateID int) (int, error) {
			return 5, nil
		},
		finishFn: func(ctx context.Context, debateID int) error {
			t.Fatal("FinishDebate should not be called")
			return nil
		},
	}
	publisher := &publisherMock{}

	svc := newTestServiceWithPublisher(
		&mockDebateVotesRepository{},
		debatesRepo,
		&mockDebateSidesRepository{},
		&mockCommentsRepository{},
		publisher,
	)

	err := svc.FinishDebate(context.Background(), 99, 10)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if len(publisher.topics) != 0 || len(publisher.events) != 0 {
		t.Fatalf("expected no events, got topics=%v events=%v", publisher.topics, publisher.events)
	}
}

type peerMock struct {
	payloads [][]byte
}

func (m *peerMock) Send(payload []byte) {
	m.payloads = append(m.payloads, payload)
}

func (m *peerMock) Shutdown() {}

// TestVote_DeliversEnvelopeToSubscribedPeer проверяет весь путь события:
// сервис -> Hub -> сериализованный конверт у подписчика.
func TestVote_DeliversEnvelopeToSubscribedPeer(t *testing.T) {
	hub := core_realtime.NewHub()
	peer := &peerMock{}
	hub.Subscribe(core_realtime.PostTopic(100), peer)

	voteRepo := &mockDebateVotesRepository{
		createFn: func(ctx context.Context, vote domain.DebateVote) (domain.DebateVote, error) {
			vote.ID = 1
			return vote, nil
		},
	}
	debatesRepo := &mockDebatesRepository{
		getByIDFn: func(ctx context.Context, debateID int) (domain.Debate, error) {
			return openDebate(10, 100), nil
		},
	}
	sidesRepo := &mockDebateSidesRepository{
		getByDebateIDFn: func(ctx context.Context, debateID int) ([]domain.DebateSide, error) {
			return []domain.DebateSide{{ID: 1}, {ID: 2}}, nil
		},
	}

	svc := NewDebateVotesService(voteRepo, debatesRepo, sidesRepo, &mockCommentsRepository{}, &mockTxManager{}, hub)

	_, err := svc.Vote(context.Background(), 5, 10, 2)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if len(peer.payloads) != 1 {
		t.Fatalf("expected 1 payload delivered, got %d", len(peer.payloads))
	}

	var envelope core_realtime.Envelope
	if err := json.Unmarshal(peer.payloads[0], &envelope); err != nil {
		t.Fatalf("expected valid JSON envelope, got error: %v", err)
	}
	if envelope.Type != core_realtime.EventDebateVoteCreated {
		t.Fatalf("expected event type %q, got %q", core_realtime.EventDebateVoteCreated, envelope.Type)
	}
	if envelope.Topic != core_realtime.PostTopic(100) {
		t.Fatalf("expected topic %q, got %q", core_realtime.PostTopic(100), envelope.Topic)
	}
	if envelope.OccurredAt.IsZero() {
		t.Fatal("expected occurred_at to be set")
	}

	data, ok := envelope.Data.(map[string]any)
	if !ok {
		t.Fatalf("expected JSON object payload, got %T", envelope.Data)
	}
	if data["debate_id"] != float64(10) || data["post_id"] != float64(100) {
		t.Fatalf("unexpected payload: %+v", data)
	}
}

func TestVote_PublishErrorPropagates(t *testing.T) {
	persisted := false
	voteRepo := &mockDebateVotesRepository{
		createFn: func(ctx context.Context, vote domain.DebateVote) (domain.DebateVote, error) {
			persisted = true
			return vote, nil
		},
	}
	debatesRepo := &mockDebatesRepository{
		getByIDFn: func(ctx context.Context, debateID int) (domain.Debate, error) {
			return openDebate(10, 100), nil
		},
	}
	sidesRepo := &mockDebateSidesRepository{
		getByDebateIDFn: func(ctx context.Context, debateID int) ([]domain.DebateSide, error) {
			return []domain.DebateSide{{ID: 1}, {ID: 2}, {ID: 3}}, nil
		},
	}
	publishErr := errors.New("publish failed")
	publisher := &publisherMock{err: publishErr}

	svc := newTestServiceWithPublisher(voteRepo, debatesRepo, sidesRepo, &mockCommentsRepository{}, publisher)

	_, err := svc.Vote(context.Background(), 5, 10, 2)
	if err == nil {
		t.Fatal("expected publish error, got nil")
	}
	if !errors.Is(err, publishErr) {
		t.Fatalf("expected wrapped publish error, got: %v", err)
	}
	if !persisted {
		t.Fatal("vote must be persisted before publish (publish error must not roll back DB write)")
	}
	if len(publisher.events) != 1 {
		t.Fatalf("expected 1 publish attempt, got %d", len(publisher.events))
	}
}
