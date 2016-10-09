package kargo

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/googleapi"
	storage "google.golang.org/api/storage/v1"
)

const (
	scope = storage.DevstorageFullControlScope
)

type UploadConfig struct {
	BucketName string
	ObjectName string
	ProjectID  string
	Path       string
}

func build(name string) (string, error) {
	tmpDir, err := ioutil.TempDir("", "")
	if err != nil {
		return "", err
	}
	output := filepath.Join(tmpDir, name)

	ldflags := `-extldflags "-static"`
	command := []string{
		"go", "build", "-o", output, "-a", "--ldflags",
		ldflags, "-tags", "netgo",
		"-installsuffix", "netgo", ".",
	}
	cmd := exec.Command(command[0], command[1:]...)

	gopath := os.Getenv("GOPATH")
	cmd.Env = []string{
		"GOOS=linux",
		"GOARCH=amd64",
		"GOPATH=" + gopath,
	}

	data, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println(string(data))
		return "", err
	}

	return output, nil
}

func Upload(config UploadConfig) (string, error) {
	if config.Path == "" {
		fmt.Printf("Building %s binary...\n", config.ObjectName)
		output, err := build(config.ObjectName)
		if err != nil {
			return "", err
		}
		config.Path = output
		fmt.Println("Created: " + config.Path)
	}

	client, err := google.DefaultClient(context.Background(), scope)
	if err != nil {
		return "", err
	}

	service, err := storage.New(client)
	if err != nil {
		return "", err
	}

	_, err = service.Buckets.Get(config.BucketName).Do()
	if err != nil {
		_, err := service.Buckets.Insert(config.ProjectID, &storage.Bucket{Name: config.BucketName}).Do()
		if err != nil {
			return "", err
		}
	}

	f, err := os.Open(config.Path)
	if err != nil {
		return "", err
	}

	data, err := ioutil.ReadAll(f)
	if err != nil {
		return "", err
	}
	f.Seek(0, 0)

	h := sha256.New()
	h.Write(data)
	checksum := hex.EncodeToString(h.Sum(nil))
	metadata := make(map[string]string)
	metadata["sha256"] = checksum

	objectName := filepath.Join(checksum, config.ObjectName)

	publicLink := fmt.Sprintf("https://storage.googleapis.com/%s/%s", config.BucketName, objectName)
	object, err := service.Objects.Get(config.BucketName, objectName).Do()

	if object != nil {
		if object.HTTPStatusCode == 200 {
			fmt.Printf("Object %s already exists, skipping upload.\n", filepath.Join(config.BucketName, objectName))
			return publicLink, nil
		}
	}

	if err != nil {
		if err.(*googleapi.Error).Code != 404 {
			return "", err
		}
	}

	acl := &storage.ObjectAccessControl{
		Bucket: config.BucketName,
		Entity: "allUsers",
		Object: objectName,
		Role:   "READER",
	}

	object = &storage.Object{
		Acl:      []*storage.ObjectAccessControl{acl},
		Name:     objectName,
		Metadata: metadata,
	}

	fmt.Printf("Uploading %s to the %s bucket...\n", object.Name, config.BucketName)
	_, err = service.Objects.Insert(config.BucketName, object).Media(f).Do()
	if err != nil {
		return "", err
	}
	fmt.Println("Upload complete.")
	return publicLink, nil
}
