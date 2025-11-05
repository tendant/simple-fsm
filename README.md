# Simple FSM

A simplified, clean implementation of a Finite State Machine in Go with pluggable storage backends.

## Overview

A finite state machine implementation that focuses on core state machine functionality while maintaining storage flexibility. It's designed to be easy to understand, test, and extend.

## Key Features

- **Simple API**: Clean, intuitive interface for state machine operations
- **Pluggable Storage**: Interface-based storage system supporting multiple backends
- **Thread-Safe**: Memory storage implementation uses proper synchronization
- **Entity Tracking**: Track state for multiple entities independently
- **Validation**: Comprehensive validation of states, events, and transitions
- **Audit Trail**: Complete history of all state transitions

## Core Concepts

### State
A state represents a distinct condition in the state machine. Example: "draft", "submitted", "approved"

### Event
An event is a trigger that causes a state transition. Example: "submit", "approve", "reject"

### Transition
A transition defines how the system moves from one state to another in response to an event.

### Entity
An entity is something being tracked by the FSM. Each entity maintains its own state independently.

## Installation

```bash
go get github.com/tendant/simple-fsm
```

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"

    fsm "github.com/tendant/simple-fsm"
)

func main() {
    // Define your workflow
    states := []fsm.State{
        {Name: "draft"},
        {Name: "published"},
    }

    events := []fsm.Event{
        {Name: "publish"},
    }

    transitions := []fsm.Transition{
        {
            From:  fsm.State{Name: "draft"},
            To:    fsm.State{Name: "published"},
            Event: fsm.Event{Name: "publish"},
        },
    }

    // Create FSM with in-memory storage
    storage := fsm.NewMemoryStorage()
    machine, err := fsm.New(states, events, transitions, storage)
    if err != nil {
        log.Fatal(err)
    }

    // Track an entity
    doc := fsm.Entity{Type: "document", ID: "doc-1"}
    ctx := context.Background()

    // Start in initial state
    err = machine.Start(ctx, doc, fsm.State{Name: "draft"}, "alice")
    if err != nil {
        log.Fatal(err)
    }

    // Trigger a transition
    err = machine.Trigger(ctx, doc, fsm.Event{Name: "publish"}, "alice")
    if err != nil {
        log.Fatal(err)
    }

    // Check current state
    state, _ := machine.GetState(ctx, doc)
    fmt.Printf("Current state: %s\n", state.Name)
}
```

## API Reference

### Creating an FSM

```go
func New(states []State, events []Event, transitions []Transition, storage Storage) (*FSM, error)
```

Creates a new FSM with validation of all states, events, and transitions.

### Starting an Entity

```go
func (f *FSM) Start(ctx context.Context, entity Entity, initialState State, createdBy string) error
```

Initializes an entity in the specified state.

### Triggering Events

```go
func (f *FSM) Trigger(ctx context.Context, entity Entity, event Event, createdBy string) error
```

Triggers an event for an entity, causing a state transition.

### Querying State

```go
func (f *FSM) GetState(ctx context.Context, entity Entity) (State, error)
func (f *FSM) GetTransitions(ctx context.Context, entity Entity) ([]EntityTransition, error)
func (f *FSM) CanTrigger(ctx context.Context, entity Entity, event Event) bool
func (f *FSM) GetAvailableEvents(ctx context.Context, entity Entity) ([]Event, error)
func (f *FSM) GetNextState(currentState State, event Event) (State, error)
```

## Storage Backends

### Memory Storage (Included)

In-memory storage for testing and simple use cases:

```go
storage := fsm.NewMemoryStorage()
```

### PostgreSQL Storage

Production-ready PostgreSQL storage backend:

```go
// Create PostgreSQL storage
storage, err := fsm.NewPostgresStorage(ctx, "postgres://user:password@localhost:5432/mydb")
if err != nil {
    log.Fatal(err)
}
defer storage.Close()

// Use with FSM
machine, err := fsm.New(states, events, transitions, storage)
```

**Setup:**

1. Run the migration to create the required table using [goose](https://github.com/pressly/goose):
```bash
# Install goose
go install github.com/pressly/goose/v3/cmd/goose@latest

# Run migration
goose -dir migrations postgres "your-connection-string" up
```

Or manually run the SQL from `migrations/20251104220000_create_entity_state_transition.sql`:
```sql
CREATE TABLE entity_state_transition (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at TIMESTAMP NOT NULL DEFAULT (NOW() AT TIME ZONE 'utc'),
    entity_type VARCHAR(255) NOT NULL,
    entity_id VARCHAR(255) NOT NULL,
    from_state VARCHAR(255),
    to_state VARCHAR(255) NOT NULL,
    event VARCHAR(255) NOT NULL,
    created_by VARCHAR(255)
);
-- Plus indexes (see migration file for complete SQL)
```

2. Install the PostgreSQL driver:
```bash
go get github.com/jackc/pgx/v5
```

3. See `examples/postgres_example.go` for a complete working example.

**Connection String Format:**
```
postgres://username:password@host:port/database?options
```

### Custom Storage

Implement the `Storage` interface for custom backends (Redis, DynamoDB, etc.):

```go
type Storage interface {
    SaveTransition(ctx context.Context, et EntityTransition) error
    GetCurrentState(ctx context.Context, entity Entity) (State, error)
    GetTransitions(ctx context.Context, entity Entity) ([]EntityTransition, error)
}
```

## Testing

Run tests:
```bash
go test -v
```

Run with coverage:
```bash
go test -cover
```

Run PostgreSQL integration tests (requires PostgreSQL connection):
```bash
# Set connection string
export POSTGRES_TEST_CONN="postgres://user:password@localhost:5432/fsm_test"

# Run all tests including PostgreSQL
go test -v

# PostgreSQL tests are automatically skipped if POSTGRES_TEST_CONN is not set
```

## Design Principles

1. **Simplicity**: Easy to understand and maintain
2. **Flexibility**: Storage backend is pluggable
3. **Type Safety**: Strong typing for states, events, and transitions
4. **Testability**: Pure functions and dependency injection
5. **Concurrency**: Thread-safe storage implementations

## Example Workflows

### Document Approval
draft → submitted → approved → published

### Order Processing
pending → processing → shipped → delivered

### User Onboarding
registered → verified → active → premium

## License

MIT
