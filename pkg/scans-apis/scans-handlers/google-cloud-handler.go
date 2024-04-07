package scans_handlers

import (
	"cloud.google.com/go/storage"
	"context"
	"fmt"
	"github.com/joho/godotenv"
	"io"
	"log"
	"mime/multipart"
	"os"
	"time"
)

func GoogleCloudHandler(username string, imageName string, image multipart.File) (string, error) {
	//loading env file
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error occurred on .env file please check")
	}

	projectID := os.Getenv("PROJECT_ID")

	//creating parent context
	parentCtx := context.Background()

	//creating a client
	client, err := storage.NewClient(parentCtx)
	if err != nil {
		log.Fatalf("Failed to create client:, %v", err)
	}
	defer client.Close()

	//creating bucket only if it does not exist
	ctx, cancel := context.WithTimeout(parentCtx, 10*time.Second)
	defer cancel()

	bucketName := "capstone-project-" + username
	bucket := client.Bucket(bucketName)
	_, err = bucket.Attrs(ctx)
	if err == storage.ErrBucketNotExist {
		// Bucket does not exist, so create it
		fmt.Printf("Bucket %s does not exist, creating it now", bucketName)
		if err := bucket.Create(ctx, projectID, nil); err != nil {
			log.Fatalf("Failed to create bucket: %v\n", err)
		}
	} else if err != nil {
		// Some other error occurred when accessing bucket attributes
		log.Fatalf("Error accessing bucket attributes: %v\n", err)
	} else {
		// No error, so the bucket already exists
		fmt.Printf("Bucket %s already exists\n", bucketName)
	}

	//uploading the object
	obj := client.Bucket(bucketName).Object(imageName)

	//uploading object with storage.Writer
	wc := obj.NewWriter(ctx)
	if _, err = io.Copy(wc, image); err != nil {
		return "", fmt.Errorf("io copy error: %v\n", err)
	}
	if err := wc.Close(); err != nil {
		return "", fmt.Errorf("Writer close: %v\n", err)
	}

	fmt.Printf("Scan Uploaded To Google Cloud\n")

	//TODO: return proper url
	url := fmt.Sprintf("https://storage.googleapis.com/%s/%s", bucket, imageName)

	return url, nil

}
