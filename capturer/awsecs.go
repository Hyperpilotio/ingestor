package capturer

import (
	"errors"
	"fmt"

	"github.com/golang/glog"
	"github.com/spf13/viper"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ecs"
)

func createSessionByRegion(viper *viper.Viper, regionName string) (*session.Session, error) {
	awsId := viper.GetString("awsId")
	awsSecret := viper.GetString("awsSecret")
	creds := credentials.NewStaticCredentials(awsId, awsSecret, "")
	config := &aws.Config{
		Region: aws.String(regionName),
	}
	config = config.WithCredentials(creds)
	sess, err := session.NewSession(config)
	if err != nil {
		glog.Errorf("Unable to create session: %s", err)
		return nil, err
	}

	return sess, nil
}

func GetContainerInstances(viper *viper.Viper) error {
	// TODO get Regions in use
	sess, sessionErr := createSessionByRegion(viper, "us-east-1")
	if sessionErr != nil {
		return errors.New("Unable to create session: " + sessionErr.Error())
	}

	ec2Svc := ec2.New(sess)
	describeInstanceOutput, _ := ec2Svc.DescribeInstances(nil)

	for _, reservation := range describeInstanceOutput.Reservations {
		for _, instance := range reservation.Instances {
			fmt.Println(*instance.InstanceType)

		}
	}
	ecsSvc := ecs.New(sess)

	resp1, _ := ecsSvc.ListContainerInstances(nil)
	fmt.Println("ListContainerInstances...")
	fmt.Println(resp1)

	resp2, _ := ecsSvc.ListServices(nil)
	fmt.Println("ListServices...")
	fmt.Println(resp2)

	resp3, _ := ecsSvc.ListTaskDefinitionFamilies(nil)
	fmt.Println("ExampleECS_ListTaskDefinitionFamilies...")
	fmt.Println(resp3)

	return nil
}
