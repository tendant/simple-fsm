# Simple FSM Examples

This directory contains example implementations demonstrating how to use the simple-fsm library.

## PostgreSQL Example

**File:** `postgres_example.go`

Demonstrates how to use the FSM with PostgreSQL storage backend.

**Prerequisites:**
1. PostgreSQL server running
2. Database created (e.g., `fsm_db`)
3. Migration applied (see `../migrations/`)

**Setup:**
```bash
# Create database
createdb fsm_db

# Run migration using goose
goose -dir ../migrations postgres "postgres://user:password@localhost:5432/fsm_db" up
```

Or manually:
```sql
-- Connect to database
\c fsm_db

-- Copy contents from ../migrations/20251104220000_create_entity_state_transition.sql
-- (Run the SQL between -- +goose Up and -- +goose Down markers)
```

**Run:**
```bash
# Update connection string in the example file first
go run postgres_example.go
```

**Expected Output:**
```
Document ID: doc-1730761200
✓ Document started in 'draft' state
✓ Document submitted
Current state: submitted
Available events: approve, reject
✓ Document approved by bob
✓ Document published by charlie
Final state: published

Transition History (4 transitions):
1.  -> draft (event: start, by: alice, at: 2025-11-04T22:00:00Z)
2. draft -> submitted (event: submit, by: alice, at: 2025-11-04T22:00:01Z)
3. submitted -> approved (event: approve, by: bob, at: 2025-11-04T22:00:02Z)
4. approved -> published (event: publish, by: charlie, at: 2025-11-04T22:00:03Z)

✓ All transitions persisted to PostgreSQL!
```

## Creating Your Own Example

```go
package main

import (
    "context"
    "log"

    fsm "github.com/tendant/simple-fsm"
)

func main() {
    // 1. Choose storage backend
    storage := fsm.NewMemoryStorage() // or NewPostgresStorage()

    // 2. Define your workflow
    states := []fsm.State{
        {Name: "pending"},
        {Name: "active"},
    }

    events := []fsm.Event{
        {Name: "activate"},
    }

    transitions := []fsm.Transition{
        {
            From:  fsm.State{Name: "pending"},
            To:    fsm.State{Name: "active"},
            Event: fsm.Event{Name: "activate"},
        },
    }

    // 3. Create FSM
    machine, err := fsm.New(states, events, transitions, storage)
    if err != nil {
        log.Fatal(err)
    }

    // 4. Use it!
    ctx := context.Background()
    entity := fsm.Entity{Type: "user", ID: "user-123"}

    machine.Start(ctx, entity, fsm.State{Name: "pending"}, "system")
    machine.Trigger(ctx, entity, fsm.Event{Name: "activate"}, "admin")
}
```
