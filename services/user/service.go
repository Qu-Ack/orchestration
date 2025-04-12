package user

import (
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type UserServiceRepo struct {
	db *sql.DB
}

type UserService struct {
	repo *UserServiceRepo
}

func newUserServiceRepo(db *sql.DB) *UserServiceRepo {
	return &UserServiceRepo{
		db: db,
	}
}

func NewUserService(db *sql.DB) *UserService {
	return &UserService{
		repo: newUserServiceRepo(db),
	}
}

func (u *UserService) CreateUser(user *User) (*User, error) {
	id := RandStringBytes(IDLENGTH)

	user.ID = id

	hashedPassword, err := HashPassword(user.Password)

	if err != nil {
		fmt.Println("ERROR WHILE HASHING THE USER PASSWORD")
		fmt.Println(err)
		return nil, err
	}

	user.Password = hashedPassword

	err = u.repo.UploadUser(user)

	if err != nil {
		fmt.Println("ERROR WHILE UPLOADING THE USER")
		fmt.Println(err)
		return nil, err
	}

	return user, nil
}

func (u *UserService) Login(user *User) (*Session, error) {
	existingUser, err := u.repo.GetUserByUsername(user.Username)

	if err != nil {
		fmt.Println("ERROR WHILE FETCHING THE USER BY USERNAME")
		fmt.Println(err)
		return nil, err
	}

	match := CheckPasswordHash(user.Password, existingUser.Password)

	if match {
		session := &Session{
			ID:         RandStringBytes(6),
			UserID:     existingUser.ID,
			LoggedInAt: time.Now(),
			ExpiresAt:  time.Now().Add(SESSION_DURATION),
		}
		err := u.repo.CreateSession(session)

		if err != nil {
			fmt.Println("ERROR WHILE CREATING SESSION FOR USER")
			fmt.Println(err)
			return nil, err
		}

		return session, nil
	} else {
		fmt.Println("ERROR PASSWORD HASHES DON'T MATCH")
		return nil, errors.New("incorrect password")
	}
}

func (u *UserService) Authenticate(sesId string) error {
	session, err := u.repo.GetSessionByID(sesId)

	if err != nil {
		fmt.Println("ERROR WHILE FETCHING THE SESSION WITH THE PROVIDED ID")
		fmt.Println(err)
		return err
	}

	if time.Now().After(session.ExpiresAt) {
		return errors.New("session expired")
	} else {
		return nil
	}
}

func (u *UserService) GetUserDeployments(userId string) ([]string, error) {
	_, err := u.repo.GetUserByID(userId)
	if err != nil {
		fmt.Println("ERROR WHILE FETCHING THE USER BY ID")
		fmt.Println(err)
		return nil, err
	}

	deployments, err := u.repo.GetDeploymentsByUserID(userId)

	if err != nil {
		fmt.Println("ERROR WHILE FETCHING THE DEPLOYMENTS BY USER ID")
		fmt.Println(err)
		return nil, err
	}

	return deployments, nil
}

func (u *UserService) GetSessionByID(sesId string) (*Session, error) {
	ses, err := u.repo.GetSessionByID(sesId)

	if err != nil {
		fmt.Println("ERROR WHILE FETCHING THE SESSION")
		fmt.Println(err)
		return nil, err
	}

	return ses, nil
}

func (u *UserService) GetUserDeployment(userId string, deploymentId string) (*UserDeployment, error) {
	ud, err := u.repo.GetUserDeployment(userId, deploymentId)

	if err != nil {
		fmt.Println("ERROR WHILE FETCHING THE USER DEPLOYMENT")
		fmt.Println(err)
		return nil, err
	}

	if ud == nil {
		fmt.Println("NO USER DEPLOYMENT FOUND")
		return nil, errors.New("no user deployment found")
	} else {
		return ud, nil
	}
}
