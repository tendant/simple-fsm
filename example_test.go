package fsm_test

import (
	"context"
	"fmt"
	"log"

	"simple-fsm"
)

func Example() {
	// Define states for a document approval workflow
	states := []fsm.State{
		{Name: "draft"},
		{Name: "submitted"},
		{Name: "approved"},
		{Name: "rejected"},
		{Name: "published"},
	}

	// Define events that trigger state transitions
	events := []fsm.Event{
		{Name: "submit"},
		{Name: "approve"},
		{Name: "reject"},
		{Name: "publish"},
		{Name: "revise"},
	}

	// Define valid transitions
	transitions := []fsm.Transition{
		{From: fsm.State{Name: "draft"}, To: fsm.State{Name: "submitted"}, Event: fsm.Event{Name: "submit"}},
		{From: fsm.State{Name: "submitted"}, To: fsm.State{Name: "approved"}, Event: fsm.Event{Name: "approve"}},
		{From: fsm.State{Name: "submitted"}, To: fsm.State{Name: "rejected"}, Event: fsm.Event{Name: "reject"}},
		{From: fsm.State{Name: "approved"}, To: fsm.State{Name: "published"}, Event: fsm.Event{Name: "publish"}},
		{From: fsm.State{Name: "rejected"}, To: fsm.State{Name: "draft"}, Event: fsm.Event{Name: "revise"}},
	}

	// Create storage backend (using in-memory storage)
	storage := fsm.NewMemoryStorage()

	// Create the FSM
	machine, err := fsm.New(states, events, transitions, storage)
	if err != nil {
		log.Fatal(err)
	}

	// Create an entity (document) to track
	document := fsm.Entity{
		Type: "document",
		ID:   "doc-123",
	}

	ctx := context.Background()

	// Start the document in draft state
	err = machine.Start(ctx, document, fsm.State{Name: "draft"}, "alice")
	if err != nil {
		log.Fatal(err)
	}

	// Check current state
	currentState, _ := machine.GetState(ctx, document)
	fmt.Printf("Current state: %s\n", currentState.Name)

	// Check what events can be triggered
	availableEvents, _ := machine.GetAvailableEvents(ctx, document)
	fmt.Printf("Available events: %v\n", availableEvents[0].Name)

	// Submit the document
	err = machine.Trigger(ctx, document, fsm.Event{Name: "submit"}, "alice")
	if err != nil {
		log.Fatal(err)
	}

	// Check new state
	currentState, _ = machine.GetState(ctx, document)
	fmt.Printf("After submit: %s\n", currentState.Name)

	// Approve the document
	err = machine.Trigger(ctx, document, fsm.Event{Name: "approve"}, "bob")
	if err != nil {
		log.Fatal(err)
	}

	currentState, _ = machine.GetState(ctx, document)
	fmt.Printf("After approve: %s\n", currentState.Name)

	// Publish the document
	err = machine.Trigger(ctx, document, fsm.Event{Name: "publish"}, "charlie")
	if err != nil {
		log.Fatal(err)
	}

	currentState, _ = machine.GetState(ctx, document)
	fmt.Printf("Final state: %s\n", currentState.Name)

	// Get all transitions
	history, _ := machine.GetTransitions(ctx, document)
	fmt.Printf("Total transitions: %d\n", len(history))

	// Output:
	// Current state: draft
	// Available events: submit
	// After submit: submitted
	// After approve: approved
	// Final state: published
	// Total transitions: 4
}

func ExampleFSM_CanTrigger() {
	// Setup FSM (abbreviated)
	states := []fsm.State{{Name: "draft"}, {Name: "submitted"}}
	events := []fsm.Event{{Name: "submit"}, {Name: "approve"}}
	transitions := []fsm.Transition{
		{From: fsm.State{Name: "draft"}, To: fsm.State{Name: "submitted"}, Event: fsm.Event{Name: "submit"}},
	}
	storage := fsm.NewMemoryStorage()
	machine, _ := fsm.New(states, events, transitions, storage)

	ctx := context.Background()
	entity := fsm.Entity{Type: "doc", ID: "1"}

	// Start in draft state
	machine.Start(ctx, entity, fsm.State{Name: "draft"}, "user")

	// Check if events can be triggered
	canSubmit := machine.CanTrigger(ctx, entity, fsm.Event{Name: "submit"})
	canApprove := machine.CanTrigger(ctx, entity, fsm.Event{Name: "approve"})

	fmt.Printf("Can submit: %v\n", canSubmit)
	fmt.Printf("Can approve: %v\n", canApprove)

	// Output:
	// Can submit: true
	// Can approve: false
}

func ExampleFSM_GetNextState() {
	// Setup FSM
	states := []fsm.State{{Name: "draft"}, {Name: "submitted"}}
	events := []fsm.Event{{Name: "submit"}}
	transitions := []fsm.Transition{
		{From: fsm.State{Name: "draft"}, To: fsm.State{Name: "submitted"}, Event: fsm.Event{Name: "submit"}},
	}
	storage := fsm.NewMemoryStorage()
	machine, _ := fsm.New(states, events, transitions, storage)

	// Get next state without triggering
	nextState, err := machine.GetNextState(
		fsm.State{Name: "draft"},
		fsm.Event{Name: "submit"},
	)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Next state: %s\n", nextState.Name)

	// Output:
	// Next state: submitted
}
