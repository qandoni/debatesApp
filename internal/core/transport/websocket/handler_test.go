package core_transport_websocket

import (
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	core_auth "github.com/qandoni/debatesApp/internal/core/auth"
	core_errors "github.com/qandoni/debatesApp/internal/core/errors"
	core_logger "github.com/qandoni/debatesApp/internal/core/logger"
	core_realtime "github.com/qandoni/debatesApp/internal/core/realtime"
	core_http_middleware "github.com/qandoni/debatesApp/internal/core/transport/http/middleware"
	"go.uber.org/zap"
)

const (
	testOrigin = "http://localhost:3000"
	testToken  = "valid-token"
	testUserID = 42
)

func newTestLogger() *core_logger.Logger {
	return &core_logger.Logger{Logger: zap.NewNop()}
}

type tokenParserStub struct {
	userID int
	err    error
}

func (s tokenParserStub) ParseAccessToken(token string) (core_auth.AuthInfo, error) {
	if s.err != nil {
		return core_auth.AuthInfo{}, s.err
	}
	return core_auth.AuthInfo{UserID: s.userID}, nil
}

func testConfig() Config {
	return Config{
		AllowedOrigins: []string{testOrigin},
		ReadLimit:      4096,
		WriteWait:      2 * time.Second,
		PongWait:       5 * time.Second,
		PingPeriod:     2 * time.Second,
		SendBufferSize: 8,
		AuthWait:       time.Second,
	}
}

type wsFixture struct {
	server *httptest.Server
	hub    *core_realtime.Hub
}

func newWSFixture(t *testing.T, parser TokenParser, config Config) *wsFixture {
	t.Helper()

	gin.SetMode(gin.TestMode)

	router := gin.New()
	// Handler.Handle берёт логгер из контекста запроса, а core_logger.FromContext
	// паникует без него — middleware обязателен, как и в main.go.
	router.Use(
		core_http_middleware.RequestID(),
		core_http_middleware.Logger(newTestLogger()),
	)

	hub := core_realtime.NewHub()
	NewHandler(hub, parser, config).Register(router.Group("/api/v1"))

	server := httptest.NewServer(router)
	t.Cleanup(server.Close)

	return &wsFixture{server: server, hub: hub}
}

func (f *wsFixture) wsURL() string {
	return "ws" + strings.TrimPrefix(f.server.URL, "http") + "/api/v1/ws"
}

