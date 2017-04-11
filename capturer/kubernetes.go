package capturer

import (
	"encoding/json"
	"errors"
	"fmt"

	"gopkg.in/mgo.v2/bson"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/tools/clientcmd"
)

type KubernetesContainer struct {
	CantainerName  string `json:"CantainerName" bson:"CantainerName"`
	ContainerImage string `json:"ContainerImage" bson:"ContainerImage"`
	// Limit          ResourceList `json:"Limit" bson:"Limit"`
}

type KubernetesPod struct {
	PodName     string                `json:"PodName" bson:"PodName"`
	NodeName    string                `json:"NodeName" bson:"NodeName"`
	ClusterName string                `json:"ClusterName" bson:"ClusterName"`
	Phase       string                `json:"Phase" bson:"Phase"`
	Containers  []KubernetesContainer `json:"Containers" bson:"Containers"`
}

type KubernetesNode struct {
	IsMaster bool            `json:"IsMaster" bson:"IsMaster"`
	NodeName string          `json:"NodeName" bson:"NodeName"`
	Pods     []KubernetesPod `json:"Pods" bson:"Pods"`
	// Conditions []NodeCondition `json:"Conditions" bson:"Conditions"`
}

type KubernetesService struct {
	ServiceName string   `json:"ServiceName" bson:"ServiceName"`
	ClusterIP   string   `json:"ClusterIP" bson:"ClusterIP"`
	ExternalIPs []string `json:"ExternalIPs" bson:"ExternalIPs"`
	// Ports       []ServicePort `json:"Ports" bson:"Ports"`
}

type KubernetesCluster struct {
	ClusterName string              `json:"ClusterName" bson:"ClusterName"`
	Nodes       []KubernetesNode    `json:"KubernetesNode" bson:"KubernetesNode"`
	Services    []KubernetesService `json:"KubernetesService" bson:"KubernetesService"`
}

type KubernetesDeployments struct {
	ID       bson.ObjectId       `json:"id" bson:"_id,omitempty"`
	Clusters []KubernetesCluster `json:"Clusters" bson:"Clusters"`
}

func GetK8SCluster(kubeconfigPath string) (*KubernetesDeployments, error) {
	// kubeconfig := flag.String("kubeconfig", filepath.Join(os.Getenv("k8sConfigDir"), "kubeconfig"), "absolute path to the kubeconfig file")
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		return nil, errors.New("Unable to build config: " + err.Error())
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, errors.New("Unable to create a new Clientset: " + err.Error())
	}
	nodes, err := clientset.CoreV1().Nodes().List(v1.ListOptions{})
	if err != nil {
		return nil, errors.New("Unable to find nodes: " + err.Error())
	}
	pods, err := clientset.CoreV1().Pods("").List(v1.ListOptions{})
	if err != nil {
		return nil, errors.New("Unable to find pods: " + err.Error())
	}

	k8sCluster := &KubernetesCluster{}
	clusterNodes := []KubernetesNode{}
	deploymentPods := []KubernetesPod{}

	for _, node := range nodes.Items {
		clusterNode := &KubernetesNode{}
		if node.ObjectMeta.Labels["kubeadm.alpha.kubernetes.io/role"] == "master" {
			clusterNode.IsMaster = true
		}
		clusterNode.NodeName = node.ObjectMeta.Name
		// clusterNode.Conditions = node.Status.Conditions

		for _, pod := range pods.Items {
			if pod.Spec.NodeName == node.ObjectMeta.Name {
				deploymentPod := &KubernetesPod{}
				deploymentPod.PodName = pod.ObjectMeta.Name
				deploymentPod.NodeName = pod.Spec.NodeName
				deploymentPod.ClusterName = pod.ObjectMeta.ClusterName
				deploymentPod.Phase = string(pod.Status.Phase)

				deploymentContainers := []KubernetesContainer{}
				for _, container := range pod.Spec.Containers {
					deploymentContainer := &KubernetesContainer{}
					deploymentContainer.CantainerName = container.Name
					deploymentContainer.ContainerImage = container.Image
					// deploymentContainer.Limit = container.Resources.Limits
					deploymentContainers = append(deploymentContainers, *deploymentContainer)
				}
				deploymentPod.Containers = deploymentContainers
				jPod, _ := json.Marshal(&deploymentPod)
				fmt.Println(string(jPod))
				deploymentPods = append(deploymentPods, *deploymentPod)
			}
		}
		clusterNode.Pods = deploymentPods
		jNode, _ := json.Marshal(&clusterNode)
		fmt.Println(string(jNode))
		clusterNodes = append(clusterNodes, *clusterNode)
	}
	k8sCluster.Nodes = clusterNodes

	services, err := clientset.CoreV1().Services("").List(v1.ListOptions{})
	if err != nil {
		return nil, errors.New("Unable to find any service: " + err.Error())
	}

	clusterServices := []KubernetesService{}
	for _, service := range services.Items {
		clusterService := &KubernetesService{}
		clusterService.ServiceName = service.ObjectMeta.Name
		clusterService.ClusterIP = service.Spec.ClusterIP
		// clusterService.Ports = service.Spec.Ports
		clusterService.ExternalIPs = service.Spec.ExternalIPs
		jService, _ := json.Marshal(&clusterService)
		fmt.Println(string(jService))
		clusterServices = append(clusterServices, *clusterService)
	}
	k8sCluster.Services = clusterServices

	clusters := []KubernetesCluster{}
	clusters = append(clusters, *k8sCluster)
	k8sDeployments := &KubernetesDeployments{}
	k8sDeployments.Clusters = clusters

	return k8sDeployments, nil
}
