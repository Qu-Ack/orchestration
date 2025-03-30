package deploy

func (r *DeployServiceRepo) addDeployment(deployment *Deployment) error {
	_, err := r.db.Exec("INSERT INTO deployments (id, subdomain, clone_url, branch, repo_name, project_type, port, project_path) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)", deployment.ID, deployment.SubDomain, deployment.CloneUrl, deployment.Branch, deployment.RepoName, deployment.ProjectType, deployment.Port, deployment.ProjectPath)

	if err != nil {
		return err
	}

	return nil
}

func (r *DeployServiceRepo) addEnvVars(deployment *Deployment) error {
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

	for _, env := range deployment.EnvVars {
		if _, err := stmt.Exec(deployment.ID, env.Key, env.Value); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *DeployServiceRepo) findDeploymentBasedOnSubdomain(subDomain string) error {
	var existingId string
	err := r.db.QueryRow("SELECT id from deployments WHERE subdomain=$1", subDomain).Scan(&existingId)
	return err
}

func (r *DeployServiceRepo) findDeploymentBasedOnCloneUrl(CloneUrl string) error {
	var existingId string
	err := r.db.QueryRow("SELECT id from deployments WHERE clone_url=$1", CloneUrl).Scan(&existingId)
	return err
}

func (r *DeployServiceRepo) GetDeploymentBasedOnCloneUrl(CloneUrl string) (*Deployment, error) {
	deploymentQuery := `
        SELECT id, subdomain, clone_url, branch, repo_name, project_path, project_type, port 
        FROM deployments 
        WHERE clone_url = $1`
	row := r.db.QueryRow(deploymentQuery, CloneUrl)

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

	envVarQuery := `
        SELECT key, value 
        FROM env_vars 
        WHERE deployment_id = $1`
	rows, err := r.db.Query(envVarQuery, dep.ID)
	if err != nil {
		return &dep, err
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

	dep.EnvVars = envVars
	return &dep, nil
}
