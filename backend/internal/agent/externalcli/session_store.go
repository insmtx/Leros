package externalcli

import (
	"context"
	"sync"
	"time"
)

const (
	externalCLISessionsMetadataKey = "external_cli_sessions"

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

// ProviderSessionBinding maps a Leros session to a provider-native CLI session.
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

// ProviderSessionMetadata stores provider-native session information in Session.Metadata.
type ProviderSessionMetadata struct {
	Provider          string    `json:"provider"`
	ProviderSessionID string    `json:"provider_session_id"`
	CreatedAt         time.Time `json:"created_at"`
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
