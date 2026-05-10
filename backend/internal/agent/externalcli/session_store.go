package externalcli

import (
	"context"
	"sync"
	"time"
)

const (
	providerSessionStatusActive = "active"
	providerSessionStatusFailed = "failed"
)

// ProviderSessionKey identifies one external CLI session binding.
type ProviderSessionKey struct {
	InternalSessionID string
	Provider          string
	WorkDir           string
	AssistantID       string
}

// ProviderSessionBinding maps a SingerOS session to a provider-native CLI session.
type ProviderSessionBinding struct {
	InternalSessionID string
	Provider          string
	ProviderSessionID string
	WorkDir           string
	AssistantID       string
	Status            string
	LastError         string
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

// ProviderSessionStore persists provider session bindings for external CLI resumes.
type ProviderSessionStore interface {
	Get(ctx context.Context, key ProviderSessionKey) (*ProviderSessionBinding, error)
	Upsert(ctx context.Context, binding *ProviderSessionBinding) error
	MarkFailed(ctx context.Context, key ProviderSessionKey, reason string) error
}

var (
	defaultProviderSessionStoreMu sync.RWMutex
	defaultProviderSessionStore   ProviderSessionStore = NewInMemoryProviderSessionStore()
)

// DefaultProviderSessionStore returns the package-level provider session store.
func DefaultProviderSessionStore() ProviderSessionStore {
	defaultProviderSessionStoreMu.RLock()
	defer defaultProviderSessionStoreMu.RUnlock()
	return defaultProviderSessionStore
}

// SetDefaultProviderSessionStore replaces the package-level provider session store.
func SetDefaultProviderSessionStore(store ProviderSessionStore) {
	if store == nil {
		return
	}
	defaultProviderSessionStoreMu.Lock()
	defer defaultProviderSessionStoreMu.Unlock()
	defaultProviderSessionStore = store
}

// InMemoryProviderSessionStore stores provider session bindings in process memory.
type InMemoryProviderSessionStore struct {
	mu       sync.RWMutex
	bindings map[ProviderSessionKey]*ProviderSessionBinding
}

// NewInMemoryProviderSessionStore creates an in-memory provider session store.
func NewInMemoryProviderSessionStore() *InMemoryProviderSessionStore {
	return &InMemoryProviderSessionStore{
		bindings: make(map[ProviderSessionKey]*ProviderSessionBinding),
	}
}

// Get returns a provider session binding for the key.
func (s *InMemoryProviderSessionStore) Get(_ context.Context, key ProviderSessionKey) (*ProviderSessionBinding, error) {
	if s == nil {
		return nil, nil
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	binding, ok := s.bindings[key]
	if !ok || binding == nil {
		return nil, nil
	}
	cloned := *binding
	return &cloned, nil
}

// Upsert creates or replaces a provider session binding.
func (s *InMemoryProviderSessionStore) Upsert(_ context.Context, binding *ProviderSessionBinding) error {
	if s == nil || binding == nil {
		return nil
	}
	key := ProviderSessionKey{
		InternalSessionID: binding.InternalSessionID,
		Provider:          binding.Provider,
		WorkDir:           binding.WorkDir,
		AssistantID:       binding.AssistantID,
	}
	if key.InternalSessionID == "" || key.Provider == "" || binding.ProviderSessionID == "" {
		return nil
	}

	now := time.Now().UTC()
	s.mu.Lock()
	defer s.mu.Unlock()
	if existing, ok := s.bindings[key]; ok && existing != nil && !existing.CreatedAt.IsZero() {
		binding.CreatedAt = existing.CreatedAt
	} else if binding.CreatedAt.IsZero() {
		binding.CreatedAt = now
	}
	binding.UpdatedAt = now
	cloned := *binding
	s.bindings[key] = &cloned
	return nil
}

// MarkFailed marks a provider session binding as failed.
func (s *InMemoryProviderSessionStore) MarkFailed(_ context.Context, key ProviderSessionKey, reason string) error {
	if s == nil {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	binding, ok := s.bindings[key]
	if !ok || binding == nil {
		return nil
	}
	cloned := *binding
	cloned.Status = providerSessionStatusFailed
	cloned.LastError = reason
	cloned.UpdatedAt = time.Now().UTC()
	s.bindings[key] = &cloned
	return nil
}
