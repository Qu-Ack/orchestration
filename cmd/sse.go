package main

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

func (s *Server) SseEvents(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Expose-Headers", "Content-Type")

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")

	for {
		select {
		case update := <-s.sseChannel:
			fmt.Fprintf(c.Writer, "update:%s\n\n", update)
			c.Writer.Flush()
		case err := <-s.errorChannel:
			fmt.Fprintf(c.Writer, "error:%s\n\n", err)
			c.Writer.Flush()
		case <-c.Request.Context().Done():
			return
		}
	}
}
