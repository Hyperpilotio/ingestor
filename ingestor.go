package main

import (
	"github.com/hyperpilotio/ingestor/config"
	"github.com/hyperpilotio/ingestor/log"
	"github.com/spf13/viper"
)

// Run start the web server
func Run() error {
	v := viper.New()

	// FIXME antipattern, should avoid using type assertion
	v = (config.Config()).(*viper.Viper)

	server := NewServer(v)
	return server.StartServer()
}

func main() {
	err := Run()
	log.Fatalln(err)
}
