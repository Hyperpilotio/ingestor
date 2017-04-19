package capturer

import (
	"errors"

	"github.com/hyperpilotio/ingestor/capturer/awsecs"
	"github.com/hyperpilotio/ingestor/capturer/kubernetes"
	"github.com/spf13/viper"
)

type Capturer interface {
	Capture() error
}

type Capturers struct {
	CapturerList []Capturer
}

func (capturers *Capturers) Run() error {
	// TODO: Each capturer has its own schedule, we eventually will need to figure out
	// another way to run all of them.
	for _, capturer := range capturers.CapturerList {
		if err := capturer.Capture(); err != nil {
			return err
		}
	}

	return nil
}

func NewCapturers(config *viper.Viper) (*Capturers, error) {
	capturers := &Capturers{
		CapturerList: make([]Capturer, 0),
	}

	aws := config.Sub("aws")
	if aws != nil {
		for _, region := range aws.GetStringSlice("regions") {
			capturer, err := awsecs.NewCapturer(aws, region)
			if err != nil {
				return nil, errors.New("Unable to create AWS capturer: " + err.Error())
			}
			capturers.CapturerList = append(capturers.CapturerList, capturer)
		}
	}

	k8sConfig := config.Sub("kubernetes")
	if k8sConfig != nil {
		configPath := k8sConfig.GetString("configPath")
		capturer, err := kubernetes.NewCapturer(configPath)
		if err != nil {
			return nil, errors.New("Unable to create Kubernetes capturer: " + err.Error())
		}
		capturers.CapturerList = append(capturers.CapturerList, capturer)
	}

	return capturers, nil
}
