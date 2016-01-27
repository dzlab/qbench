package main

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/ec2rolecreds"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/firehose"
	"log"
	"os"
)

func newFirehose(filename, profile, region string) *firehose.Firehose {
	var creds *credentials.Credentials
	if _, err := os.Stat(filename); err == nil {
		log.Printf("Connecting to AWS using credentials from '%s'", filename)
		creds = credentials.NewSharedCredentials(filename, profile)
	} else {
		log.Printf("AWS credentials file '%s' dosen't exists, I will be using EC2 Role credentials", filename)
		sess := session.New()
		creds = credentials.NewCredentials(&ec2rolecreds.EC2RoleProvider{Client: ec2metadata.New(sess)})
	}
	sess := session.New(&aws.Config{Credentials: creds, Region: aws.String(region)})
	return firehose.New(sess)
}
