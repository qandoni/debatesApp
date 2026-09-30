package core_transport_websocket

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	core_auth "github.com/qandoni/debatesApp/internal/core/auth"
	core_errors "github.com/qandoni/debatesApp/internal/core/errors"
	core_logger "github.com/qandoni/debatesApp/internal/core/logger"
	core_realtime "github.com/qandoni/debatesApp/internal/core/realtime"
	"go.uber.org/zap"
)

type Hub interface {
	Subscribe(topic string, peer core_realtime.Peer)
	Unsubscribe(topic string, peer core_realtime.Peer)
	UnsubscribeAll(peer core_realtime.Peer)
}

type TokenParser interface {
	ParseAccessToken(token string) (core_auth.AuthInfo, error)
}

func NewClient(
	conn *websocket.Conn,
	hub Hub,
	parser TokenParser,
	config Config,
	log *core_logger.Logger,
) *Client {
	conn.SetReadLimit(config.ReadLimit)

	return &Client{
		conn:   conn,
		hub:    hub,
		parser: parser,
		config: config,
		log:    log,
		send:   make(chan []byte, config.SendBufferSize),
		done:   make(chan struct{}),
		once:   new(sync.Once),
	}
}

type Client struct {
	conn   *websocket.Conn
	hub    Hub
	parser TokenParser
	config Config
	log    *core_logger.Logger

	send   chan []byte
	done   chan struct{}
	once   *sync.Once
	userID int
}

func (c *Client) Send(payload []byte) {
	select {
	case <-c.done:
		return
	default:
	}
	select {
	case c.send <- payload:
	default:
		c.log.Warn("websocket send buffer overflow, dropping client",
			zap.Int("send_buffer_size", c.config.SendBufferSize))
		c.close()
	}
}

func (c *Client) close() {
	c.once.Do(func() {
		close(c.done)
		c.hub.UnsubscribeAll(c)
	})
}

func (c *Client) Run() {
	defer c.close()
	go c.writePump()
	if err := c.authenticate(); err != nil {
		c.log.Warn("websocket authentication failed", zap.Error(err))
		c.Send(errorFrame("unauthorized", "authentication failed"))
		return
	}
	c.readPump()
}

func (c *Client) authenticate() error {
	if err := c.conn.SetReadDeadline(time.Now().Add(c.config.AuthWait)); err != nil {
		return fmt.Errorf("set read deadline: %w", err)
	}
	var message ClientMessage
	if err := c.conn.ReadJSON(&message); err != nil {
		return fmt.Errorf("read auth message: %w", err)
	}
	if message.Type != MessageTypeAuth {
		return fmt.Errorf("first message must be %q: %w", MessageTypeAuth, core_errors.ErrUnauthorized)
	}
	var payload AuthPayload
	if err := json.Unmarshal(message.Data, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal json: %w", err)
	}
	if err := payload.Validate(); err != nil {
		return err
	}
	authInfo, err := c.parser.ParseAccessToken(payload.Token)
	if err != nil {
		return fmt.Errorf("parse access token: %w", err)
	}
	c.userID = authInfo.UserID

	c.conn.SetPongHandler(func(string) error {
		return c.conn.SetReadDeadline(time.Now().Add(c.config.PongWait))
	})
	if err := c.conn.SetReadDeadline(time.Now().Add(c.config.PongWait)); err != nil {
		return fmt.Errorf("set read deadline: %w", err)
	}
	c.Send(mustMarshalServerMessage(MessageTypeReady, map[string]int{"user_id": c.userID}))
	return nil
}

func (c *Client) readPump() {
	for {
		var message ClientMessage
		if err := c.conn.ReadJSON(&message); err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
				c.log.Debug("websocket read error", zap.Error(err))
			}
			return
		}

		if err := c.conn.SetReadDeadline(time.Now().Add(c.config.PongWait)); err != nil {
			return
		}

		switch message.Type {
		case MessageTypeSubscribe:
			c.handleSubscribe(message.Data)
		case MessageTypeUnsubscribe:
			c.handleUnsubscribe(message.Data)
		case MessageTypePing:
			c.Send(mustMarshalServerMessage(MessageTypePong, nil))
		case MessageTypeAuth:
			c.Send(errorFrame("invalid_argument", "already authenticated"))
		default:
			c.Send(errorFrame("invalid_argument", "unknown message type"))
		}
	}
}

func (c *Client) handleSubscribe(raw json.RawMessage) {
	var payload SubscribePayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		c.Send(errorFrame("invalid_argument", "invalid subscribe payload"))
		return
	}
	if err := payload.Validate(); err != nil {
		c.Send(errorFrame("invalid_argument", err.Error()))
		return
	}

	c.hub.Subscribe(core_realtime.PostTopic(payload.PostID), c)
}

func mustMarshalServerMessage(messageType string, data any) []byte {
	payload, err := json.Marshal(ServerMessage{Type: messageType, Data: data})
	if err != nil {
		payload, _ = json.Marshal(ServerMessage{
			Type: MessageTypeError,
			Data: ServerErrorData{Code: "internal", Message: "failed to marshal server message"},
		})
	}
	return payload
}

func (c *Client) handleUnsubscribe(raw json.RawMessage) {
	var payload SubscribePayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		c.Send(errorFrame("invalid_argument", "invalid unsubscribe payload"))
		return
	}
	if err := payload.Validate(); err != nil {
		c.Send(errorFrame("invalid_argument", err.Error()))
		return
	}
	c.hub.Unsubscribe(core_realtime.PostTopic(payload.PostID), c)
}

func errorFrame(code, message string) []byte {
	return mustMarshalServerMessage(MessageTypeError, ServerErrorData{
		Code:    code,
		Message: message,
	})
}

func (c *Client) writePump() {
	ticker := time.NewTicker(c.config.PingPeriod)
	defer func() {
		ticker.Stop()
		_ = c.conn.WriteControl(
			websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""),
			time.Now().Add(c.config.WriteWait),
		)
		_ = c.conn.Close()
		c.close()
	}()

	for {
		select {
		case payload := <-c.send:
			if err := c.writeFrame(websocket.TextMessage, payload); err != nil {
				return
			}
		case <-ticker.C:
			if err := c.writeFrame(websocket.PingMessage, nil); err != nil {
				return
			}
		case <-c.done:
			c.drain()
			return
		}
	}
}

func (c *Client) writeFrame(messageType int, payload []byte) error {
	if err := c.conn.SetWriteDeadline(time.Now().Add(c.config.WriteWait)); err != nil {
		return err
	}
	return c.conn.WriteMessage(messageType, payload)
}

func (c *Client) drain() {
	deadline := time.Now().Add(c.config.WriteWait)
	for {
		select {
		case payload := <-c.send:
			if deadline.Before(time.Now()) {
				return
			}
			if err := c.writeFrame(websocket.TextMessage, payload); err != nil {
				return
			}
		default:
			return
		}
	}
}
