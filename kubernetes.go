package kargo

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"path/filepath"
)

var (
	apiHost             = "127.0.0.1:8001"
	replicasetsEndpoint = "/apis/extensions/v1beta1/namespaces/default/replicasets"
)

var ErrNotExist = errors.New("does not exist")

func getReplicaSet(name string) (*ReplicaSet, error) {
	var rs ReplicaSet

	path := fmt.Sprintf("%s/%s", replicasetsEndpoint, name)

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


func getScale(name string) (*Scale, error) {
	var scale Scale

	path := fmt.Sprintf("%s/%s/scale", replicasetsEndpoint, name)

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

func scaleReplicaSet(name string, replicas int) error {
	scale, err := getScale(name)
	if err != nil {
		return err
	}

	scale.Spec.Replicas = int64(replicas)
	path := fmt.Sprintf("%s/%s/scale", replicasetsEndpoint, name)

	var b []byte
	body := bytes.NewBuffer(b)
	err = json.NewEncoder(body).Encode(scale)
	if err != nil {
		return err
	}

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

		log.Println(string(data))

		return errors.New("Scale ReplicaSet error non 200 reponse: " + resp.Status)
	}
	return nil
}

func deleteReplicaSet(name string) error {
	err := scaleReplicaSet(name, 0)
	if err != nil {
		return err
	}

	path := fmt.Sprintf("%s/%s", replicasetsEndpoint, name)

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
		MountPath: "/bin",
	})

	container := Container{
		Args:         config.Args,
		Command:      []string{filepath.Join("/", config.Name)},
		Image:        "gcr.io/hightowerlabs/alpine",
		Name:         config.Name,
		VolumeMounts: volumeMounts,
	}

	annotations := config.Annotations

	binaryPath := filepath.Join("/bin", config.Name)
	initContainers := []Container{
		Container{
			Name:    "install",
			Image:   "gcr.io/hightowerlabs/alpine",
			Command: []string{"wget", "-O", binaryPath, config.BinaryURL},
			VolumeMounts: []VolumeMount{
				VolumeMount{
					Name:      "bin",
					MountPath: "/bin",
				},
			},
		},
		Container{
			Name:    "configure",
			Image:   "gcr.io/hightowerlabs/alpine",
			Command: []string{"chmod", "+x", binaryPath},
			VolumeMounts: []VolumeMount{
				VolumeMount{
					Name:      "bin",
					MountPath: "/bin",
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
		Metadata:   Metadata{Name: config.Name},
		Spec: ReplicaSetSpec{
			Replicas: int64(config.Replicas),
			Template: PodTemplate{
				Metadata: Metadata{
					Labels:      config.Labels,
					Annotations: annotations,
				},
				Spec: PodSpec{
					Containers: []Container{container},
					Volumes: volumes,
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

	request := &http.Request{
		Body:          ioutil.NopCloser(body),
		ContentLength: int64(body.Len()),
		Header:        make(http.Header),
		Method:        http.MethodPost,
		URL: &url.URL{
			Host:   apiHost,
			Path:   replicasetsEndpoint,
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

		log.Println(string(data))
		return errors.New("ReplicaSet: Unexpected HTTP status code" + resp.Status)
	}
	return nil
}
