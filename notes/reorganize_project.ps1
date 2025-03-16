# Script to reorganize the toolbox project structure

# Define paths
$projectRoot = "H:\Projects\toolbox-project"
$tempDir = "$projectRoot\temp_backup"
$toolboxDir = "$projectRoot\toolbox"

# Create a backup of everything first
Write-Host "Creating backup of current project..." -ForegroundColor Cyan
if (-not (Test-Path $tempDir)) {
    New-Item -ItemType Directory -Path $tempDir | Out-Null
}

# Ensure toolbox directory exists
if (-not (Test-Path $toolboxDir)) {
    New-Item -ItemType Directory -Path $toolboxDir | Out-Null
}

# Create standard Go project directories
$standardDirs = @(
    "$projectRoot\cmd\api",
    "$projectRoot\docs",
    "$projectRoot\examples",
    "$projectRoot\testdata"
)

foreach ($dir in $standardDirs) {
    if (-not (Test-Path $dir)) {
        New-Item -ItemType Directory -Path $dir -Force | Out-Null
        Write-Host "Created standard directory: $dir" -ForegroundColor Green
    }
}

# List of component directories to consolidate
$componentDirs = @(
    "app",
    "app-dir",
    "app-download",
    "app-json",
    "app-slug",
    "app-upload",
    "temp",
    "templates",
    "uploads"
)

