package deploy

import (
	"errors"
	"fmt"
	"maps"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type service struct {
}

func NEW() *service {
	return &service{}
}

func (s *service) NEW_DEPLOYMENT(userDeploymentRequest *User_Deployment_Request) {
	switch userDeploymentRequest.Type {
	case USER_DEP_DOCKER_COMPOSE:
		break

	case USER_DEP_DOCKER:
		break

	case USER_DEP_GO:
		break

	case USER_DEP_NEXT:
		break

	case USER_DEP_NODE:
		break

	case USER_DEP_PRISMA_NEXT:
		break

	}

}

func (s *service) ValidateDeployment(req *User_Deployment_Request) (*User_Validation_response, error) {
	clonePath := s.getClonePath()

	if err := s.gitClone(req.Repo, clonePath); err != nil {
		return nil, err
	}

	if filepath.IsAbs(req.FilePath) {
		return nil, errors.New("file_path must be relative to repo root")
	}
	fullPath := filepath.Join(clonePath, req.FilePath)

	if err := s.findFile(fullPath); err != nil {
		return nil, err
	}

	switch req.Type {
	case USER_DEP_NEXT:
		matches, _ := filepath.Glob(filepath.Join(clonePath, "next.config.*"))
		if len(matches) == 0 {
			return nil, errors.New("next.config file not found")
		}

		return &User_Validation_response{
			Status: "ok",
			Type:   req.Type.String(),
			Identified: map[string]any{
				"config":   filepath.Base(matches[0]),
				"filepath": req.FilePath,
			},
		}, nil

	case USER_DEP_DOCKER:
		dockerfile := filepath.Join(clonePath, "Dockerfile")
		if err := s.findFile(dockerfile); err != nil {
			return nil, errors.New("Dockerfile not found")
		}

		return &User_Validation_response{
			Status: "ok",
			Type:   req.Type.String(),
			Identified: map[string]any{
				"dockerfile": dockerfile,
				"filepath":   req.FilePath,
			},
		}, nil

	case USER_DEP_GO:
		gomod := filepath.Join(clonePath, "go.mod")
		if err := s.findFile(gomod); err != nil {
			return nil, errors.New("go.mod not found")
		}

		return &User_Validation_response{
			Status: "ok",
			Type:   req.Type.String(),
			Identified: map[string]any{
				"go_mod":   gomod,
				"filepath": req.FilePath,
			},
		}, nil

	case USER_DEP_NODE:
		pkg := filepath.Join(clonePath, "package.json")
		if err := s.findFile(pkg); err != nil {
			return nil, errors.New("package.json not found")
		}

		return &User_Validation_response{
			Status: "ok",
			Type:   req.Type.String(),
			Identified: map[string]any{
				"package_json": pkg,
				"filepath":     req.FilePath,
			},
		}, nil

	case USER_DEP_PRISMA_NEXT:
		prisma := filepath.Join(clonePath, "prisma", "schema.prisma")
		if err := s.findFile(prisma); err != nil {
			return nil, errors.New("prisma/schema.prisma not found")
		}

		return &User_Validation_response{
			Status: "ok",
			Type:   req.Type.String(),
			Identified: map[string]any{
				"schema":   prisma,
				"filepath": req.FilePath,
			},
		}, nil

	case USER_DEP_DOCKER_COMPOSE:
		dockerCompose := filepath.Join(clonePath, "docker-compose.yml")
		if err := s.findFile(dockerCompose); err != nil {
			return nil, errors.New("docker-compose.yml not found")
		}
		content, err := os.ReadFile(dockerCompose)
		if err != nil {
			return nil, fmt.Errorf("failed to read docker-compose.yml: %w", err)
		}

		var compose struct {
			Services map[string]any `yaml:"services"`
		}

		if err := yaml.Unmarshal(content, &compose); err != nil {
			return nil, fmt.Errorf("failed to parse docker-compose.yml: %w", err)
		}

		return &User_Validation_response{
			Status: "ok",
			Type:   req.Type.String(),
			Identified: map[string]any{
				"docker_compose": dockerCompose,
				"filepath":       req.FilePath,
				"services":       maps.Keys(compose.Services),
			},
		}, nil

	default:
		return nil, errors.New("invalid user Deployment Request type")
	}
}

func (s *service) NEW_SERVICE() {
}
