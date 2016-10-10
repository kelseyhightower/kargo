package kargo

import (
	"flag"
	"fmt"
	"io"
)

var (
	apiHost          string
	cpuLimit         string
	cpuRequest       string
	memoryLimit      string
	memoryRequest    string
	namespace        string
	replicas         int
	EnableKubernetes bool
)

func init() {
	flag.StringVar(&apiHost, "api-host", "127.0.0.1:8001", "Kubernetes API server")
	flag.StringVar(&cpuLimit, "cpu-limit", "100m", "Max CPU in milicores")
	flag.StringVar(&cpuRequest, "cpu-request", "100m", "Min CPU in milicores")
	flag.StringVar(&memoryLimit, "memory-limit", "64M", "Max memory in MB")
	flag.StringVar(&memoryRequest, "memory-request", "64M", "Min memory in MB")
	flag.StringVar(&namespace, "namespace", "default", "The Kubernetes namespace.")
	flag.IntVar(&replicas, "replicas", 1, "Number of replicas")
	flag.BoolVar(&EnableKubernetes, "kubernetes", false, "Deploy to Kubernetes.")
}

type DeploymentConfig struct {
	Annotations   map[string]string
	Args          []string
	Env           map[string]string
	BinaryURL     string
	cpuRequest    string
	cpuLimit      string
	memoryRequest string
	memoryLimit   string
	Name          string
	Namespace     string
	Replicas      int
	Labels        map[string]string
}

type DeploymentManager struct {
	apiHost string
	config  DeploymentConfig
}

func New() *DeploymentManager {
	return &DeploymentManager{apiHost: apiHost}
}

func (dm *DeploymentManager) Create(config DeploymentConfig) error {
	config.cpuRequest = cpuRequest
	config.cpuLimit = cpuLimit
	config.memoryRequest = memoryRequest
	config.memoryLimit = memoryLimit
	config.Replicas = replicas
	config.Namespace = namespace

	if config.Env == nil {
		config.Env = make(map[string]string)
	}
	if config.Annotations == nil {
		config.Annotations = make(map[string]string)
	}
	if config.Labels == nil {
		config.Labels = make(map[string]string)
	}
	dm.config = config

	fmt.Printf("Creating %s ReplicaSet...\n", config.Name)
	return createReplicaSet(dm.config)
}

func (dm *DeploymentManager) Delete() error {
	fmt.Printf("Deleting %s ReplicaSet...\n", dm.config.Name)
	return deleteReplicaSet(dm.config)
}

func (dm *DeploymentManager) Logs(w io.Writer) error {
	return getLogs(dm.config, w)
}
