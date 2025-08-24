package deploy

import (
	"fmt"
	"time"
)

func (d *DeployService) DSM_SetDeploying(deploymentId string) error {
	d.dsm.mutex.Lock()
	defer d.dsm.mutex.Unlock()

	val, ok := d.dsm.States[deploymentId]

	if !ok {
		d.dsm.States[deploymentId] = &DeploymentState{
			Status:    StatusDeploying,
			StartTime: time.Now(),
			Message:   "Deploying..",
		}
		fmt.Println("successfully added deployment to dsm")
		return nil
	} else {
		if val.Status == StatusDeploying {
			return fmt.Errorf("already deploying")
		}

		val.Status = StatusDeploying
		return nil
	}
}

func (d *DeployService) DSM_DeleteDeployment(deploymentId string) error {
	d.dsm.mutex.Lock()
	defer d.dsm.mutex.Unlock()

	_, ok := d.dsm.States[deploymentId]

	if !ok {
		return fmt.Errorf("deployment doesn't exist")
	}

	delete(d.dsm.States, deploymentId)
	fmt.Println("successfully removed deployment to dsm")
	return nil
}

func (d *DeployService) DSM_GetDeploymentState(deploymentId string) (*DeploymentState, error) {
	d.dsm.mutex.RLock()

	defer d.dsm.mutex.RUnlock()

	val, ok := d.dsm.States[deploymentId]

	if !ok {
		return nil, fmt.Errorf("no such deployment with id: %s", deploymentId)
	}

	return val, nil
}

func (d *DeployService) DSM_GetOngoingDeployments() map[string]*DeploymentState {
	d.dsm.mutex.RLock()
	defer d.dsm.mutex.RUnlock()

	ongoingDeployments := make(map[string]*DeploymentState)
	for k, v := range d.dsm.States {
		if v.Status == StatusDeploying {
			stateCopy := &DeploymentState{
				Status:    v.Status,
				StartTime: v.StartTime,
				Message:   v.Message,
			}
			ongoingDeployments[k] = stateCopy
		}
	}

	return ongoingDeployments
}
