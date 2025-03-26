-- +goose Up
CREATE TABLE deployments (
    id VARCHAR PRIMARY KEY,
    clone_url VARCHAR NOT NULL,
    branch VARCHAR NOT NULL,
    repo_name VARCHAR NOT NULL
);

-- +goose Down
DROP TABLE deployments;


