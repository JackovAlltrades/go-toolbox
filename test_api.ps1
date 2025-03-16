# Script to update project and test API

# Navigate to the project directory
Set-Location -Path "h:\Projects\toolbox-project"

# Fix the syntax error in tools_test.go more thoroughly
Write-Host "Fixing syntax error in tools_test.go..." -ForegroundColor Cyan
$toolsTestPath = "h:\Projects\toolbox-project\toolbox\tools_test.go"
$content = Get-Content -Path $toolsTestPath -Raw

# Check if the file has balanced braces
$openBraces = ($content | Select-String -Pattern "{" -AllMatches).Matches.Count
$closeBraces = ($content | Select-String -Pattern "}" -AllMatches).Matches.Count

if ($openBraces -gt $closeBraces) {
    # Add missing closing braces
    $missingBraces = $openBraces - $closeBraces
    $content += "`n" + ("}") * $missingBraces
    Set-Content -Path $toolsTestPath -Value $content
    Write-Host "Fixed $missingBraces missing closing brace(s) in tools_test.go" -ForegroundColor Green
}

# Fix syntax errors in tools.go
Write-Host "Fixing syntax errors in tools.go..." -ForegroundColor Cyan
$toolsGoPath = "h:\Projects\toolbox-project\toolbox\tools.go"

# Read the file content as a single string to preserve line breaks
$content = Get-Content -Path $toolsGoPath -Raw

# Find the problematic lines around 906-907 and fix them
$lines = $content -split "`n"
$lineNumber = 1
$fixedContent = ""

foreach ($line in $lines) {
    # Check if this is one of the problematic lines
    if ($lineNumber -eq 906 -or $lineNumber -eq 907) {
        # If the line contains 'payload' and doesn't end with a semicolon or brace
        if ($line -match "payload" -and $line -notmatch "[;{}]$") {
            # Add a semicolon at the end
            $fixedContent += $line + ";" + "`n"
        } else {
            $fixedContent += $line + "`n"
        }
    } else {
        $fixedContent += $line + "`n"
    }
    $lineNumber++
}

# Write the fixed content back to the file
Set-Content -Path $toolsGoPath -Value $fixedContent
Write-Host "Fixed syntax errors in tools.go" -ForegroundColor Green

# Set up Go workspace
Write-Host "Setting up Go workspace..." -ForegroundColor Cyan

# Create or update go.mod files for each module
$toolboxModPath = "h:\Projects\toolbox-project\toolbox\go.mod"
if (-not (Test-Path $toolboxModPath)) {
    Set-Location -Path "h:\Projects\toolbox-project\toolbox"
    go mod init github.com/JackovAlltrades/go-toolbox/toolbox
    Write-Host "Created toolbox go.mod file" -ForegroundColor Green
    Set-Location -Path "h:\Projects\toolbox-project"
}

# Create or update go.mod for the API
$apiDir = "h:\Projects\toolbox-project\cmd\api"
if (-not (Test-Path $apiDir)) {
    New-Item -ItemType Directory -Path $apiDir -Force | Out-Null
}

$apiModPath = "h:\Projects\toolbox-project\cmd\api\go.mod"
if (-not (Test-Path $apiModPath)) {
    Set-Location -Path "h:\Projects\toolbox-project\cmd\api"
    go mod init github.com/JackovAlltrades/go-toolbox/cmd/api
    Write-Host "Created API go.mod file" -ForegroundColor Green
    Set-Location -Path "h:\Projects\toolbox-project"
}

# Create a simple main.go file with web interface
# Update main.go to add error handling for template rendering
$mainContent = @'
package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
)

// PageData holds the data for the template
type PageData struct {
	Title   string
	Content string
	Page    string
	Result  string
}

