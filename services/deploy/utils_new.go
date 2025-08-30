package deploy

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/go-connections/nat"
)

func (s *Dservice) findFile(filePath string) error {
	_, err := os.Stat(filePath)

	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("file does not exist")
		}

		return errors.New("unexpected error")
	}
	return err
}

func (s *Dservice) gitClone(repo string, clonePath string) error {
	var stderr, stdout bytes.Buffer

	if err := os.MkdirAll(clonePath, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", clonePath, err)
	}

	cmd := exec.Command("git", "clone", repo, ".")
	cmd.Dir = clonePath
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("git clone failed %v: stderr: %v", err, stderr.String())
	}

	return nil
}

func (s *Dservice) removeDir(dirPath string) error {
	err := os.RemoveAll(dirPath)
	if err != nil {
		return fmt.Errorf("failed to remove directory %s: %w", dirPath, err)
	}
	return nil
}

const charMap = "abcdefghijklmnopqrstuvwxyz" + "0123456789"

func randomStringWithCharMap(length int, charset string) string {
	seededRand := rand.New(
		rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

func (s *Dservice) getRandomString(length int) string {
	return randomStringWithCharMap(length, charMap)
}

func (s *Dservice) getClonePath(id string) string {
	return fmt.Sprintf("/projects/%v", id)
}

func (s *Dservice) getCacheKey(id string, userid string) string {
	return fmt.Sprintf("deployment:%s:%s", id, userid)
}

func (s *Dservice) buildDockerImage(dockerFilePath string, imgtag string) error {
	var stderr, stdout bytes.Buffer

	cmd := exec.Command("docker", "build", "-t", imgtag, dockerFilePath)
	cmd.Dir = dockerFilePath
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("docker image build failed %v: stderr: %v", err, stderr.String())
	}

	return nil

}

func (s *Dservice) getLabelsForContainers(serv *Service) map[string]string {
	labels := make(map[string]string, 0)

	if s.env == "production" {
		labels["traefik.enable"] = "true"
		labels[fmt.Sprintf("traefik.http.routers.%v-web.rule", serv.Domain)] =
			fmt.Sprintf("Host(`%v.dakshsangal.live`)", serv.Domain)
		labels[fmt.Sprintf("traefik.http.routers.%v-web.entrypoints", serv.Domain)] = "web"

		labels[fmt.Sprintf("traefik.http.routers.%v-websecure.rule", serv.Domain)] =
			fmt.Sprintf("Host(`%v.dakshsangal.live`)", serv.Domain)
		labels[fmt.Sprintf("traefik.http.routers.%v-websecure.entrypoints", serv.Domain)] = "websecure"
		labels[fmt.Sprintf("traefik.http.routers.%v-websecure.tls", serv.Domain)] = "true"
		labels[fmt.Sprintf("traefik.http.routers.%v-websecure.tls.certresolver", serv.Domain)] = "letsencrypt"
		labels["traefik.docker.network"] = "traefik_init_default"

	} else {
		labels["traefik.enable"] = "true"
		labels[fmt.Sprintf("traefik.http.routers.%v-web.rule", serv.Domain)] =
			fmt.Sprintf("Host(`%v.localhost`)", serv.Domain)
		labels[fmt.Sprintf("traefik.http.routers.%v-web.entrypoints", serv.Domain)] = "web"
		labels["traefik.docker.network"] = "traefik_init_default"
	}

	return labels
}

func (s *Dservice) StartDockerContainer(serv *Service) error {

	containerLabels := s.getLabelsForContainers(serv)

	imageName := fmt.Sprintf("%v-%v", serv.ServiceID, serv.Domain)
	resp, err := s.dockerCli.ContainerCreate(context.Background(), &container.Config{
		Image: imageName,
		ExposedPorts: nat.PortSet{
			nat.Port(fmt.Sprintf("%v/tcp", serv.Port)): struct{}{},
		},
		Labels: containerLabels,
	}, &container.HostConfig{
		RestartPolicy: container.RestartPolicy{Name: "unless-stopped"},
	}, &network.NetworkingConfig{
		EndpointsConfig: map[string]*network.EndpointSettings{
			"db-network":           {},
			"traefik_init_default": {},
		},
	}, nil, serv.ServiceID)

	if err != nil {
		return err
	}

	if err := s.dockerCli.ContainerStart(context.Background(), resp.ID, container.StartOptions{}); err != nil {
		return err
	}

	serv.containerId = resp.ID
	serv.logFilePath = fmt.Sprintf("/var/log/%s.log", serv.ServiceID)
	return nil
}
