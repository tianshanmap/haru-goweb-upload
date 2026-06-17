package web

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"encoding/json"
	"github.com/tianshanmap/haru-goweb-upload/utils"
)

const MaxUploadSize = 10 << 20 // 10 MB in bytes
type UnzipResponse struct {
	Status string `json:"status"`
	TargetPath  string `json:"targetPath"`
	Files []string `json:"files"`
	Name string `json:"name"`
}
func UploadFileHandler(w http.ResponseWriter, r *http.Request) {
	println("uploadFileHandler-started,r.Method=" + r.Method)
	// 1. Enforce POST request method
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	println("uploadFileHandler-started.1")

	// 2. Protect server memory by restricting maximum request size
	r.Body = http.MaxBytesReader(w, r.Body, MaxUploadSize)
	if err := r.ParseMultipartForm(MaxUploadSize); err != nil {
		http.Error(w, "File too large or invalid multipart form", http.StatusBadRequest)
		return
	}
	println("uploadFileHandler-started.2")

	// 3. Retrieve file from multipart form data via its key name (e.g., "myFile")
	fileChunk, handler, err := r.FormFile("fileChunk")
	if err != nil {
		http.Error(w, "Error retrieving the file from form-data", http.StatusBadRequest)
		return
	}
	defer fileChunk.Close()

	target := r.FormValue("target")
	filename := r.FormValue("filename")
	fmt.Printf("Uploaded File: %s\n", filename)
	fmt.Printf("File Size: %d bytes\n", handler.Size)

	dstPath := filepath.Join(target, filepath.Base(filename))
	outfile, err := os.OpenFile(dstPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer outfile.Close()
	fileBytes, err := io.ReadAll(fileChunk)
	if err != nil {
		http.Error(w, "Failed to read file", http.StatusInternalServerError)
		return
	}
	outfile.Write(fileBytes)

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Successfully uploaded file: %s\n", handler.Filename)
}

func UnzipHandler(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()

	// 1. Get a single/first parameter value
	// Example URL: /search?search_term=golang
	filename := queryParams.Get("filename")
	target := queryParams.Get("target")
	dstPath := filepath.Join(target, filepath.Base(filename))
	println("UnzipHandler::dstPath=" + dstPath)
	err,topPath := utils.ExtractTarGz(dstPath, target)
	if err != nil {
		http.Error(w, "Failed to read file", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	files := utils.GetFiles(topPath,".jpeg")
	data := UnzipResponse {
		Status: "success",
		TargetPath:  topPath,
		Files: files,
		Name: files[0],
	}
	// Encode and stream data directly to the client
	json.NewEncoder(w).Encode(data)
}
