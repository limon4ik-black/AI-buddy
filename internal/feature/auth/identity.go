package auth

import "sync"

type UserStore interface {
	FindOrCreate(provider, providerUserID, email, name string) int64
}

// In-memory для разработки/демо: пропадает при рестарте.
// Позже заменяется на Mongo/SQL за тем же интерфейсом.
type MemoryStore struct {
	mu         sync.Mutex
	nextID     int64
	byProvider map[string]int64
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{byProvider: make(map[string]int64)}
}

func (s *MemoryStore) FindOrCreate(provider, providerUserID, _, _ string) int64 {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := provider + "|" + providerUserID
	if id, ok := s.byProvider[key]; ok {
		return id
	}
	s.nextID++
	s.byProvider[key] = s.nextID
	return s.nextID
}
