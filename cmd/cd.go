package main

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"text/template"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

var dockerCli *client.Client

type DockerTemplateData struct {
	BaseImage      string
	RepoIdentifier string
	Port           int
	FilePath       string
	EnvVars        []EnvVar
}

type EnvVar struct {
	Key   string
	Value string
}

const NodeDockerFileTemplate = `
FROM {{.BaseImage}}
WORKDIR /app
COPY . ./
{{range .EnvVars}}
ENV {{.Key}}={{.Value}}
{{end}}
RUN npm install
EXPOSE {{.Port}}
CMD ["node", "{{.FilePath}}"]
`

const NextjsDockerFileTemplate = `
FROM {{.BaseImage}}
WORKDIR /app

# Copy package files
COPY /projects/{{.RepoIdentifier}}/package*.json ./

{{range .EnvVars}}
ENV {{.Key}}={{.Value}}
{{end}}

# Install dependencies
RUN npm install --prefer-offline --no-audit --progress=false

# Copy project files
COPY /projects/{{.RepoIdentifier}}/ ./

# Build application
RUN npm run build

# Expose the Next.js port
EXPOSE {{.Port}}

# Start the application
CMD ["npm", "start"]
`

var NodeDockerFile = template.Must(template.New("").Parse(NodeDockerFileTemplate))
var NextDockerFile = template.Must(template.New("").Parse(NextjsDockerFileTemplate))

func ExecuteNodeTemplate(data DockerTemplateData) string {
	buf := bytes.Buffer{}
	if err := NodeDockerFile.Execute(&buf, data); err != nil {
		fmt.Println("Error", err)
	}
	return buf.String()
}

func ExecuteNextTemplate(data DockerTemplateData) string {
	buf := bytes.Buffer{}
	if err := NextDockerFile.Execute(&buf, data); err != nil {
		fmt.Println("Error", err)
	}
	return buf.String()
}

func DeploymentWorflow(deployment WebhookPayloadStruct) error {

	err := cloneUrl(deployment.Repository.CloneUrl, deployment.Id)

	if err != nil {
		fmt.Println("{SERVER}: error while cloning the repo")
		return err
	}

	dockerFileName := FindDockerFile(deployment.Id)

	if dockerFileName == "" {
		fmt.Printf("No docker file found")
		projectType, err := IdentifyProjectType(deployment.Id)

		if err != nil {
			fmt.Println("{SERVER}: error while identifying project type", err.Error())
			return err
		}

		if projectType == "node" {
			fmt.Println("node js type")
			f, err := os.Create(fmt.Sprintf("/projects/%v/Dockerfile", deployment.Id))
			if err != nil {
				fmt.Println("{SERVER}: error while creating docker file", err.Error())
				return err
			}

			defer f.Close()

			_, err = f.WriteString(ExecuteNodeTemplate(DockerTemplateData{
				BaseImage:      "node:22-alpine",
				Port:           3000,
				RepoIdentifier: deployment.Id,
				FilePath:       "./index.js",
				EnvVars:        make([]EnvVar, 0),
			}))

			if err != nil {
				return err
			}

		}

		err = BuildImage(fmt.Sprintf("%v-image", deployment.Id), fmt.Sprintf("/projects/%v", deployment.Id))

		if err != nil {
			fmt.Println("{SERVER}: error while building docker image", err.Error())
			return err
		}

		err = CreateContainer(fmt.Sprintf("%v-image", deployment.Id), deployment.Id)

		if err != nil {
			fmt.Println("{SERVER}: error while creating docker container", err.Error())
			return err
		}

		fmt.Println("docker container created successfully")

	} else {
		err = BuildImage(fmt.Sprintf("%v-image", deployment.Id), fmt.Sprintf("/projects/%v", deployment.Id))

		if err != nil {
			fmt.Println("{SERVER}: error while building docker image", err.Error())
			return err
		}

		err = CreateContainer(fmt.Sprintf("%v-image", deployment.Id), deployment.Id)
		if err != nil {
			fmt.Println("{SERVER}: error while creating docker container", err.Error())
			return err
		}
	}

	return nil
}

