package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github\.com/JackovAlltrades/go-toolbox"
)

func main() {
	// Create a simple HTTP server to test file uploads
	http.HandleFunc("/upload", handleUpload)
	fmt.Println("Starting test server on http://localhost:8090/upload")
	log.Fatal(http.ListenAndServe(":8090", nil))
}

func handleUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		// Show a simple upload form for GET requests
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, `
			<html>
				<head><title>File Upload Test</title></head>
				<body>
					<h1>Upload Test</h1>
					<form method="post" enctype="multipart/form-data">
						<input type="file" name="file" multiple>
						<button type="submit">Upload</button>
					</form>
				</body>
			</html>
		`)
		return
	}

	// Create uploads directory if it doesn't exist
	uploadDir := "./uploads"
	os.MkdirAll(uploadDir, 0755)

	// Initialize the toolbox with different configurations
	tools := toolbox.Tools{
		MaxFileSize:       10 * 1024 * 1024, // 10MB
		AllowedFileTypes:  []string{"image/jpeg", "image/png"},
		AllowUnknownTypes: false, // Only allow specified types
		MaxUploadCount:    5,
	}

	// Process the upload
	files, err := tools.UploadFiles(r, uploadDir, true)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Error: %s", err.Error())
		return
	}

	// Display results
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, "<h1>Upload Successful</h1>")
	fmt.Fprintf(w, "<p>Uploaded %d files:</p><ul>", len(files))

	for _, file := range files {
		fmt.Fprintf(w, "<li>Original: %s, New: %s, Type: %s, Size: %d bytes</li>",
			file.OriginalFileName, file.NewFileName, file.FileType, file.FileSize)
	}

	fmt.Fprintf(w, "</ul>")
}

