package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

var (
	Version  string
	reqCount uint
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "ok\n")
	})

	http.HandleFunc("/version", func(w http.ResponseWriter, r *http.Request) {
		reqCount++
		fmt.Printf("GET /version - %s, count - %d\n", Version, reqCount)
		io.WriteString(w, fmt.Sprintf("version: %s\n", Version))
	})

	fmt.Println("start server")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Printf("error serve http: %s\n", err)
		os.Exit(1)
	}
}
