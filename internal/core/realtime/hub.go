package core_realtime

import (
	"encoding/json"
	"sync"
	"time"
)

type Peer interface {
	Send(payload []byte)
}

type Publisher interface {
	Publish(topic string, event Event)
}

func NewHub() *Hub {
	return &Hub{
		mtx:        &sync.RWMutex{},
		topics:     make(map[string]map[Peer]struct{}),
		peerTopics: make(map[Peer]map[string]struct{}),
	}
}

type Hub struct {
	mtx        *sync.RWMutex
	topics     map[string]map[Peer]struct{}
	peerTopics map[Peer]map[string]struct{}
}

// Компиляторная гарантия: сигнатура Publish у Hub не разъедется с интерфейсом.
var _ Publisher = (*Hub)(nil)

func (h *Hub) Subscribe(topic string, peer Peer) {
	h.mtx.Lock()
	defer h.mtx.Unlock()

	subscribers, ok := h.topics[topic]
	if !ok {
		subscribers = make(map[Peer]struct{})
		h.topics[topic] = subscribers
	}
	subscribers[peer] = struct{}{}

	topics, ok := h.peerTopics[peer]
	if !ok {
		topics = make(map[string]struct{})
		h.peerTopics[peer] = topics
	}
	topics[topic] = struct{}{}
}

func (h *Hub) Unsubscribe(topic string, peer Peer) {
	h.mtx.Lock()
	defer h.mtx.Unlock()

	if subscribers, ok := h.topics[topic]; ok {
		delete(subscribers, peer)
		if len(subscribers) == 0 {
			delete(h.topics, topic)
		}
	}
	h.removePeerTopicLocked(peer, topic)
}

func (h *Hub) removePeerTopicLocked(peer Peer, topic string) {
	topics, ok := h.peerTopics[peer]
	if !ok {
		return
	}
	delete(topics, topic)
	if len(topics) == 0 {
		delete(h.peerTopics, peer)
	}
}

func (h *Hub) UnsubscribeAll(peer Peer) {
	h.mtx.Lock()
	defer h.mtx.Unlock()

	for topic := range h.peerTopics[peer] {
		subscribers, ok := h.topics[topic]
		if !ok {
			continue
		}
		delete(subscribers, peer)
		if len(subscribers) == 0 {
			delete(h.topics, topic)
		}
	}
	delete(h.peerTopics, peer)
}

func (h *Hub) Publish(topic string, event Event) {
	envelope := Envelope{
		Type:       event.Type,
		Topic:      topic,
		Data:       event.Data,
		OccurredAt: time.Now(),
	}

	payload, err := json.Marshal(envelope)
	if err != nil {
		return
	}
	h.mtx.RLock()
	subscribers := make([]Peer, 0, len(h.topics[topic]))
	for peer := range h.topics[topic] {
		subscribers = append(subscribers, peer)
	}
	h.mtx.RUnlock()

	for _, peer := range subscribers {
		peer.Send(payload)
	}
}
