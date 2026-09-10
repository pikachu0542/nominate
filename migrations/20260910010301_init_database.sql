-- +goose Up

CREATE TABLE IF NOT EXISTS position (
    id SERIAL PRIMARY KEY,
    position_name VARCHAR(255) NOT NULL,
);

CREATE TABLE IF NOT EXISTS nomination (
    id SERIAL PRIMARY KEY,
    position_id INT NOT NULL REFERENCES position(id),
    period_id INT NOT NULL REFERENCES period(id),
    submitted_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS nomination_period (
    id SERIAL PRIMARY KEY,
    position_id INT NOT NULL REFERENCES position(id),
    opens_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    closes_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS nominated_user (
    id SERIAL PRIMARY KEY,
    username VARCHAR(255) NOT NULL,
    nomination_id INT NOT NULL REFERENCES nomination(id),
);

CREATE TABLE IF NOT EXISTS candidate (
    id SERIAL PRIMARY KEY,
    period_id INT NOT NULL REFERENCES period(id),
    position_id INT NOT NULL REFERENCES position(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS candidate_member (
    id SERIAL PRIMARY KEY,
    candidate_id INT NOT NULL REFERENCES candidate(id),
    username VARCHAR(255) NOT NULL
);

CREATE TYPE IF NOT EXISTS decision_status AS ENUM ('pending', 'accepted', 'declined');

CREATE TABLE IF NOT EXISTS candidate_decision (
    id SERIAL PRIMARY KEY,
    candidate_id INT NOT NULL REFERENCES candidate(id),
    username VARCHAR(255) NOT NULL,
    status decision_status NOT NULL,
    notified_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    responded_at TIMESTAMPTZ,
    response_deadline TIMESTAMPTZ NOT NULL,
    UNIQUE(candidate_id, username)
);

-- +goose Down
DROP TABLE IF EXISTS candidate_decision;
DROP TABLE IF EXISTS candidate_member;
DROP TABLE IF EXISTS candidate;

DROP TABLE IF EXISTS nominated_user;
DROP TABLE IF EXISTS nomination_period;
DROP TABLE IF EXISTS nomination;
DROP TABLE IF EXISTS position;

DROP TYPE IF EXISTS decision_status;