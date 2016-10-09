package kargo

type Metadata struct {
	Name            string            `json:"name"`
	Namespace       string            `json:"namespace"`
	GenerateName    string            `json:"generateName"`
	ResourceVersion string            `json:"resourceVersion"`
	SelfLink        string            `json:"selfLink"`
	Labels          map[string]string `json:"labels"`
	Annotations     map[string]string `json:"annotations"`
	Uid             string            `json:"uid"`
}

type ListMetadata struct {
	ResourceVersion string `json:"resourceVersion"`
}

type Pod struct {
	Kind     string   `json:"kind,omitempty"`
	Metadata Metadata `json:"metadata"`
	Spec     PodSpec  `json:"spec"`
}

type PodList struct {
	ApiVersion string       `json:"apiVersion"`
	Kind       string       `json:"kind"`
	Metadata   ListMetadata `json:"metadata"`
	Items      []Pod        `json:"items"`
}

type ReplicaSet struct {
	ApiVersion string         `json:"apiVersion,omitempty"`
	Kind       string         `json:"kind,omitempty"`
	Metadata   Metadata       `json:"metadata"`
	Spec       ReplicaSetSpec `json:"spec"`
}

type ReplicaSetSpec struct {
	Replicas int64         `json:"replicas,omitempty"`
	Selector LabelSelector `json:"selector,omitempty"`
	Template PodTemplate   `json:"template,omitempty"`
}

type LabelSelector struct {
	MatchLabels map[string]string `json:"matchLabels,omitempty"`
}

type PodTemplate struct {
	Metadata Metadata `json:"metadata"`
	Spec     PodSpec  `json:"spec"`
}

type PodSpec struct {
	Containers []Container `json:"containers"`
	Volumes    []Volume    `json:"volumes,omitempty"`
}

type Container struct {
	Args         []string             `json:"args"`
	Command      []string             `json:"command"`
	Env          []EnvVar             `json:"env,omitempty"`
	Image        string               `json:"image"`
	Name         string               `json:"name"`
	VolumeMounts []VolumeMount        `json:"volumeMounts"`
	Resources    ResourceRequirements `json:"resources,omitempty"`
}

type EnvVar struct {
	Name  string `json:"name"`
	Value string `json:"value,omitempty"`
}

type ResourceRequirements struct {
	Limits   ResourceList `json:"limits,omitempty"`
	Requests ResourceList `json:"requests,omitempty"`
}

type ResourceList map[string]string

type VolumeMount struct {
	Name      string `json:"name"`
	MountPath string `json:"mountPath"`
}

type Volume struct {
	Name         string `json:"name"`
	VolumeSource `json:",inline"`
}

type VolumeSource struct {
	HostPath  *HostPathVolumeSource  `json:"hostPath,omitempty"`
	EmptyDir  *EmptyDirVolumeSource  `json:"emptyDir,omitempty"`
	Secret    *SecretVolumeSource    `json:"secret,omitempty"`
	ConfigMap *ConfigMapVolumeSource `json:"configMap,omitempty"`
}

type HostPathVolumeSource struct {
	Path string `json:"path"`
}

type EmptyDirVolumeSource struct{}

type SecretVolumeSource struct {
	SecretName string      `json:"secretName,omitempty"`
	Items      []KeyToPath `json:"items,omitempty"`
}

type ConfigMapVolumeSource struct {
	Name  string      `json:"name,omitempty"`
	Items []KeyToPath `json:"items,omitempty"`
}

type KeyToPath struct {
	Key  string `json:"key"`
	Path string `json:"path"`
}

type Scale struct {
	ApiVersion string    `json:"apiVersion,omitempty"`
	Kind       string    `json:"kind,omitempty"`
	Metadata   Metadata  `json:"metadata"`
	Spec       ScaleSpec `json:"spec,omitempty"`
}

type ScaleSpec struct {
	Replicas int64 `json:"replicas,omitempty"`
}
