package externalcli

import (
	"context"
	"sync"
	"time"
)

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
