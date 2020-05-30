package awsecs

import (
	"errors"
	"os"
	"time"

	"github.com/spf13/viper"
	"gopkg.in/mgo.v2/bson"

	"github.com/hyperpilotio/ingestor/database"
	"github.com/hyperpilotio/ingestor/log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ecs"
)

type Instance struct {
	InstanceId   string    `json:"InstanceId" bson:"InstanceId"`
	InstanceType string    `json:"InstanceType" bson:"InstanceType"`
	LaunchTime   time.Time `json:"LaunchTime" bson:"LaunchTime"`
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

type NodeInfo struct {
	Instance      Instance `json:"Instance" bson:"Instance"`
	Arn           string   `json:"Arn" bson:"Arn"`
	PublicDnsName string   `json:"PublicDnsName" bson:"PublicDnsName"`
	Tasks         []Task   `json:"Tasks" bson:"Tasks"`
}

type Service struct {
	ServiceName    string `json:"ServiceName" bson:"ServiceName"`
	ServiceArn     string `json:"ServiceArn" bson:"ServiceArn"`
	TaskDefinition string `json:"TaskDefinition" bson:"TaskDefinition"`
}

type Cluster struct {
	ClusterName string     `json:"ClusterName" bson:"ClusterName"`
	NodeInfos   []NodeInfo `json:"NodeInfos" bson:"NodeInfos"`
	Services    []Service  `json:"Services" bson:"Services"`
}

type Deployments struct {
	ID       bson.ObjectId `json:"id" bson:"_id,omitempty"`
	Region   string        `json:"Region" bson:"Region"`
	Clusters []Cluster     `json:"Clusters" bson:"Clusters"`
}

func createSessionByRegion(viper *viper.Viper, regionName string) (*session.Session, error) {
	awsId := os.Getenv("awsId")
	awsSecret := os.Getenv("awsSecret")
	creds := credentials.NewStaticCredentials(awsId, awsSecret, "")
	config := &aws.Config{
		Region: aws.String(regionName),
	}
	config = config.WithCredentials(creds)
	sess, err := session.NewSession(config)
	if err != nil {
		log.Errorf("Unable to create session: %s", err)
		return nil, err
	}

	return sess, nil
}

type AWSECSCapturer struct {
	Region string
	Sess   *session.Session
	DB     *database.MongoDB
}

func NewCapturer(config *viper.Viper, region string) (*AWSECSCapturer, error) {
	db, dbErr := database.NewDB(config)
	if dbErr != nil {
		return nil, dbErr
	}

	if session, err := createSessionByRegion(config, region); err != nil {
		return nil, err
	} else {
		return &AWSECSCapturer{
			Region: region,
			Sess:   session,
			DB:     db,
		}, nil
	}
}

func (capturer AWSECSCapturer) Capture() error {
	if deployments, err := capturer.GetClusters(); err != nil {
		return errors.New("Unable to get clusters info: " + err.Error())
	} else if deployments != nil {
		// TODO: need unique condition is required as a basis for update
		selector := bson.M{"Region": capturer.Region}
		capturer.DB.Upsert(selector, *deployments)
	}

	return nil
}

func (capturer AWSECSCapturer) GetClusters() (*Deployments, error) {
	log.Infof("GetClusters for region: %s", capturer.Region)

	ecsSvc := ecs.New(capturer.Sess)
	ec2Svc := ec2.New(capturer.Sess)

	// find clusters on region
	listClustersOutput, listClustersErr := ecsSvc.ListClusters(nil)
	if listClustersErr != nil {
		return nil, errors.New("Unable to find any clusters: " + listClustersErr.Error())
	}

	if len(listClustersOutput.ClusterArns) == 0 {
		return nil, errors.New("Unable to find any clusters: ClusterArns size is zero")
	}

	// use clusterArns get region's clusters information
	describeClustersInput := &ecs.DescribeClustersInput{
		Clusters: listClustersOutput.ClusterArns,
	}
	describeClustersOutput, _ := ecsSvc.DescribeClusters(describeClustersInput)

	deployments := &Deployments{Region: capturer.Region}
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
		ecsDescribeInstancesInput := &ecs.DescribeContainerInstancesInput{
			Cluster:            aws.String(clusterName),
			ContainerInstances: listContainerInstancesOutput.ContainerInstanceArns,
		}
		ecsDescribeInstancesOutput, _ := ecsSvc.DescribeContainerInstances(ecsDescribeInstancesInput)

		// use Ec2InstanceId get instance information
		nodeInfos := []NodeInfo{}
		for _, containerInstance := range ecsDescribeInstancesOutput.ContainerInstances {
			ec2InstanceId := *containerInstance.Ec2InstanceId
			containerInstanceArn := *containerInstance.ContainerInstanceArn
			ec2DescribeInstancesInput := &ec2.DescribeInstancesInput{
				InstanceIds: []*string{
					aws.String(ec2InstanceId),
				},
			}
			ec2DescribeInstanceOutput, _ := ec2Svc.DescribeInstances(ec2DescribeInstancesInput)

			reservation := ec2DescribeInstanceOutput.Reservations[0]
			instance := reservation.Instances[0]

			deployInstance := &Instance{}
			deployInstance.InstanceId = *instance.InstanceId
			deployInstance.InstanceType = *instance.InstanceType
			deployInstance.LaunchTime = *instance.LaunchTime

			nodeInfo := &NodeInfo{}
			nodeInfo.PublicDnsName = *instance.PublicDnsName
			nodeInfo.Arn = containerInstanceArn
			nodeInfo.Instance = *deployInstance

			// use clusterName get TaskArns
			listTasksInput := &ecs.ListTasksInput{
				Cluster:           aws.String(clusterName),
				ContainerInstance: aws.String(containerInstanceArn),
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
			nodeInfo.Tasks = clusterTasks
			nodeInfos = append(nodeInfos, *nodeInfo)
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
		deployClusters = append(deployClusters, *deployCluster)
	}
	deployments.Clusters = deployClusters

	return deployments, nil
}
