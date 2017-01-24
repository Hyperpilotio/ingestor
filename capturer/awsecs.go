package capturer

import (
	"errors"
	"time"

	"github.com/golang/glog"
	"github.com/spf13/viper"
	"gopkg.in/mgo.v2/bson"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ecs"
)

var Regions = []string{
	"us-east-1",
	"us-east-2",
	"us-west-1",
	"us-west-2",
	"eu-west-1",
	"eu-central-1",
	"ap-northeast-1",
	"ap-southeast-1",
	"ap-southeast-2",
}

type Instance struct {
	InstanceId   string    `json:"InstanceId" bson:"InstanceId"`
	InstanceType string    `json:"InstanceType" bson:"InstanceType"`
	LaunchTime   time.Time `json:"LaunchTime" bson:"LaunchTime"`
}

type NodeInfo struct {
	Instance      Instance `json:"Instance" bson:"Instance"`
	Arn           string   `json:"Arn" bson:"Arn"`
	PublicDnsName string   `json:"PublicDnsName" bson:"PublicDnsName"`
}

type Deployments struct {
	ID             bson.ObjectId    `json:"id" bson:"_id"`
	Region         string           `json:"Region" bson:"Region"`
	InstanceNumber int              `json:"InstanceNumber" bson:"InstanceNumber"`
	NodeInfos      map[int]NodeInfo `json:"NodeInfos" bson:"NodeInfos"`
	Tasks          []string         `json:"Tasks" bson:"Tasks"`
}

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

func GetClusters(viper *viper.Viper, regionName string) (*Deployments, error) {
	glog.V(1).Infof("GetClusters")
	deployments := &Deployments{ID: bson.NewObjectId(), Region: regionName}
	sess, sessionErr := createSessionByRegion(viper, regionName)
	if sessionErr != nil {
		return nil, errors.New("Unable to create session: " + sessionErr.Error())
	}

	ec2Svc := ec2.New(sess)
	ecsSvc := ecs.New(sess)

	describeInstanceOutput, _ := ec2Svc.DescribeInstances(nil)
	if describeInstanceOutput.Reservations == nil {
		return nil, errors.New("Can not find instance in the " + regionName)
	}

	deployments.InstanceNumber = len(describeInstanceOutput.Reservations)
	nodeInfos := map[int]NodeInfo{}

	for idx, reservation := range describeInstanceOutput.Reservations {
		nodeInfo := &NodeInfo{}
		deployInstance := &Instance{}

		for _, instance := range reservation.Instances {
			nodeInfo.PublicDnsName = *instance.PublicDnsName
			nodeInfo.Arn = *instance.IamInstanceProfile.Arn

			deployInstance.InstanceId = *instance.InstanceId
			deployInstance.InstanceType = *instance.InstanceType
			deployInstance.LaunchTime = *instance.LaunchTime
		}

		nodeInfo.Instance = *deployInstance
		nodeInfos[idx+1] = *nodeInfo
	}
	deployments.NodeInfos = nodeInfos

	tasks := []string{}
	listTaskDefinitionFamiliesOutput, _ := ecsSvc.ListTaskDefinitionFamilies(nil)
	for _, familie := range listTaskDefinitionFamiliesOutput.Families {
		tasks = append(tasks, *familie)
	}
	deployments.Tasks = tasks

	return deployments, nil
}
