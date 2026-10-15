package core_realtime

import "sync"

type syncMap struct {
	entries sync.Map
}

type mapEntry struct {
	mtx   sync.RWMutex
	items map[Peer]struct{}
}

func newMapEntry() *mapEntry {
	return &mapEntry{items: make(map[Peer]struct{})}
}

func (s *syncMap) Add(key string, item Peer) {
	for {
		value, _ := s.entries.LoadOrStore(key, newMapEntry())
		entry := value.(*mapEntry)

		entry.mtx.Lock()

		if current, ok := s.entries.Load(key); !ok || current != entry {
			entry.mtx.Unlock()
			continue
		}
		entry.items[item] = struct{}{}
		entry.mtx.Unlock()
		return
	}
}

func (s *syncMap) Remove(key string, item Peer) {
	value, ok := s.entries.Load(key)
	if !ok {
		return
	}
	entry := value.(*mapEntry)

	entry.mtx.Lock()
	defer entry.mtx.Unlock()

	s.removeLocked(key, entry, item)
}

func (s *syncMap) RemoveFromAll(item Peer) {
	s.entries.Range(func(key, value any) bool {
		entry := value.(*mapEntry)

		entry.mtx.Lock()
		s.removeLocked(key.(string), entry, item)
		entry.mtx.Unlock()
		return true
	})
}

func (s *syncMap) removeLocked(key string, entry *mapEntry, item Peer) {
	delete(entry.items, item)
	if len(entry.items) == 0 {
		s.entries.CompareAndDelete(key, entry)
	}
}

func (s *syncMap) Snapshot(key string) []Peer {
	value, ok := s.entries.Load(key)
	if !ok {
		return nil
	}
	entry := value.(*mapEntry)

	entry.mtx.RLock()
	defer entry.mtx.RUnlock()

	items := make([]Peer, 0, len(entry.items))
	for item := range entry.items {
		items = append(items, item)
	}
	return items
}

func (s *syncMap) All() []Peer {
	seen := make(map[Peer]struct{})

	s.entries.Range(func(_, value any) bool {
		entry := value.(*mapEntry)
		entry.mtx.RLock()
		for item := range entry.items {
			seen[item] = struct{}{}
		}
		entry.mtx.RUnlock()
		return true
	})

	items := make([]Peer, 0, len(seen))
	for item := range seen {
		items = append(items, item)
	}
	return items
}
