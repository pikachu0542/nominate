-- +goose Up
CREATE TYPE decision_status AS ENUM ('NA', 'Accepted', 'Declined');

CREATE TABLE IF NOT EXISTS open_position (
    position_id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    position VARCHAR(25),
    open_time TIMESTAMP,
    close_time TIMESTAMP,
    decision_deadline TIMESTAMP
);

CREATE TABLE nomination (
    id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    nominee_username VARCHAR(32),
    position_id INT REFERENCES open_position(position_id),
    status decision_status,
    submitted_time TIMESTAMP
);

-- +goose Down
DROP TABLE IF EXISTS open_position;
DROP TABLE IF EXISTS nomination;
DROP TYPE IF EXISTS decision_status;