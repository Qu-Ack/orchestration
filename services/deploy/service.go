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

	err = d.repo.addEnvVars(deployment)

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

func (d *DeployService) Deploy(deployment *Deployment, dockerCli *client.Client) error {
	err := d.GetCodeBase(deployment)

	if err != nil {
		log.Println("{SERVER}: ERROR IN FETCHING CODEBASE")
		log.Println(err.Error())
		return err
	}

	dockerFileExists := d.FindDockerFile(deployment)

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
			FilePath:       fmt.Sprintf("%v/src/index.js", deployment.ProjectPath),
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
