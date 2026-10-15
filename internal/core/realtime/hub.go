package core_realtime

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

type Peer interface {
	Send(payload []byte)
	Shutdown()
}

type Publisher interface {
	Publish(topic string, event Event) error
}

func NewHub() *Hub {
	return &Hub{}
}

type Hub struct {
	peers syncMap
}

func (h *Hub) Subscribe(topic string, peer Peer) {
	h.peers.Add(topic, peer)
}

func (h *Hub) Unsubscribe(topic string, peer Peer) {
	h.peers.Remove(topic, peer)
}

func (h *Hub) UnsubscribeAll(peer Peer) {
	h.peers.RemoveFromAll(peer)
}

func (h *Hub) Shutdown() {
	for _, peer := range h.peers.All() {
		peer.Shutdown()
	}
}

func (h *Hub) Publish(topic string, event Event) error {
	envelope := Envelope{
		Type:       event.Type,
		Topic:      topic,
		Data:       event.Data,
		OccurredAt: time.Now(),
	}

	payload, err := json.Marshal(envelope)
	if err != nil {
		return fmt.Errorf("marshal event envelope: %w", err)
	}

	peers := h.peers.Snapshot(topic)

	var wg sync.WaitGroup
	wg.Add(len(peers))
	for _, peer := range peers {
		go func(peer Peer) {
			defer wg.Done()
			peer.Send(payload)
		}(peer)
	}
	wg.Wait()
	return nil
}
