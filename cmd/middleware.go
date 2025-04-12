package main

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (s *Server) AuthMiddleware() gin.HandlerFunc {

	return func(c *gin.Context) {
		sesToken := c.GetHeader("Authorization")

		err := s.userService.Authenticate(sesToken)

		if err != nil {
			fmt.Println(err)
			c.JSON(http.StatusForbidden, gin.H{
				"status": "failiure",
				"error":  "Bad Authentication",
			})
			c.Abort()
			return
		}

		c.Next()
	}

}
