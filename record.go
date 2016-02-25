package main

import (
	"log"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

type RecordUploader struct {
	sk     string
	fk     string
	dk     string
	sent   uint32
	failed uint32
	svc    Queue
	ticker *time.Ticker
}

func NewRecordUploader(svc Queue, size int) *RecordUploader {
	return &RecordUploader{
		svc:    svc,
		sk:     newKey("rate-" + strconv.Itoa(size) + "-sent"),
		fk:     newKey("rate-" + strconv.Itoa(size) + "-failed"),
		dk:     newKey("duration-" + strconv.Itoa(size)),
		sent:   0,
		failed: 0,
	}
}

// do something before starting the upload
func (this *RecordUploader) PreUpload() {
	this.ticker = time.NewTicker(time.Second * 1)
	go func() {
		for _ = range this.ticker.C {
			// report metrics at each tick
			report(this.sk, &(this.sent))
			report(this.fk, &(this.failed))
		}
	}()
}

// do the real upload
func (this *RecordUploader) Upload(channel string, data []byte, total int, delay int) {
	var wg sync.WaitGroup
	wg.Add(total)
	if delay > 0 {
		// send requests after delay (ms)
		ticker := time.NewTicker(time.Duration(delay) * time.Millisecond)
		count := 0
	TickLoop:
		for t := range ticker.C {
			log.Printf("Ticked at %s", t)
			go this.SyncUpload(&wg, channel, data)
			count = count + 1
			if count == total {
				break TickLoop
			}
		}
		ticker.Stop()
	} else {
		// send burst requests
		for i := 0; i < total; i++ {
			go this.SyncUpload(&wg, channel, data)
		}
	}
	wg.Wait()
}

// execute a synchronous upload
func (this *RecordUploader) SyncUpload(wg *sync.WaitGroup, channel string, data []byte) {
	defer wg.Done()
	duration, err := Time(func() error {
		err := this.svc.PutRecord(channel, data)
		return err
	})
	if err != nil {
		// Print the error, cast err to awserr.Error to get the Code and Message from an error.
		log.Println("Failed to put record", err)
		atomic.AddUint32(&(this.failed), 1)
	} else {
		atomic.AddUint32(&(this.sent), 1)
	}
	reportFloat64(this.dk, float64(duration.Nanoseconds())/float64(time.Millisecond))
}

// do something after finishing upload
func (this *RecordUploader) PostUpload() {
	this.ticker.Stop()
	// report metrics
	report(this.sk, &(this.sent))
	report(this.fk, &(this.failed))
}
