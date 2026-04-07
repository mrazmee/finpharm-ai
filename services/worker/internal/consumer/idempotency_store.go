package consumer

import "sync"

type InMemoryProcessedStore struct {
	mu   sync.RWMutex
	data map[string]struct{}
}

func NewInMemoryProcessedStore() *InMemoryProcessedStore {
	return &InMemoryProcessedStore{
		data: make(map[string]struct{}),
	}
}

func (s *InMemoryProcessedStore) Exists(key string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, ok := s.data[key]
	return ok
}

func (s *InMemoryProcessedStore) Mark(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key] = struct{}{}
}