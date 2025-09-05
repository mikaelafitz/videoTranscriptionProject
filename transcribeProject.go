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
	"github.com/aws/aws-sdk-go-v2/service/mediaconvert"
	"github.com/aws/aws-sdk-go-v2/service/mediaconvert/types"
)

func main(){
	//ctx stops code when AWS takes more than 10 seconds to load
	ctx, cancel := context.WithTimeout(context.Background(), 180*time.Second)
	defer cancel()

	//log into aws
	cfg := logIn(ctx)

	//get original video file
	filePath, err := getVideo()
	if err != nil {
		log.Fatal("Error: ", err)
	}

	
	//upload video file to s3 bucket
	fileName, err := uploadFile(cfg, ctx, filePath)
	if err != nil {
		log.Fatal("Error: ", err)
	}

	//use MediaConvert to convert to mp4
	mediaConvert(cfg, ctx, fileName)

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

func uploadFile(cfg aws.Config, ctx context.Context, filePath string) (string, error){
	//create S3 client
	client := s3.NewFromConfig(cfg)

	//opens the file, returns error if cannot
	file, err := os.Open(filePath)
	if err != nil {
		return "",fmt.Errorf("failed to open video file: %w")
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
		return "",fmt.Errorf("failed to upload video into s3 bucket: %w", fileName, err)
	}

	//successfully uplaoded
	fmt.Println("Successfully uploaded video into s3 bucket")
	return fileName,nil
}

func mediaConvert(cfg aws.Config, ctx context.Context, fileName string) {
	//create MediaConvert client
	client := mediaconvert.NewFromConfig(cfg)

	//get account endpoint
	endpointsResp, err := client.DescribeEndpoints(ctx, &mediaconvert.DescribeEndpointsInput{})
	if err != nil {
		log.Fatalf("failed to describe endpoints: %v", err)
	}
	if len(endpointsResp.Endpoints) == 0 {
		log.Fatal("no MediaConvert endpoints found")
	}
	endpointURL := *endpointsResp.Endpoints[0].Url

	//rebuild client
	client = mediaconvert.NewFromConfig(cfg, func(o *mediaconvert.Options) {
		o.BaseEndpoint = aws.String(endpointURL)
	})

	//job settings
	loc := "s3://transcription-job-original-files/" + fileName
	fmt.Println("file location: ", loc)
	jobInput := &mediaconvert.CreateJobInput{
		Role: aws.String("arn:aws:iam::294560508449:role/service-role/MediaConvert_Role_With_Permissions"), // IAM role
		Settings: &types.JobSettings{
			Inputs: []types.Input{
				{
					FileInput: aws.String(loc),
					AudioSelectors: map[string]types.AudioSelector{
						"Audio Selector 1": {
							DefaultSelection: types.AudioDefaultSelectionDefault,
						},
					},
				},
			},
			OutputGroups: []types.OutputGroup{
				{
					Name: aws.String("mp4VideoFiles"),
					OutputGroupSettings: &types.OutputGroupSettings{
						Type: types.OutputGroupTypeFileGroupSettings,
						FileGroupSettings: &types.FileGroupSettings{
							Destination: aws.String("s3://transcription-job-mp4-files/"),
							DestinationSettings: &types.DestinationSettings{
								S3Settings: &types.S3DestinationSettings{
									StorageClass: types.S3StorageClassStandard,
								},
							},
						},
					},
					Outputs: []types.Output{
						{
							ContainerSettings: &types.ContainerSettings{
								Container: types.ContainerTypeMp4,
							},
							VideoDescription: &types.VideoDescription{
								CodecSettings: &types.VideoCodecSettings{
									Codec: types.VideoCodecH264,
									H264Settings: &types.H264Settings{
										RateControlMode: types.H264RateControlModeQvbr,
										QvbrSettings: &types.H264QvbrSettings{
											QvbrQualityLevel: aws.Int32(7),
										},
										MaxBitrate: aws.Int32(5000000),
									},
								},
							},
							AudioDescriptions: []types.AudioDescription{
								{
									AudioSourceName: aws.String("Audio Selector 1"),
									CodecSettings: &types.AudioCodecSettings{
										Codec: types.AudioCodecAac,
										AacSettings: &types.AacSettings{
											Bitrate: aws.Int32(96000),
											CodingMode: types.AacCodingModeCodingMode20,
											SampleRate: aws.Int32(48000),
											Specification: types.AacSpecificationMpeg4,
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	//submit job
	jobResp, err := client.CreateJob(ctx, jobInput)
	if err != nil {
		log.Fatalf("failed to create MediaConvert job: %v", err)
	}

	fmt.Printf("MediacConvert job created! ID: %s\n", *jobResp.Job.Id)
}