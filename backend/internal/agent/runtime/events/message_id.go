package events

import (
	"strings"

	"github.com/google/uuid"
)

// MessageIDMapper maps provider message identifiers to Leros message IDs within one run.
type MessageIDMapper struct {
	byProvider map[string]string
	current    string
}

// NewMessageIDMapper creates a mapper for one runtime execution.
func NewMessageIDMapper() *MessageIDMapper {
	return &MessageIDMapper{
		byProvider: make(map[string]string),
	}
}

// ForProvider returns a stable Leros message ID for a provider-local message ID.
func (m *MessageIDMapper) ForProvider(providerID string) string {
	if m == nil {
		return uuid.NewString()
	}
	providerID = strings.TrimSpace(providerID)
	if providerID == "" {
		return m.CurrentOrNew()
	}
	if m.byProvider == nil {
		m.byProvider = make(map[string]string)
	}
	if id := m.byProvider[providerID]; id != "" {
		m.current = id
		return id
	}
	id := uuid.NewString()
	m.byProvider[providerID] = id
	m.current = id
	return id
}

// CurrentOrNew returns the current message ID, creating one if needed.
func (m *MessageIDMapper) CurrentOrNew() string {
	if m == nil {
		return uuid.NewString()
	}
	if m.current == "" {
		m.current = uuid.NewString()
	}
	return m.current
}

// StartNew creates a fresh message ID and marks it as current.
func (m *MessageIDMapper) StartNew() string {
	if m == nil {
		return uuid.NewString()
	}
	m.current = uuid.NewString()
	return m.current
}
