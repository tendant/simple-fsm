package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"simple-fsm"
)

func main() {
	ctx := context.Background()

	// PostgreSQL connection string
	// Format: postgres://username:password@host:port/database
	connString := "postgres://fsm:password@localhost:5432/fsm_db"

	// Create PostgreSQL storage
	storage, err := fsm.NewPostgresStorage(ctx, connString)
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}
	defer storage.Close()

	// Define workflow states
	states := []fsm.State{
		{Name: "draft"},
		{Name: "submitted"},
		{Name: "approved"},
		{Name: "rejected"},
		{Name: "published"},
	}

	// Define events
	events := []fsm.Event{
		{Name: "submit"},
		{Name: "approve"},
		{Name: "reject"},
		{Name: "publish"},
		{Name: "revise"},
	}

	// Define transitions
	transitions := []fsm.Transition{
		{From: fsm.State{Name: "draft"}, To: fsm.State{Name: "submitted"}, Event: fsm.Event{Name: "submit"}},
		{From: fsm.State{Name: "submitted"}, To: fsm.State{Name: "approved"}, Event: fsm.Event{Name: "approve"}},
		{From: fsm.State{Name: "submitted"}, To: fsm.State{Name: "rejected"}, Event: fsm.Event{Name: "reject"}},
		{From: fsm.State{Name: "approved"}, To: fsm.State{Name: "published"}, Event: fsm.Event{Name: "publish"}},
		{From: fsm.State{Name: "rejected"}, To: fsm.State{Name: "draft"}, Event: fsm.Event{Name: "revise"}},
	}

	// Create FSM with PostgreSQL storage
	machine, err := fsm.New(states, events, transitions, storage)
	if err != nil {
		log.Fatalf("Failed to create FSM: %v", err)
	}

	// Create a document entity
	document := fsm.Entity{
		Type: "document",
		ID:   fmt.Sprintf("doc-%d", time.Now().Unix()),
	}

	fmt.Printf("Document ID: %s\n", document.ID)

	// Start document in draft state
	err = machine.Start(ctx, document, fsm.State{Name: "draft"}, "alice")
	if err != nil {
		log.Fatalf("Failed to start: %v", err)
	}
	fmt.Println("✓ Document started in 'draft' state")

	// Submit the document
	err = machine.Trigger(ctx, document, fsm.Event{Name: "submit"}, "alice")
	if err != nil {
		log.Fatalf("Failed to submit: %v", err)
	}
	fmt.Println("✓ Document submitted")

	// Check current state
	currentState, err := machine.GetState(ctx, document)
	if err != nil {
		log.Fatalf("Failed to get state: %v", err)
	}
	fmt.Printf("Current state: %s\n", currentState.Name)

	// Check available events
	availableEvents, err := machine.GetAvailableEvents(ctx, document)
	if err != nil {
		log.Fatalf("Failed to get available events: %v", err)
	}
	fmt.Printf("Available events: ")
	for i, event := range availableEvents {
		if i > 0 {
			fmt.Print(", ")
		}
		fmt.Print(event.Name)
	}
	fmt.Println()

	// Approve the document
	err = machine.Trigger(ctx, document, fsm.Event{Name: "approve"}, "bob")
	if err != nil {
		log.Fatalf("Failed to approve: %v", err)
	}
	fmt.Println("✓ Document approved by bob")

	// Publish the document
	err = machine.Trigger(ctx, document, fsm.Event{Name: "publish"}, "charlie")
	if err != nil {
		log.Fatalf("Failed to publish: %v", err)
	}
	fmt.Println("✓ Document published by charlie")

	// Get final state
	finalState, err := machine.GetState(ctx, document)
	if err != nil {
		log.Fatalf("Failed to get final state: %v", err)
	}
	fmt.Printf("Final state: %s\n", finalState.Name)

	// Get complete transition history
	history, err := machine.GetTransitions(ctx, document)
	if err != nil {
		log.Fatalf("Failed to get history: %v", err)
	}

	fmt.Printf("\nTransition History (%d transitions):\n", len(history))
	for i, t := range history {
		fmt.Printf("%d. %s -> %s (event: %s, by: %s, at: %s)\n",
			i+1,
			t.Transition.From.Name,
			t.Transition.To.Name,
			t.Transition.Event.Name,
			t.Transition.CreatedBy,
			t.Transition.CreatedAt.Format(time.RFC3339),
		)
	}

	fmt.Println("\n✓ All transitions persisted to PostgreSQL!")
}
