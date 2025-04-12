package user

import "database/sql"

func (r *UserServiceRepo) UploadUser(user *User) error {
	query := `
		INSERT INTO users (id, username, password, created_at)
		VALUES ($1, $2, $3, $4)
	`
	_, err := r.db.Exec(query, user.ID, user.Username, user.Password, user.CreatedAt)
	return err
}

func (r *UserServiceRepo) GetUserByID(id string) (*User, error) {
	query := `
		SELECT id, username, password, created_at 
		FROM users 
		WHERE id = $1
	`
	row := r.db.QueryRow(query, id)
	var user User
	if err := row.Scan(&user.ID, &user.Username, &user.Password, &user.CreatedAt); err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserServiceRepo) GetUserByUsername(username string) (*User, error) {
	query := `
		SELECT id, username, password, created_at 
		FROM users 
		WHERE username = $1
	`
	row := r.db.QueryRow(query, username)
	var user User
	if err := row.Scan(&user.ID, &user.Username, &user.Password, &user.CreatedAt); err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserServiceRepo) CreateSession(session *Session) error {
	query := `
		INSERT INTO sessions (user_id, logged_in_at, expires_at)
		VALUES ($1, $2, $3)
		RETURNING id
	`
	return r.db.QueryRow(query, session.UserID, session.LoggedInAt, session.ExpiresAt).Scan(&session.ID)
}

func (r *UserServiceRepo) GetSessionByID(id string) (*Session, error) {
	query := `
		SELECT id, user_id, logged_in_at, expires_at
		FROM sessions
		WHERE id = $1
	`
	row := r.db.QueryRow(query, id)
	var session Session
	if err := row.Scan(&session.ID, &session.UserID, &session.LoggedInAt, &session.ExpiresAt); err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *UserServiceRepo) AddUserDeployment(ud *UserDeployment) error {
	query := `
		INSERT INTO user_deployments (user_id, deployment_id)
		VALUES ($1, $2)
	`
	_, err := r.db.Exec(query, ud.UserID, ud.DeploymentID)
	return err
}

func (r *UserServiceRepo) GetDeploymentsByUserID(userID string) ([]string, error) {
	query := `
		SELECT deployment_id
		FROM user_deployments
		WHERE user_id = $1
	`
	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var deployments []string
	for rows.Next() {
		var deploymentID string
		if err := rows.Scan(&deploymentID); err != nil {
			return nil, err
		}
		deployments = append(deployments, deploymentID)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return deployments, nil
}

func (r *UserServiceRepo) GetUserDeployment(userID string, deploymentID string) (*UserDeployment, error) {
	query := `
        SELECT user_id, deployment_id
        FROM user_deployments
        WHERE user_id = $1 AND deployment_id = $2
    `
	row := r.db.QueryRow(query, userID, deploymentID)
	var ud UserDeployment
	if err := row.Scan(&ud.UserID, &ud.DeploymentID); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &ud, nil
}