func (f *wsFixture) dial(t *testing.T, origin string) *websocket.Conn {
	t.Helper()

	header := http.Header{}
	if origin != "" {
		header.Set("Origin", origin)
	}

	conn, response, err := websocket.DefaultDialer.Dial(f.wsURL(), header)
	if response != nil {
		defer response.Body.Close()
	}
	if err != nil {
		t.Fatalf("failed to dial websocket: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	return conn
}

func (f *wsFixture) dialStatus(t *testing.T, origin string) int {
	t.Helper()

	header := http.Header{}
	if origin != "" {
		header.Set("Origin", origin)
	}

	conn, response, _ := websocket.DefaultDialer.Dial(f.wsURL(), header)
	if conn != nil {
		_ = conn.Close()
	}
	if response == nil {
		t.Fatal("expected HTTP response from failed handshake")
	}
	defer response.Body.Close()

	return response.StatusCode
}

func writeClientMessage(t *testing.T, conn *websocket.Conn, messageType string, data any) {
	t.Helper()

	payload, err := json.Marshal(map[string]any{"type": messageType, "data": data})
	if err != nil {
		t.Fatalf("failed to marshal client message: %v", err)
	}
	if err := conn.WriteMessage(websocket.TextMessage, payload); err != nil {
		t.Fatalf("failed to write client message: %v", err)
	}
}

func readRaw(t *testing.T, conn *websocket.Conn, timeout time.Duration) []byte {
	t.Helper()

	if err := conn.SetReadDeadline(time.Now().Add(timeout)); err != nil {
		t.Fatalf("failed to set read deadline: %v", err)
	}
	_, payload, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("failed to read server frame: %v", err)
	}

	return payload
}

func readServerMessage(t *testing.T, conn *websocket.Conn, timeout time.Duration) ServerMessage {
	t.Helper()

	payload := readRaw(t, conn, timeout)

	var message ServerMessage
	if err := json.Unmarshal(payload, &message); err != nil {
		t.Fatalf("failed to unmarshal server message: %v", err)
	}

	return message
}

func readEnvelope(t *testing.T, conn *websocket.Conn, timeout time.Duration) core_realtime.Envelope {
	t.Helper()

	payload := readRaw(t, conn, timeout)

	var envelope core_realtime.Envelope
	if err := json.Unmarshal(payload, &envelope); err != nil {
		t.Fatalf("failed to unmarshal envelope: %v", err)
	}

	return envelope
}

// authAndReady выполняет рукопожатие и проверяет кадр ready.
func authAndReady(t *testing.T, conn *websocket.Conn) {
	t.Helper()

	writeClientMessage(t, conn, MessageTypeAuth, AuthPayload{Token: testToken})

	message := readServerMessage(t, conn, time.Second)
	if message.Type != MessageTypeReady {
		t.Fatalf("expected %q frame, got %q", MessageTypeReady, message.Type)
	}

	data, ok := message.Data.(map[string]any)
	if !ok {
		t.Fatalf("expected ready payload object, got %T", message.Data)
	}
	if data["user_id"] != float64(testUserID) {
		t.Fatalf("expected user_id %d, got %v", testUserID, data["user_id"])
	}
}

// expectAck читает кадр-подтверждение подписки и сверяет его тип и post_id.
func expectAck(t *testing.T, conn *websocket.Conn, messageType string, postID int) {
	t.Helper()

	message := readServerMessage(t, conn, time.Second)
	if message.Type != messageType {
		t.Fatalf("expected %q frame, got %q", messageType, message.Type)
	}

	data, ok := message.Data.(map[string]any)
	if !ok {
		t.Fatalf("expected ack payload object, got %T", message.Data)
	}
	if data["post_id"] != float64(postID) {
		t.Fatalf("expected post_id %d, got %v", postID, data["post_id"])
	}
}

// subscribeAndWait подтверждает подписку: сначала приходит кадр "subscribed",
// затем pong на отправленный ping (кадры обрабатываются строго по порядку).
func subscribeAndWait(t *testing.T, conn *websocket.Conn, postID int) {
	t.Helper()

	writeClientMessage(t, conn, MessageTypeSubscribe, SubscribePayload{PostID: postID})
	writeClientMessage(t, conn, MessageTypePing, nil)

	expectAck(t, conn, MessageTypeSubscribed, postID)

	if message := readServerMessage(t, conn, time.Second); message.Type != MessageTypePong {
		t.Fatalf("expected %q frame, got %q", MessageTypePong, message.Type)
	}
}

func expectNoFrame(t *testing.T, conn *websocket.Conn, timeout time.Duration) {
	t.Helper()

	if err := conn.SetReadDeadline(time.Now().Add(timeout)); err != nil {
		t.Fatalf("failed to set read deadline: %v", err)
	}
	if _, payload, err := conn.ReadMessage(); err == nil {
		t.Fatalf("expected no frame, got: %s", payload)
	}
}

func TestHandle_SubscribeReceivesPublishedEvent(t *testing.T) {
	fixture := newWSFixture(t, tokenParserStub{userID: testUserID}, testConfig())
	conn := fixture.dial(t, testOrigin)

	authAndReady(t, conn)
	subscribeAndWait(t, conn, 100)

	fixture.hub.Publish(
		core_realtime.PostTopic(100),
		core_realtime.NewEvent(core_realtime.EventCommentCreated, map[string]any{"comment_id": 7}),
	)

	envelope := readEnvelope(t, conn, time.Second)
	if envelope.Type != core_realtime.EventCommentCreated {
		t.Fatalf("expected event %q, got %q", core_realtime.EventCommentCreated, envelope.Type)
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
	if data["comment_id"] != float64(7) {
		t.Fatalf("unexpected payload: %+v", data)
	}
}

func TestHandle_InvalidTokenReceivesErrorBeforeClose(t *testing.T) {
	fixture := newWSFixture(t, tokenParserStub{err: core_errors.ErrInvalidToken}, testConfig())
	conn := fixture.dial(t, testOrigin)

	writeClientMessage(t, conn, MessageTypeAuth, AuthPayload{Token: testToken})

	// Кадр с причиной отказа должен дойти до закрытия соединения:
	// writePump дочищает очередь перед закрытием.
	message := readServerMessage(t, conn, time.Second)
	if message.Type != MessageTypeError {
		t.Fatalf("expected %q frame, got %q", MessageTypeError, message.Type)
	}

	data, ok := message.Data.(map[string]any)
	if !ok {
		t.Fatalf("expected error payload object, got %T", message.Data)
	}
	if data["code"] != "unauthorized" {
		t.Fatalf("expected code 'unauthorized', got %v", data["code"])
	}
}

func TestHandle_FirstMessageNotAuth(t *testing.T) {
	fixture := newWSFixture(t, tokenParserStub{userID: testUserID}, testConfig())
	conn := fixture.dial(t, testOrigin)

	writeClientMessage(t, conn, MessageTypeSubscribe, SubscribePayload{PostID: 100})

	message := readServerMessage(t, conn, time.Second)
	if message.Type != MessageTypeError {
		t.Fatalf("expected %q frame, got %q", MessageTypeError, message.Type)
	}

	data, ok := message.Data.(map[string]any)
	if !ok {
		t.Fatalf("expected error payload object, got %T", message.Data)
	}
	if data["code"] != "unauthorized" {
		t.Fatalf("expected code 'unauthorized', got %v", data["code"])
	}
}

func TestHandle_AuthTimeout(t *testing.T) {
	config := testConfig()
	config.AuthWait = 100 * time.Millisecond

	fixture := newWSFixture(t, tokenParserStub{userID: testUserID}, config)
	conn := fixture.dial(t, testOrigin)

	message := readServerMessage(t, conn, 2*time.Second)
	if message.Type != MessageTypeError {
		t.Fatalf("expected %q frame, got %q", MessageTypeError, message.Type)
	}
}

func TestHandle_RepeatedAuthRejected(t *testing.T) {
	fixture := newWSFixture(t, tokenParserStub{userID: testUserID}, testConfig())
	conn := fixture.dial(t, testOrigin)

	authAndReady(t, conn)

	writeClientMessage(t, conn, MessageTypeAuth, AuthPayload{Token: testToken})

	message := readServerMessage(t, conn, time.Second)
	if message.Type != MessageTypeError {
		t.Fatalf("expected %q frame, got %q", MessageTypeError, message.Type)
	}

	data, ok := message.Data.(map[string]any)
	if !ok {
		t.Fatalf("expected error payload object, got %T", message.Data)
	}
	if data["message"] != "already authenticated" {
		t.Fatalf("unexpected error message: %v", data["message"])
	}
}

func TestHandle_OriginNotAllowed(t *testing.T) {
	fixture := newWSFixture(t, tokenParserStub{userID: testUserID}, testConfig())

	if status := fixture.dialStatus(t, "http://evil.example.com"); status != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d", http.StatusForbidden, status)
	}
}

func TestHandle_PingPong(t *testing.T) {
	fixture := newWSFixture(t, tokenParserStub{userID: testUserID}, testConfig())
	conn := fixture.dial(t, testOrigin)

	authAndReady(t, conn)

	writeClientMessage(t, conn, MessageTypePing, nil)

	message := readServerMessage(t, conn, time.Second)
	if message.Type != MessageTypePong {
		t.Fatalf("expected %q frame, got %q", MessageTypePong, message.Type)
	}
}

func TestHandle_UnsubscribeStopsEvents(t *testing.T) {
	fixture := newWSFixture(t, tokenParserStub{userID: testUserID}, testConfig())
	conn := fixture.dial(t, testOrigin)

	authAndReady(t, conn)
	subscribeAndWait(t, conn, 100)

	fixture.hub.Publish(core_realtime.PostTopic(100), core_realtime.NewEvent("test.event", nil))
	if envelope := readEnvelope(t, conn, time.Second); envelope.Type != "test.event" {
		t.Fatalf("expected event %q, got %q", "test.event", envelope.Type)
	}

	writeClientMessage(t, conn, MessageTypeUnsubscribe, SubscribePayload{PostID: 100})
	writeClientMessage(t, conn, MessageTypePing, nil)
	expectAck(t, conn, MessageTypeUnsubscribed, 100)
	if message := readServerMessage(t, conn, time.Second); message.Type != MessageTypePong {
		t.Fatalf("expected %q frame, got %q", MessageTypePong, message.Type)
	}

	fixture.hub.Publish(core_realtime.PostTopic(100), core_realtime.NewEvent("test.event", nil))

	expectNoFrame(t, conn, 300*time.Millisecond)
}

func TestHandle_SubscribeSendsAck(t *testing.T) {
	fixture := newWSFixture(t, tokenParserStub{userID: testUserID}, testConfig())
	conn := fixture.dial(t, testOrigin)

	authAndReady(t, conn)

	writeClientMessage(t, conn, MessageTypeSubscribe, SubscribePayload{PostID: 100})
	expectAck(t, conn, MessageTypeSubscribed, 100)

	fixture.hub.Publish(core_realtime.PostTopic(100), core_realtime.NewEvent("test.event", nil))
	if envelope := readEnvelope(t, conn, time.Second); envelope.Type != "test.event" {
		t.Fatalf("expected event %q, got %q", "test.event", envelope.Type)
	}
}

func TestHandle_SubscribeInvalidPayloadSendsErrorOnly(t *testing.T) {
	fixture := newWSFixture(t, tokenParserStub{userID: testUserID}, testConfig())
	conn := fixture.dial(t, testOrigin)

	authAndReady(t, conn)

	// post_id = 0 не проходит Validate: приходит только server.error, ack быть не должно.
	writeClientMessage(t, conn, MessageTypeSubscribe, SubscribePayload{PostID: 0})
	writeClientMessage(t, conn, MessageTypePing, nil)

	message := readServerMessage(t, conn, time.Second)
	if message.Type != MessageTypeError {
		t.Fatalf("expected %q frame, got %q", MessageTypeError, message.Type)
	}

	data, ok := message.Data.(map[string]any)
	if !ok {
		t.Fatalf("expected error payload object, got %T", message.Data)
	}
	if data["code"] != "invalid_argument" {
		t.Fatalf("expected code 'invalid_argument', got %v", data["code"])
	}

	if message := readServerMessage(t, conn, time.Second); message.Type != MessageTypePong {
		t.Fatalf("expected %q frame, got %q", MessageTypePong, message.Type)
	}

	// Подписка не применилась: события по этому топику не приходят.
	fixture.hub.Publish(core_realtime.PostTopic(0), core_realtime.NewEvent("test.event", nil))

	expectNoFrame(t, conn, 300*time.Millisecond)
}

func TestHandle_UnsubscribeSendsAck(t *testing.T) {
	fixture := newWSFixture(t, tokenParserStub{userID: testUserID}, testConfig())
	conn := fixture.dial(t, testOrigin)

	authAndReady(t, conn)
	subscribeAndWait(t, conn, 100)

	writeClientMessage(t, conn, MessageTypeUnsubscribe, SubscribePayload{PostID: 100})
	writeClientMessage(t, conn, MessageTypePing, nil)

	expectAck(t, conn, MessageTypeUnsubscribed, 100)
	if message := readServerMessage(t, conn, time.Second); message.Type != MessageTypePong {
		t.Fatalf("expected %q frame, got %q", MessageTypePong, message.Type)
	}
}

func TestHandle_RepeatedSubscribeDeliversEventOnce(t *testing.T) {
	fixture := newWSFixture(t, tokenParserStub{userID: testUserID}, testConfig())
	conn := fixture.dial(t, testOrigin)

	authAndReady(t, conn)
	subscribeAndWait(t, conn, 100)

	// Повторная подписка идемпотентна: hub хранит подписчиков множеством,
	// но ack приходит на каждый subscribe.
	writeClientMessage(t, conn, MessageTypeSubscribe, SubscribePayload{PostID: 100})
	expectAck(t, conn, MessageTypeSubscribed, 100)

	fixture.hub.Publish(core_realtime.PostTopic(100), core_realtime.NewEvent("test.event", nil))

	if envelope := readEnvelope(t, conn, time.Second); envelope.Type != "test.event" {
		t.Fatalf("expected event %q, got %q", "test.event", envelope.Type)
	}

	expectNoFrame(t, conn, 300*time.Millisecond)
}

func TestHandle_MultipleClientsSameTopic(t *testing.T) {
	fixture := newWSFixture(t, tokenParserStub{userID: testUserID}, testConfig())

	first := fixture.dial(t, testOrigin)
	second := fixture.dial(t, testOrigin)

	authAndReady(t, first)
	authAndReady(t, second)
	subscribeAndWait(t, first, 100)
	subscribeAndWait(t, second, 100)

	fixture.hub.Publish(core_realtime.PostTopic(100), core_realtime.NewEvent("test.event", nil))

	for _, conn := range []*websocket.Conn{first, second} {
		if envelope := readEnvelope(t, conn, time.Second); envelope.Type != "test.event" {
			t.Fatalf("expected event %q, got %q", "test.event", envelope.Type)
		}
	}
}

func TestHandle_OversizedFrameDropsConnection(t *testing.T) {
	config := testConfig()
	config.ReadLimit = 256

	fixture := newWSFixture(t, tokenParserStub{userID: testUserID}, config)
	conn := fixture.dial(t, testOrigin)

	authAndReady(t, conn)

	writeClientMessage(t, conn, MessageTypeSubscribe, SubscribePayload{PostID: 100})
	if err := conn.WriteMessage(websocket.TextMessage, make([]byte, 1024)); err != nil {
		t.Fatalf("failed to write oversized frame: %v", err)
	}

	if err := conn.SetReadDeadline(time.Now().Add(2 * time.Second)); err != nil {
		t.Fatalf("failed to set read deadline: %v", err)
	}
	// До обрыва успевает прийти ack подписки — читаем кадры, пока соединение не закроется.
	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			var netErr net.Error
			if errors.As(err, &netErr) && netErr.Timeout() {
				t.Fatal("connection was not dropped after oversized frame (read deadline exceeded)")
			}
			return
		}
	}
}

