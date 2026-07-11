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
	"archive/zip"

)

func IsDirectory(path string) (bool, error) {
	// Get file details; returns an error if the path does not exist
	fileInfo, err := os.Stat(path)
	if err != nil {
		return false, err 
	}

	// Return true if the path points to a directory
	return fileInfo.IsDir(), nil
}

// ExtractTarGz securely extracts a .tar.gz file, mitigating Zip Slip vulnerabilities.
func ExtractTarGz(archivePath, targetDir string) (error,string) {
	println("ExtractTarGz::targetDir=" + targetDir)
	file, err := os.Open(archivePath)
	if err != nil {
		println("ExtractTarGz::failed to open=" + archivePath)
		return err,""
	}
	defer file.Close()

	gzipReader, err := gzip.NewReader(file)
	if err != nil {
		println("ExtractTarGz::failed to open gzipReader")
		return err,""
	}
	defer gzipReader.Close()

	println("ExtractTarGz::open a tar")
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
		println("ExtractTarGz::targetPath=" + targetPath)
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

func ZipFolder(folder string) string {
	fmt.Println("ZipFolder-start-" + folder);	
	foldername := filepath.Base(folder)
	parentPath := filepath.Dir(folder)
	filename := filepath.Join(parentPath,foldername + ".zip")
    file, err := os.Create(filename)
    if err != nil {
        panic(err)
    }
    defer file.Close()

    zipFile := zip.NewWriter(file)
    defer zipFile.Close()

    walker := func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }
        if info.IsDir() {
            return nil
        }
        file, err := os.Open(path)
        if err != nil {
            return err
        }
        defer file.Close()

        // Ensure that `path` is not absolute; it should not start with "/".
        // This snippet happens to work because I don't use 
        // absolute paths, but ensure your real-world code 
        // transforms path into a zip-root relative path.
		subpath := path[len(folder) + 1:]
        f, err := zipFile.Create(subpath)
        if err != nil {
            return err
        }

        _, err = io.Copy(f, file)
        if err != nil {
            return err
        }

        return nil
    }
    err = filepath.Walk(folder, walker)
    if err != nil {
        panic(err)
    }
	fmt.Println("ZipFolder-end-" + folder + ",filename=" + filename);	
	return filename 
}

func ExtractZip(archivePath, dst string) {
	Log.Info("ExtractZip-started,archivePath=%s,dst=%s",archivePath,dst)
    archive, err := zip.OpenReader(archivePath)
    if err != nil {
        panic(err)
    }
    defer archive.Close()

    for _, f := range archive.File {
        filePath := filepath.Join(dst, f.Name)
        fmt.Println("unzipping file ", filePath)

        if !strings.HasPrefix(filePath, filepath.Clean(dst)+string(os.PathSeparator)) {
            fmt.Println("invalid file path")
            panic("invalid file path")
        }
        if f.FileInfo().IsDir() {
            fmt.Println("creating directory...")
            os.MkdirAll(filePath, os.ModePerm)
            continue
        }

        if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
            panic(err)
        }

        dstFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
        if err != nil {
            panic(err)
        }

        fileInArchive, err := f.Open()
        if err != nil {
            panic(err)
        }

        if _, err := io.Copy(dstFile, fileInArchive); err != nil {
            panic(err)
        }

        dstFile.Close()
        fileInArchive.Close()
    }
}