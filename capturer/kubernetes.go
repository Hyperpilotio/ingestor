package capturer

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/tools/clientcmd"
)

func getK8SCluster(viper *viper.Viper, regionName string) (*Deployments, error) {

	kubeconfig := flag.String("kubeconfig", filepath.Join(os.Getenv("k8sConfigDir"), "kubeconfig"), "absolute path to the kubeconfig file")
	flag.Parse()
	// uses the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())
	}
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	services, err := clientset.CoreV1().Services("").List(v1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}
	fmt.Printf("There are %d services in the cluster\n", len(services.Items))

	pods, err := clientset.CoreV1().Pods("").List(v1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}
	fmt.Printf("There are %d pods in the cluster\n", len(pods.Items))

	for _, s := range services.Items {
		for p, _ := range s.Spec.Ports {
			fmt.Println("Port:", s.Spec.Ports[p].Port)
			fmt.Println("NodePort:", s.Spec.Ports[p].NodePort)
		}
	}

	return deployments, nil
}
