package main

import (
	"database/sql"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/docker/docker/client"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

const charset = "abcdefghijklmnopqrstuvwxyz" +
	"0123456789"

var seededRand = rand.New(
	rand.NewSource(time.Now().UnixNano()))

func StringWithCharset(length int, charset string) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

func String(length int) string {
	return StringWithCharset(length, charset)
}

type WebhookPayloadCommitAuthorStruct struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type WebhookPayloadCommitStruct struct {
	CommitHash    string                           `json:"id"`
	TreeHash      string                           `json:"tree_id"`
	CommitMessage string                           `json:"message"`
	CommitUrl     string                           `json:"url"`
	Author        WebhookPayloadCommitAuthorStruct `json:"author"`
}
type WebhookPayloadRepositoryStruct struct {
	Name     string `json:"full_name"`
	Url      string `json:"url"`
	CloneUrl string `json:"clone_url"`
}

type WebhookPayloadStruct struct {
	Ref        string                         `json:"ref"`
	Repository WebhookPayloadRepositoryStruct `json:"repository"`
	Commits    []WebhookPayloadCommitStruct   `json:"commits"`
}

func main() {
	r := gin.Default()

	connStr := "user=postgres dbname=orchestration password=postgres sslmode=disable"
	db, err := sql.Open("postgres", connStr)

	if err != nil {
		panic(err)
	}
	defer db.Close()

	fmt.Println("DB successfully connected")

	dockerCli, err = client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		fmt.Printf("Failed to create Docker client: %v\n", err)
		return
	}

	fmt.Println("{SERVER}: Docker Client Successfully created")

	r.GET("/ping", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	r.GET("/dbtest", func(ctx *gin.Context) {
		var existingID string
		err := db.QueryRow(
			"SELECT id FROM deployments LIMIT 1",
		).Scan(&existingID)

		if err != nil {
			fmt.Println("error in accessing db: ", err)
			return
		}

		ctx.JSON(http.StatusOK, gin.H{
			"data": existingID,
		})

	})

	r.POST("/deploy", func(ctx *gin.Context) {

	})

	r.POST("/webhook", func(ctx *gin.Context) {
		var json WebhookPayloadStruct

		if err := ctx.ShouldBindJSON(&json); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		ctx.Writer.WriteHeader(202)
		ctx.Writer.Write([]byte("accepted"))

		githubEvent := ctx.Request.Header.Get("x-github-event")

		var existingID string
		err := db.QueryRow(
			"SELECT id FROM deployments WHERE clone_url = $1",
			json.Repository.CloneUrl,
		).Scan(&existingID)

		fmt.Println(existingID)

		if err != nil {
			if err == sql.ErrNoRows {
				newID := String(6)
				_, err := db.Exec(
					"INSERT INTO deployments (id, repo_name, clone_url, branch) VALUES ($1, $2, $3, $4)",
					newID,
					json.Repository.Name,
					json.Repository.CloneUrl,
					json.Ref,
				)
				if err != nil {
					fmt.Println("Error creating deployment:", err)
					return
				}
				json.Id = newID
			} else {
				fmt.Println("Error checking deployments:", err)
				return
			}
		} else {
			json.Id = existingID
		}

		if githubEvent == "push" {
			fmt.Println("push event")
			err := DeploymentWorflow(json)
			if err != nil {
				fmt.Println("{SERVER}: ERROR IN DEPLOYMENT WORKFLOW", err.Error())
				return
			}
		} else {
			fmt.Println("received other event:", githubEvent)
		}
	})
	r.Run("localhost:5000")
}
