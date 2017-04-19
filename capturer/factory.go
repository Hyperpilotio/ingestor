package capturer

import (
	"errors"

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
		for _, regionName := range AWSRegions {
			capturer, err := NewAWSECSCapturer(aws, regionName)
			if err != nil {
				return nil, errors.New("Unable to create AWS capturer: " + err.Error())
			}
			capturers.CapturerList = append(capturers.CapturerList, capturer)
		}
	}

	return capturers, nil
}
