-- +goose Up
CREATE TABLE users (
    id VARCHAR PRIMARY KEY,
    username VARCHAR NOT NULL,
    password VARCHAR NOT NULL,
    created_at TIMESTAMP DEFAULT now()
);

CREATE TABLE sessions (
    id VARCHAR PRIMARY KEY,
    user_id VARCHAR NOT NULL,
    logged_in_at TIMESTAMP DEFAULT now(),
    expires_at TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE user_deployments (
    user_id VARCHAR NOT NULL,
    deployment_id VARCHAR NOT NULL,
    PRIMARY KEY (user_id, deployment_id),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (deployment_id) REFERENCES deployments(id) ON DELETE CASCADE
);

-- +goose Down
DROP TABLE IF EXISTS user_deployments;
DROP TABLE IF EXISTS sessions;
DROP TABLE IF EXISTS users;

