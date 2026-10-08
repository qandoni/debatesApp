package core_realtime

import (
	"encoding/json"
	"sync"
	"testing"
	"time"

	"github.com/qandoni/debatesApp/internal/core/domain"
)

type peerMock struct {
	mtx       sync.Mutex
	payloads  [][]byte
	shutdowns int
}

func (m *peerMock) Send(payload []byte) {
	m.mtx.Lock()
	defer m.mtx.Unlock()

	m.payloads = append(m.payloads, payload)
}

func (m *peerMock) Shutdown() {
	m.mtx.Lock()
	defer m.mtx.Unlock()

	m.shutdowns++
}

func (m *peerMock) count() int {
	m.mtx.Lock()
	defer m.mtx.Unlock()

	return len(m.payloads)
}

func (m *peerMock) last(t *testing.T) []byte {
	t.Helper()

	m.mtx.Lock()
	defer m.mtx.Unlock()

	if len(m.payloads) == 0 {
		t.Fatal("expected at least one payload")
	}

	return m.payloads[len(m.payloads)-1]
}

func intPtr(value int) *int {
	return &value
}

func TestPostTopic(t *testing.T) {
	if got := PostTopic(42); got != "post:42" {
		t.Fatalf("expected %q, got %q", "post:42", got)
	}
}

func TestHub_PublishDeliversOnlyToTopicSubscribers(t *testing.T) {
	hub := NewHub()
	subscriber := &peerMock{}
	otherTopicSubscriber := &peerMock{}

	hub.Subscribe(PostTopic(1), subscriber)
	hub.Subscribe(PostTopic(2), otherTopicSubscriber)

	hub.Publish(PostTopic(1), NewEvent("test.event", map[string]int{"id": 1}))

	if subscriber.count() != 1 {
		t.Fatalf("expected 1 payload for subscriber, got %d", subscriber.count())
	}
	if otherTopicSubscriber.count() != 0 {
		t.Fatalf("expected no payload for other topic, got %d", otherTopicSubscriber.count())
	}

	var envelope Envelope
	if err := json.Unmarshal(subscriber.last(t), &envelope); err != nil {
		t.Fatalf("failed to unmarshal envelope: %v", err)
	}
	if envelope.Type != "test.event" {
		t.Fatalf("expected type %q, got %q", "test.event", envelope.Type)
	}
	if envelope.Topic != PostTopic(1) {
		t.Fatalf("expected topic %q, got %q", PostTopic(1), envelope.Topic)
	}
	if envelope.OccurredAt.IsZero() {
		t.Fatal("expected occurred_at to be set")
	}
}

func TestHub_UnsubscribeStopsDelivery(t *testing.T) {
	hub := NewHub()
	peer := &peerMock{}

	hub.Subscribe(PostTopic(1), peer)
	hub.Unsubscribe(PostTopic(1), peer)
	hub.Publish(PostTopic(1), NewEvent("test.event", nil))

	if peer.count() != 0 {
		t.Fatalf("expected no payload after unsubscribe, got %d", peer.count())
	}

	// повторная отписка не должна паниковать
	hub.Unsubscribe(PostTopic(1), peer)
}

func TestHub_UnsubscribeAllStopsDelivery(t *testing.T) {
	hub := NewHub()
	peer := &peerMock{}

	hub.Subscribe(PostTopic(1), peer)
	hub.Subscribe(PostTopic(2), peer)

	hub.UnsubscribeAll(peer)

	hub.Publish(PostTopic(1), NewEvent("test.event", nil))
	hub.Publish(PostTopic(2), NewEvent("test.event", nil))

	if peer.count() != 0 {
		t.Fatalf("expected no payload after unsubscribe all, got %d", peer.count())
	}

	// повторный вызов идемпотентен
	hub.UnsubscribeAll(peer)
}

func TestHub_PublishWithoutSubscribers(t *testing.T) {
	hub := NewHub()

	hub.Publish(PostTopic(99), NewEvent("test.event", nil))
}

func TestHub_PublishReturnsMarshalError(t *testing.T) {
	hub := NewHub()

	if err := hub.Publish(PostTopic(1), NewEvent("test.event", make(chan int))); err == nil {
		t.Fatal("expected marshal error, got nil")
	}
}

type reentrantPeer struct {
	hub   *Hub
	sends int
}

func (p *reentrantPeer) Send(payload []byte) {
	p.sends++
	p.hub.UnsubscribeAll(p)
}

func (p *reentrantPeer) Shutdown() {}

func TestHub_PublishAllowsSendToUnsubscribe(t *testing.T) {
	hub := NewHub()
	peer := &reentrantPeer{hub: hub}
	hub.Subscribe(PostTopic(1), peer)

	done := make(chan error, 1)
	go func() {
		done <- hub.Publish(PostTopic(1), NewEvent("test.event", nil))
	}()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Publish deadlocked: Send must be able to call UnsubscribeAll")
	}

	if peer.sends != 1 {
		t.Fatalf("expected 1 send, got %d", peer.sends)
	}

	// пир отписался во время Send — повторная рассылка ему не идёт
	if err := hub.Publish(PostTopic(1), NewEvent("test.event", nil)); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if peer.sends != 1 {
		t.Fatalf("expected no delivery after unsubscribe, got %d sends", peer.sends)
	}
}

