-- +goose Up
CREATE TABLE deployments (
    id VARCHAR(255) NOT NULL,
    subdomain VARCHAR(255) NOT NULL,
    clone_url VARCHAR(255) NOT NULL,
    branch VARCHAR(255) NOT NULL,
    repo_name VARCHAR(255) NOT NULL,
    project_type INT,
    port INT,
    PRIMARY KEY (id),
    UNIQUE (subdomain),
    UNIQUE (clone_url)
);

CREATE TABLE env_vars (
    deployment_id VARCHAR(255) NOT NULL,
    key VARCHAR(255) NOT NULL,
    value VARCHAR(255),
    PRIMARY KEY (deployment_id, key),
    FOREIGN KEY (deployment_id) REFERENCES deployments(id) ON DELETE CASCADE
);

-- +goose Down
DROP TABLE env_vars;
DROP TABLE deployments;

