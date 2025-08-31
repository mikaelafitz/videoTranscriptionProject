package main

import (
	"context"
	"time"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
)

func main(){
	//ctx stops code when AWS takes more than 10 seconds to load
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	//check if filename is in the command line
	if len(os.Args) < 2 {
		fmt.Println("Please put file name in command line.")
		return
	}

	//access the video file
	videoFile := os.Args[1]

	//check if file is valid
	info, err := os.Stat(videoFile)
	if os.IsNotExist(err) {
		fmt.Println("file does not exist")
		return
	}
	if info.IsDir() {
		fmt.Println("this is a directory, not a file")
		return
	}
	fmt.Println("file is valid!")

	//log into aws
	cfg := logIn(ctx)

	//upload video file to s3 bucket
	uploadFile(cfg, ctx)
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

func uploadFile(cfg aws.Config, ctx context.Context) {

}