func TestHub_ConcurrentPublishSubscribe(t *testing.T) {
	const (
		subscribers = 8
		publishers  = 8
		iterations  = 50
	)

	hub := NewHub()
	peers := make([]*peerMock, subscribers)
	for i := range peers {
		peers[i] = &peerMock{}
	}

	var wg sync.WaitGroup

	for _, peer := range peers {
		wg.Add(1)
		go func(peer *peerMock) {
			defer wg.Done()
			for i := 0; i < iterations; i++ {
				hub.Subscribe(PostTopic(1), peer)
				hub.Publish(PostTopic(1), NewEvent("test.event", i))
				hub.Unsubscribe(PostTopic(1), peer)
				hub.UnsubscribeAll(peer)
			}
		}(peer)
	}

	for i := 0; i < publishers; i++ {
		wg.Add(1)
		go func(number int) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				hub.Publish(PostTopic(1), NewEvent("test.event", number))
				hub.Publish(PostTopic(2), NewEvent("test.event", number))
			}
		}(i)
	}

	wg.Wait()
}

func assertJSONKeys(t *testing.T, data any, keys ...string) {
	t.Helper()

	payload, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}

	var raw map[string]any
	if err := json.Unmarshal(payload, &raw); err != nil {
		t.Fatalf("failed to unmarshal payload: %v", err)
	}

	for _, key := range keys {
		if _, ok := raw[key]; !ok {
			t.Fatalf("expected key %q in %s", key, payload)
		}
	}
}

func TestNewCommentCreatedData(t *testing.T) {
	createdAt := time.Now()
	comment := domain.Comment{
		ID:              1,
		PostID:          100,
		ParentCommentID: intPtr(7),
		AuthorID:        5,
		DebateSideID:    intPtr(2),
		AuthorLiked:     true,
		Content:         "hello",
		CreatedAt:       createdAt,
	}

	data := NewCommentCreatedData(comment)

	if data.CommentID != 1 || data.PostID != 100 || data.AuthorID != 5 || data.Content != "hello" {
		t.Fatalf("unexpected payload: %+v", data)
	}
	if data.ParentCommentID == nil || *data.ParentCommentID != 7 {
		t.Fatalf("unexpected parent comment id: %+v", data.ParentCommentID)
	}
	if data.DebateSideID == nil || *data.DebateSideID != 2 {
		t.Fatalf("unexpected debate side id: %+v", data.DebateSideID)
	}
	if !data.AuthorLiked || !data.CreatedAt.Equal(createdAt) {
		t.Fatalf("unexpected payload: %+v", data)
	}

	assertJSONKeys(
		t,
		data,
		"comment_id",
		"post_id",
		"parent_comment_id",
		"author_id",
		"debate_side_id",
		"author_liked",
		"content",
		"created_at",
	)
}

func TestNewCommentUpdatedData(t *testing.T) {
	updatedAt := time.Now()
	comment := domain.Comment{
		ID:        1,
		PostID:    100,
		AuthorID:  5,
		Content:   "updated",
		UpdatedAt: &updatedAt,
	}

	data := NewCommentUpdatedData(comment)

	if data.CommentID != 1 || data.PostID != 100 || data.AuthorID != 5 || data.Content != "updated" {
		t.Fatalf("unexpected payload: %+v", data)
	}
	if data.UpdatedAt == nil || !data.UpdatedAt.Equal(updatedAt) {
		t.Fatalf("unexpected updated at: %+v", data.UpdatedAt)
	}

	assertJSONKeys(t, data, "comment_id", "post_id", "author_id", "content", "updated_at")
}

func TestNewAuthorLikeData(t *testing.T) {
	comment := domain.Comment{ID: 1, PostID: 100, AuthorLiked: true}

	data := NewAuthorLikeData(comment, 5)

	if data.CommentID != 1 || data.PostID != 100 || data.DebateAuthorID != 5 || !data.AuthorLiked {
		t.Fatalf("unexpected payload: %+v", data)
	}

	assertJSONKeys(t, data, "comment_id", "post_id", "debate_author_id", "author_liked")
}

func TestNewRatingData(t *testing.T) {
	rating := domain.CommentRating{ID: 1, CommentID: 2, UserID: 5, Rating: 4}

	data := NewRatingData(rating, 100, true)

	if data.CommentID != 2 || data.PostID != 100 || data.UserID != 5 || data.Score != 4 || !data.IsNew {
		t.Fatalf("unexpected payload: %+v", data)
	}

	assertJSONKeys(t, data, "comment_id", "post_id", "user_id", "score", "is_new")
}

func TestNewVoteData(t *testing.T) {
	vote := domain.DebateVote{ID: 1, DebateID: 10, UserID: 5, DebateSideID: 3}

	data := NewVoteData(vote, 100, true)

	if data.DebateID != 10 || data.PostID != 100 || data.UserID != 5 || data.DebateSideID != 3 {
		t.Fatalf("unexpected payload: %+v", data)
	}
	if !data.IsChanged {
		t.Fatalf("expected is_changed=true, got: %+v", data)
	}

	assertJSONKeys(t, data, "debate_id", "post_id", "user_id", "debate_side_id", "is_changed")
}

func TestNewDebateFinishedData(t *testing.T) {
	debate := domain.Debate{ID: 10, PostID: 100, WinnerSideID: intPtr(2)}

	data := NewDebateFinishedData(debate, 5)

	if data.DebateID != 10 || data.PostID != 100 || data.FinishedByUserID != 5 {
		t.Fatalf("unexpected payload: %+v", data)
	}
	if data.WinnerSideID == nil || *data.WinnerSideID != 2 {
		t.Fatalf("unexpected winner side id: %+v", data.WinnerSideID)
	}

	assertJSONKeys(t, data, "debate_id", "post_id", "winner_side_id", "finished_by_user_id")
}

func TestHub_ShutdownClosesEachPeerOnce(t *testing.T) {
	hub := NewHub()
	peer := &peerMock{}

	hub.Subscribe(PostTopic(1), peer)
	hub.Subscribe(PostTopic(2), peer)

	hub.Shutdown()

	if peer.shutdowns != 1 {
		t.Fatalf("expected peer to be shut down once, got %d", peer.shutdowns)
	}
}
