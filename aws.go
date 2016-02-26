package main

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/ec2rolecreds"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"io/ioutil"
	"log"
	"os"
)

type S3 struct {
	svc *s3.S3
}

func NewS3(filename, profile, region string) *S3 {
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
	return &S3{svc: s3.New(sess)}
}

func (this *S3) List(bucket string) error {
	params := &s3.ListObjectsInput{
		Bucket: aws.String(bucket),
	}

	resp, err := this.svc.ListObjects(params)
	if err != nil {
		log.Println("Failed to list bucket keys", err)
		return err
	}
	for _, key := range resp.Contents {
		log.Println(*key.Key)
	}
	return nil
}

func (this *S3) GetObject(bucket, key string) error {
	params := &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}
	out, err := this.svc.GetObject(params)
	if err != nil {
		log.Println("Failed to get object from S3", err)
		return err
	}
	bytes, err := ioutil.ReadAll(out.Body)
	if err != nil {
		log.Println("Failed to read S3 response body", err)
		return err
	}
	err = ioutil.WriteFile("/tmp/"+bucket, bytes, os.ModePerm)
	if err != nil {
		log.Println("Failed to write S3 file locally", err)
		return err
	}
	return nil
}

func main_s3() {
	bucket := "adomik-firehose-dump"
	key := "2016/02/23/14/auction_stream-2-2016-02-23-14-41-29-f39a8176-9227-4e96-9fbd-41cfbc1a924a"
	sss := NewS3("", "default", "eu-west-1")
	sss.GetObject(bucket, key)
}
