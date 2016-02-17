package main

import (
	"bytes"
	"math/rand"
	"text/template"
	"time"
)

type RecordGenerator interface {
	Generate(input interface{}) []byte
}

type BytesGenerator struct {
	generator *rand.Rand
}

// Returns a random message generated from the chars byte slice.
// Message length of m bytes as defined by msgSize.
func (this *BytesGenerator) Generate(size int) []byte {
	m := make([]byte, size)
	for i := range m {
		m[i] = chars[this.generator.Intn(len(chars))]
	}
	return m
}

func NewJsonGenerator(name, templ string) (*JsonGenerator, error) {
	t := template.New(name)
	t, err := t.Parse(templ)
	if err != nil {
		return nil, err
		//buff := bytes.NewBufferString("")
		//t.Execute(buff, person)
	}
	return &JsonGenerator{templ: t}, nil
}

type JsonGenerator struct {
	templ *template.Template
}

func (this *JsonGenerator) Generate(data interface{}) []byte {
	buff := bytes.NewBufferString("")
	this.templ.Execute(buff, data)
	return buff.Bytes()
}

func RandomString(strlen int) string {
	rand.Seed(time.Now().UTC().UnixNano())
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, strlen)
	for i := 0; i < strlen; i++ {
		result[i] = chars[rand.Intn(len(chars))]
	}
	return string(result)
}
