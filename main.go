package main

import (
	"fmt"
	"net/http"
    "github.com/alecthomas/log4go"
	"github.com/tianshanmap/haru-goweb-upload/utils"
	"github.com/tianshanmap/haru-goweb-upload/web"
)

func main() {
	log4go.LoadConfiguration("./conf/log4go.xml")
	defer log4go.Close()
	utils.SetLogger(log4go.Global)

	utils.Yaml_init("./conf/application.yaml")

	// Mount the router pattern to our handler logic
	http.HandleFunc("/goweb/filesystem/upload_chunk", web.UploadFileHandler)
	http.HandleFunc("/goweb/filesystem/unzip", web.UnzipHandler)
	http.HandleFunc("/goweb/filesystem/download", web.Download)
	http.HandleFunc("/goweb/filesystem/download-chunk", web.ChunkDownloadHandler)

	utils.Log.Info("Server starting on " + utils.YamlConfig.Server.Address + "...")
	if err := http.ListenAndServe(utils.YamlConfig.Server.Address, nil); err != nil {
		fmt.Printf("Server failed: %v\n", err)
	}
}
