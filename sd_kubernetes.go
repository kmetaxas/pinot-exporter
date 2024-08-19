package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// A Pinot Controller cache from Kubernetes discovery mechanism
type KubePinotControllerCache struct {
	knownControllers   []PinotController
	kubernetesClient   *kubernetes.Clientset
	KubeApiEndpoint    string `json:"kube_api_url" yaml:"kube_api_url"`
	ServiceAccountName string `json:"serviceaccount" yaml:"serviceaccount"`
	KubeConfigPath     string `json:"kubeconfig_path" yaml:"kubeconfig_path"`
	discoveryConfig    ServiceDiscoveryConfigK8S
}

// Creates a new KubePinotControllerCache with defaults.
func NewKubePinotControllerCache(discoveryConfig ServiceDiscoveryConfigK8S) *KubePinotControllerCache {
	c := KubePinotControllerCache{
		// Add defaults
		KubeApiEndpoint:    "https://kubernetes:443",
		ServiceAccountName: "default",
		KubeConfigPath:     "~/.kube/config",
		discoveryConfig:    discoveryConfig,
	}
	return &c
}

// Create and return a new Kubernetes client
func getKubernetesConfig() (*rest.Config, error) {
	// If we are running inside the cluster, get the in-cluster config
	if isRunningInsideKubernetes() {
		config, err := rest.InClusterConfig()
		if err != nil {
			logger.Errorf("Failed to get inClusterConfig with error %s", err)
			return nil, err
		}
		return config, nil
	} else {
		// Running outside of kubernetes, so get using ~/.kube/config (as per config)

		context := "minikube" // TODO make this config
		kubeconfig := filepath.Join(homeDir(), ".kube", "config")
		config, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
			&clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfig},
			&clientcmd.ConfigOverrides{CurrentContext: context},
		).ClientConfig()
		if err != nil {
			panic(err.Error())
		}
		return config, nil
	}
	return nil, nil

}

func createKubernetesClient(config *rest.Config) (*kubernetes.Clientset, error) {
	config, err := getKubernetesConfig()
	if err != nil {
		return nil, err
	}
	// Create a Kubernetes client
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		logger.Errorf("Failed to create ClientSet with error: %s", err)
		return clientset, err
	}
	return clientset, err

}

// Connect to kubernetes. Creates a clientset object
func (k *KubePinotControllerCache) Connect() error {
	config, err := getKubernetesConfig()
	if err != nil {
		return err
	}

	clientset, err := createKubernetesClient(config)
	if err != nil {
		return err
	}
	k.kubernetesClient = clientset
	return nil

}
func isRunningInsideKubernetes() bool {
	_, err := os.Stat("/var/run/secrets/kubernetes.io/serviceaccount/token")
	return err == nil
}

// return a label selector string suitable for using in the Kubernetes client LabelSelector field
func GetLabelSelectorString(labels map[string]string) string {
	var selector string
	var pairs []string
	for labelName, labelValue := range labels {
		pairs = append(pairs, labelName+"="+labelValue)
	}
	// Sort pairs to get consistent strings.
	slices.Sort(pairs)
	selector = strings.Join(pairs, ",")
	return selector

}
func (k *KubePinotControllerCache) refreshPinotClustersList() []string {
	var endpoints []string
	var knownControllers []PinotController
	// List PinotCluster resources
	//labelSelector := "app=pinot,nodeType=controller"
	labelSelector := GetLabelSelectorString(k.discoveryConfig.Labels)

	// TODO extract hardcoded timeout to a config
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	services, err := k.kubernetesClient.CoreV1().Services("").List(ctx, metav1.ListOptions{
		LabelSelector: labelSelector,
	})

	if err != nil {
		logger.Errorf("Error fetching Pinot services: %v\n", err)
		return endpoints
	}

	//logger.Infof("Fetched Pinot services: %v\n", services)
	for _, service := range services.Items {
		//logger.Debugf("Discovered service %+v\n", service)
		// TODO http or https?
		// TODO (2) first port is used. How to check which port if a service has multiple ports?
		endpoints = append(endpoints, fmt.Sprintf("http://%s.%s.svc:%d", service.ObjectMeta.Name, service.ObjectMeta.Namespace, service.Spec.Ports[0].Port))
	}

	// update cache
	// Build and update the known controllers
	for _, endpoint := range endpoints {
		knownControllers = append(knownControllers, PinotController{
			URL: endpoint,
		})
	}
	logger.Debugf("We have our endpoints: %+v\n", endpoints)
	k.knownControllers = knownControllers
	return endpoints
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}
