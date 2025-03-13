package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/JackovAlltrades/go-toolbox"
)

func main() {
	mux := routes()

	log.Println("Starting server on port 8090")

	err := http.ListenAndServe(":8090", mux)
	if err != nil {
		log.Fatal(err)
	}
}

func routes() http.Handler {
	mux := http.NewServeMux()

	mux.Handle("/", http.StripPrefix("/", http.FileServer(http.Dir("."))))
	mux.HandleFunc("/upload", uploadFiles)
	mux.HandleFunc("/upload-one", uploadOneFile)

	return mux
}

func uploadFiles(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Create uploads directory with absolute path
	uploadsDir := filepath.Join("H:", "Projects", "toolbox-project", "app-upload", "uploads")
	err := os.MkdirAll(uploadsDir, os.ModePerm)
	if err != nil {
		http.Error(w, "Unable to create upload directory", http.StatusInternalServerError)
		return
	}

	t := toolbox.Tools{
		MaxFileSize:      1024 * 1024 * 1024,
		AllowedFileTypes: []string{"image/jpeg", "image/png", "image/gif"},
	}

	files, err := t.UploadFiles(r, uploadsDir)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	out := ""
	for _, item := range files {
		out += fmt.Sprintf("Uploaded %s to the uploads folder, renamed to %s\n", item.OriginalFileName, item.NewFileName)
	}

	_, _ = w.Write([]byte(out))
}

func uploadOneFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Create uploads directory with absolute path
	uploadsDir := filepath.Join("H:", "Projects", "toolbox-project", "app-upload", "uploads")
	err := os.MkdirAll(uploadsDir, os.ModePerm)
	if err != nil {
		http.Error(w, "Unable to create upload directory", http.StatusInternalServerError)
		return
	}

	t := toolbox.Tools{
		MaxFileSize:      2048 * 2048 * 2048,
		AllowedFileTypes: []string{"image/jpeg", "image/png", "image/gif"},
	}

	f, err := t.UploadOneFile(r, uploadsDir)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	_, _ = w.Write([]byte(fmt.Sprintf("Uploaded 1 file, %s, to the uploads folder", f.OriginalFileName)))
}
