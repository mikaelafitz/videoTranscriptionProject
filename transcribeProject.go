package main

import (
	"context"
	"time"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func main(){
	//ctx stops code when AWS takes more than 10 seconds to load
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	//log into aws
	cfg := logIn(ctx)

	//get original video file
	filePath, err := getVideo()
	if err != nil {
		log.Fatal("Error: ", err)
	}

	//upload video file to s3 bucket
	err = uploadFile(cfg, ctx, filePath)
	if err != nil {
		log.Fatal("Error: ", err)
	}

	fmt.Println("made it to the end")
}

func logIn(ctx context.Context) aws.Config{
	// in order for it to login, must do this in terminal:
		/*
		$$ aws configure
		IAM user's Access Key ID
		IAM user's Secret Access Key
		us-west-2
		text
		*/

	// Create a custom AWS configuration
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Fatal("failed to login to config: ", err)
	}

	fmt.Println("successfully logged into AWS IAM")
	return cfg
}

func getVideo() (string, error) {
	//check if filename is in the command line
	if len(os.Args) < 2 {
		return "", fmt.Errorf("Please put file name in command line")
	}

	//access the video file
	filePath := os.Args[1]

	//check if file is valid
	info, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		return "", fmt.Errorf("file does not exist")
	}
	if info.IsDir() {
		return "", fmt.Errorf("this is a directory, not a file")
	}
	fmt.Println("file is valid!")

	return filePath, err
}

func uploadFile(cfg aws.Config, ctx context.Context, filePath string) error{
	//create S3 client
	client := s3.NewFromConfig(cfg)

	//opens the file, returns error if cannot
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open video file: %w")
	}
	defer file.Close()

	//fileName
	fileName := filepath.Base(filePath)

	//upload file into s3 bucket
	_, err = client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String("transcription-job-original-files"),
		Key: aws.String(fileName),
		Body: file,
	})

	//check for errors
	if err != nil {
		return fmt.Errorf("failed to upload video into s3 bucket: %w", fileName, err)
	}

	//successfully uplaoded
	fmt.Println("Successfully uploaded video into s3 bucket")
	return nil
}