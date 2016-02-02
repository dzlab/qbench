package main

import (
	"flag"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/firehose"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"sync/atomic"
	"time"
)

const (
	channel = "auction_stream"
)

var (
	chars = []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890!@#$^&*(){}][:<>.")
	creds = flag.String("credentials", "", "filename of AWS credentials")
)

// Returns a random message generated from the chars byte slice.
// Message length of m bytes as defined by msgSize.
func randMsg(m []byte, generator rand.Rand) {
	for i := range m {
		m[i] = chars[generator.Intn(len(chars))]
	}
}

func putRecord(svc Queue, data []byte, total int) {
	//func putRecord(svc *firehose.Firehose, data []byte, total int) {
	// pre-put
	size := len(data)
	sk := newKey("rate-" + strconv.Itoa(size) + "-sent")
	fk := newKey("rate-" + strconv.Itoa(size) + "-failed")
	dk := newKey("duration-" + strconv.Itoa(size))
	var sent, failed uint32
	ticker := time.NewTicker(time.Second * 1)
	go func() {
		for _ = range ticker.C {
			//fmt.Println("Tick")
			report(sk, &sent)
			report(fk, &failed)
		}
	}()
	// put
	//ts := time.Now()
	for i := 0; i < total; i++ {
		duration, err := Time(func() error {
			err := svc.PutRecord(channel, data)
			return err
		})
		reportFloat64(dk, float64(duration)/float64(time.Millisecond))
		if err != nil {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			log.Println("Failed to put record", err)
			atomic.AddUint32(&failed, 1)
		} else {
			atomic.AddUint32(&sent, 1)
		}
	}
	//duration := float64(time.Since(ts)) / float64(time.Millisecond)
	// post-put
	ticker.Stop()
	report(sk, &sent)
	report(fk, &failed)
	//log.Printf("Pushed a total of %d messages (msg size %d) in %fms", total, len(data), duration)
	//fmt.Printf("%d,%d,%d,%f\n", total, len(data), 0, duration)
	fmt.Printf("Pushed a message of %d bytes %d times\n", len(data), total)
}

func putRecordBuffered(svc *firehose.Firehose, data []byte, total, batch int) {
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

func putRecordBatch(svc *firehose.Firehose, data []byte, total, batch int) {
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

func run(brokers string) {
	var svc Queue
	if brokers == "" {
		svc, _ = newFirehose(*creds, "default", "eu-west-1")
	} else {
		svc, _ = NewKafkaSyncProducer(brokers)
	}

	// Instantiate rand per producer to avoid mutex contention.
	source := rand.NewSource(time.Now().UnixNano())
	generator := rand.New(source)

	totals := []int{1000} //, 5000, 10000, 20000, 50000}
	//batchs := []int{50, 100, 200, 350, 500}
	sizes := []int{1000000, 2500, 1200, 600, 300, 100}
	//fmt.Printf("time,size,batch,duration\n")
	for _, size := range sizes {
		data := make([]byte, size)
		randMsg(data, *generator)
		for _, total := range totals {
			// testing put
			putRecord(svc, data, total)
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

func main() {
	brokers := os.Getenv("BROKERS")
	// run bench task
	go run(brokers)
	// serve http (for aws beanstalk)
	port := os.Getenv("PORT")
	log.Printf("Listening on port %s..\n", port)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, %q", r.URL.Path)
	})
	http.ListenAndServe(":"+port, nil)
}
