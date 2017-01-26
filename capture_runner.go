package main

import (
	"github.com/hyperpilotio/ingestor/capturer"
	"github.com/spf13/viper"
)

func RunCapture(config *viper.Viper) error {
	db, dbErr := capturer.NewDB(config)
	if dbErr != nil {
		return dbErr
	}

	for _, regionName := range capturer.Regions {
		// capture AWS ECS Clusters
		deployments, _ := capturer.GetClusters(config, regionName)
		if deployments != nil {
			db.Insert(*deployments)
		}

		// TODO: capture other....
	}

	return nil
}
