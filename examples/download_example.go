package main

import (
    "fmt"
    "net/http"
    "github.com/JackovAlltrades/go-toolbox/toolbox"
)

func main() {
    http.HandleFunc("/download", downloadHandler)
    fmt.Println("Starting server on :8080")
    http.ListenAndServe(":8080", nil)
}

func downloadHandler(w http.ResponseWriter, r *http.Request) {
    tools := &toolbox.Tools{}
    tools.DownloadStaticFile(w, r, "./files", "document.pdf", "my-document.pdf")
}
