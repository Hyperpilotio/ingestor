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

	// TODO: Refactor capturer so we can support a interface instead of adding each one here.
	// We should be able to get a list of capturers, iterate through it and run capture.
	// ECS regions should be iterated in the awsecs capturer instead of here.
	// Also configuration we can just pass through a map[string]{}interface so each capturer
	// can know how to get these config on its own.
	for _, regionName := range capturer.Regions {
		// capture AWS ECS Clusters
		deployments, _ := capturer.GetClusters(config, regionName)
		if deployments != nil {
			// TODO: need unique condition is required as a basis for update
			selector := bson.M{"Region": regionName}
			db.Upsert(selector, *deployments)
		}
	}

	// capture Kubernetes Cluster
	_, err := capturer.GetK8SCluster(config.GetString("kubeconfig"))
	if err != nil {
		// selector := bson.M{"Region": regionName}
		// db.Upsert(selector, *k8sDeployments)
	}

	return nil
}
