package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"github.com/Qu-Ack/orchestration/services/deploy"
	"github.com/Qu-Ack/orchestration/services/user"
	"github.com/docker/docker/client"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	_ "github.com/lib/pq"
)

type cfg struct {
	env string
}

type Server struct {
	r               *gin.Engine
	dockerCli       *client.Client
	cfg             *cfg
	redisCli        *redis.Client
	db              *sql.DB
	deployService   *deploy.DeployService
	userService     *user.UserService
	deployServicev2 *deploy.Dservice
	sseChannel      chan string
	errorChannel    chan string
}

func NewDockerClient() *client.Client {

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())

	if err != nil {
		log.Println(err.Error())
		log.Panic("{SERVER}: Error while initializing docker client")
	}

	return cli
}

func NewDB() *sql.DB {
	host := "localhost"
	port := 5433
	user := "postgres"
	password := "postgres"
	dbname := "orchestration"

	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Println(err.Error())
		log.Panic("Error while connecting to database")
	}

	err = db.Ping()
	if err != nil {
		log.Println(err.Error())
		log.Panic("Error while verifying connection to database")
	}

	return db
}

func (s *Server) ServerCleanUp() {
	s.db.Close()
}

func NewRedisClient() *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	return client
}

func NewServer() *Server {
	return &Server{
		r:          gin.Default(),
		dockerCli:  NewDockerClient(),
		redisCli:   NewRedisClient(),
		cfg:        &config,
		db:         NewDB(),
		sseChannel: make(chan string, 100),
	}
}

func (s *Server) InstanitateServerServices() {
	s.deployService = deploy.NewDeployService(s.db, s.cfg.env)
	s.deployServicev2 = deploy.NEW(s.dockerCli, s.cfg.env, s.db, s.redisCli)
	s.userService = user.NewUserService(s.db)
}

func (s *Server) SetUpRoutes() {
	s.r.GET("/ping", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	s.r.GET("/active/:deploymentid", s.GetOngoingDeployments)
	s.r.GET("/events", s.SseEvents)
	s.r.POST("/webhook", s.PostWebHook)
	s.r.POST("/deploy", s.AuthMiddleware(), s.PostDeploy)
	s.r.PUT("/env/:deploymentid/:envid", s.AuthMiddleware(), s.PutEnv)
	s.r.DELETE("/env/:deploymentid/:envid", s.AuthMiddleware(), s.DeleteEnv)
	s.r.POST("/env/:deploymentid", s.AuthMiddleware(), s.PostEnv)
	s.r.PUT("/redeploy/:deploymentid", s.AuthMiddleware(), s.REDeploy)
	s.r.POST("/login", s.PostLogin)
	//	s.r.POST("/register", s.PostUser)
	s.r.GET("/deployments/:userid", s.AuthMiddleware(), s.GetUserDeployments)
	s.r.GET("/deployment/:deploymentid", s.AuthMiddleware(), s.GetDeployment)
	s.r.GET("/deployment/:deploymentid/stats", s.AuthMiddleware(), s.GetContainerStats)
	s.r.GET("/deployment/:deploymentid/logs", s.AuthMiddleware(), s.GetContainerLogs)
	s.r.POST("/v2/deployment/validate", s.ValidateDeployment)
	s.r.POST("/v2/deployment/confirm", s.ConfirmDeployment)

}
