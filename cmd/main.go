package main

import (
	"flag"
	"fmt"
)

func main() {

	var cfg cfg

	flag.StringVar(&cfg.env, "env", "development", "server type")
	flag.Parse()
	fmt.Printf("started a %s server\n", cfg.env)

	server := NewServer()
	server.InstanitateServerServices()
	server.r.Use(server.corsMiddleware())
	server.SetUpRoutes()
	defer server.ServerCleanUp()
	server.r.Run(":5000")
}
