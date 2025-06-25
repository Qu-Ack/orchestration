package deploy

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"math/rand"
	"os"
	"os/exec"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

const charset = "abcdefghijklmnopqrstuvwxyz" +
	"0123456789"

func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, fs.ErrNotExist) {
		return false, nil
	}
	return false, err
}

func StringWithCharset(length int, charset string) string {
	seededRand := rand.New(
		rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

func String(length int) string {
	return StringWithCharset(length, charset)
}

func constructProjectPath(id string) string {
	return fmt.Sprintf("/projects/%v", id)
}

func (d *DeployService) CreateDockerFile(deployment *Deployment, data DockerTemplateData, ProjectType int) error {

	fmt.Println("Project Type is ", ProjectType)

	f, err := os.Create(fmt.Sprintf("%v/Dockerfile", deployment.ProjectPath))

	if err != nil {
		return err
	}

	defer f.Close()

	switch ProjectType {
	case node:
		_, err = f.WriteString(ExecuteNodeTemplate(data))
		return err
	case next:
		_, err = f.WriteString(ExecuteNextTemplate(data))
		return err
	case golang:
		_, err := f.WriteString(ExecuteGoTemplate(data))
		return err
	case react:
		_, err := f.WriteString(ExecuteViteReactTemplate(data))
		return err
	default:
		return errors.New("invalid project type")
	}

}

func (d *DeployService) BuildImage(deployment *Deployment) error {
	cmd := exec.Command("docker", "build", "-t", fmt.Sprintf("%v-image", deployment.ID), deployment.ProjectPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (d *DeployService) ServiceDiscovery(deployment *Deployment) (int, error) {
	services := map[string]int{
		"src/index.js":    node,
		"index.js":        node,
		"next.config.ts":  next,
		"next.config.mjs": next,
		"go.mod":          golang,
		"vite.config.js":  react,
	}

	for k, v := range services {
		_, err := os.Stat(fmt.Sprintf("%v/%v", deployment.ProjectPath, k))

		if errors.Is(err, os.ErrNotExist) {
			continue
		}

		return v, nil
	}

	return -1, errors.New("no known service found")
}

func (d *DeployService) FindDockerFile(deployment *Deployment) bool {
	files := []string{"Dockerfile"}

	for _, val := range files {
		_, err := os.Stat(fmt.Sprintf("%v/%v", deployment.ProjectPath, val))

		if err == nil {
			return true
		}

	}

	return false
}

func (d *DeployService) GetCodeBase(deployment *Deployment) error {
	baseDir := "/projects"
	gitDirPath := fmt.Sprintf("%s/.git", deployment.ProjectPath)

	dirExists, err := exists(baseDir)
	if err != nil {
		return err
	}
	if !dirExists {
		err := os.Mkdir(baseDir, 0777)
		if err != nil {
			return err
		}
	}

	repoExists, err := exists(gitDirPath)
	if err != nil {
		return err
	}

	if repoExists {
		cmd := exec.Command("git", "-C", deployment.ProjectPath, "pull")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	} else {
		cmd := exec.Command("git", "clone", deployment.CloneUrl, deployment.ProjectPath)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}
}

func (d *DeployService) ContainerCreate(deployment *Deployment, dockerCli *client.Client) error {
	ctx := context.Background()

	_, err := dockerCli.ContainerInspect(ctx, deployment.ID)
	if err == nil {
		fmt.Println("{SERVER}: Removing existing container:", deployment.ID)
		err = dockerCli.ContainerRemove(ctx, deployment.ID, container.RemoveOptions{Force: true})
		if err != nil {
			fmt.Println("{SERVER}: Failed to remove existing container:", err.Error())
			return err
		}
	}

	// the orchestration_default network is the default network that traefik starts on in docker.
	// for other containers to be accessible by traefik they need to be on the same network.
	// therefore we assign the orchestration_default network to every network so that traefik can access it.

	// db-network will be a future plan to host a database container and making all the containers communicate to that db.

	labels := make(map[string]string, 0)

	labels["traefik.enable"] = "true"
	labels[fmt.Sprintf("traefik.http.routers.%v.rule", deployment.SubDomain)] = fmt.Sprintf("Host(`%v.localhost`)", deployment.SubDomain)
	labels[fmt.Sprintf("traefik.http.routers.%v.entrypoints", deployment.SubDomain)] = "web"
	labels["traefik.docker.network"] = "traefik_init_default"

	resp, err := dockerCli.ContainerCreate(ctx, &container.Config{
		Image: fmt.Sprintf("%v-image", deployment.ID),
		ExposedPorts: nat.PortSet{
			nat.Port(fmt.Sprintf("%v/tcp", deployment.Port)): struct{}{},
		},
		Labels: labels,
	}, &container.HostConfig{
		RestartPolicy: container.RestartPolicy{Name: "unless-stopped"},
	}, &network.NetworkingConfig{
		EndpointsConfig: map[string]*network.EndpointSettings{
			"db-network":           {},
			"traefik_init_default": {},
		},
	}, nil, deployment.ID)

	if err != nil {
		return err
	}

	err = dockerCli.ContainerStart(ctx, resp.ID, container.StartOptions{})
	if err != nil {
		return err
	}

	return nil
}
