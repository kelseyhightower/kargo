package kargo

import (
	"bytes"
	"flag"
	"io"
)

var (
	apiServer        string
	cpuLimit         int
	cpuRequest       int
	memoryLimit      int
	memoryRequest    int
	replicas         int
	EnableKubernetes bool
)

func init() {
	flag.StringVar(&apiServer, "api-server", "http://127.0.0.1:8080", "Kubernetes API server")
	flag.IntVar(&cpuLimit, "cpu-limit", 100, "Max CPU in milicores")
	flag.IntVar(&cpuRequest, "cpu-request", 100, "Min CPU in milicores")
	flag.IntVar(&memoryLimit, "memory-limit", 64, "Max memory in MB")
	flag.IntVar(&memoryRequest, "memory-request", 64, "Min memory in MB")
	flag.IntVar(&replicas, "replicas", 1, "Number of replicas")
	flag.BoolVar(&EnableKubernetes, "kubernetes", false, "Deploy to Kubernetes.")
}

type DeploymentConfig struct {
	Annotations   map[string]string
	Args          []string
	BinaryURL     string
	ConfigMap     map[string]string
	CPURequest    int
	CPULimit      int
	MemoryRequest int
	MemoryLimit   int
	Name          string
	Replicas      int
	Secrets       map[string]string
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
	config.Replicas = replicas
	if config.Annotations == nil {
		config.Annotations = make(map[string]string)
	}
	if config.ConfigMap == nil {
		config.ConfigMap = make(map[string]string)
	}
	if config.Secrets == nil {
		config.Secrets = make(map[string]string)
	}
	if config.Labels == nil {
		config.Labels = make(map[string]string)
	}

	dm.config = config
	return createReplicaSet(config)
}

func (dm *DeploymentManager) Logs() (io.Reader, error) {
	var b bytes.Buffer
	return &b, nil
}

func (dm *DeploymentManager) Delete() error {
	return deleteReplicaSet(dm.config.Name)
}
