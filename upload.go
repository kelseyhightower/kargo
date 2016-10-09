package kargo

import (
	"crypto/sha256"
	"encoding/hex"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
	storage "google.golang.org/api/storage/v1"
)

const (
	scope = storage.DevstorageFullControlScope
)

type UploadConfig struct {
	BucketName string
	ObjectName string
	Path       string
	ProjectID  string
}

func Build(name string) (string, error) {
	tmpDir, err := ioutil.TempDir("", "")
	if err != nil {
		return "", err
	}
	output := filepath.Join(tmpDir, name)

	command := []string{
		"go", "build", "-o", output, "-a", "--ldflags",
		"'-extldflags \"-static\"'", "-tags", "netgo",
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
		return "", err
	}

	return string(data), nil
}

func Upload(config UploadConfig) (string, error) {
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

	acl := &storage.ObjectAccessControl{
		Bucket: config.BucketName,
		Entity: "allUsers",
		Object: objectName,
		Role:   "READER",
	}

	object := &storage.Object{
		Acl:      []*storage.ObjectAccessControl{acl},
		Name:     objectName,
		Metadata: metadata,
	}

	response, err := service.Objects.Insert(config.BucketName, object).Media(f).Do()
	if err != nil {
		return "", err
	}

	return response.SelfLink, nil
}
