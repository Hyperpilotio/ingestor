package main

import (
	"github.com/hyperpilotio/ingestor/config"
	"github.com/hyperpilotio/ingestor/log"
)

// Run start the web server
func Run() error {
	v := config.Config()

	server := NewServer(v)
	return server.StartServer()
}

func main() {
	err := Run()
	log.Fatalln(err)
}
