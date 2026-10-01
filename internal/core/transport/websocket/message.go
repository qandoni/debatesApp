package core_transport_websocket

import (
	"encoding/json"
	"fmt"

	core_errors "github.com/qandoni/debatesApp/internal/core/errors"
)

const (
	MessageTypeAuth        = "auth"
	MessageTypeSubscribe   = "subscribe"
	MessageTypeUnsubscribe = "unsubscribe"
	MessageTypePing        = "ping"

	MessageTypeReady        = "ready"
	MessageTypePong         = "pong"
	MessageTypeSubscribed   = "subscribed"
	MessageTypeUnsubscribed = "unsubscribed"
	MessageTypeError        = "server.error"
)

type ClientMessage struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

type AuthPayload struct {
	Token string `json:"token"`
}

func (p AuthPayload) Validate() error {
	if p.Token == "" {
		return fmt.Errorf("AuthPayload token is NULL: %w", core_errors.ErrUnauthorized)
	}
	return nil
}

type SubscribePayload struct {
	PostID int `json:"post_id"`
}

func (p SubscribePayload) Validate() error {
	if p.PostID <= 0 {
		return fmt.Errorf("'PostID' cannot be <=0: %w", core_errors.ErrInvalidArgument)
	}
	return nil
}

type SubscriptionAckData struct {
	PostID int `json:"post_id"`
}

type ServerErrorData struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type ServerMessage struct {
	Type string `json:"type"`
	Data any    `json:"data,omitempty"`
}
