package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	ec2 "github.com/aws/aws-sdk-go/service/ec2"
	lambda "github.com/aws/aws-sdk-go/service/lambda"
)

func check(e error) {

	if e != nil {
		panic(e)
	}
}

var (
	// ErrNameNotProvided is thrown when a name is not provided
	ErrNameNotProvided = errors.New("no name was provided in the HTTP body")
)

var region = "us-east-1"
var id string
/*type CloudWatchEvent struct {
	Version    string          `json:"version"`
	ID         string          `json:"id"`
	DetailType string          `json:"detail-type"`
	Source     string          `json:"source"`
	AccountID  string          `json:"account"`
	Time       time.Time       `json:"time"`
	Region     string          `json:"region"`
	Resources  []string        `json:"resources"`
	Detail     json.RawMessage `json:"detail"`
}*/

func main() {
	lambda.Start(Handler)

	now := time.Now()
	t := fmt.Sprintln(now.String())
	fmt.Println(now.Format("2006-01-02 15:04:05"))

	// config aws region and pull Volume from instance
	svc := ec2.New(session.New(&aws.Config{Region: aws.String(region)}))
	input := &ec2.DescribeVolumesInput{
		Filters: []*ec2.Filter{
			{
				Name: aws.String("attachment.instance-id"),
				Values: []*string{
					aws.String(id),
				},
			},
		},
	}

	result, err := svc.DescribeVolumes(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			fmt.Println(err.Error())
		}
		return
	}

	fmt.Println(*result.Volumes[0].VolumeId)

	// Add tags to the created instance
	_, errtag := svc.CreateTags(&ec2.CreateTagsInput{
		Resources: []*string{result.Volumes[0].VolumeId},
		Tags: []*ec2.Tag{
			{
				Key:   aws.String("Name"),
				Value: aws.String("jjackson-tomcat9"),
			},
			{
				Key:   aws.String("env"),
				Value: aws.String("staging"),
			},
			{
				Key:   aws.String("date"),
				Value: aws.String(t),
			},
		},
	})

	if errtag != nil {
		log.Println("Could not create tags for volume", result.Volumes[0].VolumeId, errtag)
		return
	}
}
