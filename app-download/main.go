package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/JackovAlltrades/go-toolbox"
)

func main() {
	// Serve static files
	http.Handle("/", http.FileServer(http.Dir(".")))

	// Handle download requests
	http.HandleFunc("/download", downloadHandler)

	// Handle toolkit download
	http.HandleFunc("/download-file", downloadFile)

	// Start the server
	fmt.Println("Server started at http://localhost:8090")
	http.ListenAndServe(":8090", nil)
}

func downloadHandler(w http.ResponseWriter, r *http.Request) {
	// Path to the file you want to download
	filePath := filepath.Join(".", "files", "sample.pdf")

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	// Set headers for file download
	w.Header().Set("Content-Disposition", "attachment; filename=sample.pdf")
	w.Header().Set("Content-Type", "application/pdf")

	// Serve the file
	http.ServeFile(w, r, filePath)
}

func downloadFile(w http.ResponseWriter, r *http.Request) {
	t := toolbox.Tools{}
	t.DownloadStaticFile(w, r, "./files", "pic.png", "pic2.png")
}
