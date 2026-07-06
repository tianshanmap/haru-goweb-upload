package main

import (
	"fmt"
	"net/http"

	"github.com/tianshanmap/haru-goweb-upload/web"
)

func main() {
	// Mount the router pattern to our handler logic
	http.HandleFunc("/goweb/filesystem/upload_chunk", web.UploadFileHandler)
	http.HandleFunc("/goweb/filesystem/unzip", web.UnzipHandler)
	http.HandleFunc("/goweb/filesystem/download", web.Download)
	http.HandleFunc("/goweb/filesystem/download-chunk", web.ChunkDownloadHandler)

	fmt.Println("Server starting on port :9082...")
	if err := http.ListenAndServe(":9082", nil); err != nil {
		fmt.Printf("Server failed: %v\n", err)
	}
}