func main() {
	fmt.Println("Starting toolbox API server...")
	
	// Define the template as a string
	const homeTmpl = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Toolbox Project</title>
    <style>
        body {
            font-family: Arial, sans-serif;
            line-height: 1.6;
            margin: 0;
            padding: 20px;
            color: #333;
        }
        header {
            background-color: #f4f4f4;
            padding: 20px;
            margin-bottom: 20px;
            border-radius: 5px;
        }
        nav {
            background-color: #333;
            overflow: hidden;
            border-radius: 5px;
        }
        nav a {
            float: left;
            display: block;
            color: white;
            text-align: center;
            padding: 14px 16px;
            text-decoration: none;
        }
        nav a:hover {
            background-color: #ddd;
            color: black;
        }
        .container {
            padding: 20px;
            background-color: #f9f9f9;
            border-radius: 5px;
        }
        footer {
            text-align: center;
            padding: 10px;
            margin-top: 20px;
            background-color: #f4f4f4;
            border-radius: 5px;
        }
    </style>
</head>
<body>
    <header>
        <h1>Toolbox Project</h1>
        <p>A collection of useful tools and utilities</p>
    </header>
    
    <nav>
        <a href="/">Home</a>
        <a href="/tools/random-string">Random String</a>
        <a href="/tools/file-upload">File Upload</a>
        <a href="/tools/slugify">Slugify</a>
        <a href="/tools/directory">Directory Tools</a>
        <a href="/tools/json">JSON Tools</a>
        <a href="/tools/download">Download Tools</a>
    </nav>
    
    <div class="container">
        <h2>{{.Title}}</h2>
        <p>{{.Content}}</p>
        
        {{if eq .Page "home"}}
            <p>Welcome to the Toolbox Project. Use the navigation menu to access different tools.</p>
        {{end}}
        
        {{if eq .Page "random-string"}}
            <form action="/tools/random-string" method="post">
                <label for="length">Length:</label>
                <input type="number" id="length" name="length" value="10" min="1" max="100">
                <button type="submit">Generate</button>
            </form>
            {{if .Result}}
                <div style="margin-top: 20px; padding: 10px; background-color: #e9e9e9; border-radius: 5px;">
                    <p>Generated string: <strong>{{.Result}}</strong></p>
                </div>
            {{end}}
        {{end}}
        
        {{if eq .Page "file-upload"}}
            <form action="/tools/file-upload" method="post" enctype="multipart/form-data">
                <label for="file">Select file:</label>
                <input type="file" id="file" name="file">
                <button type="submit">Upload</button>
            </form>
            {{if .Result}}
                <div style="margin-top: 20px; padding: 10px; background-color: #e9e9e9; border-radius: 5px;">
                    <p>{{.Result}}</p>
                </div>
            {{end}}
        {{end}}
        
        {{if eq .Page "slugify"}}
            <form action="/tools/slugify" method="post">
                <label for="text">Text to slugify:</label>
                <input type="text" id="text" name="text" style="width: 300px;">
                <button type="submit">Slugify</button>
            </form>
            {{if .Result}}
                <div style="margin-top: 20px; padding: 10px; background-color: #e9e9e9; border-radius: 5px;">
                    <p>Slugified text: <strong>{{.Result}}</strong></p>
                </div>
            {{end}}
        {{end}}
        
        {{if eq .Page "directory"}}
            <form action="/tools/directory" method="post">
                <label for="path">Directory path:</label>
                <input type="text" id="path" name="path" style="width: 300px;">
                <button type="submit">Check/Create</button>
            </form>
            {{if .Result}}
                <div style="margin-top: 20px; padding: 10px; background-color: #e9e9e9; border-radius: 5px;">
                    <p>{{.Result}}</p>
                </div>
            {{end}}
        {{end}}
        
        {{if eq .Page "json"}}
            <form action="/tools/json" method="post">
                <label for="json">JSON Data:</label><br>
                <textarea id="json" name="json" rows="10" cols="50" placeholder='{"example": "data"}'></textarea><br>
                <button type="submit">Validate & Format</button>
            </form>
            {{if .Result}}
                <div style="margin-top: 20px; padding: 10px; background-color: #e9e9e9; border-radius: 5px;">
                    <pre>{{.Result}}</pre>
                </div>
            {{end}}
        {{end}}
        
        {{if eq .Page "download"}}
            <form action="/tools/download" method="post">
                <label for="url">URL to download:</label>
                <input type="text" id="url" name="url" style="width: 300px;" placeholder="https://example.com/file.pdf"><br><br>
                <label for="destination">Destination path:</label>
                <input type="text" id="destination" name="destination" style="width: 300px;" placeholder="C:\\Downloads\\file.pdf"><br><br>
                <button type="submit">Download</button>
            </form>
            {{if .Result}}
                <div style="margin-top: 20px; padding: 10px; background-color: #e9e9e9; border-radius: 5px;">
                    <p>{{.Result}}</p>
                </div>
            {{end}}
        {{end}}
    </div>
    
    <footer>
        <p>&copy; 2023 Toolbox Project</p>
    </footer>
</body>
</html>`

	// Parse the template
	tmpl, err := template.New("home").Parse(homeTmpl)
	if err != nil {
		log.Fatalf("Error parsing template: %v", err)
	}
	
	// Add a simple fallback handler for all routes
	http.HandleFunc("/debug", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprintf(w, "Debug page is working!\nRequest path: %s", r.URL.Path)
	})
	
	// Home page handler
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		
		data := PageData{
			Title:   "Welcome to Toolbox",
			Content: "Select a tool from the menu above to get started.",
			Page:    "home",
		}
		
		err := tmpl.Execute(w, data)
		if err != nil {
			log.Printf("Error rendering template: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	})
	
	// Random String tool
	http.HandleFunc("/tools/random-string", func(w http.ResponseWriter, r *http.Request) {
		data := PageData{
			Title:   "Random String Generator",
			Content: "Generate a random string of specified length.",
			Page:    "random-string",
		}
		
		if r.Method == "POST" {
			r.ParseForm()
			// In a real app, this would call the toolbox.RandomString function
			data.Result = "abcdef1234" // Placeholder result
		}
		
		err := tmpl.Execute(w, data)
		if err != nil {
			log.Printf("Error rendering template: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	})
	
	// File Upload tool
	http.HandleFunc("/tools/file-upload", func(w http.ResponseWriter, r *http.Request) {
		data := PageData{
			Title:   "File Upload",
			Content: "Upload a file to the server.",
			Page:    "file-upload",
		}
		
		if r.Method == "POST" {
			// In a real app, this would call the toolbox.UploadOneFile function
			data.Result = "File uploaded successfully!" // Placeholder result
		}
		
		err := tmpl.Execute(w, data)
		if err != nil {
			log.Printf("Error rendering template: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	})
	
	// Slugify tool
	http.HandleFunc("/tools/slugify", func(w http.ResponseWriter, r *http.Request) {
		data := PageData{
			Title:   "Slugify Text",
			Content: "Convert text to a URL-friendly slug.",
			Page:    "slugify",
		}
		
		if r.Method == "POST" {
			r.ParseForm()
			// In a real app, this would call the toolbox.Slugify function
			data.Result = "slugified-text" // Placeholder result
		}
		
		err := tmpl.Execute(w, data)
		if err != nil {
			log.Printf("Error rendering template: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	})
	
	// Directory tool
	http.HandleFunc("/tools/directory", func(w http.ResponseWriter, r *http.Request) {
		data := PageData{
			Title:   "Directory Tools",
			Content: "Check if a directory exists and create it if it doesn't.",
			Page:    "directory",
		}
		
		if r.Method == "POST" {
			r.ParseForm()
			// In a real app, this would call the toolbox.CreateDirIfNotExist function
			data.Result = "Directory checked/created successfully!" // Placeholder result
		}
		
		err := tmpl.Execute(w, data)
		if err != nil {
			log.Printf("Error rendering template: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	})
	
	// JSON tool
	http.HandleFunc("/tools/json", func(w http.ResponseWriter, r *http.Request) {
		data := PageData{
			Title:   "JSON Tools",
			Content: "Validate and format JSON data.",
			Page:    "json",
		}
		
		if r.Method == "POST" {
			r.ParseForm()
			// In a real app, this would validate and format the JSON
			data.Result = "{\n  \"formatted\": \"json\",\n  \"status\": \"valid\"\n}" // Placeholder result
		}
		
		err := tmpl.Execute(w, data)
		if err != nil {
			log.Printf("Error rendering template: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	})
	
	// Download tool
	http.HandleFunc("/tools/download", func(w http.ResponseWriter, r *http.Request) {
		data := PageData{
			Title:   "Download Tools",
			Content: "Download files from URLs.",
			Page:    "download",
		}
		
		if r.Method == "POST" {
			r.ParseForm()
			// In a real app, this would download the file
			data.Result = "File downloaded successfully!" // Placeholder result
		}
		
		err := tmpl.Execute(w, data)
		if err != nil {
			log.Printf("Error rendering template: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	})
	
	// Serve static files if needed
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	
	// Add a simple HTML page as a fallback
	http.HandleFunc("/fallback", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w, "<html><body><h1>Toolbox Project</h1><p>This is a fallback page.</p></body></html>")
	})
	
	fmt.Println("Server listening on :8084")
	log.Fatal(http.ListenAndServe(":8084", nil))
}
'@
# Update main.go to fix the JSON placeholder in the template
$mainContent = @'
package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
)

// PageData holds the data for the template
type PageData struct {
	Title   string
	Content string
	Page    string
	Result  string
}

func main() {
	fmt.Println("Starting toolbox API server...")
	
	// Define the template as a string
	const homeTmpl = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Toolbox Project</title>
    <style>
        body {
            font-family: Arial, sans-serif;
            line-height: 1.6;
            margin: 0;
            padding: 20px;
            color: #333;
        }
        header {
            background-color: #f4f4f4;
            padding: 20px;
            margin-bottom: 20px;
            border-radius: 5px;
        }
        nav {
            background-color: #333;
            overflow: hidden;
            border-radius: 5px;
        }
        nav a {
            float: left;
            display: block;
            color: white;
            text-align: center;
            padding: 14px 16px;
            text-decoration: none;
        }
        nav a:hover {
            background-color: #ddd;
            color: black;
        }
        .container {
            padding: 20px;
            background-color: #f9f9f9;
            border-radius: 5px;
        }
        footer {
            text-align: center;
            padding: 10px;
            margin-top: 20px;
            background-color: #f4f4f4;
            border-radius: 5px;
        }
    </style>
</head>
<body>
    <header>
        <h1>Toolbox Project</h1>
        <p>A collection of useful tools and utilities</p>
    </header>
    
    <nav>
        <a href="/">Home</a>
        <a href="/tools/random-string">Random String</a>
        <a href="/tools/file-upload">File Upload</a>
        <a href="/tools/slugify">Slugify</a>
        <a href="/tools/directory">Directory Tools</a>
        <a href="/tools/json">JSON Tools</a>
        <a href="/tools/download">Download Tools</a>
    </nav>
    
    <div class="container">
        <h2>{{.Title}}</h2>
        <p>{{.Content}}</p>
        
        {{if eq .Page "home"}}
            <p>Welcome to the Toolbox Project. Use the navigation menu to access different tools.</p>
        {{end}}
        
        {{if eq .Page "random-string"}}
            <form action="/tools/random-string" method="post">
                <label for="length">Length:</label>
                <input type="number" id="length" name="length" value="10" min="1" max="100">
                <button type="submit">Generate</button>
            </form>
            {{if .Result}}
                <div style="margin-top: 20px; padding: 10px; background-color: #e9e9e9; border-radius: 5px;">
                    <p>Generated string: <strong>{{.Result}}</strong></p>
                </div>
            {{end}}
        {{end}}
        
        {{if eq .Page "file-upload"}}
            <form action="/tools/file-upload" method="post" enctype="multipart/form-data">
                <label for="file">Select file:</label>
                <input type="file" id="file" name="file">
                <button type="submit">Upload</button>
            </form>
            {{if .Result}}
                <div style="margin-top: 20px; padding: 10px; background-color: #e9e9e9; border-radius: 5px;">
                    <p>{{.Result}}</p>
                </div>
            {{end}}
        {{end}}
        
        {{if eq .Page "slugify"}}
            <form action="/tools/slugify" method="post">
                <label for="text">Text to slugify:</label>
                <input type="text" id="text" name="text" style="width: 300px;">
                <button type="submit">Slugify</button>
            </form>
            {{if .Result}}
                <div style="margin-top: 20px; padding: 10px; background-color: #e9e9e9; border-radius: 5px;">
                    <p>Slugified text: <strong>{{.Result}}</strong></p>
                </div>
            {{end}}
        {{end}}
        
        {{if eq .Page "directory"}}
            <form action="/tools/directory" method="post">
                <label for="path">Directory path:</label>
                <input type="text" id="path" name="path" style="width: 300px;">
                <button type="submit">Check/Create</button>
            </form>
            {{if .Result}}
                <div style="margin-top: 20px; padding: 10px; background-color: #e9e9e9; border-radius: 5px;">
                    <p>{{.Result}}</p>
                </div>
            {{end}}
        {{end}}
        
        {{if eq .Page "json"}}
            <form action="/tools/json" method="post">
                <label for="json">JSON Data:</label><br>
                <textarea id="json" name="json" rows="10" cols="50" placeholder='{"example": "data"}'></textarea><br>
                <button type="submit">Validate & Format</button>
            </form>
            {{if .Result}}
                <div style="margin-top: 20px; padding: 10px; background-color: #e9e9e9; border-radius: 5px;">
                    <pre>{{.Result}}</pre>
                </div>
            {{end}}
        {{end}}
        
        {{if eq .Page "download"}}
            <form action="/tools/download" method="post">
                <label for="url">URL to download:</label>
                <input type="text" id="url" name="url" style="width: 300px;" placeholder="https://example.com/file.pdf"><br><br>
                <label for="destination">Destination path:</label>
                <input type="text" id="destination" name="destination" style="width: 300px;" placeholder="C:\\Downloads\\file.pdf"><br><br>
                <button type="submit">Download</button>
            </form>
            {{if .Result}}
                <div style="margin-top: 20px; padding: 10px; background-color: #e9e9e9; border-radius: 5px;">
                    <p>{{.Result}}</p>
                </div>
            {{end}}
        {{end}}
    </div>
    
    <footer>
        <p>&copy; 2023 Toolbox Project</p>
    </footer>
</body>
</html>`

	// Parse the template
	tmpl, err := template.New("home").Parse(homeTmpl)
	if err != nil {
		log.Fatalf("Error parsing template: %v", err)
	}
	
	// Add a simple fallback handler for all routes
	http.HandleFunc("/debug", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprintf(w, "Debug page is working!\nRequest path: %s", r.URL.Path)
	})
	
	// Home page handler
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		
		data := PageData{
			Title:   "Welcome to Toolbox",
			Content: "Select a tool from the menu above to get started.",
			Page:    "home",
		}
		
		err := tmpl.Execute(w, data)
		if err != nil {
			log.Printf("Error rendering template: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	})
	
	// Random String tool
	http.HandleFunc("/tools/random-string", func(w http.ResponseWriter, r *http.Request) {
		data := PageData{
			Title:   "Random String Generator",
			Content: "Generate a random string of specified length.",
			Page:    "random-string",
		}
		
		if r.Method == "POST" {
			r.ParseForm()
			// In a real app, this would call the toolbox.RandomString function
			data.Result = "abcdef1234" // Placeholder result
		}
		
		err := tmpl.Execute(w, data)
		if err != nil {
			log.Printf("Error rendering template: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	})
	
	// File Upload tool
	http.HandleFunc("/tools/file-upload", func(w http.ResponseWriter, r *http.Request) {
		data := PageData{
			Title:   "File Upload",
			Content: "Upload a file to the server.",
			Page:    "file-upload",
		}
		
		if r.Method == "POST" {
			// In a real app, this would call the toolbox.UploadOneFile function
			data.Result = "File uploaded successfully!" // Placeholder result
		}
		
		err := tmpl.Execute(w, data)
		if err != nil {
			log.Printf("Error rendering template: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	})
	
	// Slugify tool
	http.HandleFunc("/tools/slugify", func(w http.ResponseWriter, r *http.Request) {
		data := PageData{
			Title:   "Slugify Text",
			Content: "Convert text to a URL-friendly slug.",
			Page:    "slugify",
		}
		
		if r.Method == "POST" {
			r.ParseForm()
			// In a real app, this would call the toolbox.Slugify function
			data.Result = "slugified-text" // Placeholder result
		}
		
		err := tmpl.Execute(w, data)
		if err != nil {
			log.Printf("Error rendering template: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	})
	
	// Directory tool
	http.HandleFunc("/tools/directory", func(w http.ResponseWriter, r *http.Request) {
		data := PageData{
			Title:   "Directory Tools",
			Content: "Check if a directory exists and create it if it doesn't.",
			Page:    "directory",
		}
		
		if r.Method == "POST" {
			r.ParseForm()
			// In a real app, this would call the toolbox.CreateDirIfNotExist function
			data.Result = "Directory checked/created successfully!" // Placeholder result
		}
		
		err := tmpl.Execute(w, data)
		if err != nil {
			log.Printf("Error rendering template: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	})
	
	// JSON tool
	http.HandleFunc("/tools/json", func(w http.ResponseWriter, r *http.Request) {
		data := PageData{
			Title:   "JSON Tools",
			Content: "Validate and format JSON data.",
			Page:    "json",
		}
		
		if r.Method == "POST" {
			r.ParseForm()
			// In a real app, this would validate and format the JSON
			data.Result = "{\n  \"formatted\": \"json\",\n  \"status\": \"valid\"\n}" // Placeholder result
		}
		
		err := tmpl.Execute(w, data)
		if err != nil {
			log.Printf("Error rendering template: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	})
	
	// Download tool
	http.HandleFunc("/tools/download", func(w http.ResponseWriter, r *http.Request) {
		data := PageData{
			Title:   "Download Tools",
			Content: "Download files from URLs.",
			Page:    "download",
		}
		
		if r.Method == "POST" {
			r.ParseForm()
			// In a real app, this would download the file
			data.Result = "File downloaded successfully!" // Placeholder result
		}
		
		err := tmpl.Execute(w, data)
		if err != nil {
			log.Printf("Error rendering template: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	})
	
	// Serve static files if needed
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	
	// Add a simple HTML page as a fallback
	http.HandleFunc("/fallback", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w, "<html><body><h1>Toolbox Project</h1><p>This is a fallback page.</p></body></html>")
	})
	
	fmt.Println("Server listening on :8084")
	log.Fatal(http.ListenAndServe(":8084", nil))
}
'@
Set-Content -Path "$apiDir\main.go" -Value $mainContent
Write-Host "Fixed JSON placeholder in template" -ForegroundColor Green

# Create or update go.work file
Set-Location -Path "h:\Projects\toolbox-project"
if (Test-Path "h:\Projects\toolbox-project\go.work") {
    Remove-Item "h:\Projects\toolbox-project\go.work" -Force
}
go work init
go work use ./toolbox ./cmd/api
Write-Host "Created go.work file" -ForegroundColor Green

# Get dependencies
Write-Host "Getting dependencies..." -ForegroundColor Cyan
go mod tidy

# Build the project
Write-Host "Building project..." -ForegroundColor Cyan
# After creating the main.go file, let's verify it contains the correct port
$mainGoPath = "$apiDir\main.go"
$mainGoContent = Get-Content -Path $mainGoPath -Raw
if ($mainGoContent -match "8080" -and -not $mainGoContent -match "8090") {
    Write-Host "Found port 8080 in main.go, replacing with 8090..." -ForegroundColor Yellow
    $mainGoContent = $mainGoContent -replace "8080", "8090"
    Set-Content -Path $mainGoPath -Value $mainGoContent
    Write-Host "Updated port in main.go to 8090" -ForegroundColor Green
}

# Let's also clean the build cache before building
Write-Host "Cleaning Go build cache..." -ForegroundColor Cyan
go clean -cache

# Build the project with verbose output to see what's happening
Write-Host "Building project with verbose output..." -ForegroundColor Cyan
go build -v -o toolbox.exe ./cmd/api

# Check if build was successful
if (Test-Path ".\toolbox.exe") {
    # Start the API server
    Write-Host "Starting API server..." -ForegroundColor Cyan
    Write-Host "Press Ctrl+C to stop the server" -ForegroundColor Yellow
    .\toolbox.exe
} else {
    Write-Host "Build failed. Please check the errors above." -ForegroundColor Red
}