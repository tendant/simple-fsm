package fsm

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresStorage implements Storage interface using PostgreSQL
type PostgresStorage struct {
	pool *pgxpool.Pool
}

// NewPostgresStorage creates a new PostgreSQL storage instance
// connString format: "postgres://username:password@localhost:5432/database_name"
func NewPostgresStorage(ctx context.Context, connString string) (*PostgresStorage, error) {
	pool, err := pgxpool.New(ctx, connString)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Test the connection
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &PostgresStorage{
		pool: pool,
	}, nil
}

// Close closes the database connection pool
func (p *PostgresStorage) Close() {
	p.pool.Close()
}

// SaveTransition saves a state transition to PostgreSQL
func (p *PostgresStorage) SaveTransition(ctx context.Context, et EntityTransition) error {
	query := `
		INSERT INTO entity_state_transition
		(entity_type, entity_id, from_state, to_state, event, created_by, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := p.pool.Exec(ctx, query,
		et.Entity.Type,
		et.Entity.ID,
		et.Transition.From.Name,
		et.Transition.To.Name,
		et.Transition.Event.Name,
		et.Transition.CreatedBy,
		et.Transition.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to save transition: %w", err)
	}

	return nil
}

// GetCurrentState retrieves the current state of an entity from PostgreSQL
func (p *PostgresStorage) GetCurrentState(ctx context.Context, entity Entity) (State, error) {
	query := `
		SELECT to_state
		FROM entity_state_transition
		WHERE entity_type = $1 AND entity_id = $2
		ORDER BY created_at DESC
		LIMIT 1
	`

	var stateName string
	err := p.pool.QueryRow(ctx, query, entity.Type, entity.ID).Scan(&stateName)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return State{}, ErrEntityNotFound
		}
		return State{}, fmt.Errorf("failed to get current state: %w", err)
	}

	return State{Name: stateName}, nil
}

// GetTransitions retrieves all transitions for an entity from PostgreSQL
func (p *PostgresStorage) GetTransitions(ctx context.Context, entity Entity) ([]EntityTransition, error) {
	query := `
		SELECT from_state, to_state, event, created_by, created_at
		FROM entity_state_transition
		WHERE entity_type = $1 AND entity_id = $2
		ORDER BY created_at ASC
	`

	rows, err := p.pool.Query(ctx, query, entity.Type, entity.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to query transitions: %w", err)
	}
	defer rows.Close()

	var transitions []EntityTransition
	for rows.Next() {
		var (
			fromState string
			toState   string
			event     string
			createdBy string
			createdAt time.Time
		)

		err := rows.Scan(&fromState, &toState, &event, &createdBy, &createdAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan transition row: %w", err)
		}

		transitions = append(transitions, EntityTransition{
			Entity: entity,
			Transition: Transition{
				From:      State{Name: fromState},
				To:        State{Name: toState},
				Event:     Event{Name: event},
				CreatedBy: createdBy,
				CreatedAt: createdAt,
			},
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating transition rows: %w", err)
	}

	return transitions, nil
}
