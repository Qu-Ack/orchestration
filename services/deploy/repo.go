package deploy

import "fmt"

func (r *DeployServiceRepo) addDeployment(deployment *Deployment) error {
	_, err := r.db.Exec("INSERT INTO deployments (id, subdomain, clone_url, branch, repo_name, project_type, port, project_path) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)", deployment.ID, deployment.SubDomain, deployment.CloneUrl, deployment.Branch, deployment.RepoName, deployment.ProjectType, deployment.Port, deployment.ProjectPath)

	if err != nil {
		return err
	}

	return nil
}

func (r *DeployServiceRepo) addEnvVars(deployment *Deployment, envs []EnvVar) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	stmt, err := tx.Prepare("INSERT INTO env_vars (deployment_id, key, value) VALUES ($1, $2, $3)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, env := range envs {
		fmt.Println(env.Key)
		if _, err := stmt.Exec(deployment.ID, env.Key, env.Value); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *DeployServiceRepo) updateEnvVar(deployment *Deployment, env EnvVar, newEnv EnvVar) error {
	_, err := r.db.Exec("UPDATE env_vars SET value = $1 WHERE deployment_id = $2 AND key = $3",
		newEnv.Value, deployment.ID, env.Key)
	return err
}

func (r *DeployServiceRepo) deleteEnvVar(deployment *Deployment, env EnvVar) error {
	_, err := r.db.Exec("DELETE FROM env_vars WHERE deployment_id = $1 AND key = $2",
		deployment.ID, env.Key)
	return err
}

func (r *DeployServiceRepo) getEnv(deployment *Deployment, env *EnvVar) error {
	var value string
	err := r.db.QueryRow("SELECT value FROM env_vars WHERE deployment_id = $1 AND key = $2",
		deployment.ID, env.Key).Scan(&value)
	if err != nil {
		return err
	}

	env.Value = value
	return nil
}

func (r *DeployServiceRepo) findDeploymentBasedOnSubdomain(subDomain string) error {
	var existingId string
	err := r.db.QueryRow("SELECT id from deployments WHERE subdomain=$1", subDomain).Scan(&existingId)
	return err
}

func (r *DeployServiceRepo) findDeploymentBasedOnId(id string) error {
	var existingId string
	err := r.db.QueryRow("SELECT id from deployments WHERE id=$1", id).Scan(&existingId)
	return err
}

func (r *DeployServiceRepo) findDeploymentBasedOnCloneUrl(CloneUrl string) error {
	var existingId string
	err := r.db.QueryRow("SELECT id from deployments WHERE clone_url=$1", CloneUrl).Scan(&existingId)
	return err
}

func (r *DeployServiceRepo) GetEnvVarsForDeployment(deploymentID string) ([]EnvVar, error) {
	envVarQuery := `
        SELECT key, value 
        FROM env_vars 
        WHERE deployment_id = $1`
	rows, err := r.db.Query(envVarQuery, deploymentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var envVars []EnvVar
	for rows.Next() {
		var ev EnvVar
		if err := rows.Scan(&ev.Key, &ev.Value); err != nil {
			continue
		}
		envVars = append(envVars, ev)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return envVars, nil
}

func (r *DeployServiceRepo) GetDeploymentBasedOnCloneUrl(cloneUrl string) (*Deployment, error) {
	deploymentQuery := `
        SELECT id, subdomain, clone_url, branch, repo_name, project_path, project_type, port 
        FROM deployments 
        WHERE clone_url = $1`
	row := r.db.QueryRow(deploymentQuery, cloneUrl)

	var dep Deployment
	err := row.Scan(
		&dep.ID,
		&dep.SubDomain,
		&dep.CloneUrl,
		&dep.Branch,
		&dep.RepoName,
		&dep.ProjectPath,
		&dep.ProjectType,
		&dep.Port,
	)
	if err != nil {
		return nil, err
	}

	envVars, err := r.GetEnvVarsForDeployment(dep.ID)
	if err != nil {
		return &dep, err
	}

	dep.EnvVars = envVars
	return &dep, nil
}

func (r *DeployServiceRepo) GetDeploymentByID(id string) (*Deployment, error) {
	deploymentQuery := `
        SELECT id, subdomain, clone_url, branch, repo_name, project_path, project_type, port 
        FROM deployments 
        WHERE id = $1`
	row := r.db.QueryRow(deploymentQuery, id)

	var dep Deployment
	err := row.Scan(
		&dep.ID,
		&dep.SubDomain,
		&dep.CloneUrl,
		&dep.Branch,
		&dep.RepoName,
		&dep.ProjectPath,
		&dep.ProjectType,
		&dep.Port,
	)
	if err != nil {
		return nil, err
	}

	envVars, err := r.GetEnvVarsForDeployment(dep.ID)
	if err != nil {
		return &dep, err
	}

	dep.EnvVars = envVars
	return &dep, nil
}
