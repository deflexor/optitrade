package dashboard

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"
)

type previewPayload struct {
	kind    string
	expires time.Time
	data    map[string]any
}

type previewStore struct {
	mu sync.Mutex
	m  map[string]previewPayload
}

func newPreviewStore() *previewStore {
	return &previewStore{m: make(map[string]previewPayload)}
}

func (s *previewStore) issue(kind string, data map[string]any, ttl time.Duration) string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	tok := hex.EncodeToString(b)
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cleanupLocked()
	s.m[tok] = previewPayload{kind: kind, expires: time.Now().Add(ttl), data: data}
	return tok
}

func (s *previewStore) take(kind, token string) (map[string]any, bool) {
	if token == "" {
		return nil, false
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	p, ok := s.m[token]
	if !ok || p.kind != kind || time.Now().After(p.expires) {
		return nil, false
	}
	delete(s.m, token)
	return p.data, true
}

func (s *previewStore) cleanupLocked() {
	now := time.Now()
	for k, v := range s.m {
		if now.After(v.expires) {
			delete(s.m, k)
		}
	}
}
