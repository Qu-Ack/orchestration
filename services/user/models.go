package user

import "time"

const (
	IDLENGTH         = 6
	SESSION_DURATION = time.Hour * 24
)

type User struct {
	ID        string
	Username  string
	Password  string
	CreatedAt time.Time
}

type Session struct {
	ID         string
	UserID     string
	LoggedInAt time.Time
	ExpiresAt  time.Time
	User       User
}

type UserDeployment struct {
	UserID       string
	DeploymentID string
	User         User
}
