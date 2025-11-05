package fsm

import (
	"context"
	"testing"
)

// Test data - simple document approval workflow
var (
	testStates = []State{
		{Name: "draft"},
		{Name: "submitted"},
		{Name: "approved"},
		{Name: "rejected"},
		{Name: "published"},
	}

	testEvents = []Event{
		{Name: "submit"},
		{Name: "approve"},
		{Name: "reject"},
		{Name: "publish"},
		{Name: "revise"},
	}

	testTransitions = []Transition{
		{From: State{Name: "draft"}, To: State{Name: "submitted"}, Event: Event{Name: "submit"}},
		{From: State{Name: "submitted"}, To: State{Name: "approved"}, Event: Event{Name: "approve"}},
		{From: State{Name: "submitted"}, To: State{Name: "rejected"}, Event: Event{Name: "reject"}},
		{From: State{Name: "approved"}, To: State{Name: "published"}, Event: Event{Name: "publish"}},
		{From: State{Name: "rejected"}, To: State{Name: "draft"}, Event: Event{Name: "revise"}},
	}
)

func newTestFSM(t *testing.T) *FSM {
	storage := NewMemoryStorage()
	fsm, err := New(testStates, testEvents, testTransitions, storage)
	if err != nil {
		t.Fatalf("failed to create FSM: %v", err)
	}
	return fsm
}

