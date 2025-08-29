package deploy

import (
	"database/sql"
	"errors"
	"fmt"
	"maps"
	"os"
	"path/filepath"

	"github.com/go-redis/redis"
	"gopkg.in/yaml.v3"
)

type service struct {
	repo struct {
		db    *sql.DB
		redis *redis.Client
	}
}

func NEW() *service {
	return &service{}
}

func (s *service) NEW_DEPLOYMENT(deployment *Deployment_New) error {

	switch deployment.Type {
	case DEP_DOCKER_COMPOSE:
		break
	default:
		for _, service := range deployment.Services {
			clonePath := s.getClonePath(service.ServiceID)
			err := s.gitClone(service.CodeRepo, clonePath)

			if err != nil {
				return err
			}

		}

	}

	return nil

}

func (s *service) DeployService(serv *Service) error {
	switch serv.ServiceType {
	case SER_NODE:
		err := s.gitClone(serv.CodeRepo, serv.ClonePath)

		if err != nil {
			return err
		}

		dockerFilePath := filepath.Join(serv.ClonePath, "Dockerfile")

		err = s.findFile(dockerFilePath)

		if err != nil {
			return err
		}

		dockerFile, err := os.Create(dockerFilePath)

		if err != nil {
			return err
		}

		_, err = dockerFile.WriteString(ExecuteNodeTemplate(DockerTemplateData{
			RepoIdentifier: serv.ServiceID,
			Port:           serv.Port,
			EnvVars:        serv.EnvVars,
		}))

		if err != nil {
			return err
		}

		err = s.buildDockerImage(serv.ClonePath, fmt.Sprintf("%v-%v", serv.ServiceID, serv.Domain))

		if err != nil {
			return err
		}

		break
	case SER_NEXT:
		break
	case SER_NEXT_PRISMA:
		break
	case SER_GO:
		break
	case SER_DOCKER:
		break
	default:
		return fmt.Errorf("Invalid service type")
	}
}

func (s *service) ValidateDeployment(req *User_Deployment_Request) (*User_Validation_response, error) {
	randId := s.getRandomString(6)
	clonePath := s.getClonePath(randId)

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

		deployment := &Deployment_New{
			ID:     randId,
			UserID: req.UserID,
			Name:   req.Name,
			Status: STATUS_PENDING,
			Services: []Service{
				{
					CodeRepo:       req.Repo,
					CodeRepoBranch: req.RepoBranch,
				},
			},
		}
		err := s.cacheDeployment(deployment)

		if err != nil {
			return nil, err
		}

		return &User_Validation_response{
			Status: "ok",
			Type:   req.Type.String(),
			Identified: map[string]any{
				"config":   filepath.Base(matches[0]),
				"filepath": req.FilePath,
			},
			Deployment: deployment,
		}, nil

	case USER_DEP_DOCKER:
		dockerfile := filepath.Join(clonePath, "Dockerfile")
		if err := s.findFile(dockerfile); err != nil {
			return nil, errors.New("Dockerfile not found")
		}

		deployment := &Deployment_New{
			ID:     randId,
			UserID: req.UserID,
			Name:   req.Name,
			Type:   DEP_DOCKER_FILE,
			Status: STATUS_PENDING,
			Services: []Service{
				{
					CodeRepo:       req.Repo,
					CodeRepoBranch: req.RepoBranch,
				},
			},
		}

		err := s.cacheDeployment(deployment)

		if err != nil {
			return nil, err
		}

		return &User_Validation_response{
			Status: "ok",
			Type:   req.Type.String(),
			Identified: map[string]any{
				"dockerfile": dockerfile,
				"filepath":   req.FilePath,
			},
			Deployment: deployment,
		}, nil

	case USER_DEP_GO:
		gomod := filepath.Join(clonePath, "go.mod")
		if err := s.findFile(gomod); err != nil {
			return nil, errors.New("go.mod not found")
		}

		deployment := &Deployment_New{
			UserID: req.UserID,
			ID:     randId,
			Name:   req.Name,
			Status: STATUS_PENDING,
			Services: []Service{
				{
					CodeRepo:       req.Repo,
					CodeRepoBranch: req.RepoBranch,
				},
			},
		}

		err := s.cacheDeployment(deployment)

		if err != nil {
			return nil, err
		}

		return &User_Validation_response{
			Status: "ok",
			Type:   req.Type.String(),
			Identified: map[string]any{
				"go_mod":   gomod,
				"filepath": req.FilePath,
			},
			Deployment: deployment,
		}, nil

	case USER_DEP_NODE:
		pkg := filepath.Join(clonePath, "package.json")
		if err := s.findFile(pkg); err != nil {
			return nil, errors.New("package.json not found")
		}

		deployment := &Deployment_New{
			ID:     randId,
			Name:   req.Name,
			UserID: req.UserID,
			Status: STATUS_PENDING,
			Services: []Service{
				{
					CodeRepo:       req.Repo,
					CodeRepoBranch: req.RepoBranch,
				},
			},
		}

		err := s.cacheDeployment(deployment)

		if err != nil {
			return nil, err
		}

		return &User_Validation_response{
			Status: "ok",
			Type:   req.Type.String(),
			Identified: map[string]any{
				"package_json": pkg,
				"filepath":     req.FilePath,
			},
			Deployment: deployment,
		}, nil

	case USER_DEP_PRISMA_NEXT:
		prisma := filepath.Join(clonePath, "prisma", "schema.prisma")
		if err := s.findFile(prisma); err != nil {
			return nil, errors.New("prisma/schema.prisma not found")
		}

		deployment := &Deployment_New{
			ID:     randId,
			UserID: req.UserID,
			Name:   req.Name,
			Status: STATUS_PENDING,
			Services: []Service{
				{
					CodeRepo:       req.Repo,
					CodeRepoBranch: req.RepoBranch,
				},
			},
		}

		err := s.cacheDeployment(deployment)

		if err != nil {
			return nil, err
		}

		return &User_Validation_response{
			Status: "ok",
			Type:   req.Type.String(),
			Identified: map[string]any{
				"schema":   prisma,
				"filepath": req.FilePath,
			},
			Deployment: deployment,
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

		var services []Service
		for serviceName := range compose.Services {
			services = append(services, Service{
				Name:           serviceName,
				CodeRepo:       req.Repo,
				CodeRepoBranch: req.RepoBranch,
			})
		}

		deployment := &Deployment_New{
			ID:       randId,
			Name:     req.Name,
			UserID:   req.UserID,
			Type:     DEP_DOCKER_COMPOSE,
			Status:   STATUS_PENDING,
			Services: services,
		}

		err = s.cacheDeployment(deployment)
		if err != nil {
			return nil, err
		}

		return &User_Validation_response{
			Status: "ok",
			Type:   req.Type.String(),
			Identified: map[string]any{
				"docker_compose": dockerCompose,
				"filepath":       req.FilePath,
				"services":       maps.Keys(compose.Services),
			},
			Deployment: deployment,
		}, nil
	default:
		return nil, errors.New("invalid user Deployment Request type")
	}
}

func (s *service) GET_PENDING_DEPLOYMENTS(userid string) ([]*Deployment_New, error) {

	deployments, err := s.getAllCachedDeploymentsOfUser(userid)

	if err != nil {
		return nil, err
	}

	return deployments, nil
}

func (s *service) NEW_SERVICE() {
}
