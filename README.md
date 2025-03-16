# Go Toolbox

A collection of utility tools for Go web applications.

## Features

- File upload handling with validation
- Static file downloads
- Chunked file uploads for large files
- Concurrent upload support
- File type verification

## Usage

### File Downloads

```go
// Example of downloading a static file
func downloadHandler(w http.ResponseWriter, r *http.Request) {
    tools := &toolbox.Tools{}
    tools.DownloadStaticFile(w, r, "./files", "document.pdf", "my-document.pdf")
}
```

## Installation

```bash
go get github.com/JackovAlltrades/toolbox
```
## Usage 
```
import "github.com/JackovAlltrades/toolbox"

// Create a new toolbox
t := toolbox.Tools{
    MaxFileSize: 2048 * 2048 * 2048, // 2GB
    AllowedFileTypes: []string{"image/jpeg", "image/png", "image/gif"},
}

// Generate a random string
randomString := t.RandomString(10)

// Upload files
files, err := t.UploadFiles(request, "./uploads")

// Upload a single file
file, err := t.UploadOneFile(request, "./uploads")

// Create a URL-friendly slug
slug, err := t.Slugify("Hello World!")
// Returns: "hello-world"
```
## License
This project is licensed under the MIT License - see the LICENSE.MD file for details.