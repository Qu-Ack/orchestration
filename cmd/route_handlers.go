package main

import (
	"github.com/Qu-Ack/orchestration/services/deploy"
	"github.com/gin-gonic/gin"
	"net/http"
)

func (s *Server) PostWebHook(c *gin.Context) {
	var json WebhookPayloadStruct

	err := c.ShouldBindJSON(&json)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	exists := s.deployService.CheckDeploymentExistenceBasedOnCloneUrl(json.Repository.CloneUrl)

	if exists {
		dep, err := s.deployService.GetDeploymentBasedOnCloneUrl(json.Repository.CloneUrl)

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}

		err = s.deployService.Deploy(dep, s.dockerCli)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

	} else {
		deployment, err := s.deployService.NewDeploymentFromWebhook(json.Repository.Name, json.Repository.CloneUrl, json.Ref)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		err = s.deployService.Deploy(deployment, s.dockerCli)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
	}

}

type EnvVar struct {
	Key   string
	Value string
}

func convertEnvvarsToDeployEnvvars(envs []EnvVar) []deploy.EnvVar {
	result := make([]deploy.EnvVar, len(envs))

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

	err = s.deployService.Deploy(deployment, s.dockerCli)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
}