func TestOriginChecker(t *testing.T) {
	checker := originChecker([]string{testOrigin, "https://debates.app"})

	tests := []struct {
		name   string
		origin string
		want   bool
	}{
		{name: "empty origin is allowed", origin: "", want: true},
		{name: "allowed origin", origin: testOrigin, want: true},
		{name: "another allowed origin", origin: "https://debates.app", want: true},
		{name: "unknown origin", origin: "http://evil.example.com", want: false},
		{name: "same host different scheme", origin: "https://localhost:3000", want: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, "/api/v1/ws", nil)
			if test.origin != "" {
				request.Header.Set("Origin", test.origin)
			}

			if got := checker(request); got != test.want {
				t.Fatalf("expected %t, got %t", test.want, got)
			}
		})
	}
}

func TestSubscribePayload_Validate(t *testing.T) {
	tests := []struct {
		name    string
		payload SubscribePayload
		wantErr bool
	}{
		{name: "valid", payload: SubscribePayload{PostID: 1}, wantErr: false},
		{name: "zero", payload: SubscribePayload{PostID: 0}, wantErr: true},
		{name: "negative", payload: SubscribePayload{PostID: -5}, wantErr: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.payload.Validate()
			if test.wantErr && !errors.Is(err, core_errors.ErrInvalidArgument) {
				t.Fatalf("expected ErrInvalidArgument, got: %v", err)
			}
			if !test.wantErr && err != nil {
				t.Fatalf("expected no error, got: %v", err)
			}
		})
	}
}
