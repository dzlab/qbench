package main

import (
	"flag"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/firehose"
	"github.com/satori/go.uuid"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"
)

var (
	chars   = []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890!@#$^&*(){}][:<>.")
	creds   = flag.String("credentials", "", "filename of AWS credentials")
	channel = flag.String("c", "auction_stream", "Channel/topic on which the event will be pushed")
	port    = flag.String("p", "", "Port number to listen on")
	brokers = flag.String("b", "", "Kafka brokers")
	msg     = flag.String("m", "", "Type of messages: bytes, json")
	total   = flag.Int("t", 1000, "Total number of messages to send upstream")
	size    = flag.Int("s", -1, "Size of messages to send upstream")
	delay   = flag.Int("d", 0, "Delay in milliseconds between two subsequent requests")
)

type Msg struct {
	id        string
	timestamp string
	size      int
}

func putRecord(svc Queue, channel string, data []byte, total int, delay int) {
	uploader := NewRecordUploader(svc, len(data))
	// pre-put
	uploader.PreUpload()
	// put
	uploader.Upload(channel, data, total, delay)
	// post-put
	uploader.PostUpload()
	fmt.Printf("Pushed a message of %d bytes %d times\n", len(data), total)
}

func putRecordBuffered(svc *firehose.Firehose, channel string, data []byte, total, batch int) {
	ts := time.Now()
	for i := 0; i < total/batch; i++ {
		var buf []byte
		for j := 0; j < batch; j++ {
			buf = append(buf, data...)
			buf = append(buf, byte('\n'))
		}
		params := &firehose.PutRecordInput{
			DeliveryStreamName: aws.String(channel),
			Record:             &firehose.Record{Data: buf},
		}
		svc.PutRecord(params)
	}
	duration := float64(time.Since(ts)) / float64(time.Millisecond)
	//log.Printf("Pushed a total of %d (msg size %d) buffered into  %d in %fms", total, len(data), batch, duration)
	fmt.Printf("%d,%d,%d,%f\n", total, len(data), batch, duration)
}

func putRecordBatch(svc *firehose.Firehose, channel string, data []byte, total, batch int) {
	ts := time.Now()
	for i := 0; i < total/batch; i++ {
		var records []*firehose.Record
		for j := 0; j < batch; j++ {
			records = append(records, &firehose.Record{Data: data})
		}
		params := &firehose.PutRecordBatchInput{
			DeliveryStreamName: aws.String(channel),
			Records:            records,
		}
		svc.PutRecordBatch(params)
	}
	duration := float64(time.Since(ts)) / float64(time.Millisecond)
	//log.Printf("Pushed a total of %d (msg size %d) in batch of %d in %fms", total, len(data), batch, duration)
	fmt.Printf("%d,%d,%d,%f\n", total, len(data), batch, duration)
}

func run(brokers, channel string) {
	var svc Queue
	if brokers == "firehose" {
		svc, _ = newFirehose(*creds, "default", "eu-west-1")
	} else if strings.HasPrefix(brokers, "http") {
		svc = NewEndpoint(brokers)
	} else {
		svc, _ = NewKafkaSyncProducer(brokers)
	}
	// Instantiate rand per producer to avoid mutex contention.
	var generator RecordGenerator
	if *msg == "" {
		source := rand.NewSource(time.Now().UnixNano())
		generator = &BytesGenerator{generator: rand.New(source)}
	} else if *msg == "json" {
		generator, _ = NewJsonGenerator("dump", "{'id': '{{.id}}', 'timestamp': {{.timestamp}}}")
	}
	//totals := []int{1000} //, 5000, 10000, 20000, 50000}
	totals := []int{*total}
	//batchs := []int{50, 100, 200, 350, 500}
	var sizes []int
	if *size > -1 {
		sizes = []int{*size}
	} else {
		sizes = []int{100, 300, 600, 1200, 2500, 10000}
	}
	var data []byte
	for _, size := range sizes {
		if *msg == "" {
			data = generator.Generate(size)
		} else if *msg == "json" {
			msg := Msg{id: uuid.NewV4().String(), timestamp: time.Now().Format(time.RFC3339), size: size}
			data = generator.Generate(msg)
		}
		for _, total := range totals {
			// testing put
			putRecord(svc, channel, data, total, *delay)
			/*for _, batch := range batchs {
				// testing put batch
				putRecordBuffered(svc, data, total, batch)
				putRecordBatch(svc, data, total, batch)
			}*/
		}
	}
	output()
	log.Println("Finished benchs")
}

func main2() {
	bucket := "adomik-firehose-dump"
	key := "2016/02/23/14/auction_stream-2-2016-02-23-14-41-29-f39a8176-9227-4e96-9fbd-41cfbc1a924a"
	sss := NewS3("./credentials", "default", "eu-west-1")
	log.Println(sss.GetObject(bucket, key))
	log.Println(sss.List(bucket))
}

func main() {
	flag.Parse()
	if b := os.Getenv("BROKERS"); b != "" && *brokers == "" {
		*brokers = b
	}
	if p := os.Getenv("PORT"); p != "" && *port == "" {
		*port = p
	}
	// run bench task
	go run(*brokers, *channel)
	// serve http (for aws beanstalk)
	log.Printf("Listening on port %s..\n", *port)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, %q", r.URL.Path)
	})
	err := http.ListenAndServe(":"+*port, nil)
	if err != nil {
		log.Fatal(err)
	}
}
