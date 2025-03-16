package main

import (
    "fmt"
    "net/http"
    "github\.com/JackovAlltrades/go-toolbox"
)

func main() {
    http.HandleFunc("/upload", uploadHandler)
    fmt.Println("Starting server on :8080")
    http.ListenAndServe(":8080", nil)
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
    tools := &toolbox.Tools{
        MaxFileSize: 10 * 1024 * 1024, // 10MB
        AllowedFileTypes: []string{"image/jpeg", "image/png"},
        AllowUnknownTypes: true,
    }
    
    files, err := tools.UploadFiles(r, "./uploads", true)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    
    fmt.Fprintf(w, "Uploaded %d files successfully", len(files))
}

