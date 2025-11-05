package fsm

import (
	"context"
	"errors"
	"fmt"
	"time"
)

var (
	ErrInvalidState      = errors.New("invalid state")
	ErrInvalidEvent      = errors.New("invalid event")
	ErrInvalidTransition = errors.New("invalid transition")
)

// State represents a state in the FSM
type State struct {
	Name string
}

// Event represents an event that triggers a transition
type Event struct {
	Name string
}

// Transition defines how states change in response to events
type Transition struct {
	From      State
	To        State
	Event     Event
	CreatedAt time.Time
	CreatedBy string
}

// Entity represents something being tracked by the FSM
type Entity struct {
	Type string
	ID   string
}

// EntityState represents the current state of an entity
type EntityState struct {
	Entity Entity
	State  State
}

// EntityTransition represents a state transition for an entity
type EntityTransition struct {
	Entity     Entity
	Transition Transition
}

// Storage defines the interface for persisting FSM state
type Storage interface {
	SaveTransition(ctx context.Context, et EntityTransition) error
	GetCurrentState(ctx context.Context, entity Entity) (State, error)
	GetTransitions(ctx context.Context, entity Entity) ([]EntityTransition, error)
}

// FSM represents a simple finite state machine
type FSM struct {
	states      []State
	events      []Event
	transitions []Transition
	storage     Storage
}

// New creates a new FSM instance
func New(states []State, events []Event, transitions []Transition, storage Storage) (*FSM, error) {
	if len(states) == 0 {
		return nil, errors.New("no states defined")
	}

	if len(events) == 0 {
		return nil, errors.New("no events defined")
	}

	if len(transitions) == 0 {
		return nil, errors.New("no transitions defined")
	}

	if storage == nil {
		return nil, errors.New("storage cannot be nil")
	}

	// Validate transitions
	for _, t := range transitions {
		if err := validateState(t.From, states); err != nil {
			return nil, fmt.Errorf("invalid from state in transition: %w", err)
		}
		if err := validateState(t.To, states); err != nil {
			return nil, fmt.Errorf("invalid to state in transition: %w", err)
		}
		if err := validateEvent(t.Event, events); err != nil {
			return nil, fmt.Errorf("invalid event in transition: %w", err)
		}
	}

	return &FSM{
		states:      states,
		events:      events,
		transitions: transitions,
		storage:     storage,
	}, nil
}

// Start initializes an entity in the given state
func (f *FSM) Start(ctx context.Context, entity Entity, initialState State, createdBy string) error {
	if err := validateState(initialState, f.states); err != nil {
		return err
	}

	et := EntityTransition{
		Entity: entity,
		Transition: Transition{
			From:      State{Name: ""},
			To:        initialState,
			Event:     Event{Name: "start"},
			CreatedAt: time.Now().UTC(),
			CreatedBy: createdBy,
		},
	}

	return f.storage.SaveTransition(ctx, et)
}

// Trigger attempts to trigger an event for an entity, causing a state transition
func (f *FSM) Trigger(ctx context.Context, entity Entity, event Event, createdBy string) error {
	// Get current state
	currentState, err := f.storage.GetCurrentState(ctx, entity)
	if err != nil {
		return fmt.Errorf("failed to get current state: %w", err)
	}

	// Validate event
	if err := validateEvent(event, f.events); err != nil {
		return err
	}

	// Find valid transition
	nextState, err := f.findNextState(currentState, event)
	if err != nil {
		return err
	}

	// Save transition
	et := EntityTransition{
		Entity: entity,
		Transition: Transition{
			From:      currentState,
			To:        nextState,
			Event:     event,
			CreatedAt: time.Now().UTC(),
			CreatedBy: createdBy,
		},
	}

	return f.storage.SaveTransition(ctx, et)
}

// GetState returns the current state of an entity
func (f *FSM) GetState(ctx context.Context, entity Entity) (State, error) {
	return f.storage.GetCurrentState(ctx, entity)
}

// GetTransitions returns all transitions for an entity
func (f *FSM) GetTransitions(ctx context.Context, entity Entity) ([]EntityTransition, error) {
	return f.storage.GetTransitions(ctx, entity)
}

// CanTrigger checks if an event can be triggered from the entity's current state
func (f *FSM) CanTrigger(ctx context.Context, entity Entity, event Event) bool {
	currentState, err := f.storage.GetCurrentState(ctx, entity)
	if err != nil {
		return false
	}

	_, err = f.findNextState(currentState, event)
	return err == nil
}

// GetAvailableEvents returns all events that can be triggered from the entity's current state
func (f *FSM) GetAvailableEvents(ctx context.Context, entity Entity) ([]Event, error) {
	currentState, err := f.storage.GetCurrentState(ctx, entity)
	if err != nil {
		return nil, err
	}

	var events []Event
	for _, t := range f.transitions {
		if t.From.Name == currentState.Name {
			events = append(events, t.Event)
		}
	}

	return events, nil
}

// GetNextState returns the next state for a given current state and event without triggering
func (f *FSM) GetNextState(currentState State, event Event) (State, error) {
	return f.findNextState(currentState, event)
}

// findNextState finds the next state for a given state and event
func (f *FSM) findNextState(from State, event Event) (State, error) {
	for _, t := range f.transitions {
		if t.From.Name == from.Name && t.Event.Name == event.Name {
			return t.To, nil
		}
	}

	return State{}, fmt.Errorf("%w: no transition from %q with event %q",
		ErrInvalidTransition, from.Name, event.Name)
}

// Helper validation functions
func validateState(state State, validStates []State) error {
	for _, s := range validStates {
		if s.Name == state.Name {
			return nil
		}
	}
	return fmt.Errorf("%w: state %q is not valid", ErrInvalidState, state.Name)
}

func validateEvent(event Event, validEvents []Event) error {
	for _, e := range validEvents {
		if e.Name == event.Name {
			return nil
		}
	}
	return fmt.Errorf("%w: event %q is not valid", ErrInvalidEvent, event.Name)
}
