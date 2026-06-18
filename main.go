package main

import (
	"fmt"
	"net/http"
)

import "github.com/tianshanmap/haru-goweb-upload/web"

func main() {
	// Mount the router pattern to our handler logic
	http.HandleFunc("/goweb/filesystem/upload_chunk", web.UploadFileHandler)
	http.HandleFunc("/goweb/filesystem/unzip", web.UnzipHandler)

	fmt.Println("Server starting on port :8081...")
	if err := http.ListenAndServe(":8081", nil); err != nil {
		fmt.Printf("Server failed: %v\n", err)
	}
}
