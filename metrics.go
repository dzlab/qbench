package main

import (
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

var (
	metrics = make(map[string][]float64)
	keys    []string
)

func report(key string, value *uint32) {
	metrics[key] = append(metrics[key], float64(atomic.LoadUint32(value)))
	atomic.StoreUint32(value, 0)
}

func reportFloat64(key string, value float64) {
	metrics[key] = append(metrics[key], value)
}

func newKey(key string) string {
	keys = append(keys, key)
	return key
}

func Time(f func() error) (time.Duration, error) {
	ts := time.Now()
	err := f()
	duration := time.Since(ts)
	//float64(time.Since(ts)) / float64(time.Millisecond)
	//reportFloat64("duration", duration)
	return duration, err
}

// print out metrics
func output() {
	var wg sync.WaitGroup
	wg.Add(2)
	// print rates
	go func() {
		defer wg.Done() // signal task done
		rk := Filter(keys, func(k string) bool { return strings.HasPrefix(k, "rate") })
		store("/tmp/rates.dat", append([]string{"time"}, rk...))
	}()
	// print durations
	go func() {
		defer wg.Done() // signal task done
		dk := Filter(keys, func(k string) bool { return strings.HasPrefix(k, "duration") })
		store("/tmp/duration.dat", dk)
	}()
	// wait for all task to finish
	wg.Wait()
}

// store a collection of metrics into a file
func store(filename string, keys []string) {
	f, err := os.Create(filename)
	if err != nil {
		log.Printf("Failed to open %s\n", filename)
		log.Println(err)
	}
	defer f.Close()
	// print header
	header := ""
	for index, key := range keys {
		if index > 0 {
			header = header + ","
		}
		header = header + key
	}
	f.WriteString(header + "\n")
	log.Println(header)
	// print values
	t := 1
	loop := true
	for loop {
		row := strconv.Itoa(t)
		loop = false
		for index, key := range keys {
			if index > 0 {
				row = row + ","
			}
			value := metrics[key]
			if t <= len(value) {
				row = row + strconv.FormatFloat(value[t-1], 'f', 2, 32)
				loop = true
			}
		}
		if loop {
			f.WriteString(row + "\n")
			log.Println(row)
			t++
		}
	}
	f.Sync()
}
