package main

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"time"

	"golang.org/x/exp/mmap"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("too few arguments")
		os.Exit(1)
	}

	sleep, err := strconv.Atoi(os.Args[2])
	if err != nil {
		fmt.Printf("error parse sleep time: %s", err)
	}

	for {
		data, err := mmap.Open(os.Args[1])
		if err != nil {
			fmt.Printf("error read file: %s\n", err.Error())
			time.Sleep(time.Second * time.Duration(sleep))
			continue
		}
		for i := 0; i < data.Len(); i++ {
			fmt.Fprint(io.Discard, data.At(i))
		}
		fmt.Printf("size: %d\n", data.Len())
		time.Sleep(time.Second * time.Duration(sleep))
		data.Close()
	}
}
