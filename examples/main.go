package main

import (
	qbench "github.com/dzlab/qbench/bench"
	"log"
)

var ()

func main() {
	p := qbench.NewParser()
	object, err := p.Parse("main.yml")
	if err != nil {
		panic(err)
	}
	// generate some data
	for i := 0; i < 10; i++ {
		log.Println(i, ">", object.GetKV("=", "&"))
		log.Println(i, ">", object.GetJSON())

	}
}
