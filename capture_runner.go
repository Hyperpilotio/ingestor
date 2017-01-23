package main

import (
	"github.com/hyperpilotio/ingestor/capturer"
	"github.com/spf13/viper"
)

func RunCapture(config *viper.Viper) {
	capturer.GetContainerInstances(config)
}
