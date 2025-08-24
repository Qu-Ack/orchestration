package main

import (
	"fmt"
	"net/http"
	"strings"

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
		}

		ses, err := s.userService.GetSessionByID(sesToken)

		if err != nil {
			fmt.Println(err)
			c.JSON(http.StatusForbidden, gin.H{
				"status": "failiure",
				"error":  "Bad Authentication",
			})
			return
		}
		c.Set("session", ses.UserID)
		c.Next()
	}
}

func corsMiddleware() gin.HandlerFunc {
	originsString := "http://localhost:3000,http://orchestration.localhost,http://orchestration.test,http://orchestration.dakshsangal.live,https://orchestration.dakshsangal.live"
	var allowedOrigins []string
	if originsString != "" {
		allowedOrigins = strings.Split(originsString, ",")
	}

	return func(c *gin.Context) {
		isOriginAllowed := func(origin string, allowedOrigins []string) bool {
			for _, allowedOrigin := range allowedOrigins {
				if origin == allowedOrigin {
					return true
				}
			}
			return false
		}

		origin := c.Request.Header.Get("Origin")

		if isOriginAllowed(origin, allowedOrigins) {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
			c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
			c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")
		}

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
