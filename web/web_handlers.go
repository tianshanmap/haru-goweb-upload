package web

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
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

func removeExtension(filename string) string {
	return strings.TrimSuffix(filename, filepath.Ext(filename))
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
	w.Header().Set("Access-Control-Expose-Headers", "*")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Successfully uploaded file: %s\n", handler.Filename)
}

func UnzipHandler(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()

	// 1. Get a single/first parameter value
	// Example URL: /search?search_term=golang
	filename := queryParams.Get("filename")
	parentName := removeExtension(filepath.Base(filename)) // Returns: "myapp"

	target_from_web := queryParams.Get("target")
	target := filepath.Join(target_from_web, parentName);
	// dstPath := filepath.Join(target, parentName,filepath.Base(filename))
	// println("UnzipHandler::dstPath=" + dstPath)
	utils.Log.Info("UnzipHandler - filename=%s,target_from_web=%s",filename,target_from_web)	
	if strings.HasSuffix(filename, ".zip"){
		utils.ExtractZip(filename,target)
	} else {
		err,_ := utils.ExtractTarGz(filename, target)
		if err != nil {
			http.Error(w, "Failed to read file", http.StatusInternalServerError)
			return
		}
	}
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	// files := utils.GetFiles(target,".jpeg")
	data := UnzipResponse {
		Status: "success",
		TargetPath:  target_from_web,
		Files: []string{},
		Name: "",
	}
	// Encode and stream data directly to the client
	json.NewEncoder(w).Encode(data)
}

func Download(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()
	filename := queryParams.Get("name")
	isDir,_ := utils.IsDirectory(filename)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	filePath := filename
	if isDir {
		filePath = utils.ZipFolder(filename)	
		downloadName := filepath.Base(filePath)
		w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, downloadName))
		w.Header().Set("Content-Type", "application/zip")
	} else {
		downloadName := filepath.Base(filename)
		w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, downloadName))
		w.Header().Set("Content-Type", "application/octet-stream")
	}
	http.ServeFile(w, r, filePath)	
}

func ChunkDownloadHandler(w http.ResponseWriter, r *http.Request) {
    fmt.Printf("ChunkDownloadHandler-started\n")
	queryParams := r.URL.Query()
	filename := queryParams.Get("name")
	isDir,_ := utils.IsDirectory(filename)
	filePath := filename
	if isDir {
		filePath = utils.ZipFolder(filename)	
	}

	// Open the file
	file, err := os.Open(filePath)
	if err != nil {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}
	defer file.Close()

	// Get file information for size and modification time
	fileInfo, err := file.Stat()
	if err != nil {
		http.Error(w, "Could not get file info", http.StatusInternalServerError)
		return
	}

	// ServeContent handles HTTP Range requests and chunking automatically
	w.Header().Set("Access-Control-Allow-Origin", "*")
	//allow client to see all the headers
	w.Header().Set("Access-Control-Expose-Headers", "*")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filepath.Base(filePath)))
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("filename", filepath.Base(filePath))
	
	http.ServeContent(w, r, fileInfo.Name(), fileInfo.ModTime(), file)
}
