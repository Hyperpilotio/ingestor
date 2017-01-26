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

type Service struct {
	ServiceName    string `json:"ServiceName" bson:"ServiceName"`
	ServiceArn     string `json:"ServiceArn" bson:"ServiceArn"`
	TaskDefinition string `json:"TaskDefinition" bson:"TaskDefinition"`
}

type Container struct {
	ContainerArn string `json:"ContainerArn" bson:"ContainerArn"`
	Name         string `json:"Name" bson:"Name"`
}

type Task struct {
	Containers        []Container `json:"Containers" bson:"Containers"`
	TaskArn           string      `json:"TaskArn" bson:"TaskArn"`
	TaskDefinitionArn string      `json:"TaskDefinitionArn" bson:"TaskDefinitionArn"`
}

type Cluster struct {
	ClusterName string           `json:"ClusterName" bson:"ClusterName"`
	NodeInfos   map[int]NodeInfo `json:"NodeInfos" bson:"NodeInfos"`
	Services    []Service        `json:"Services" bson:"Services"`
	Tasks       []Task           `json:"Tasks" bson:"Tasks"`
}

type Deployments struct {
	ID       bson.ObjectId `json:"id" bson:"_id"`
	Region   string        `json:"Region" bson:"Region"`
	Clusters []Cluster     `json:"Clusters" bson:"Clusters"`
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
	glog.V(1).Infof("GetClusters for region: %s", regionName)
	deployments := &Deployments{ID: bson.NewObjectId(), Region: regionName}
	sess, sessionErr := createSessionByRegion(viper, regionName)
	if sessionErr != nil {
		return nil, errors.New("Unable to create session: " + sessionErr.Error())
	}

	ecsSvc := ecs.New(sess)
	ec2Svc := ec2.New(sess)

	// find clusters on region
	listClustersOutput, listClustersErr := ecsSvc.ListClusters(nil)
	if listClustersErr != nil {
		return nil, errors.New("Unable to find any clusters: " + listClustersErr.Error())
	}

	// use clusterArns get region's clusters information
	describeClustersInput := &ecs.DescribeClustersInput{
		Clusters: listClustersOutput.ClusterArns,
	}
	describeClustersOutput, _ := ecsSvc.DescribeClusters(describeClustersInput)

	deployClusters := []Cluster{}
	for _, cluster := range describeClustersOutput.Clusters {
		clusterName := *cluster.ClusterName
		deployCluster := &Cluster{}
		deployCluster.ClusterName = clusterName

		// use clusterName get instancess arns of Container
		listInstancesInput := &ecs.ListContainerInstancesInput{
			Cluster: aws.String(clusterName),
		}
		listContainerInstancesOutput, _ := ecsSvc.ListContainerInstances(listInstancesInput)

		// use containerInstanceArns get ContainerInstances
		describeInstancesInput := &ecs.DescribeContainerInstancesInput{
			Cluster:            aws.String(clusterName),
			ContainerInstances: listContainerInstancesOutput.ContainerInstanceArns,
		}
		describeInstancesOutput, _ := ecsSvc.DescribeContainerInstances(describeInstancesInput)

		// use Ec2InstanceId get instance information
		nodeInfos := map[int]NodeInfo{}
		for idx, containerInstance := range describeInstancesOutput.ContainerInstances {
			ec2InstanceId := *containerInstance.Ec2InstanceId
			describeInstancesInput := &ec2.DescribeInstancesInput{
				InstanceIds: []*string{
					aws.String(ec2InstanceId),
				},
			}
			describeInstanceOutput, _ := ec2Svc.DescribeInstances(describeInstancesInput)

			reservation := describeInstanceOutput.Reservations[0]
			instance := reservation.Instances[0]

			deployInstance := &Instance{}
			deployInstance.InstanceId = *instance.InstanceId
			deployInstance.InstanceType = *instance.InstanceType
			deployInstance.LaunchTime = *instance.LaunchTime

			nodeInfo := &NodeInfo{}
			nodeInfo.PublicDnsName = *instance.PublicDnsName
			nodeInfo.Arn = *instance.IamInstanceProfile.Arn
			nodeInfo.Instance = *deployInstance
			nodeInfos[idx+1] = *nodeInfo
		}
		deployCluster.NodeInfos = nodeInfos

		// use clusterName get ServiceArns
		listServicesInput := &ecs.ListServicesInput{
			Cluster: aws.String(clusterName),
		}
		listServicesOutput, _ := ecsSvc.ListServices(listServicesInput)

		// use ServiceArns get service information
		describeServicesInput := &ecs.DescribeServicesInput{
			Services: listServicesOutput.ServiceArns,
			Cluster:  aws.String(clusterName),
		}
		describeServicesOutput, _ := ecsSvc.DescribeServices(describeServicesInput)

		clusterServices := []Service{}
		for _, service := range describeServicesOutput.Services {
			clusterService := &Service{}
			clusterService.ServiceArn = *service.ServiceArn
			clusterService.ServiceName = *service.ServiceName
			clusterService.TaskDefinition = *service.TaskDefinition
			clusterServices = append(clusterServices, *clusterService)
		}
		deployCluster.Services = clusterServices

		// use clusterName get TaskArns
		listTasksInput := &ecs.ListTasksInput{
			Cluster: aws.String(clusterName),
		}
		listTasksOutput, _ := ecsSvc.ListTasks(listTasksInput)

		// use TaskArns get task information
		describeTasksInput := &ecs.DescribeTasksInput{
			Tasks:   listTasksOutput.TaskArns,
			Cluster: aws.String(clusterName),
		}
		describeTasksOutput, _ := ecsSvc.DescribeTasks(describeTasksInput)

		clusterTasks := []Task{}
		for _, task := range describeTasksOutput.Tasks {
			clusterTask := &Task{}
			clusterTask.TaskArn = *task.TaskArn
			clusterTask.TaskDefinitionArn = *task.TaskDefinitionArn

			taskContainers := []Container{}
			for _, container := range task.Containers {
				taskContainer := &Container{}
				taskContainer.ContainerArn = *container.ContainerArn
				taskContainer.Name = *container.Name
				taskContainers = append(taskContainers, *taskContainer)
			}
			clusterTask.Containers = taskContainers
			clusterTasks = append(clusterTasks, *clusterTask)
		}
		deployCluster.Tasks = clusterTasks
		deployClusters = append(deployClusters, *deployCluster)
	}
	deployments.Clusters = deployClusters

	return deployments, nil
}
