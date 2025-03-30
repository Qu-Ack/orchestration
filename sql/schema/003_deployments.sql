-- +goose Up
ALTER table deployments ADD COLUMN project_path varchar(1000) NOT NULL;


-- +goose Down
ALTER table deployments drop column if exists project_path;

