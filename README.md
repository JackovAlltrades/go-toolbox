# Go Toolbox

A collection of useful Go utilities for web development.

## Features

- File upload handling
- Random string generation
- More utilities coming soon...
- Create a directory, including all parent directories if they don't exist

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
```
## License
This project is licensed under the MIT License - see the LICENSE.MD file for details.