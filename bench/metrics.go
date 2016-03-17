package bench

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
	mutex   = &sync.Mutex{}
)

func report(key string, value *uint32) {
	metrics[key] = append(metrics[key], float64(atomic.LoadUint32(value)))
	atomic.StoreUint32(value, 0)
}

func reportFloat64(key string, value float64) {
	mutex.Lock()
	metrics[key] = append(metrics[key], value)
	mutex.Unlock()
}

func newKey(key string) string {
	keys = append(keys, key)
	return key
}

func Time(f func() ResultType) (time.Duration, ResultType) {
	ts := time.Now()
	res := f()
	duration := time.Since(ts)
	return duration, res
}

// print out metrics
func output(workdir string) {
	var wg sync.WaitGroup
	wg.Add(2)
	// print rates
	go func() {
		defer wg.Done() // signal task done
		rk := Filter(keys, func(k string) bool { return strings.HasPrefix(k, "rate") })
		store(workdir+"/rates.dat", append([]string{"time"}, rk...))
	}()
	// print durations
	go func() {
		defer wg.Done() // signal task done
		dk := Filter(keys, func(k string) bool { return strings.HasPrefix(k, "duration") })
		store(workdir+"/duration.dat", append([]string{"time"}, dk...))
	}()
	// wait for all task to finish
	wg.Wait()
}

type FileWriter struct {
	f      *os.File
	Output chan []byte
}

func NewFileWriter(filename string) (*FileWriter, error) {
	f, err := os.Create(filename)
	if err != nil {
		log.Printf("Failed to open %s\n", filename)
		return nil, err
	}
	writer := &FileWriter{f: f}
	writer.Initialize()
	return writer, nil
}

func (this *FileWriter) Initialize() {
	channel := make(chan []byte)
	go func() {
		nw := []byte("\n")
		for {
			data := <-channel
			this.f.Write(data)
			this.f.Write(nw)
		}
	}()
	this.Output = channel
}

func (this *FileWriter) Close() {
	this.f.Close()
	close(this.Output)
}

// store a collection of metrics into a file
func store(filename string, keys []string) {
	f, err := os.Create(filename)
	if err != nil {
		log.Printf("Failed to open %s\n", filename)
		log.Println(err)
		return
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
	//log.Println(header)
	// print values
	t := 1
	loop := true
	for loop {
		row := strconv.Itoa(t - 1)
		loop = false
		for index, key := range keys {
			if index > 0 {
				row = row + ","
			}
			value := metrics[key]
			if t <= len(value) {
				row = row + strconv.FormatFloat(value[t-1], 'f', 2, 64)
				loop = true
			}
		}
		if loop {
			f.WriteString(row + "\n")
			//log.Println(row)
			t++
		}
	}
	f.Sync()
}
