package fsm

import (
	"context"
	"errors"
	"sync"
)

var (
	ErrEntityNotFound = errors.New("entity not found")
)

// MemoryStorage implements Storage interface using in-memory data structures
type MemoryStorage struct {
	mu          sync.RWMutex
	transitions []EntityTransition
}

// NewMemoryStorage creates a new in-memory storage instance
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		transitions: make([]EntityTransition, 0),
	}
}

// SaveTransition saves a transition to memory
func (m *MemoryStorage) SaveTransition(ctx context.Context, et EntityTransition) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.transitions = append(m.transitions, et)
	return nil
}

// GetCurrentState retrieves the current state of an entity
func (m *MemoryStorage) GetCurrentState(ctx context.Context, entity Entity) (State, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Find the most recent transition for this entity
	for i := len(m.transitions) - 1; i >= 0; i-- {
		t := m.transitions[i]
		if t.Entity.Type == entity.Type && t.Entity.ID == entity.ID {
			return t.Transition.To, nil
		}
	}

	return State{}, ErrEntityNotFound
}

// GetTransitions retrieves all transitions for an entity
func (m *MemoryStorage) GetTransitions(ctx context.Context, entity Entity) ([]EntityTransition, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []EntityTransition
	for _, t := range m.transitions {
		if t.Entity.Type == entity.Type && t.Entity.ID == entity.ID {
			result = append(result, t)
		}
	}

	return result, nil
}
