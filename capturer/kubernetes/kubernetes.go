package kubernetes

import (
	"errors"

	"gopkg.in/mgo.v2/bson"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type KubernetesCapturer struct {
	config *rest.Config
}

type KubernetesContainer struct {
	CantainerName  string            `json:"CantainerName" bson:"CantainerName"`
	ContainerImage string            `json:"ContainerImage" bson:"ContainerImage"`
	Limit          map[string]string `json:"Limit" bson:"Limit"`
}

type KubernetesPod struct {
	PodName     string                `json:"PodName" bson:"PodName"`
	NodeName    string                `json:"NodeName" bson:"NodeName"`
	ClusterName string                `json:"ClusterName" bson:"ClusterName"`
	Phase       string                `json:"Phase" bson:"Phase"`
	Containers  []KubernetesContainer `json:"Containers" bson:"Containers"`
}

type KubernetesNode struct {
	IsMaster   bool               `json:"IsMaster" bson:"IsMaster"`
	NodeName   string             `json:"NodeName" bson:"NodeName"`
	Pods       []KubernetesPod    `json:"Pods" bson:"Pods"`
	Conditions []v1.NodeCondition `json:"Conditions" bson:"Conditions"`
}

type KubernetesService struct {
	ServiceName string           `json:"ServiceName" bson:"ServiceName"`
	ClusterIP   string           `json:"ClusterIP" bson:"ClusterIP"`
	ExternalIPs []string         `json:"ExternalIPs" bson:"ExternalIPs"`
	Ports       []v1.ServicePort `json:"Ports" bson:"Ports"`
}

type KubernetesDeployment struct {
	Name         string            `json:"Name" bson:"Name"`
	Namespace    string            `json:"Namespace" bson:"Namespace"`
	SelfLink     string            `json:"SelfLink" bson:"SelfLink"`
	Replicas     int32             `json:"Replicas" bson:"Replicas"`
	Labels       map[string]string `json:"Labels" bson:"Labels"`
	Selector     map[string]string `json:"Selector" bson:"Selector"`
	NodeSelector map[string]string `json:"NodeSelector" bson:"NodeSelector"`
}

type KubernetesCluster struct {
	ClusterName string                 `json:"ClusterName" bson:"ClusterName"`
	ID          bson.ObjectId          `json:"id" bson:"_id,omitempty"`
	Nodes       []KubernetesNode       `json:"Nodes" bson:"Nodes"`
	Services    []KubernetesService    `json:"Services" bson:"Services"`
	Deployments []KubernetesDeployment `json:"Deployments" bson:"Deployments"`
}

type K8sDeployments struct {
	ID       bson.ObjectId       `json:"id" bson:"_id,omitempty"`
	Clusters []KubernetesCluster `json:"Clusters" bson:"Clusters"`
}

func NewCapturer(kubeconfigPath string) (*KubernetesCapturer, error) {
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		return nil, errors.New("Unable to build config: " + err.Error())
	}

	return &KubernetesCapturer{
		config: config,
	}, nil
}

func (capturer *KubernetesCapturer) Capture() error {
	clientset, err := kubernetes.NewForConfig(capturer.config)
	if err != nil {
		return errors.New("Unable to create a new Clientset: " + err.Error())
	}

	deployments, err := clientset.ExtensionsV1beta1().Deployments("").List(v1.ListOptions{})
	if err != nil {
		return errors.New("Unable to find deployments: " + err.Error())
	}

	nodes, err := clientset.CoreV1().Nodes().List(v1.ListOptions{})
	if err != nil {
		return errors.New("Unable to find nodes: " + err.Error())
	}

	pods, err := clientset.CoreV1().Pods("").List(v1.ListOptions{})
	if err != nil {
		return errors.New("Unable to find pods: " + err.Error())
	}

	k8sCluster := &KubernetesCluster{}
	clusterDeployments := []KubernetesDeployment{}
	for _, deployment := range deployments.Items {
		clusterDeployment := &KubernetesDeployment{}
		clusterDeployment.Name = deployment.Name
		clusterDeployment.Namespace = deployment.Namespace
		clusterDeployment.SelfLink = deployment.SelfLink
		clusterDeployment.Labels = deployment.Labels
		clusterDeployment.Replicas = *deployment.Spec.Replicas
		clusterDeployment.Selector = deployment.Spec.Selector.MatchLabels
		clusterDeployment.NodeSelector = deployment.Spec.Template.Spec.NodeSelector
		clusterDeployments = append(clusterDeployments, *clusterDeployment)
	}

	k8sCluster.Deployments = clusterDeployments
	clusterNodes := []KubernetesNode{}
	for _, node := range nodes.Items {
		clusterNode := &KubernetesNode{}
		if node.Labels["kubeadm.alpha.kubernetes.io/role"] == "master" {
			clusterNode.IsMaster = true
		}
		clusterNode.NodeName = node.Name
		clusterNode.Conditions = node.Status.Conditions

		deploymentPods := []KubernetesPod{}
		for _, pod := range pods.Items {
			if pod.Spec.NodeName == node.Name {
				deploymentPod := &KubernetesPod{}
				deploymentPod.PodName = pod.Name
				deploymentPod.NodeName = pod.Spec.NodeName
				deploymentPod.ClusterName = pod.ClusterName
				deploymentPod.Phase = string(pod.Status.Phase)

				deploymentContainers := []KubernetesContainer{}
				for _, container := range pod.Spec.Containers {
					deploymentContainer := &KubernetesContainer{}
					deploymentContainer.CantainerName = container.Name
					deploymentContainer.ContainerImage = container.Image
					deploymentContainer.Limit = make(map[string]string)
					for k, v := range container.Resources.Limits {
						limitJson, error := v.MarshalJSON()
						if error == nil {
							deploymentContainer.Limit[string(k)] = string(limitJson)
						}
					}
					deploymentContainers = append(deploymentContainers, *deploymentContainer)
				}
				deploymentPod.Containers = deploymentContainers
				deploymentPods = append(deploymentPods, *deploymentPod)
				clusterNode.Pods = deploymentPods
			}
		}

		clusterNodes = append(clusterNodes, *clusterNode)
	}

	k8sCluster.Nodes = clusterNodes

	clusters := []KubernetesCluster{}
	clusters = append(clusters, *k8sCluster)
	k8sDeployments := &K8sDeployments{}
	k8sDeployments.Clusters = clusters

	return nil
}
