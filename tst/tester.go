package main

import (
	"fmt"
	"github.com/XANi/go-filereopen"
	"time"
)

func main() {
	fd, err := filereopen.OpenFileForAppend("test.log", 0644)
	if err != nil {
		fmt.Printf("can't open: %s\n", err)
	}
	for {
		time.Sleep(time.Second)
		n, err := fmt.Fprintf(fd, "date: %s\n", time.Now().String())
		if err != nil {
			fmt.Printf("err: %s[%d]\n", err, n)
		}
	}
}
