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

type peerSet struct {
	mtx   sync.RWMutex
	peers map[Peer]struct{}
}

func newPeerSet() *peerSet {
	return &peerSet{peers: make(map[Peer]struct{})}
}

func NewHub() *Hub {
	return &Hub{}
}

type Hub struct {
	topics sync.Map
}

func (h *Hub) Subscribe(topic string, peer Peer) {
	for {
		value, _ := h.topics.LoadOrStore(topic, newPeerSet())
		set := value.(*peerSet)

		set.mtx.Lock()
		if current, ok := h.topics.Load(topic); !ok || current != set {
			set.mtx.Unlock()
			continue
		}
		set.peers[peer] = struct{}{}
		set.mtx.Unlock()
		return
	}
}

func (h *Hub) Unsubscribe(topic string, peer Peer) {
	value, ok := h.topics.Load(topic)
	if !ok {
		return
	}
	set := value.(*peerSet)

	set.mtx.Lock()
	defer set.mtx.Unlock()

	h.removePeerLocked(topic, set, peer)
}

func (h *Hub) removePeerLocked(topic string, set *peerSet, peer Peer) {
	delete(set.peers, peer)
	if len(set.peers) == 0 {
		h.topics.CompareAndDelete(topic, set)
	}
}

func (h *Hub) UnsubscribeAll(peer Peer) {
	h.topics.Range(func(key, value any) bool {
		set := value.(*peerSet)

		set.mtx.Lock()
		h.removePeerLocked(key.(string), set, peer)
		set.mtx.Unlock()
		return true
	})
}

func (h *Hub) subscribers(topic string) []Peer {
	value, ok := h.topics.Load(topic)
	if !ok {
		return nil
	}
	set := value.(*peerSet)

	set.mtx.RLock()
	defer set.mtx.RUnlock()

	peers := make([]Peer, 0, len(set.peers))
	for peer := range set.peers {
		peers = append(peers, peer)
	}
	return peers
}

func (h *Hub) allPeers() []Peer {
	seen := make(map[Peer]struct{})

	h.topics.Range(func(_, value any) bool {
		set := value.(*peerSet)
		set.mtx.RLock()
		for peer := range set.peers {
			seen[peer] = struct{}{}
		}
		set.mtx.RUnlock()
		return true
	})

	peers := make([]Peer, 0, len(seen))
	for peer := range seen {
		peers = append(peers, peer)
	}
	return peers
}

func (h *Hub) Shutdown() {
	for _, peer := range h.allPeers() {
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

	peers := h.subscribers(topic)

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
