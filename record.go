package main

import (
	"fmt"
	//"github.com/streamrail/concurrent-map"
	"log"
	//"strconv"
	"sync"
	//"sync/atomic"
	"time"
)

type RecordUploader struct {
	rates   map[string]int
	rks     []string
	dks     []string
	size    int
	svc     Queue
	ticker  *time.Ticker
	channel <-chan []byte
}

func NewRecordUploader(svc Queue, channel <-chan []byte, size int) *RecordUploader {
	rks := []string{}
	dks := []string{}
	rates := make(map[string]int)
	for _, elm := range svc.PutResults() {
		rk := newKey(fmt.Sprintf("rate-%d-%s", size, elm))
		rks = append(rks, rk)
		rates[rk] = 0
		dk := newKey(fmt.Sprintf("duration-%d-%s", size, elm))
		dks = append(dks, dk)
	}
	return &RecordUploader{
		rates:   rates,
		svc:     svc,
		rks:     rks,
		dks:     dks,
		size:    size,
		channel: channel,
	}
}

// do something before starting the upload
func (this *RecordUploader) PreUpload() {
	this.ticker = time.NewTicker(time.Second * 1)
	go func() {
		for _ = range this.ticker.C {
			// report metrics at each tick
			for _, key := range this.rks {
				value, _ := this.rates[key]
				this.rates[key] = 0
				//log.Println("Tick", key, value, ok)
				reportFloat64(key, float64(value))
			}
		}
	}()
}

// do the real upload
func (this *RecordUploader) Upload(topic string, total int, delay int64) {
	var wg sync.WaitGroup
	wg.Add(total)
	if delay > 0 {
		// send requests after delay (ns)
		ticker := time.NewTicker(time.Duration(delay))
		count := 0
	TickLoop:
		for _ = range ticker.C {
			// do the upload
			go this.SyncUpload(&wg, topic)
			count = count + 1
			if count == total {
				break TickLoop
			}
		}
		ticker.Stop()
	} else {
		// send burst requests
		for i := 0; i < total; i++ {
			go this.SyncUpload(&wg, topic)
		}
	}
	wg.Wait()
	log.Println("All scheduled go-routines for upload have finished.")
}

// execute a synchronous upload
func (this *RecordUploader) SyncUpload(wg *sync.WaitGroup, topic string) {
	defer wg.Done()
	data := <-this.channel
	duration, res := Time(func() ResultType {
		res := this.svc.PutRecord(topic, data)
		return res
	})
	// increment rates
	rk := fmt.Sprintf("rate-%d-%s", this.size, res)
	current, _ := this.rates[rk]
	this.rates[rk] = current + 1
	// report duration
	dk := fmt.Sprintf("duration-%d-%s", this.size, res)
	reportFloat64(dk, float64(duration.Nanoseconds())/float64(time.Millisecond))
}

// do something after finishing upload
func (this *RecordUploader) PostUpload() {
	this.ticker.Stop()
	// report metrics
	for _, key := range this.rks {
		value, _ := this.rates[key]
		this.rates[key] = 0
		//log.Println("PostUpload", key, value)
		reportFloat64(key, float64(value))
	}
}
