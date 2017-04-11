package main

import (
	"gopkg.in/mgo.v2/bson"

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
			// TODO: need unique condition is required as a basis for update
			selector := bson.M{"Region": regionName}
			db.Upsert(selector, *deployments)
		}

		// TODO: capture other....
		// capture Kubernetes Clusters
		k8sDeployments, _ := capturer.GetK8SCluster(regionName)
		if k8sDeployments != nil {
			// selector := bson.M{"Region": regionName}
			// db.Upsert(selector, *k8sDeployments)
		}
	}

	return nil
}
