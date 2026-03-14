-- +goose Up
CREATE TABLE visits (
    id SERIAL PRIMARY KEY,
    link_id SERIAL references links(id) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    ip TEXT NOT NULL,
    user_agent TEXT NOT NULL,
    status SMALLINT NOT NULL
);

-- +goose Down
DROP TABLE visits;