func CreateContainer(ImageName string, Id string) error {
	ctx := context.Background()

	_, err := dockerCli.ContainerInspect(ctx, Id)
	if err == nil {
		fmt.Println("{SERVER}: Removing existing container:", Id)
		err = dockerCli.ContainerRemove(ctx, Id, container.RemoveOptions{Force: true})
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
	labels["traefik.http.routers.my.rule"] = fmt.Sprintf("Host(`%v.localhost`)", Id)
	labels["traefik.http.routers.my.entrypoints"] = "web"
	labels["traefik.docker.network"] = "orchestration_default"

	resp, err := dockerCli.ContainerCreate(ctx, &container.Config{
		Image: ImageName,
		ExposedPorts: nat.PortSet{
			"3000/tcp": struct{}{},
		},
		Labels: labels,
	}, &container.HostConfig{
		RestartPolicy: container.RestartPolicy{Name: "unless-stopped"},
	}, &network.NetworkingConfig{
		EndpointsConfig: map[string]*network.EndpointSettings{
			"db-network":            {},
			"orchestration_default": {},
		},
	}, nil, Id)

	if err != nil {
		fmt.Println("{SERVER}: Error creating container:", err.Error())
		return err
	}

	err = dockerCli.ContainerStart(ctx, resp.ID, container.StartOptions{})
	if err != nil {
		fmt.Println("{SERVER}: Error starting container:", err.Error())
		return err
	}

	fmt.Println(fmt.Sprintf("{SERVER}: Container started on domain: http://%v.localhost", Id))

	return StreamContainerLogs(ctx, resp.ID)
}

func StreamContainerLogs(ctx context.Context, containerID string) error {
	out, err := dockerCli.ContainerLogs(ctx, containerID, container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     true,
	})

	if err != nil {
		fmt.Println("{SERVER}: Error getting logs:", err.Error())
		return err
	}
	defer out.Close()

	reader := bufio.NewReader(out)
	done := make(chan bool)

	go func() {
		for {
			line, err := reader.ReadString('\n')
			if err == io.EOF {
				break
			}
			if err != nil {
				fmt.Println("{SERVER}: Error reading log:", err.Error())
				break
			}
			fmt.Print(line)
		}
		done <- true
	}()

	select {
	case <-done:
		fmt.Println("{SERVER}: Log stream ended.")
	case <-time.After(30 * time.Second):
		fmt.Println("{SERVER}: Log streaming timeout, returning.")
	}

	return nil
}

func BuildImage(imageName string, projectPath string) error {
	cmd := exec.Command("docker", "build", "-t", imageName, projectPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func FindDockerFile(repoName string) string {
	files := []string{"docker-compose.yaml", "docker-compose.yml", "Dockerfile", "compose.yaml", "compose.yml"}

	for _, val := range files {
		_, err := os.Stat(fmt.Sprintf("/projects/%v/%v", repoName, val))

		if err == nil {
			return val
		}

	}

	return ""
}

func IdentifyProjectType(repoName string) (string, error) {
	services := map[string]string{
		"src/index.js":    "node",
		"next.config.ts":  "next",
		"next.config.mjs": "next",
		"vite.config.js":  "react",
	}

	for k, v := range services {
		_, err := os.Stat(fmt.Sprintf("/projects/%v/%v", repoName, k))

		if errors.Is(err, os.ErrNotExist) {
			continue
		}

		return v, nil
	}

	return "", errors.New("no known service found")
}

func cloneUrl(repoUrl string, deploymentId string) error {
	baseDir := "/projects"
	deploymentPath := fmt.Sprintf("%s/%s", baseDir, deploymentId)

	dirExists, err := exists(baseDir)
	if err != nil {
		fmt.Println("{SERVER}: Something happened", err.Error())
	}
	if !dirExists {
		err := os.Mkdir(baseDir, 0777)
		if err != nil {
			fmt.Println("{SERVER}: Something happened while creating directory", err.Error())
			return err
		}
	}

	repoExists, err := exists(deploymentPath)
	if err != nil {
		fmt.Println("{SERVER}: Something happened", err.Error())
	}

	if repoExists {
		cmd := exec.Command("git", "-C", deploymentPath, "pull")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	} else {
		cmd := exec.Command("git", "clone", repoUrl, deploymentPath)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}
}

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