# Process each component directory
foreach ($dir in $componentDirs) {
    $sourcePath = "$projectRoot\$dir"
    if (Test-Path $sourcePath) {
        Write-Host "Processing $dir..." -ForegroundColor Cyan
        
        # Backup the directory
        Copy-Item -Path $sourcePath -Destination "$tempDir\$dir" -Recurse -Force
        
        # Check for Go files and move them to appropriate locations
        $goFiles = Get-ChildItem -Path $sourcePath -Filter "*.go" -Recurse
        foreach ($file in $goFiles) {
            $fileName = $file.Name
            $relativePath = $file.FullName.Replace($sourcePath, "").TrimStart("\")
            
            # Determine destination based on file content and name
            $fileContent = Get-Content -Path $file.FullName -Raw
            
            if ($fileName -like "*_test.go") {
                # Test files go to toolbox/tests
                $destDir = "$toolboxDir\tests"
                if (-not (Test-Path $destDir)) {
                    New-Item -ItemType Directory -Path $destDir | Out-Null
                }
                
                # Create a new file with appropriate imports
                $newContent = $fileContent -replace 'package \w+', 'package toolbox_test'
                $newContent = $newContent -replace 'import \(', "import (
    `"github.com/JackovAlltrades/go-toolbox/toolbox`""
                
                Set-Content -Path "$destDir\$fileName" -Value $newContent
                Write-Host "  Moved test file: $fileName to toolbox/tests" -ForegroundColor Green
            }
            elseif ($fileContent -match "func main\(\)") {
                # Main application files go to cmd/api
                $destDir = "$projectRoot\cmd\api"
                if (-not (Test-Path $destDir)) {
                    New-Item -ItemType Directory -Path $destDir -Force | Out-Null
                }
                
                Copy-Item -Path $file.FullName -Destination "$destDir\$fileName" -Force
                Write-Host "  Moved main file: $fileName to cmd/api" -ForegroundColor Green
            }
            else {
                # Library files go to toolbox
                Copy-Item -Path $file.FullName -Destination "$toolboxDir\$fileName" -Force
                Write-Host "  Moved library file: $fileName to toolbox" -ForegroundColor Green
            }
        }
        
        # Handle special directories
        if ($dir -eq "templates") {
            # Move templates to examples directory
            $destDir = "$projectRoot\examples\templates"
            if (-not (Test-Path $destDir)) {
                New-Item -ItemType Directory -Path $destDir -Force | Out-Null
            }
            
            # Copy template files
            Get-ChildItem -Path $sourcePath -File -Recurse | ForEach-Object {
                $relativePath = $_.FullName.Replace($sourcePath, "").TrimStart("\")
                $destPath = Join-Path -Path $destDir -ChildPath $relativePath
                $destParent = Split-Path -Path $destPath -Parent
                
                if (-not (Test-Path $destParent)) {
                    New-Item -ItemType Directory -Path $destParent -Force | Out-Null
                }
                
                Copy-Item -Path $_.FullName -Destination $destPath -Force
                Write-Host "  Moved template file: $relativePath to examples/templates" -ForegroundColor Green
            }
        }
        elseif ($dir -eq "uploads") {
            # Create testdata directory for upload examples
            $destDir = "$projectRoot\testdata\uploads"
            if (-not (Test-Path $destDir)) {
                New-Item -ItemType Directory -Path $destDir -Force | Out-Null
            }
            
            # Copy any existing files
            Get-ChildItem -Path $sourcePath -File -Recurse | ForEach-Object {
                $relativePath = $_.FullName.Replace($sourcePath, "").TrimStart("\")
                $destPath = Join-Path -Path $destDir -ChildPath $relativePath
                $destParent = Split-Path -Path $destPath -Parent
                
                if (-not (Test-Path $destParent)) {
                    New-Item -ItemType Directory -Path $destParent -Force | Out-Null
                }
                
                Copy-Item -Path $_.FullName -Destination $destPath -Force
                Write-Host "  Moved upload file: $relativePath to testdata/uploads" -ForegroundColor Green
            }
        }
        
        # Check for other important files (README, config, etc.)
        $otherFiles = Get-ChildItem -Path $sourcePath -Exclude "*.go" -File
        foreach ($file in $otherFiles) {
            if ($file.Name -like "README*" -or $file.Name -like "*.md") {
                # Consolidate documentation
                $docsDir = "$projectRoot\docs"
                if (-not (Test-Path $docsDir)) {
                    New-Item -ItemType Directory -Path $docsDir -Force | Out-Null
                }
                
                Copy-Item -Path $file.FullName -Destination "$docsDir\$($dir)_$($file.Name)" -Force
                Write-Host "  Moved documentation: $($file.Name) to docs" -ForegroundColor Green
            }
            elseif ($file.Name -like "*.json" -or $file.Name -like "*.yaml" -or $file.Name -like "*.yml") {
                # Configuration files
                $configDir = "$projectRoot\examples\config"
                if (-not (Test-Path $configDir)) {
                    New-Item -ItemType Directory -Path $configDir -Force | Out-Null
                }
                Copy-Item -Path $file.FullName -Destination "$configDir\$($dir)_$($file.Name)" -Force
                Write-Host "  Moved config file: $($file.Name) to examples/config" -ForegroundColor Green
            }
        }
    }
}

# Create example files
$exampleDir = "$projectRoot\examples"
$exampleFiles = @{
    "upload_example.go" = @"
package main

import (
    "fmt"
    "net/http"
    "github.com/JackovAlltrades/go-toolbox/toolbox"
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
"@

    "download_example.go" = @"
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
"@

    "slugify_example.go" = @"
package main

import (
    "fmt"
    "github.com/JackovAlltrades/go-toolbox/toolbox"
)

func main() {
    tools := &toolbox.Tools{}
    
    examples := []string{
        "Hello World!",
        "This is a test",
        "Special Characters: !@#$%^&*()",
        "Multiple   Spaces",
        "Dashes-and_underscores",
    }
    
    for _, example := range examples {
        slug, err := tools.Slugify(example)
        if err != nil {
            fmt.Printf("Error slugifying '%s': %v\n", example, err)
            continue
        }
        
        fmt.Printf("Original: '%s' -> Slug: '%s'\n", example, slug)
    }
}
"@

    "json_response_example.go" = @"
package main

import (
    "fmt"
    "net/http"
    "github.com/JackovAlltrades/go-toolbox/toolbox"
)

type Person struct {
    Name    string `json:"name"`
    Age     int    `json:"age"`
    Email   string `json:"email"`
}

func main() {
    http.HandleFunc("/json", jsonHandler)
    fmt.Println("Starting server on :8080")
    http.ListenAndServe(":8080", nil)
}

func jsonHandler(w http.ResponseWriter, r *http.Request) {
    tools := &toolbox.Tools{}
    
    person := Person{
        Name:  "John Doe",
        Age:   30,
        Email: "john@example.com",
    }
    
    err := tools.WriteJSON(w, http.StatusOK, person, nil)
    if err != nil {
        http.Error(w, "Error writing JSON", http.StatusInternalServerError)
    }
}
"@
}

foreach ($fileName in $exampleFiles.Keys) {
    $filePath = "$exampleDir\$fileName"
    Set-Content -Path $filePath -Value $exampleFiles[$fileName]
    Write-Host "Created example file: $fileName" -ForegroundColor Green
}

# Update go.work file
$goWorkContent = @"
go 1.24.1

use (
    ./cmd/api
    ./toolbox
)
"@

Set-Content -Path "$projectRoot\go.work" -Value $goWorkContent
Write-Host "Updated go.work file" -ForegroundColor Green

# Create a new go.mod for toolbox if it doesn't exist
$toolboxModPath = "$toolboxDir\go.mod"
if (-not (Test-Path $toolboxModPath)) {
    $goModContent = @"
module github.com/JackovAlltrades/go-toolbox/toolbox

go 1.24.1
"@
    Set-Content -Path $toolboxModPath -Value $goModContent
    Write-Host "Created toolbox go.mod file" -ForegroundColor Green
}

# Create a new go.mod for cmd/api if it doesn't exist
$cmdApiModPath = "$projectRoot\cmd\api\go.mod"
if (-not (Test-Path $cmdApiModPath)) {
    $goModContent = @"
module github.com/JackovAlltrades/go-toolbox/cmd/api

go 1.24.1

require github.com/JackovAlltrades/go-toolbox/toolbox v0.0.0-00010101000000-000000000000

replace github.com/JackovAlltrades/go-toolbox/toolbox => ../../toolbox
"@
    Set-Content -Path $cmdApiModPath -Value $goModContent
    Write-Host "Created cmd/api go.mod file" -ForegroundColor Green
}

# Create a new main.go in cmd/api if it doesn't exist
$mainApiPath = "$projectRoot\cmd\api\main.go"
if (-not (Test-Path $mainApiPath)) {
    $mainApiContent = @"
package main

import (
    "fmt"
    "github.com/JackovAlltrades/go-toolbox/toolbox"
)

func main() {
    t := &toolbox.Tools{}
    
    // Example of random string generation
    randomString := t.RandomString(10)
    fmt.Println("Generated random string:", randomString)
    
    // Example of slug generation
    slug, err := t.Slugify("Hello World!")
    if err != nil {
        fmt.Println("Error generating slug:", err)
    } else {
        fmt.Println("Generated slug:", slug)
    }
    
    fmt.Println("Toolbox API is ready to use!")
}
"@
    Set-Content -Path $mainApiPath -Value $mainApiContent
    Write-Host "Created example main.go in cmd/api" -ForegroundColor Green
}

# Update README.md from external file
$readmePath = "$projectRoot\README.md"
$readmeContentPath = "$projectRoot\readme_content.txt"
if (Test-Path $readmeContentPath) {
    Copy-Item -Path $readmeContentPath -Destination $readmePath -Force
    Write-Host "Updated README.md from readme_content.txt" -ForegroundColor Green
} else {
    Write-Host "Warning: readme_content.txt not found. README.md not updated." -ForegroundColor Yellow
}

# Create a LICENSE.md file if it doesn't exist
$licensePath = "$projectRoot\LICENSE.md"
if (-not (Test-Path $licensePath)) {
    $licenseContent = @"
MIT License

Copyright (c) $(Get-Date -Format yyyy) JackovAlltrades

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
"@
    Set-Content -Path $licensePath -Value $licenseContent
    Write-Host "Created LICENSE.md" -ForegroundColor Green
}

Write-Host "Project reorganization complete!" -ForegroundColor Cyan
Write-Host "New structure:" -ForegroundColor Cyan
Write-Host "  - /cmd/api: Main application code" -ForegroundColor White
Write-Host "  - /toolbox: Core library code" -ForegroundColor White
Write-Host "  - /toolbox/tests: Test files" -ForegroundColor White
Write-Host "  - /examples: Example usage code" -ForegroundColor White
Write-Host "  - /examples/config: Configuration examples" -ForegroundColor White
Write-Host "  - /examples/templates: Template examples" -ForegroundColor White
Write-Host "  - /docs: Documentation" -ForegroundColor White
Write-Host "  - /testdata: Test data files" -ForegroundColor White
Write-Host "  - /temp_backup: Backup of original files" -ForegroundColor White
Write-Host "You can run 'go work sync' to update dependencies" -ForegroundColor Yellow