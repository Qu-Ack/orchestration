package deploy

import (
	"database/sql"
	"errors"
	"fmt"
	"log"

	"github.com/docker/docker/client"
)

type DeployServiceRepo struct {
	db *sql.DB
}

type DeployService struct {
	repo DeployServiceRepo
}

func newDeployServiceRepo(db *sql.DB) *DeployServiceRepo {
	return &DeployServiceRepo{
		db: db,
	}
}

func NewDeployService(db *sql.DB) *DeployService {
	return &DeployService{
		repo: *newDeployServiceRepo(db),
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
		return nil, err
	}

	err = d.repo.addEnvVars(deployment, deployment.EnvVars)

	if err != nil {
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

func (d *DeployService) Deploy(deployment *Deployment, dockerCli *client.Client, redeploy bool) error {
	err := d.GetCodeBase(deployment)

	if err != nil {
		log.Println("{SERVER}: ERROR IN FETCHING CODEBASE")
		log.Println(err.Error())
		return err
	}

	dockerFileExists := d.FindDockerFile(deployment)

	if redeploy {
		dockerFileExists = false
	}

	if dockerFileExists {
		err := d.BuildImage(deployment)

		if err != nil {
			log.Println("{SERVER}: ERROR IN BUILDING IMAGE")
			log.Println(err.Error())
			return err
		}

		err = d.ContainerCreate(deployment, dockerCli)

		if err != nil {
			log.Println("{SERVER}: ERROR IN STARTING CONTAINER")
			log.Println(err.Error())
			return err
		}

		fmt.Println(fmt.Sprintf("{SERVER}: APPLICATION STARTED ON URL : http://%v.localhost", deployment.SubDomain))

		return nil
	} else {

		service, err := d.ServiceDiscovery(deployment)

		if err != nil {
			log.Println("{SERVER}: ERROR IN SERVICE DISCOVERY")
			log.Println(err.Error())
			return err
		}

		err = d.CreateDockerFile(deployment, DockerTemplateData{
			Port:           deployment.Port,
			RepoIdentifier: deployment.ID,
			EnvVars:        deployment.EnvVars,
		}, service)

		if err != nil {
			log.Println("{SERVER}: ERROR IN CREATING DOCKER FILE")
			log.Println(err.Error())
			return err
		}

		err = d.BuildImage(deployment)

		if err != nil {
			log.Println("{SERVER}: ERROR IN BUILDING IMAGE")
			log.Println(err.Error())
			return err
		}

		err = d.ContainerCreate(deployment, dockerCli)

		if err != nil {
			log.Println("{SERVER}: ERROR IN STARTING CONTAINER")
			log.Println(err.Error())
			return err
		}

		fmt.Println(fmt.Sprintf("{SERVER}: APPLICATION STARTED ON URL : http://%v.localhost", deployment.SubDomain))

		return nil
	}
}
