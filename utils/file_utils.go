package utils

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"io/fs"
)

// ExtractTarGz securely extracts a .tar.gz file, mitigating Zip Slip vulnerabilities.
func ExtractTarGz(archivePath, targetDir string) (error,string) {
	println("ExtractTarGz::targetDir=" + targetDir)
	file, err := os.Open(archivePath)
	if err != nil {
		return err,""
	}
	defer file.Close()

	gzipReader, err := gzip.NewReader(file)
	if err != nil {
		return err,""
	}
	defer gzipReader.Close()

	tarReader := tar.NewReader(gzipReader)
	topPath := ""
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err,""
		}

		// Securely determine path
		targetPath := filepath.Join(targetDir, filepath.Clean(header.Name))
		if topPath == "" {
			topPath = targetPath		
		}
		println("ExtractTarGz::targetPath=" + targetPath)
		
		if !strings.HasPrefix(targetPath, filepath.Clean(targetDir)+string(filepath.Separator)) {
			return fmt.Errorf("illegal path: %s", header.Name),""
		}

		switch header.Typeflag {
		case tar.TypeDir:
			os.MkdirAll(targetPath, header.FileInfo().Mode())
		case tar.TypeReg:
			os.MkdirAll(filepath.Dir(targetPath), 0755) // Ensure parent exists
			outFile, _ := os.OpenFile(targetPath, os.O_CREATE|os.O_RDWR|os.O_TRUNC, header.FileInfo().Mode())
			io.Copy(outFile, tarReader)
			outFile.Close()
		}
	}
	println("ExtractTarGz::topPath=" + topPath)
	return nil,topPath
}

func GetFiles(root string,ext string)([] string) {

	// Slice to store the file paths
	var files []string

	// WalkDir traverses the directory tree recursively
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err // Return error to stop walking, or return nil to skip
		}

		// Check if the current path is a file (not a directory)
		if !d.IsDir() && strings.HasSuffix(d.Name(),ext) && !strings.HasPrefix(d.Name(), ".") {
			files = append(files, path)
		}

		return nil
	})

	if err != nil {
		fmt.Printf("Error walking the path: %v\n", err)
		return nil
	}

	// Print all collected files
	for _, file := range files {
		fmt.Println(file)
	}
	return files
}