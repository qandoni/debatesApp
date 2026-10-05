package core_realtime

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

type Peer interface {
	Send(payload []byte)
}

type Publisher interface {
	Publish(topic string, event Event) error
}

func NewHub() *Hub {
	return &Hub{
		mtx:    &sync.RWMutex{},
		topics: make(map[string]map[Peer]struct{}),
	}
}

// Hub хранит все подписки в единственной map «топик → подписчики»: она
// является единственным источником истины, поэтому вторая обратная map и
// ручная синхронизация между ними не нужны. Пустые топики удаляются при
// отписке, поэтому UnsubscribeAll обходит только активные топики.
type Hub struct {
	mtx    *sync.RWMutex
	topics map[string]map[Peer]struct{}
}

func (h *Hub) Subscribe(topic string, peer Peer) {
	h.mtx.Lock()
	defer h.mtx.Unlock()

	subscribers, ok := h.topics[topic]
	if !ok {
		subscribers = make(map[Peer]struct{})
		h.topics[topic] = subscribers
	}
	subscribers[peer] = struct{}{}
}

func (h *Hub) Unsubscribe(topic string, peer Peer) {
	h.mtx.Lock()
	defer h.mtx.Unlock()

	subscribers, ok := h.topics[topic]
	if !ok {
		return
	}
	delete(subscribers, peer)
	if len(subscribers) == 0 {
		delete(h.topics, topic)
	}
}

func (h *Hub) UnsubscribeAll(peer Peer) {
	h.mtx.Lock()
	defer h.mtx.Unlock()

	for topic, subscribers := range h.topics {
		delete(subscribers, peer)
		if len(subscribers) == 0 {
			delete(h.topics, topic)
		}
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

	// Снапшот подписчиков снимается под RLock, а отправка идёт уже без
	// блокировки: Peer.Send может обращаться к Hub (например, отписываться
	// при переполнении буфера), удержание RLock здесь приводило к само-дедлоку.
	h.mtx.RLock()
	subscribers := make([]Peer, 0, len(h.topics[topic]))
	for peer := range h.topics[topic] {
		subscribers = append(subscribers, peer)
	}
	h.mtx.RUnlock()

	for _, peer := range subscribers {
		peer.Send(payload)
	}
	return nil
}
