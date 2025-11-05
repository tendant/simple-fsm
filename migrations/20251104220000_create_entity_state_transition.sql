-- +goose Up
-- +goose StatementBegin
-- Create entity_state_transition table for FSM state tracking
CREATE TABLE IF NOT EXISTS entity_state_transition (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at TIMESTAMP NOT NULL DEFAULT (NOW() AT TIME ZONE 'utc'),
    entity_type VARCHAR(255) NOT NULL,
    entity_id VARCHAR(255) NOT NULL,
    from_state VARCHAR(255),
    to_state VARCHAR(255) NOT NULL,
    event VARCHAR(255) NOT NULL,
    created_by VARCHAR(255)
);

-- Create index for fast entity lookups
CREATE INDEX IF NOT EXISTS idx_entity_state_transition_entity
    ON entity_state_transition(entity_type, entity_id, created_at DESC);

-- Create index for timeline queries
CREATE INDEX IF NOT EXISTS idx_entity_state_transition_created_at
    ON entity_state_transition(created_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Drop indexes
DROP INDEX IF EXISTS idx_entity_state_transition_created_at;
DROP INDEX IF EXISTS idx_entity_state_transition_entity;

-- Drop table
DROP TABLE IF EXISTS entity_state_transition;
-- +goose StatementEnd
