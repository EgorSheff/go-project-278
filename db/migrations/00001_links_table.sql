-- +goose Up
CREATE TABLE links (
    id SERIAL PRIMARY KEY,
    original_url text NOT NULL,
    short_name text NOT NULL UNIQUE
);

-- +goose Down
DROP TABLE links;
