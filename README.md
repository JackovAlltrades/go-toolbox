# Go Toolbox

A collection of utility tools for Go web applications.

## Table of Contents
- [Features](#features)
- [Installation](#installation)
- [Usage](#usage)
  - [File Uploads](#file-uploads)
  - [Chunked Uploads](#chunked-uploads)
  - [JSON Handling](#json-handling)
  - [String Utilities](#string-utilities)
- [Configuration](#configuration)
- [Benchmarks](#benchmarks)
- [Examples](#examples)
- [Contributing](#contributing)
- [License](#license)

## Features

- File upload handling with comprehensive validation
- Static file downloads with content disposition
- Chunked file uploads for handling large files
- Concurrent upload support with batch processing
- File type verification and MIME type detection
- URL-friendly slug generation with transliteration
- Random string generation for secure filenames
- JSON response helpers with error handling
- Progress tracking for uploads

## Installation

```bash
go get github.com/JackovAlltrades/go-toolbox
```

## Usage

### File Uploads

Basic file upload handling:

```go
// Create a new Tools instance
tools := toolbox.Tools{
    MaxFileSize:       10 * 1024 * 1024, // 10MB
    AllowedFileTypes:  []string{"image/jpeg", "image/png", "application/pdf"},
    MaxUploadCount:    5,
    UploadPath:        "./uploads",
}

// Handle a single file upload
file, err := tools.UploadOneFile(r, "", true)
if err != nil {
    // Handle error
}

// Handle multiple file uploads
files, err := tools.UploadFiles(r, "", true)
if err != nil {
    // Handle error
}
```

### Chunked Uploads

For large file uploads using chunks:

```go
// Configure for chunked uploads
tools := toolbox.Tools{
    ChunkSize:       1024 * 1024, // 1MB chunks
    ChunksDirectory: "./temp_chunks",
    UploadPath:      "./uploads",
}

// Upload a chunk
err := tools.UploadChunk(uploadID, fileName, chunkNumber, totalChunks, chunkData)
if err != nil {
    // Handle error
}

// Complete the upload when all chunks are received
file, err := tools.CompleteChunkedUpload(uploadID, originalFileName)
if err != nil {
    // Handle error
}

// Check upload progress
progress, err := tools.GetUploadProgress(uploadID)
if err != nil {
    // Handle error
}
```

### JSON Handling

Working with JSON requests and responses:

```go
// Read JSON from request
var data YourStruct
err := tools.ReadJSON(w, r, &data)
if err != nil {
    tools.ErrorJSON(w, err, http.StatusBadRequest)
    return
}

// Write JSON response
response := struct {
    ID      int    `json:"id"`
    Message string `json:"message"`
}{
    ID:      123,
    Message: "Success!",
}
err = tools.WriteJSON(w, http.StatusOK, response)
if err != nil {
    // Handle error
}
```

### String Utilities

Generating slugs and random strings:

```go
// Generate a URL-friendly slug
slug, err := tools.Slugify("Hello World! Special Characters: @#$%")
// Output: "hello-world-special-characters"

// Generate a random string
randomStr := tools.RandomString(20)
// Output: "a1b2c3d4e5f6g7h8i9j0"
```

## Configuration

The `Tools` struct supports various configuration options:

```go
type Tools struct {
    // File size limits
    MaxFileSize            int
    TypeSpecificSizeLimits map[string]int
    DefaultSizeLimits      map[string]int
    MaxBatchSize           int64
    
    // File type validation
    AllowedFileTypes       []string
    AllowUnknownTypes      bool
    ValidationCallback     func(file *UploadedFile) error
    
    // Upload configuration
    MaxUploadCount         int
    UploadPath             string
    TempFilePath           string
    
    // Chunked upload configuration
    ChunkSize              int64
    ChunksDirectory        string
    
    // JSON handling
    MaxJSONSize            int
    AllowUnknownFields     bool
}
```

## Benchmarks

The package includes benchmarks for performance testing different upload scenarios:

- Single file uploads of various sizes
- Chunked uploads with different chunk sizes
- Concurrent uploads with varying concurrency levels

Run the benchmarks:

```bash
go test -bench=. ./toolbox/benchmarks
```

## Examples

Check the `examples` directory for complete usage examples:

- Basic file upload server
- Chunked upload implementation
- Advanced validation examples
- Concurrent upload handling

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the LICENSE file for details.