func TestNew(t *testing.T) {
	tests := []struct {
		name        string
		states      []State
		events      []Event
		transitions []Transition
		wantErr     bool
	}{
		{
			name:        "valid FSM",
			states:      testStates,
			events:      testEvents,
			transitions: testTransitions,
			wantErr:     false,
		},
		{
			name:        "no states",
			states:      []State{},
			events:      testEvents,
			transitions: testTransitions,
			wantErr:     true,
		},
		{
			name:        "no events",
			states:      testStates,
			events:      []Event{},
			transitions: testTransitions,
			wantErr:     true,
		},
		{
			name:        "no transitions",
			states:      testStates,
			events:      testEvents,
			transitions: []Transition{},
			wantErr:     true,
		},
		{
			name:   "invalid from state",
			states: testStates,
			events: testEvents,
			transitions: []Transition{
				{From: State{Name: "invalid"}, To: State{Name: "submitted"}, Event: Event{Name: "submit"}},
			},
			wantErr: true,
		},
		{
			name:   "invalid to state",
			states: testStates,
			events: testEvents,
			transitions: []Transition{
				{From: State{Name: "draft"}, To: State{Name: "invalid"}, Event: Event{Name: "submit"}},
			},
			wantErr: true,
		},
		{
			name:   "invalid event",
			states: testStates,
			events: testEvents,
			transitions: []Transition{
				{From: State{Name: "draft"}, To: State{Name: "submitted"}, Event: Event{Name: "invalid"}},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := NewMemoryStorage()
			_, err := New(tt.states, tt.events, tt.transitions, storage)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFSM_Start(t *testing.T) {
	fsm := newTestFSM(t)
	ctx := context.Background()

	entity := Entity{Type: "document", ID: "doc-1"}
	initialState := State{Name: "draft"}

	err := fsm.Start(ctx, entity, initialState, "user1")
	if err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	// Verify state was saved
	currentState, err := fsm.GetState(ctx, entity)
	if err != nil {
		t.Fatalf("GetState() error = %v", err)
	}

	if currentState.Name != initialState.Name {
		t.Errorf("GetState() = %v, want %v", currentState.Name, initialState.Name)
	}
}

func TestFSM_StartInvalidState(t *testing.T) {
	fsm := newTestFSM(t)
	ctx := context.Background()

	entity := Entity{Type: "document", ID: "doc-2"}
	invalidState := State{Name: "invalid"}

	err := fsm.Start(ctx, entity, invalidState, "user1")
	if err == nil {
		t.Fatal("Start() should fail with invalid state")
	}
}

func TestFSM_Trigger(t *testing.T) {
	fsm := newTestFSM(t)
	ctx := context.Background()

	entity := Entity{Type: "document", ID: "doc-3"}

	// Start in draft state
	err := fsm.Start(ctx, entity, State{Name: "draft"}, "user1")
	if err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	// Trigger submit event
	err = fsm.Trigger(ctx, entity, Event{Name: "submit"}, "user1")
	if err != nil {
		t.Fatalf("Trigger() error = %v", err)
	}

	// Verify new state
	currentState, err := fsm.GetState(ctx, entity)
	if err != nil {
		t.Fatalf("GetState() error = %v", err)
	}

	if currentState.Name != "submitted" {
		t.Errorf("GetState() = %v, want submitted", currentState.Name)
	}
}

func TestFSM_TriggerInvalidEvent(t *testing.T) {
	fsm := newTestFSM(t)
	ctx := context.Background()

	entity := Entity{Type: "document", ID: "doc-4"}

	// Start in draft state
	err := fsm.Start(ctx, entity, State{Name: "draft"}, "user1")
	if err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	// Try to trigger an event that's not valid from draft state
	err = fsm.Trigger(ctx, entity, Event{Name: "approve"}, "user1")
	if err == nil {
		t.Fatal("Trigger() should fail with invalid transition")
	}
}

func TestFSM_CompleteWorkflow(t *testing.T) {
	fsm := newTestFSM(t)
	ctx := context.Background()

	entity := Entity{Type: "document", ID: "doc-5"}

	// Test complete workflow: draft -> submitted -> approved -> published
	steps := []struct {
		event     Event
		wantState string
	}{
		{Event{Name: "submit"}, "submitted"},
		{Event{Name: "approve"}, "approved"},
		{Event{Name: "publish"}, "published"},
	}

	// Start in draft
	err := fsm.Start(ctx, entity, State{Name: "draft"}, "user1")
	if err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	// Execute each step
	for _, step := range steps {
		err = fsm.Trigger(ctx, entity, step.event, "user1")
		if err != nil {
			t.Fatalf("Trigger(%v) error = %v", step.event.Name, err)
		}

		currentState, err := fsm.GetState(ctx, entity)
		if err != nil {
			t.Fatalf("GetState() error = %v", err)
		}

		if currentState.Name != step.wantState {
			t.Errorf("After %v: got state %v, want %v", step.event.Name, currentState.Name, step.wantState)
		}
	}

	// Verify all transitions were recorded
	transitions, err := fsm.GetTransitions(ctx, entity)
	if err != nil {
		t.Fatalf("GetTransitions() error = %v", err)
	}

	// Should have 4 transitions: start + 3 events
	if len(transitions) != 4 {
		t.Errorf("GetTransitions() count = %v, want 4", len(transitions))
	}
}

func TestFSM_CanTrigger(t *testing.T) {
	fsm := newTestFSM(t)
	ctx := context.Background()

	entity := Entity{Type: "document", ID: "doc-6"}

	// Start in draft state
	err := fsm.Start(ctx, entity, State{Name: "draft"}, "user1")
	if err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	tests := []struct {
		event Event
		want  bool
	}{
		{Event{Name: "submit"}, true},
		{Event{Name: "approve"}, false},
		{Event{Name: "reject"}, false},
		{Event{Name: "publish"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.event.Name, func(t *testing.T) {
			got := fsm.CanTrigger(ctx, entity, tt.event)
			if got != tt.want {
				t.Errorf("CanTrigger(%v) = %v, want %v", tt.event.Name, got, tt.want)
			}
		})
	}
}

func TestFSM_GetAvailableEvents(t *testing.T) {
	fsm := newTestFSM(t)
	ctx := context.Background()

	entity := Entity{Type: "document", ID: "doc-7"}

	// Start in submitted state
	err := fsm.Start(ctx, entity, State{Name: "submitted"}, "user1")
	if err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	events, err := fsm.GetAvailableEvents(ctx, entity)
	if err != nil {
		t.Fatalf("GetAvailableEvents() error = %v", err)
	}

	// From submitted state, should be able to approve or reject
	if len(events) != 2 {
		t.Errorf("GetAvailableEvents() count = %v, want 2", len(events))
	}

	hasApprove := false
	hasReject := false
	for _, e := range events {
		if e.Name == "approve" {
			hasApprove = true
		}
		if e.Name == "reject" {
			hasReject = true
		}
	}

	if !hasApprove || !hasReject {
		t.Errorf("GetAvailableEvents() missing expected events, got %v", events)
	}
}

func TestFSM_GetNextState(t *testing.T) {
	fsm := newTestFSM(t)

	tests := []struct {
		name      string
		from      State
		event     Event
		want      string
		wantErr   bool
	}{
		{
			name:    "valid transition",
			from:    State{Name: "draft"},
			event:   Event{Name: "submit"},
			want:    "submitted",
			wantErr: false,
		},
		{
			name:    "invalid transition",
			from:    State{Name: "draft"},
			event:   Event{Name: "approve"},
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := fsm.GetNextState(tt.from, tt.event)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetNextState() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got.Name != tt.want {
				t.Errorf("GetNextState() = %v, want %v", got.Name, tt.want)
			}
		})
	}
}

func TestFSM_RejectionRevisionWorkflow(t *testing.T) {
	fsm := newTestFSM(t)
	ctx := context.Background()

	entity := Entity{Type: "document", ID: "doc-8"}

	// Start in draft
	err := fsm.Start(ctx, entity, State{Name: "draft"}, "user1")
	if err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	// Submit
	err = fsm.Trigger(ctx, entity, Event{Name: "submit"}, "user1")
	if err != nil {
		t.Fatalf("Trigger(submit) error = %v", err)
	}

	// Reject
	err = fsm.Trigger(ctx, entity, Event{Name: "reject"}, "user2")
	if err != nil {
		t.Fatalf("Trigger(reject) error = %v", err)
	}

	// Verify in rejected state
	currentState, err := fsm.GetState(ctx, entity)
	if err != nil {
		t.Fatalf("GetState() error = %v", err)
	}
	if currentState.Name != "rejected" {
		t.Errorf("GetState() = %v, want rejected", currentState.Name)
	}

	// Revise back to draft
	err = fsm.Trigger(ctx, entity, Event{Name: "revise"}, "user1")
	if err != nil {
		t.Fatalf("Trigger(revise) error = %v", err)
	}

	// Verify back in draft state
	currentState, err = fsm.GetState(ctx, entity)
	if err != nil {
		t.Fatalf("GetState() error = %v", err)
	}
	if currentState.Name != "draft" {
		t.Errorf("GetState() = %v, want draft", currentState.Name)
	}
}

func TestMemoryStorage_EntityNotFound(t *testing.T) {
	storage := NewMemoryStorage()
	ctx := context.Background()

	entity := Entity{Type: "document", ID: "nonexistent"}

	_, err := storage.GetCurrentState(ctx, entity)
	if err != ErrEntityNotFound {
		t.Errorf("GetCurrentState() error = %v, want ErrEntityNotFound", err)
	}
}
