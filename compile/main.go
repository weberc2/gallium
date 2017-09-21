package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

func main() {
	data, err := ioutil.ReadFile(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	r := ParseFile(Input(string(data)))
	if r.IsErr() {
		log.Fatal(r.Err)
	}
	fmt.Println(r.Node)
}
