package kargo

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"path/filepath"
	"time"
)

var (
	replicasetsEndpoint = "/apis/extensions/v1beta1/namespaces/%s/replicasets"
	replicasetEndpoint  = "/apis/extensions/v1beta1/namespaces/%s/replicasets/%s"
	scaleEndpoint       = "/apis/extensions/v1beta1/namespaces/%s/replicasets/%s/scale"
	logsEndpoint        = "/api/v1/namespaces/%s/pods/%s/log"
	podsEndpoint        = "/api/v1/namespaces/%s/pods"
)

var ErrNotExist = errors.New("does not exist")

func getPods(namespace, labelSelector string) (*PodList, error) {
	var podList *PodList

	v := url.Values{}
	v.Set("labelSelector", labelSelector)

	path := fmt.Sprintf(podsEndpoint, namespace)
	request := &http.Request{
		Header: make(http.Header),
		Method: http.MethodGet,
		URL: &url.URL{
			Host:     apiHost,
			Path:     path,
			Scheme:   "http",
			RawQuery: v.Encode(),
		},
	}
	request.Header.Set("Accept", "application/json, */*")

	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == 404 {
		fmt.Println("No pods found using selector: ", labelSelector)
		return nil, ErrNotExist
	}
	if resp.StatusCode != 200 {
		return nil, errors.New("Get pods error non 200 reponse: " + resp.Status)
	}

	err = json.NewDecoder(resp.Body).Decode(&podList)
	if err != nil {
		return nil, err
	}
	return podList, nil

}

func getLogs(config DeploymentConfig, w io.Writer) error {
	time.Sleep(10 * time.Second)
	rs, err := getReplicaSet(config.Namespace, config.Name)
	if err != nil {
		return err
	}

	var labelSelector bytes.Buffer
	for key, value := range rs.Spec.Selector.MatchLabels {
		labelSelector.WriteString(fmt.Sprintf("%s=%s", key, value))
	}

	podList, err := getPods(config.Namespace, labelSelector.String())
	if err != nil {
		return err
	}

	for _, pod := range podList.Items {
		v := url.Values{}
		v.Set("follow", "true")

		path := fmt.Sprintf(logsEndpoint, config.Namespace, pod.Metadata.Name)
		request := &http.Request{
			Header: make(http.Header),
			Method: http.MethodGet,
			URL: &url.URL{
				Host:     apiHost,
				Path:     path,
				Scheme:   "http",
				RawQuery: v.Encode(),
			},
		}
		request.Header.Set("Accept", "application/json, */*")

		go func() {
			for {
				resp, err := http.DefaultClient.Do(request)
				if err != nil {
					fmt.Println(err)
					time.Sleep(5 * time.Second)
					continue
				}

				if resp.StatusCode == 404 {
					data, err := ioutil.ReadAll(resp.Body)
					if err != nil {
						fmt.Println(err)
						time.Sleep(5 * time.Second)
						continue
					}
					fmt.Println(string(data))
					fmt.Println("GET pod logs error: ", ErrNotExist)
					time.Sleep(5 * time.Second)
					continue
				}
				if resp.StatusCode != 200 {
					fmt.Println(errors.New("Get replica set error non 200 reponse: " + resp.Status))
					time.Sleep(5 * time.Second)
					continue
				}

				if _, err := io.Copy(w, resp.Body); err != nil {
					fmt.Println(err)
				}
			}
		}()
	}

	return nil
}

func getReplicaSet(namespace, name string) (*ReplicaSet, error) {
	var rs ReplicaSet

	path := fmt.Sprintf(replicasetEndpoint, namespace, name)
	request := &http.Request{
		Header: make(http.Header),
		Method: http.MethodGet,
		URL: &url.URL{
			Host:   apiHost,
			Path:   path,
			Scheme: "http",
		},
	}
	request.Header.Set("Accept", "application/json, */*")

	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == 404 {
		return nil, ErrNotExist
	}
	if resp.StatusCode != 200 {
		return nil, errors.New("Get deployment error non 200 reponse: " + resp.Status)
	}

	err = json.NewDecoder(resp.Body).Decode(&rs)
	if err != nil {
		return nil, err
	}
	return &rs, nil
}

func getScale(namespace, name string) (*Scale, error) {
	var scale Scale

	path := fmt.Sprintf(scaleEndpoint, namespace, name)
	request := &http.Request{
		Header: make(http.Header),
		Method: http.MethodGet,
		URL: &url.URL{
			Host:   apiHost,
			Path:   path,
			Scheme: "http",
		},
	}
	request.Header.Set("Accept", "application/json, */*")

	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == 404 {
		return nil, ErrNotExist
	}
	if resp.StatusCode != 200 {
		return nil, errors.New("Get scale error non 200 reponse: " + resp.Status)
	}

	err = json.NewDecoder(resp.Body).Decode(&scale)
	if err != nil {
		return nil, err
	}
	return &scale, nil
}

