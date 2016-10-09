package kargo

import (
	"bytes"
	"flag"
	"io"
)

var (
	apiServer        string
	cpuLimit         string
	cpuRequest       string
	memoryLimit      string
	memoryRequest    string
	replicas         int
	EnableKubernetes bool
)

func init() {
	flag.StringVar(&apiServer, "api-server", "http://127.0.0.1:8080", "Kubernetes API server")
	flag.StringVar(&cpuLimit, "cpu-limit", "100m", "Max CPU in milicores")
	flag.StringVar(&cpuRequest, "cpu-request", "100m", "Min CPU in milicores")
	flag.StringVar(&memoryLimit, "memory-limit", "64M", "Max memory in MB")
	flag.StringVar(&memoryRequest, "memory-request", "64M", "Min memory in MB")
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
	Replicas      int
	Labels        map[string]string
}

type DeploymentManager struct {
	APIServerURL string
	config       DeploymentConfig
}

func New(url string) *DeploymentManager {
	return &DeploymentManager{APIServerURL: url}
}

func (dm *DeploymentManager) Create(config DeploymentConfig) error {
	config.cpuRequest = cpuRequest
	config.cpuLimit = cpuLimit
	config.memoryRequest = memoryRequest
	config.memoryLimit = memoryLimit
	config.Replicas = replicas

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
	return createReplicaSet(config)
}

func (dm *DeploymentManager) Delete() error {
	return deleteReplicaSet(dm.config.Name)
}

func (dm *DeploymentManager) Logs() (io.Reader, error) {
	var b bytes.Buffer
	return &b, nil
}
