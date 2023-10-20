package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"time"
)

type Bucket struct {
	Kind             string    `json:"kind"`
	SelfLink         string    `json:"selfLink"`
	ID               string    `json:"id"`
	Name             string    `json:"name"`
	ProjectNumber    string    `json:"projectNumber"`
	Metageneration   string    `json:"metageneration"`
	Location         string    `json:"location"`
	StorageClass     string    `json:"storageClass"`
	Etag             string    `json:"etag"`
	TimeCreated      time.Time `json:"timeCreated"`
	UpdateTime       time.Time `json:"updated"`
	IamConfiguration struct {
		BucketPolicyOnly struct {
			Enabled bool `json:"enabled"`
		} `json:"bucketPolicyOnly"`
		UniformBucketLevelAccess struct {
			Enabled bool `json:"enabled"`
		} `json:"uniformBucketLevelAccess"`
		PublicAccessPrevention string `json:"publicAccessPrevention"`
	} `json:"iamConfiguration"`
	LocationType          string `json:"locationType"`
	Rpo                   string `json:"rpo,omitempty"`
	DefaultEventBasedHold bool   `json:"defaultEventBasedHold,omitempty"`
	Cors                  []struct {
		Origin         []string `json:"origin"`
		Method         []string `json:"method"`
		ResponseHeader []string `json:"responseHeader"`
		MaxAgeSeconds  int      `json:"maxAgeSeconds"`
	} `json:"cors,omitempty"`
	Lifecycle struct {
		Rule []struct {
			Action struct {
				Type string `json:"type"`
			} `json:"action"`
			Condition struct {
				Age int `json:"age"`
			} `json:"condition"`
		} `json:"rule"`
	} `json:"lifecycle,omitempty"`
	SatisfiesPZS bool `json:"satisfiesPZS,omitempty"`
	Versioning   struct {
		Enabled bool `json:"enabled"`
	} `json:"versioning,omitempty"`
}

func SyncingBuckets(sourceBucket string, targetBucket string, delete bool) error {
	sourceBucketUrl := fmt.Sprintf("gs://%s", sourceBucket)
	targetBucketUrl := fmt.Sprintf("gs://%s", targetBucket)
	var cmd *exec.Cmd
	if delete {
		cmd = exec.Command("gsutil", "-o", "GSUtil:use_magicfile=True", "-o", "GSUtil:parallel_composite_upload_threshold=150M", "-o", "GSUtil:sliced_object_download_threshold=20M", "-o", "GSUtil:sliced_object_download_max_components=16", "-m", "rsync", "-d", "-r", sourceBucketUrl, targetBucketUrl)
	} else {
		cmd = exec.Command("gsutil", "-o", "GSUtil:use_magicfile=True", "-o", "GSUtil:parallel_composite_upload_threshold=150M", "-o", "GSUtil:sliced_object_download_threshold=20M", "-o", "GSUtil:sliced_object_download_max_components=16", "-m", "rsync", "-r", sourceBucketUrl, targetBucketUrl)
	}
	// Create a pipe to capture stdout and stderr
	stdoutPipe, _ := cmd.StdoutPipe()
	stderrPipe, _ := cmd.StderrPipe()

	// Start the command
	err := cmd.Start()
	if err != nil {
		return err
	}

	// Print the logs from stdout and stderr in real-time
	go io.Copy(os.Stdout, stdoutPipe)
	go io.Copy(os.Stderr, stderrPipe)

	// Wait for the command to finish
	err = cmd.Wait()
	if err != nil {
		return err
	}
	return nil
}
func CheckBucket(bucket string, project string) (Bucket, error) {
	url := fmt.Sprintf("https://storage.googleapis.com/storage/v1/b/%s?project=%s", bucket, project)
	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return Bucket{}, err
	}

	res, err := client.Do(req)
	if err != nil {
		return Bucket{}, err
	}
	defer res.Body.Close()

	var result Bucket
	err = json.NewDecoder(res.Body).Decode(&result)
	if err != nil {
		return Bucket{}, err
	}
	return result, nil
}

func main() {
	log.SetFlags(0)

	targetProject := flag.String("target-project", "", "the target project id to sync assets")
	targetBucket := flag.String("target-bucket", "", "the target bucket to sync assets")
	sourceProject := flag.String("source-project", "", "the source project id to sync assets from")
	sourceBucket := flag.String("source-bucket", "", "the source bucket to sync assets from")
	enableDelete := flag.Bool("delete", false, "delete the item when it's not contained in the source bucket")
	flag.Parse()

	if *targetProject == "" || *targetBucket == "" || *sourceProject == "" || *sourceBucket == "" {
		log.Fatal("Error: Parameter 'target-project', 'target-bucket', 'source-project' and 'source-bucket' are required")
	}

	var err error
	_, err = CheckBucket(*sourceBucket, *sourceProject)
	if err != nil {
		log.Fatal(err)
	}
	_, err = CheckBucket(*targetBucket, *targetProject)
	if err != nil {
		log.Fatal(err)
	}

	err = SyncingBuckets(*sourceBucket, *targetBucket, *enableDelete)
	if err != nil {
		log.Fatal(err)
	}

	_, err = CheckBucket(*targetBucket, *targetProject)
	if err != nil {
		log.Fatal(err)
	}
}
