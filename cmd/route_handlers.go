package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/Qu-Ack/orchestration/services/deploy"
	"github.com/gin-gonic/gin"
)

func (s *Server) PostWebHook(c *gin.Context) {

	c.JSON(http.StatusAccepted, gin.H{"message": "accepted"})

	var json WebhookPayloadStruct

	if err := c.ShouldBindJSON(&json); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	githubEvent := c.Request.Header.Get("x-github-event")
	fmt.Println(githubEvent)
	fmt.Println(json)

	go func() {
		if githubEvent == "push" {
			exists := s.deployService.CheckDeploymentExistenceBasedOnCloneUrl(json.Repository.CloneUrl)
			if exists {
				dep, err := s.deployService.GetDeploymentBasedOnCloneUrl(json.Repository.CloneUrl)
				if err != nil {
					fmt.Printf("Error fetching deployment: %v\n", err)
					return
				}
				if err := s.deployService.Deploy(dep, s.dockerCli, true); err != nil {
					fmt.Printf("Error deploying: %v\n", err)
					return
				}
			} else {
				deployment, err := s.deployService.NewDeploymentFromWebhook(json.Repository.Name, json.Repository.CloneUrl, json.Ref)
				if err != nil {
					fmt.Printf("Error creating new deployment: %v\n", err)
					return
				}
				if err := s.deployService.Deploy(deployment, s.dockerCli, true); err != nil {
					fmt.Printf("Error deploying new deployment: %v\n", err)
					return
				}
			}
		}
	}()
}

type EnvVar struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func convertEnvvarsToDeployEnvvars(envs []EnvVar) []deploy.EnvVar {
	result := make([]deploy.EnvVar, 0)

	for _, env := range envs {
		result = append(result, deploy.EnvVar{
			Key:   env.Key,
			Value: env.Value,
		})
	}

	return result
}

func (s *Server) PostDeploy(c *gin.Context) {
	type body struct {
		CloneUrl  string   `json:"clone_url"`
		RepoName  string   `json:"repo_name"`
		Branch    string   `json:"branch"`
		SubDomain string   `json:"subdomain"`
		EnvVars   []EnvVar `json:"envs"`
		Port      int      `json:"port"`
	}

	var json body
	err := c.ShouldBindJSON(&json)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "bad body",
		})
		return
	}

	deployment, err := s.deployService.NewDeployment(&deploy.Deployment{
		SubDomain: json.SubDomain,
		CloneUrl:  json.CloneUrl,
		Branch:    json.Branch,
		RepoName:  json.RepoName,
		EnvVars:   convertEnvvarsToDeployEnvvars(json.EnvVars),
		Port:      json.Port,
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	err = s.deployService.Deploy(deployment, s.dockerCli, false)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
}

func (s *Server) PutEnv(c *gin.Context) {

	type body struct {
		Value string `json:"value"`
	}
	var json body

	if err := c.ShouldBindJSON(&json); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "bad body",
		})
		return

	}

	deploymentId := c.Param("deploymentid")
	envId := c.Param("envid")

	err := s.deployService.UpdateEnvVar(&deploy.Deployment{
		ID: deploymentId,
	}, deploy.EnvVar{
		Key: envId,
	}, deploy.EnvVar{
		Value: json.Value,
	})

	if err != nil {
		fmt.Println("{SERVER}: Error In Updating Env")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Internal Server Error",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
	})
}

func (s *Server) DeleteEnv(c *gin.Context) {

	deploymentId := c.Param("deploymentid")
	envId := c.Param("envid")

	err := s.deployService.DeleteEnvVar(&deploy.Deployment{
		ID: deploymentId,
	}, deploy.EnvVar{
		Key: envId,
	})

	if err != nil {
		fmt.Println("{SERVER}: Error In Updating Env")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Internal Server Error",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
	})
}

func (s *Server) PostEnv(c *gin.Context) {

	type body struct {
		Envs []EnvVar `json:"envs"`
	}

	var json body

	if err := c.ShouldBindJSON(&json); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "bad body",
		})
		return
	}

	deploymentId := c.Param("deploymentid")

	exists := s.deployService.CheckDeploymentExistenceBasedOnId(deploymentId)

	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "deployment doesn't exist",
		})
		return
	}

	deployenvs := convertEnvvarsToDeployEnvvars(json.Envs)
	fmt.Println(deployenvs)
	err := s.deployService.AddEnvs(&deploy.Deployment{
		ID: deploymentId,
	}, deployenvs)

	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Internal Server Error",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
	})
}

func (s *Server) REDeploy(c *gin.Context) {
	deploymentId := c.Param("deploymentid")

	dep, err := s.deployService.GetDeploymentBasedOnID(deploymentId)

	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Internal Server Error",
		})
	}

	err = s.deployService.Deploy(dep, s.dockerCli, true)

	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Internal Server Error",
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
	})
}
