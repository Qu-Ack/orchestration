package main

func main() {
	server := NewServer()
	server.InstanitateServerServices()
	server.SetUpRoutes()
	defer server.ServerCleanUp()
	server.r.Run(":5000")
}
