package fsm

import (
	"context"
	"os"
	"testing"
	"time"
)

// getTestPostgresConnString returns the connection string for testing
// Set environment variable POSTGRES_TEST_CONN to run these tests
// Example: POSTGRES_TEST_CONN="postgres://user:password@localhost:5432/fsm_test"
func getTestPostgresConnString(t *testing.T) string {
	connString := os.Getenv("POSTGRES_TEST_CONN")
	if connString == "" {
		t.Skip("Skipping PostgreSQL tests: POSTGRES_TEST_CONN not set")
	}
	return connString
}

func setupTestPostgresDB(t *testing.T) *PostgresStorage {
	ctx := context.Background()
	connString := getTestPostgresConnString(t)

	storage, err := NewPostgresStorage(ctx, connString)
	if err != nil {
		t.Fatalf("Failed to create PostgreSQL storage: %v", err)
	}

	// Clean up the test table
	_, err = storage.pool.Exec(ctx, "TRUNCATE TABLE entity_state_transition")
	if err != nil {
		t.Fatalf("Failed to clean test database: %v", err)
	}

	return storage
}

func TestPostgresStorage_SaveTransition(t *testing.T) {
	storage := setupTestPostgresDB(t)
	defer storage.Close()

	ctx := context.Background()
	entity := Entity{Type: "document", ID: "doc-1"}

	transition := EntityTransition{
		Entity: entity,
		Transition: Transition{
			From:      State{Name: ""},
			To:        State{Name: "draft"},
			Event:     Event{Name: "start"},
			CreatedBy: "user1",
			CreatedAt: time.Now().UTC(),
		},
	}

	err := storage.SaveTransition(ctx, transition)
	if err != nil {
		t.Fatalf("SaveTransition() error = %v", err)
	}

	// Verify it was saved
	state, err := storage.GetCurrentState(ctx, entity)
	if err != nil {
		t.Fatalf("GetCurrentState() error = %v", err)
	}

	if state.Name != "draft" {
		t.Errorf("GetCurrentState() = %v, want draft", state.Name)
	}
}

func TestPostgresStorage_GetCurrentState(t *testing.T) {
	storage := setupTestPostgresDB(t)
	defer storage.Close()

	ctx := context.Background()
	entity := Entity{Type: "document", ID: "doc-2"}

	// Save multiple transitions
	transitions := []EntityTransition{
		{
			Entity: entity,
			Transition: Transition{
				From:      State{Name: ""},
				To:        State{Name: "draft"},
				Event:     Event{Name: "start"},
				CreatedBy: "user1",
				CreatedAt: time.Now().UTC().Add(-2 * time.Hour),
			},
		},
		{
			Entity: entity,
			Transition: Transition{
				From:      State{Name: "draft"},
				To:        State{Name: "submitted"},
				Event:     Event{Name: "submit"},
				CreatedBy: "user1",
				CreatedAt: time.Now().UTC().Add(-1 * time.Hour),
			},
		},
		{
			Entity: entity,
			Transition: Transition{
				From:      State{Name: "submitted"},
				To:        State{Name: "approved"},
				Event:     Event{Name: "approve"},
				CreatedBy: "user2",
				CreatedAt: time.Now().UTC(),
			},
		},
	}

	for _, tr := range transitions {
		err := storage.SaveTransition(ctx, tr)
		if err != nil {
			t.Fatalf("SaveTransition() error = %v", err)
		}
	}

	// Should return the most recent state
	state, err := storage.GetCurrentState(ctx, entity)
	if err != nil {
		t.Fatalf("GetCurrentState() error = %v", err)
	}

	if state.Name != "approved" {
		t.Errorf("GetCurrentState() = %v, want approved", state.Name)
	}
}

func TestPostgresStorage_GetCurrentState_NotFound(t *testing.T) {
	storage := setupTestPostgresDB(t)
	defer storage.Close()

	ctx := context.Background()
	entity := Entity{Type: "document", ID: "nonexistent"}

	_, err := storage.GetCurrentState(ctx, entity)
	if err != ErrEntityNotFound {
		t.Errorf("GetCurrentState() error = %v, want ErrEntityNotFound", err)
	}
}

