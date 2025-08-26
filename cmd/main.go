package main

import (
	"flag"
	"fmt"
)

var config cfg

func main() {

	flag.StringVar(&config.env, "env", "development", "server type")
	flag.Parse()
	fmt.Printf("started a %s server\n", config.env)

	server := NewServer()
	server.InstanitateServerServices()
	server.r.Use(server.corsMiddleware())
	server.SetUpRoutes()
	defer server.ServerCleanUp()
	server.r.Run(":5000")
}
