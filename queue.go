package main

import (
	"github.com/Shopify/sarama"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/ec2rolecreds"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/firehose"
	"log"
	"os"
	"strings"
)

type Queue interface {
	PutRecord(channel string, data []byte) error
}

type Firehose struct {
	svc *firehose.Firehose
}

func (this *Firehose) PutRecord(channel string, data []byte) error {
	params := &firehose.PutRecordInput{
		DeliveryStreamName: aws.String(channel),
		Record: &firehose.Record{
			Data: data,
		},
	}
	_, err := this.svc.PutRecord(params)
	return err
}

func newFirehose(filename, profile, region string) (*Firehose, error) {
	return &Firehose{svc: newFirehoseService(filename, profile, region)}, nil
}

func newFirehoseService(filename, profile, region string) *firehose.Firehose {
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

//
type Kafka struct {
	sp sarama.SyncProducer
}

// Create a new Kafka synchronized producer
func NewKafkaSyncProducer(brokers string) (*Kafka, error) {
	brokerList := strings.Split(brokers, ",")
	log.Printf("Kafka brokers: %s", strings.Join(brokerList, ", "))
	// create Kafka client configuration
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll // Wait for all in-sync replicas to ack the message
	config.Producer.Retry.Max = 10                   // Retry up to 10 times to produce the message
	// tlsConfig := createTlsConfiguration()
	producer, err := sarama.NewSyncProducer(brokerList, config)
	if err != nil {
		log.Println("Failed to start Sarama producer:", err)
		return nil, err
	}
	return &Kafka{sp: producer}, nil
}

// publish an array of data on a Kafka channel
func (this *Kafka) PutRecord(channel string, data []byte) error {
	//key := RandomString(8)
	msg := sarama.ProducerMessage{Topic: channel,
		//Key:   sarama.StringEncoder(key),
		Value: sarama.ByteEncoder(data)}
	partition, offset, err := this.sp.SendMessage(&msg)
	log.Printf("Message is stored in topic(%s)/partition(%d)/offset(%d)\n", channel, partition, offset)
	return err
}
