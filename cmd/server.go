package main

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/Qu-Ack/orchestration/services/deploy"
	"github.com/docker/docker/client"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

type Server struct {
	r             *gin.Engine
	dockerCli     *client.Client
	db            *sql.DB
	deployService *deploy.DeployService
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
	connStr := "user=postgres dbname=orchestration password=postgres sslmode=disable"
	db, err := sql.Open("postgres", connStr)

	if err != nil {
		log.Println(err.Error())
		log.Panic("{SERVER}: Error while connecting to database")
	}

	return db
}

func (s *Server) ServerCleanUp() {
	s.db.Close()
}

func NewServer() *Server {
	return &Server{
		r:         gin.Default(),
		dockerCli: NewDockerClient(),
		db:        NewDB(),
	}
}

func (s *Server) InstanitateServerServices() {
	s.deployService = deploy.NewDeployService(s.db)
}

func (s *Server) SetUpRoutes() {
	s.r.GET("/ping", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	s.r.POST("/webhook", s.PostWebHook)
	s.r.POST("/deploy", s.PostDeploy)
	s.r.PUT("/env/:deploymentid/:envid", s.PutEnv)
	s.r.DELETE("/env/:deploymentid/:envid", s.DeleteEnv)
	s.r.POST("/env/:deploymentid", s.PostEnv)
	s.r.PUT("/redeploy/:deploymentid", s.REDeploy)
}
