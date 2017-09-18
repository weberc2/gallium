package main

import (
	"fmt"
	"io/ioutil"
)

var fileStr = `package main

fn main() {
	println("hello, world!")
}`

func main() {
	file, _, err := parseFile(NewStringInput(fileStr))
	if err != nil {
		err.Error()
		fmt.Println(err)
	}
	if err := ioutil.WriteFile(
		"/tmp/test.go",
		[]byte(CompileFile(file)),
		0644,
	); err != nil {
		fmt.Println(err)
	}
}
