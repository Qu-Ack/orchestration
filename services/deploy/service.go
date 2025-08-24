package deploy

import (
	"context"
	"database/sql"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"sync"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

type DeployServiceRepo struct {
	db *sql.DB
}

type DeployService struct {
	repo DeployServiceRepo
	dsm  *DeploymentStateManager
}

func newDeployServiceRepo(db *sql.DB) *DeployServiceRepo {
	return &DeployServiceRepo{
		db: db,
	}
}

func newDeploymentStateManager() *DeploymentStateManager {
	return &DeploymentStateManager{
		mutex:  sync.RWMutex{},
		States: make(map[string]*DeploymentState),
	}
}

func NewDeployService(db *sql.DB) *DeployService {
	return &DeployService{
		repo: *newDeployServiceRepo(db),
		dsm:  newDeploymentStateManager(),
	}
}

func (d *DeployService) NewDeployment(deployment *Deployment) (*Deployment, error) {
	if d.CheckDeploymentExistenceBasedOnSubDomain(deployment.SubDomain) {
		return nil, errors.New("Deployment already exists")
	}

	deployment.ID = String(6)
	deployment.ProjectPath = constructProjectPath(deployment.ID)

	err := d.repo.addDeployment(deployment)
	if err != nil {
		fmt.Println("ERROR WHILE ADDING DEPLOYMENT")
		fmt.Println(err)
		return nil, err
	}

	err = d.repo.addEnvVars(deployment, deployment.EnvVars)

	if err != nil {
		fmt.Println("ERROR WHILE ADDING ENV VARS")
		fmt.Println(err)
		return nil, err
	}

	return deployment, nil
}

func (d *DeployService) NewDeploymentFromWebhook(RepoName string, CloneUrl string, Branch string) (*Deployment, error) {

	Id := String(6)

	deployment := Deployment{
		ID:          Id,
		SubDomain:   Id,
		CloneUrl:    CloneUrl,
		Branch:      Branch,
		RepoName:    RepoName,
		ProjectPath: constructProjectPath(Id),
		Port:        3000,
	}

	err := d.repo.addDeployment(&deployment)

	if err != nil {
		return nil, err
	}

	return &deployment, nil
}

func (d *DeployService) CheckDeploymentExistenceBasedOnCloneUrl(CloneUrl string) bool {
	err := d.repo.findDeploymentBasedOnCloneUrl(CloneUrl)

	if err != nil {
		log.Println(err)
		return false
	}
	return true
}

func (d *DeployService) CheckDeploymentExistenceBasedOnId(id string) bool {
	err := d.repo.findDeploymentBasedOnId(id)

	if err != nil {
		log.Println(err)
		return false
	}

	return true
}

func (d *DeployService) CheckDeploymentExistenceBasedOnSubDomain(SubDomain string) bool {
	err := d.repo.findDeploymentBasedOnSubdomain(SubDomain)

	if err != nil {
		log.Println(err)
		return false
	}
	return true
}

func (d *DeployService) GetDeploymentBasedOnCloneUrl(CloneUrl string) (*Deployment, error) {
	dep, err := d.repo.GetDeploymentBasedOnCloneUrl(CloneUrl)

	if err != nil {
		return nil, err
	}

	return dep, nil
}

func (d *DeployService) GetDeploymentBasedOnID(deploymentId string) (*Deployment, error) {
	dep, err := d.repo.GetDeploymentByID(deploymentId)

	if err != nil {
		return nil, err
	}

	return dep, nil
}

func (d *DeployService) UpdateEnvVar(deployment *Deployment, oldEnv EnvVar, newEnv EnvVar) error {

	err := d.repo.getEnv(deployment, &oldEnv)

	if err != nil {
		return err
	}

	err = d.repo.updateEnvVar(deployment, oldEnv, newEnv)

	if err != nil {
		return nil
	}

	return err
}

func (d *DeployService) AddEnvs(deployment *Deployment, envs []EnvVar) error {
	err := d.repo.addEnvVars(deployment, envs)

	if err != nil {
		return err
	}

	return nil
}

func (d *DeployService) DeleteEnvVar(deployment *Deployment, env EnvVar) error {

	err := d.repo.getEnv(deployment, &env)

	if err != nil {
		return err
	}

	err = d.repo.deleteEnvVar(deployment, env)

	if err != nil {
		return nil
	}

	return err
}

func (d *DeployService) Deploy(deployment *Deployment, dockerCli *client.Client, redeploy bool, sse chan string, errsse chan string) error {
	err := d.GetCodeBase(deployment)
	if err != nil {
		log.Println("{SERVER}: ERROR IN FETCHING CODEBASE")
		log.Println(err.Error())
		sendEvent(errsse, fmt.Sprintf("%s:%s:%s", deployment.ID, deployment.SubDomain, "codebase clone failed"))
		return err
	}
	sendEvent(sse, fmt.Sprintf("%s:%s:%s", deployment.ID, deployment.SubDomain, "codebase cloned"))

	dockerFileExists := d.FindDockerFile(deployment)
	if redeploy {
		dockerFileExists = false
	}

	if dockerFileExists {
		err := d.BuildImage(deployment)
		if err != nil {
			log.Println("{SERVER}: ERROR IN BUILDING IMAGE")
			log.Println(err.Error())
			sendEvent(errsse, fmt.Sprintf("%s:%s:%s", deployment.ID, deployment.SubDomain, "docker image build failed"))
			return err
		}
		sendEvent(sse, fmt.Sprintf("%s:%s:%s", deployment.ID, deployment.SubDomain, "docker image built"))

		err = d.ContainerCreate(deployment, dockerCli)
		if err != nil {
			log.Println("{SERVER}: ERROR IN STARTING CONTAINER")
			log.Println(err.Error())
			sendEvent(errsse, fmt.Sprintf("%s:%s:%s", deployment.ID, deployment.SubDomain, "container creation failed"))
			return err
		}
		sendEvent(sse, fmt.Sprintf("%s:%s:%s", deployment.ID, deployment.SubDomain, "deployment successful"))
		return nil
	} else {
		service, err := d.ServiceDiscovery(deployment)
		if err != nil {
			log.Println("{SERVER}: ERROR IN SERVICE DISCOVERY")
			log.Println(err.Error())
			sendEvent(errsse, fmt.Sprintf("%s:%s:%s", deployment.ID, deployment.SubDomain, "service discovery failed"))
			return err
		}
		sendEvent(sse, fmt.Sprintf("%s:%s:%s", deployment.ID, deployment.SubDomain, "service discovered"))

		err = d.CreateDockerFile(deployment, DockerTemplateData{
			Port:           deployment.Port,
			RepoIdentifier: deployment.ID,
			EnvVars:        deployment.EnvVars,
		}, service)
		if err != nil {
			log.Println("{SERVER}: ERROR IN CREATING DOCKER FILE")
			log.Println(err.Error())
			sendEvent(errsse, fmt.Sprintf("%s:%s:%s", deployment.ID, deployment.SubDomain, "docker file creation failed"))
			return err
		}
		sendEvent(sse, fmt.Sprintf("%s:%s:%s", deployment.ID, deployment.SubDomain, "docker file created"))

		err = d.BuildImage(deployment)
		if err != nil {
			log.Println("{SERVER}: ERROR IN BUILDING IMAGE")
			log.Println(err.Error())
			sendEvent(errsse, fmt.Sprintf("%s:%s:%s", deployment.ID, deployment.SubDomain, "docker image build failed"))
			return err
		}
		sendEvent(sse, fmt.Sprintf("%s:%s:%s", deployment.ID, deployment.SubDomain, "docker image built"))

		err = d.ContainerCreate(deployment, dockerCli)
		if err != nil {
			log.Println("{SERVER}: ERROR IN STARTING CONTAINER")
			log.Println(err.Error())
			sendEvent(errsse, fmt.Sprintf("%s:%s:%s", deployment.ID, deployment.SubDomain, "container creation failed"))
			return err
		}
		sendEvent(sse, fmt.Sprintf("%s:%s:%s", deployment.ID, deployment.SubDomain, "deployment successful"))
		err = d.DSM_DeleteDeployment(deployment.ID)

		if err != nil {
			log.Println("{SERVER}: ERROR IN DEL FROM DSM")
			log.Println(err.Error())
			sendEvent(errsse, fmt.Sprintf("%s:%s:%s", deployment.ID, deployment.SubDomain, "status update failed"))
			return err
		}
		return nil
	}
}

func (d *DeployService) GetDeploymentStats(deployment *Deployment, dockercli *client.Client, ctx context.Context) (*ContainerStats, error) {
	containers, err := dockercli.ContainerList(ctx, container.ListOptions{})

	if err != nil {
		fmt.Println("ERROR WHILE GETTING CONTAINER LIST")
		fmt.Println(err)
		return nil, err
	}

	var containerId string
	var state string
	for _, container := range containers {
		if container.Image == fmt.Sprintf("%v-image", deployment.ID) {
			containerId = container.ID
			state = container.State
			break
		}
	}

	if containerId == "" {
		return &ContainerStats{
			Status: "STOPPED",
		}, nil
	}

	containerStats := &ContainerStats{
		Status: state,
	}

	if state != "running" {
		return containerStats, nil
	}

	statsReader, err := dockercli.ContainerStats(ctx, containerId, false)

	if err != nil {
		fmt.Println("ERROR WHILE READING STATS")
		fmt.Println(err)
		return nil, err
	}

	readbytes, err := io.ReadAll(statsReader.Body)

	if err != nil {
		fmt.Println("ERROR WHILE READING STATS")
		fmt.Println(err)
		return nil, err
	}

	var dockerStats container.StatsResponse

	err = json.Unmarshal(readbytes, &dockerStats)

	if err != nil {
		fmt.Println("ERROR WHILE UNMARSHLING INTO JSON")
		fmt.Println(err)
		return nil, err
	}
	cpuDelta := float64(dockerStats.CPUStats.CPUUsage.TotalUsage - dockerStats.PreCPUStats.CPUUsage.TotalUsage)
	systemDelta := float64(dockerStats.CPUStats.SystemUsage - dockerStats.PreCPUStats.SystemUsage)

	if systemDelta > 0 && cpuDelta > 0 {
		numCPUs := float64(len(dockerStats.CPUStats.CPUUsage.PercpuUsage))
		if numCPUs == 0 {
			numCPUs = 1
		}
		containerStats.CPUUsage = (cpuDelta / systemDelta) * numCPUs * 100.0
	}

	containerStats.MemoryUsage = int64(dockerStats.MemoryStats.Usage)
	containerStats.MemoryLimit = int64(dockerStats.MemoryStats.Limit)

	containerStats.NetworkRx = 0
	containerStats.NetworkTx = 0

	for _, network := range dockerStats.Networks {
		containerStats.NetworkRx += int64(network.RxBytes)
		containerStats.NetworkTx += int64(network.TxBytes)
	}

	return containerStats, nil
}

func (d *DeployService) GetContainerLogs(deployment *Deployment, dockercli *client.Client) ([]string, error) {
	containerName := deployment.ID

	options := container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     false,
		Timestamps: true,
		Tail:       "100",
	}

	logReader, err := dockercli.ContainerLogs(context.Background(), containerName, options)
	if err != nil {
		return nil, fmt.Errorf("error getting container logs: %v", err)
	}
	defer logReader.Close()

	logs, err := processMuxedLogs(logReader)
	if err != nil {
		return nil, fmt.Errorf("error processing container logs: %v", err)
	}

	return logs, nil
}

func processMuxedLogs(reader io.Reader) ([]string, error) {
	var logs []string

	header := make([]byte, 8)
	for {
		_, err := io.ReadFull(reader, header)
		if err != nil {
			if err == io.EOF {
				break
			}
			return logs, err
		}

		size := binary.BigEndian.Uint32(header[4:])

		logEntry := make([]byte, size)
		_, err = io.ReadFull(reader, logEntry)
		if err != nil {
			return logs, err
		}

		logs = append(logs, string(logEntry))
	}

	return logs, nil
}
