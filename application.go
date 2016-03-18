package main // import "github.com/dzlab/qbench"

import (
	"flag"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/firehose"
	"github.com/dzlab/qbench/bench"
	"github.com/satori/go.uuid"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"
)

var (
	// common flags
	commons = flag.NewFlagSet("", flag.ContinueOnError)
	creds   = commons.String("credentials", "", "filename of AWS credentials")
	topic   = commons.String("c", "auction_stream", "Channel/topic on which the event will be pushed")
	port    = commons.String("p", "", "Port number to listen on")
	msg     = commons.String("m", "", "Type of messages: bytes, json")
	total   = commons.Int("t", 1000, "Total number of messages to send upstream")
	size    = commons.Int("s", -1, "Size of messages to send upstream")
	delay   = commons.Float64("d", 0.0, "Delay in milliseconds between two subsequent requests")
	workdir = commons.String("o", "/tmp", "Working directory where metrics files will be stored")
	// http related args
	httpCmd    = flag.NewFlagSet("http", flag.ContinueOnError)
	httpUrl    = httpCmd.String("u", "", "http endpoint url")
	httpMethod = httpCmd.String("m", "", "HTTP endpoint method")
	// firehose related args
	firehoseCmd = flag.NewFlagSet("firehose", flag.ContinueOnError)
	// kfka related args
	kafkaCmd = flag.NewFlagSet("kafka", flag.ContinueOnError)
	kBrokers = kafkaCmd.String("b", "", "Kafka brokers")
)

type Msg struct {
	id        string
	timestamp string
	size      int
}

func putRecord(svc bench.Queue, topic string, input <-chan []byte, ro chan<- []byte, do chan<- []byte, size int, total int, delay int64) {
	uploader := bench.NewRecordUploader(svc, input, ro, do, size)
	// pre-put
	uploader.PreUpload()
	// put
	uploader.Upload(topic, total, delay)
	// post-put
	uploader.PostUpload()
	log.Printf("Pushed a message of %d bytes %d times\n", size, total)
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

func run(svc bench.Queue, topic string) { //brokers, topic string) {
	// Instantiate rand per producer to avoid mutex contention.
	var generator bench.RecordGenerator
	//totals := []int{1000} //, 5000, 10000, 20000, 50000}
	totals := []int{*total}
	//batchs := []int{50, 100, 200, 350, 500}
	var sizes []int
	if *size > -1 {
		sizes = []int{*size}
	} else {
		sizes = []int{100, 300, 600, 1200, 2500, 10000}
	}
	var delay_ns int64 = int64(*delay * 1e6)
	var input <-chan []byte
	rw, _ := bench.NewFileWriter(*workdir + "/rates.dat")
	dw, _ := bench.NewFileWriter(*workdir + "/duration.dat")
	for _, size := range sizes {
		if *msg == "" {
			generator, _ = qbench.NewFixedSizeStringGenerator(size)
			input = generator.Generate()
		} else if *msg == "json" {
			msg := Msg{id: uuid.NewV4().String(), timestamp: time.Now().Format(time.RFC3339), size: size}
			generator, _ = bench.NewJsonGenerator("dump", "{'id': '{{.id}}', 'timestamp': {{.timestamp}}}", msg)
			input = generator.Generate()
		}
		for _, total := range totals {
			// testing put
			putRecord(svc, topic, input, rw.Output, dw.Output, size, total, delay_ns)
			/*for _, batch := range batchs {
				// testing put batch
				putRecordBuffered(svc, data, total, batch)
				putRecordBatch(svc, data, total, batch)
			}*/
		}
	}
	rw.Close()
	dw.Close()
	log.Println("Finished benchs")
}

func main2() {
	bucket := "adomik-firehose-dump"
	key := "2016/02/23/14/auction_stream-2-2016-02-23-14-41-29-f39a8176-9227-4e96-9fbd-41cfbc1a924a"
	sss := bench.NewS3("./credentials", "default", "eu-west-1")
	log.Println(sss.GetObject(bucket, key))
	log.Println(sss.List(bucket))
}

func main() {
	// check env variables
	if b := os.Getenv("BROKERS"); b != "" && *kBrokers == "" {
		*kBrokers = b
	}
	if p := os.Getenv("PORT"); p != "" && *port == "" {
		*port = p
	}
	// parse command line args
	for i := 2; !(firehoseCmd.Parsed() || httpCmd.Parsed() || kafkaCmd.Parsed()) && i < len(os.Args); i++ {
		switch os.Args[i-1] {
		case "http":
			httpCmd.Parse(os.Args[i:])
		case "firehose":
			firehoseCmd.Parse(os.Args[i:])
		case "kafka":
			kafkaCmd.Parse(os.Args[i:])
		default:
			//fmt.Printf("%q is not valid command.\n", os.Args[1])
		}
	}
	// common arguments should be provided at first
	commons.Parse(os.Args[1:])

	// create the endpoint to be tested
	var svc bench.Queue
	if firehoseCmd.Parsed() { // brokers == "firehose" {
		log.Println("Uploading to Firehose")
		svc, _ = bench.NewFirehose(*creds, "default", "eu-west-1")
	} else if httpCmd.Parsed() { //strings.HasPrefix(brokers, "http") {
		log.Println("Uploading to an HTTP endpoint")
		svc = bench.NewEndpoint(*httpUrl, *httpMethod, "Basic YWRtaW46YWRtaW4=")
	} else if kafkaCmd.Parsed() {
		log.Println("Uploading to a Kafka cluster")
		svc, _ = bench.NewKafkaSyncProducer(*kBrokers)
	}
	// run bench task
	go run(svc, *topic) //*brokers, *topic)
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