func scaleReplicaSet(namespace, name string, replicas int) error {
	scale, err := getScale(namespace, name)
	if err != nil {
		return err
	}
	scale.Spec.Replicas = int64(replicas)

	var b []byte
	body := bytes.NewBuffer(b)
	err = json.NewEncoder(body).Encode(scale)
	if err != nil {
		return err
	}

	path := fmt.Sprintf(scaleEndpoint, namespace, name)
	request := &http.Request{
		Body:          ioutil.NopCloser(body),
		ContentLength: int64(body.Len()),
		Header:        make(http.Header),
		Method:        http.MethodPut,
		URL: &url.URL{
			Host:   apiHost,
			Path:   path,
			Scheme: "http",
		},
	}
	request.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		return err
	}

	if resp.StatusCode == 404 {
		return ErrNotExist
	}
	if resp.StatusCode != 200 {
		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		fmt.Println(string(data))
		return errors.New("Scale ReplicaSet error non 200 reponse: " + resp.Status)
	}

	return nil
}

func deleteReplicaSet(config DeploymentConfig) error {
	err := scaleReplicaSet(config.Namespace, config.Name, 0)
	if err != nil {
		return err
	}

	path := fmt.Sprintf(replicasetEndpoint, config.Namespace, config.Name)
	request := &http.Request{
		Header: make(http.Header),
		Method: http.MethodDelete,
		URL: &url.URL{
			Host:   apiHost,
			Path:   path,
			Scheme: "http",
		},
	}
	request.Header.Set("Accept", "application/json, */*")

	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		return err
	}

	if resp.StatusCode == 404 {
		return ErrNotExist
	}
	if resp.StatusCode != 200 {
		return errors.New("Delete ReplicaSet error non 200 reponse: " + resp.Status)
	}

	return nil
}

func createReplicaSet(config DeploymentConfig) error {
	volumes := make([]Volume, 0)
	volumes = append(volumes, Volume{
		Name:         "bin",
		VolumeSource: VolumeSource{},
	})

	volumeMounts := make([]VolumeMount, 0)
	volumeMounts = append(volumeMounts, VolumeMount{
		Name:      "bin",
		MountPath: "/opt/bin",
	})

	container := Container{
		Args:            config.Args,
		Command:         []string{filepath.Join("/opt/bin", config.Name)},
		Image:           "gcr.io/hightowerlabs/alpine",
		ImagePullPolicy: "Always",
		Name:            config.Name,
		VolumeMounts:    volumeMounts,
	}

	resourceLimits := make(ResourceList)
	if config.cpuLimit != "" {
		resourceLimits["cpu"] = config.cpuLimit
	}
	if config.memoryLimit != "" {
		resourceLimits["memory"] = config.memoryLimit
	}

	resourceRequests := make(ResourceList)
	if config.cpuRequest != "" {
		resourceRequests["cpu"] = config.cpuRequest
	}
	if config.memoryRequest != "" {
		resourceRequests["memory"] = config.memoryRequest
	}

	if len(resourceLimits) > 0 {
		container.Resources.Limits = resourceLimits
	}
	if len(resourceRequests) > 0 {
		container.Resources.Requests = resourceRequests
	}

	if len(config.Env) > 0 {
		env := make([]EnvVar, 0)
		for name, value := range config.Env {
			env = append(env, EnvVar{Name: name, Value: value})
		}
		container.Env = env
	}

	annotations := config.Annotations

	binaryPath := filepath.Join("/opt/bin", config.Name)
	initContainers := []Container{
		Container{
			Name:            "install",
			Image:           "gcr.io/hightowerlabs/alpine",
			ImagePullPolicy: "Always",
			Command:         []string{"wget", "-O", binaryPath, config.BinaryURL},
			VolumeMounts: []VolumeMount{
				VolumeMount{
					Name:      "bin",
					MountPath: "/opt/bin",
				},
			},
		},
		Container{
			Name:            "configure",
			Image:           "gcr.io/hightowerlabs/alpine",
			ImagePullPolicy: "Always",
			Command:         []string{"chmod", "+x", binaryPath},
			VolumeMounts: []VolumeMount{
				VolumeMount{
					Name:      "bin",
					MountPath: "/opt/bin",
				},
			},
		},
	}

	ic, err := json.MarshalIndent(&initContainers, "", " ")
	if err != nil {
		return err
	}
	annotations["pod.alpha.kubernetes.io/init-containers"] = string(ic)

	config.Labels["run"] = config.Name

	rs := ReplicaSet{
		ApiVersion: "extensions/v1beta1",
		Kind:       "ReplicaSet",
		Metadata: Metadata{
			Name:      config.Name,
			Namespace: config.Namespace,
		},
		Spec: ReplicaSetSpec{
			Replicas: int64(config.Replicas),
			Selector: LabelSelector{
				MatchLabels: config.Labels,
			},
			Template: PodTemplate{
				Metadata: Metadata{
					Labels:      config.Labels,
					Annotations: annotations,
				},
				Spec: PodSpec{
					Containers: []Container{container},
					Volumes:    volumes,
				},
			},
		},
	}

	var b []byte
	body := bytes.NewBuffer(b)
	err = json.NewEncoder(body).Encode(rs)
	if err != nil {
		return err
	}

	path := fmt.Sprintf(replicasetsEndpoint, config.Namespace)
	request := &http.Request{
		Body:          ioutil.NopCloser(body),
		ContentLength: int64(body.Len()),
		Header:        make(http.Header),
		Method:        http.MethodPost,
		URL: &url.URL{
			Host:   apiHost,
			Path:   path,
			Scheme: "http",
		},
	}
	request.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		return err
	}

	if resp.StatusCode != 201 {
		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		fmt.Println(string(data))
		return errors.New("ReplicaSet: Unexpected HTTP status code" + resp.Status)
	}

	return nil
}