func TestPostgresStorage_GetTransitions(t *testing.T) {
	storage := setupTestPostgresDB(t)
	defer storage.Close()

	ctx := context.Background()
	entity := Entity{Type: "document", ID: "doc-3"}

	// Save multiple transitions
	transitions := []EntityTransition{
		{
			Entity: entity,
			Transition: Transition{
				From:      State{Name: ""},
				To:        State{Name: "draft"},
				Event:     Event{Name: "start"},
				CreatedBy: "user1",
				CreatedAt: time.Now().UTC().Add(-2 * time.Hour),
			},
		},
		{
			Entity: entity,
			Transition: Transition{
				From:      State{Name: "draft"},
				To:        State{Name: "submitted"},
				Event:     Event{Name: "submit"},
				CreatedBy: "user1",
				CreatedAt: time.Now().UTC().Add(-1 * time.Hour),
			},
		},
	}

	for _, tr := range transitions {
		err := storage.SaveTransition(ctx, tr)
		if err != nil {
			t.Fatalf("SaveTransition() error = %v", err)
		}
	}

	// Get all transitions
	result, err := storage.GetTransitions(ctx, entity)
	if err != nil {
		t.Fatalf("GetTransitions() error = %v", err)
	}

	if len(result) != 2 {
		t.Errorf("GetTransitions() count = %v, want 2", len(result))
	}

	// Verify order (should be ascending by created_at)
	if result[0].Transition.To.Name != "draft" {
		t.Errorf("First transition to_state = %v, want draft", result[0].Transition.To.Name)
	}

	if result[1].Transition.To.Name != "submitted" {
		t.Errorf("Second transition to_state = %v, want submitted", result[1].Transition.To.Name)
	}
}

func TestPostgresStorage_MultipleEntities(t *testing.T) {
	storage := setupTestPostgresDB(t)
	defer storage.Close()

	ctx := context.Background()

	// Create transitions for multiple entities
	entity1 := Entity{Type: "document", ID: "doc-4"}
	entity2 := Entity{Type: "document", ID: "doc-5"}

	tr1 := EntityTransition{
		Entity: entity1,
		Transition: Transition{
			From:      State{Name: ""},
			To:        State{Name: "draft"},
			Event:     Event{Name: "start"},
			CreatedBy: "user1",
			CreatedAt: time.Now().UTC(),
		},
	}

	tr2 := EntityTransition{
		Entity: entity2,
		Transition: Transition{
			From:      State{Name: ""},
			To:        State{Name: "submitted"},
			Event:     Event{Name: "start"},
			CreatedBy: "user2",
			CreatedAt: time.Now().UTC(),
		},
	}

	storage.SaveTransition(ctx, tr1)
	storage.SaveTransition(ctx, tr2)

	// Verify each entity has its own state
	state1, err := storage.GetCurrentState(ctx, entity1)
	if err != nil {
		t.Fatalf("GetCurrentState(entity1) error = %v", err)
	}
	if state1.Name != "draft" {
		t.Errorf("Entity1 state = %v, want draft", state1.Name)
	}

	state2, err := storage.GetCurrentState(ctx, entity2)
	if err != nil {
		t.Fatalf("GetCurrentState(entity2) error = %v", err)
	}
	if state2.Name != "submitted" {
		t.Errorf("Entity2 state = %v, want submitted", state2.Name)
	}

	// Verify transitions are isolated
	transitions1, _ := storage.GetTransitions(ctx, entity1)
	if len(transitions1) != 1 {
		t.Errorf("Entity1 transitions count = %v, want 1", len(transitions1))
	}

	transitions2, _ := storage.GetTransitions(ctx, entity2)
	if len(transitions2) != 1 {
		t.Errorf("Entity2 transitions count = %v, want 1", len(transitions2))
	}
}

func TestPostgresStorage_WithFSM(t *testing.T) {
	storage := setupTestPostgresDB(t)
	defer storage.Close()

	// Create FSM with PostgreSQL storage
	states := []State{{Name: "draft"}, {Name: "submitted"}, {Name: "approved"}}
	events := []Event{{Name: "submit"}, {Name: "approve"}}
	transitions := []Transition{
		{From: State{Name: "draft"}, To: State{Name: "submitted"}, Event: Event{Name: "submit"}},
		{From: State{Name: "submitted"}, To: State{Name: "approved"}, Event: Event{Name: "approve"}},
	}

	fsm, err := New(states, events, transitions, storage)
	if err != nil {
		t.Fatalf("Failed to create FSM: %v", err)
	}

	ctx := context.Background()
	entity := Entity{Type: "document", ID: "doc-6"}

	// Start entity
	err = fsm.Start(ctx, entity, State{Name: "draft"}, "user1")
	if err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	// Trigger transition
	err = fsm.Trigger(ctx, entity, Event{Name: "submit"}, "user1")
	if err != nil {
		t.Fatalf("Trigger() error = %v", err)
	}

	// Verify state
	state, err := fsm.GetState(ctx, entity)
	if err != nil {
		t.Fatalf("GetState() error = %v", err)
	}

	if state.Name != "submitted" {
		t.Errorf("GetState() = %v, want submitted", state.Name)
	}

	// Verify transition history
	history, err := fsm.GetTransitions(ctx, entity)
	if err != nil {
		t.Fatalf("GetTransitions() error = %v", err)
	}

	if len(history) != 2 {
		t.Errorf("Transition history count = %v, want 2", len(history))
	}
}
