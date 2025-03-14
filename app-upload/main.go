package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/JackovAlltrades/go-toolbox"
	"github.com/joho/godotenv"
	"flag"
)

// Default values for configuration
const (
	defaultUploadDir      = "./uploads"
	defaultTempDir        = "./temp"
	defaultMaxFileSize    = 100 * 1024 * 1024  // 100MB
	defaultMaxBatchSize   = 1 * 1024 * 1024 * 1024 // 1GB
	defaultServerPort     = "8090"
	defaultTemplatePath   = "./templates"
	defaultMaxUploadCount = 5 // Maximum number of files per upload
)

// Config holds the application configuration
type Config struct {
	UploadDir      string
	TempDir        string
	MaxFileSize    int
	MaxBatchSize   int64
	ServerPort     string
	TemplatePath   string
	MaxUploadCount int
}

// AppConfig holds the application configuration
type AppConfig struct {
	Tools  *toolbox.Tools
	Config *Config
}

// loadConfig loads configuration from environment variables
// Add these imports
import (
	"encoding/json"
	"io/ioutil"
)

// Add a function to load config from file
func loadConfigFromFile(filePath string) (*Config, error) {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	
	var config struct {
		UploadDir      string `json:"upload_dir"`
		TempDir        string `json:"temp_dir"`
		MaxFileSize    int    `json:"max_file_size"`
		MaxBatchSize   int64  `json:"max_batch_size"`
		ServerPort     string `json:"server_port"`
		TemplatePath   string `json:"template_path"`
		MaxUploadCount int    `json:"max_upload_count"`
	}
	
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}
	
	return &Config{
		UploadDir:      config.UploadDir,
		TempDir:        config.TempDir,
		MaxFileSize:    config.MaxFileSize,
		MaxBatchSize:   config.MaxBatchSize,
		ServerPort:     config.ServerPort,
		TemplatePath:   config.TemplatePath,
		MaxUploadCount: config.MaxUploadCount,
	}, nil
}

// Update loadConfig to try config file first, then env vars
func loadConfig() *Config {
	// Try to load from config file first
	config, err := loadConfigFromFile("config.json")
	if err == nil {
		return config
	}
	
	// Fall back to environment variables
	// Define command line flags
	uploadDir := flag.String("upload-dir", getEnv("UPLOAD_DIR", defaultUploadDir), "Directory for uploaded files")
	tempDir := flag.String("temp-dir", getEnv("TEMP_DIR", defaultTempDir), "Directory for temporary files")
	maxFileSize := flag.Int("max-file-size", getEnvInt("MAX_FILE_SIZE", defaultMaxFileSize), "Maximum size of individual files in bytes")
	maxBatchSize := flag.Int64("max-batch-size", getEnvInt64("MAX_BATCH_SIZE", defaultMaxBatchSize), "Maximum total size of all files in a batch in bytes")
	serverPort := flag.String("port", getEnv("SERVER_PORT", defaultServerPort), "Port for the HTTP server")
	templatePath := flag.String("template-path", getEnv("TEMPLATE_PATH", defaultTemplatePath), "Path to the templates directory")
	maxUploadCount := flag.Int("max-upload-count", getEnvInt("MAX_UPLOAD_COUNT", defaultMaxUploadCount), "Maximum number of files per upload")
	
	// Parse flags
	flag.Parse()
	
	return &Config{
		UploadDir:      *uploadDir,
		TempDir:        *tempDir,
		MaxFileSize:    *maxFileSize,
		MaxBatchSize:   *maxBatchSize,
		ServerPort:     *serverPort,
		TemplatePath:   *templatePath,
		MaxUploadCount: *maxUploadCount,
	}
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// getEnvInt gets an environment variable as an integer or returns a default value
func getEnvInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		log.Printf("Warning: Invalid value for %s, using default: %v", key, err)
		return defaultValue
	}
	return value
}

// getEnvInt64 gets an environment variable as an int64 or returns a default value
func getEnvInt64(key string, defaultValue int64) int64 {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.ParseInt(valueStr, 10, 64)
	if err != nil {
		log.Printf("Warning: Invalid value for %s, using default: %v", key, err)
		return defaultValue
	}
	return value
}

// UploadHandler handles file uploads
func (app *AppConfig) UploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse the multipart form to access files
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		http.Error(w, "Error parsing form: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Check if any files were uploaded
	if r.MultipartForm == nil || r.MultipartForm.File == nil {
		http.Error(w, "No files were uploaded", http.StatusBadRequest)
		return
	}

	// Calculate total batch size and check individual file sizes
	totalBatchSize := int64(0)
	fileCount := 0

	for _, fHeaders := range r.MultipartForm.File {
		fileCount += len(fHeaders)
		
		for _, header := range fHeaders {
			totalBatchSize += header.Size
			
			// Check individual file size
			if header.Size > int64(app.Config.MaxFileSize) {
				http.Error(w, fmt.Sprintf(
					"File '%s' exceeds the maximum allowed size (%d MB)", 
					header.Filename, 
					app.Config.MaxFileSize/(1024*1024),
				), http.StatusBadRequest)
				return
			}
		}
	}

	// Check total number of files
	if fileCount > app.Config.MaxUploadCount {
		http.Error(w, fmt.Sprintf(
			"Maximum %d files allowed per upload, got %d", 
			app.Config.MaxUploadCount, 
			fileCount,
		), http.StatusBadRequest)
		return
	}

	// Check total batch size
	if totalBatchSize > app.Config.MaxBatchSize {
		http.Error(w, fmt.Sprintf(
			"Total batch size exceeds limit (%d MB / %d GB)", 
			totalBatchSize/(1024*1024), 
			app.Config.MaxBatchSize/(1024*1024*1024),
		), http.StatusBadRequest)
		return
	}

	// After all checks pass, proceed with upload
	files, err := app.Tools.UploadFiles(r, "", true)
	if err != nil {
		http.Error(w, "Upload failed: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Render the results using a template
	data := struct {
		Files []*toolbox.UploadedFile
		TotalSize int64
	}{
		Files: files,
		TotalSize: totalBatchSize,
	}

	// Create template with formatFileSize function
	tmpl := template.New("result.html").Funcs(template.FuncMap{
		"formatFileSize": func(size int64) string {
			const (
				KB = 1024
				MB = 1024 * KB
				GB = 1024 * MB
			)

			switch {
			case size >= GB:
				return fmt.Sprintf("%.2f GB", float64(size)/float64(GB))
			case size >= MB:
				return fmt.Sprintf("%.2f MB", float64(size)/float64(MB))
			case size >= KB:
				return fmt.Sprintf("%.2f KB", float64(size)/float64(KB))
			default:
				return fmt.Sprintf("%d bytes", size)
			}
		},
	})
	
	tmpl, err = tmpl.ParseFiles(filepath.Join(templatePath, "result.html"))
	if err != nil {
		log.Printf("Error parsing template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	
	if err := tmpl.Execute(w, data); err != nil {
		log.Printf("Error rendering template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// Helper function to format file sizes in a human-readable way
func formatSize(bytes int) string {
	const (
		KB = 1024
		MB = 1024 * KB
		GB = 1024 * MB
	)

	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.2f KB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

// HomeHandler renders the home page with the upload form
func (app *AppConfig) HomeHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	// Pass the limits to the template
	data := struct {
		MaxUploadCount int
		MaxFileSize    string
		MaxBatchSize   string
		TypeLimits     map[string]string
	}{
		MaxUploadCount: app.Config.MaxUploadCount,
		MaxFileSize:    formatSize(app.Config.MaxFileSize),
		MaxBatchSize:   formatSize(int(app.Config.MaxBatchSize)),
		TypeLimits: map[string]string{
			"document": formatSize(20 * 1024 * 1024),  // 20MB
			"image":    formatSize(10 * 1024 * 1024),  // 10MB
			"text":     formatSize(5 * 1024 * 1024),   // 5MB
			"audio":    formatSize(50 * 1024 * 1024),  // 50MB
			"video":    formatSize(100 * 1024 * 1024), // 100MB
			"archive":  formatSize(50 * 1024 * 1024),  // 50MB
		},
	}

	tmpl := template.Must(template.ParseFiles(filepath.Join(templatePath, "index.html")))
	if err := tmpl.Execute(w, data); err != nil {
		log.Printf("Error rendering template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// setupRoutes configures the HTTP routes
func (app *AppConfig) setupRoutes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", app.HomeHandler)
	mux.HandleFunc("/upload", app.UploadHandler)
	
	return mux
}

// setupTemplates ensures the template directory exists
func setupTemplates(templatePath string) error {
	if err := os.MkdirAll(templatePath, 0755); err != nil {
		return err
	}
	
	// Create index.html template if it doesn't exist
	indexPath := filepath.Join(templatePath, "index.html")
	if _, err := os.Stat(indexPath); os.IsNotExist(err) {
		indexContent := `<!DOCTYPE html>
<html>
<head>
	<title>File Upload Example</title>
	<style>
		body {
			font-family: Arial, sans-serif;
			max-width: 800px;
			margin: 0 auto;
			padding: 20px;
		}
		.upload-form {
			border: 1px solid #ddd;
			padding: 20px;
			border-radius: 5px;
		}
		.info {
			background-color: #f8f9fa;
			padding: 15px;
			margin-top: 20px;
			border-left: 4px solid #17a2b8;
		}
		.limits-table {
			width: 100%;
			border-collapse: collapse;
			margin-top: 10px;
		}
		.limits-table th, .limits-table td {
			border: 1px solid #ddd;
			padding: 8px;
			text-align: left;
		}
		.limits-table th {
			background-color: #f2f2f2;
		}
		.progress-container {
			width: 100%;
			background-color: #f1f1f1;
			margin-top: 10px;
			display: none;
		}
		.progress-bar {
			width: 0%;
			height: 20px;
			background-color: #4CAF50;
			text-align: center;
			line-height: 20px;
			color: white;
		}
	</style>
</head>
<body>
	<h1>File Upload Example</h1>
	<div class="upload-form">
		<form id="uploadForm" action="/upload" method="post" enctype="multipart/form-data">
			<h3>Select files to upload:</h3>
			<input type="file" name="file" id="fileInput" multiple><br>
			<small>Maximum {{.MaxUploadCount}} files, individual file limit {{.MaxFileSize}}, total batch size limit {{.MaxBatchSize}}</small><br><br>
			<input type="submit" value="Upload" id="submitBtn">
		</form>
		<div class="progress-container" id="progressContainer">
			<div class="progress-bar" id="progressBar">0%</div>
		</div>
	</div>
	<div class="info">
		<h3>Allowed File Types:</h3>
		<ul>
			<li>Images: JPEG, PNG, GIF, WebP</li>
			<li>Documents: PDF, TXT, DOC, DOCX</li>
			<li>Archives: ZIP, RAR</li>
		</ul>
		
		<h3>File Size Limits:</h3>
		<table class="limits-table">
			<tr>
				<th>File Type</th>
				<th>Size Limit</th>
			</tr>
			<tr>
				<td>Documents (PDF, DOCX, etc.)</td>
				<td>{{.TypeLimits.document}}</td>
			</tr>
			<tr>
				<td>Images (JPEG, PNG, etc.)</td>
				<td>{{.TypeLimits.image}}</td>
			</tr>
			<tr>
				<td>Text Files</td>
				<td>{{.TypeLimits.text}}</td>
			</tr>
			<tr>
				<td>Audio Files</td>
				<td>{{.TypeLimits.audio}}</td>
			</tr>
			<tr>
				<td>Video Files</td>
				<td>{{.TypeLimits.video}}</td>
			</tr>
			<tr>
				<td>Archive Files</td>
				<td>{{.TypeLimits.archive}}</td>
			</tr>
		</table>
		
		<p>Total batch size limit: {{.MaxBatchSize}}</p>
	</div>

	<script>
		// Add upload progress bar functionality
		document.getElementById('uploadForm').onsubmit = function(e) {
			e.preventDefault();
			
			var fileInput = document.getElementById('fileInput');
			var files = fileInput.files;
			
			// Check if files were selected
			if (files.length === 0) {
				alert('Please select at least one file');
				return false;
			}
			
			// Check file count
			if (files.length > {{.MaxUploadCount}}) {
				alert('Maximum {{.MaxUploadCount}} files allowed');
				return false;
			}
			
			// Calculate total size
			var totalSize = 0;
			for (var i = 0; i < files.length; i++) {
				totalSize += files[i].size;
			}
			
			// Show progress bar
			document.getElementById('progressContainer').style.display = 'block';
			document.getElementById('submitBtn').disabled = true;
			
			var formData = new FormData(this);
			var xhr = new XMLHttpRequest();
			
			xhr.open('POST', '/upload', true);
			
			xhr.upload.onprogress = function(e) {
				if (e.lengthComputable) {
					var percent = Math.round((e.loaded / e.total) * 100);
					document.getElementById('progressBar').style.width = percent + '%';
					document.getElementById('progressBar').textContent = percent + '%';
				}
			};
			
			xhr.onload = function() {
				if (xhr.status === 200) {
					document.open();
					document.write(xhr.responseText);
					document.close();
				} else {
					alert('Upload failed: ' + xhr.responseText);
					document.getElementById('progressContainer').style.display = 'none';
					document.getElementById('submitBtn').disabled = false;
				}
			};
			
			xhr.onerror = function() {
				alert('Upload failed. Please try again.');
				document.getElementById('progressContainer').style.display = 'none';
				document.getElementById('submitBtn').disabled = false;
			};
			
			xhr.send(formData);
			return false;
		};
	</script>
</body>
</html>`
		if err := os.WriteFile(indexPath, []byte(indexContent), 0644); err != nil {
			return err
		}
	}
	
	// Create result.html template if it doesn't exist
	resultPath := filepath.Join(templatePath, "result.html")
	if _, err := os.Stat(resultPath); os.IsNotExist(err) {
		resultContent := `<!DOCTYPE html>
<html>
<head>
	<title>Upload Results</title>
	<style>
		body {
			font-family: Arial, sans-serif;
			max-width: 800px;
			margin: 0 auto;
			padding: 20px;
		}
		.result {
			background-color: #f8f9fa;
			padding: 15px;
			margin-top: 10px;
			border-left: 4px solid #28a745;
		}
		.back-link {
			margin-top: 20px;
		}
		.summary {
			background-color: #e9f7ef;
			padding: 15px;
			margin-bottom: 20px;
			border-radius: 5px;
		}
	</style>
</head>
<body>
	<h1>Uploaded Files</h1>
	
	<div class="summary">
		<p><strong>Files uploaded:</strong> {{len .Files}}</p>
		<p><strong>Total size:</strong> {{if .TotalSize}}{{formatFileSize .TotalSize}}{{else}}0 bytes{{end}}</p>
	</div>
	
	{{range .Files}}
	<div class="result">
		<p><strong>Original:</strong> {{.OriginalFileName}}</p>
		<p><strong>Saved as:</strong> {{.NewFileName}}</p>
		<p><strong>Size:</strong> {{formatFileSize .FileSize}}</p>
		<p><strong>Type:</strong> {{.FileType}}</p>
	</div>
	{{end}}
	<div class="back-link">
		<a href="/">Back to Upload Form</a>
	</div>
</body>
</html>`
		if err := os.WriteFile(resultPath, []byte(resultContent), 0644); err != nil {
			return err
		}
	}
	
	return nil
}

// In the main function, update the tools initialization:
func main() {
	// Setup templates
	if err := setupTemplates(); err != nil {
		log.Fatal("Error setting up templates:", err)
	}

	// Create directories
	absUploadDir, err := filepath.Abs(uploadDir)
	if err != nil {
		log.Fatal("Error getting absolute path for upload directory:", err)
	}

	absTempDir, err := filepath.Abs(tempDir)
	if err != nil {
		log.Fatal("Error getting absolute path for temp directory:", err)
	}

	// Create a new instance of the toolbox with type-specific limits
	tools := &toolbox.Tools{
		MaxFileSize: maxFileSize,
		AllowedFileTypes: []string{
			// Images
			"image/jpeg", "image/png", "image/gif", "image/webp",
			// Documents
			"application/pdf", "text/plain", "application/msword",
			"application/vnd.openxmlformats-officedocument.wordprocessingml.document",
			// Archives
			"application/zip", "application/x-rar-compressed",
		},
		AllowUnknownTypes: false,
		MaxUploadCount:    maxUploadCount,
		UploadPath:        absUploadDir,
		TempFilePath:      absTempDir,
		// Add type-specific size limits if supported by the toolbox
		TypeSpecificSizeLimits: map[string]int{
			"image/jpeg": 10 * 1024 * 1024, // 10MB for JPEG
			"image/png":  10 * 1024 * 1024, // 10MB for PNG
			"application/pdf": 20 * 1024 * 1024, // 20MB for PDF
			"text/plain": 5 * 1024 * 1024, // 5MB for text
		},
	}

	// Create upload and temp directories
	if err := tools.CreateDirIfNotExist(absUploadDir); err != nil {
		log.Fatal("Could not create uploads directory:", err)
	}

	if err := tools.CreateDirIfNotExist(absTempDir); err != nil {
		log.Fatal("Could not create temp directory:", err)
	}

	// Initialize application
	app := &AppConfig{
		Tools: tools,
		Config: config,
	}

	// Setup routes
	handler := app.setupRoutes()

	// Start the server
	fmt.Printf("Starting server on http://localhost:%s\n", config.ServerPort)
	fmt.Println("- Upload page: http://localhost:" + config.ServerPort + "/")
	fmt.Println("- Upload endpoint: http://localhost:" + config.ServerPort + "/upload")
	fmt.Printf("- Upload directory: %s\n", absUploadDir)
	fmt.Printf("- Temporary directory: %s\n", absTempDir)
	fmt.Printf("- Maximum upload count: %d files\n", config.MaxUploadCount)
	fmt.Printf("- Maximum file size: %s\n", formatSize(config.MaxFileSize))
	fmt.Printf("- Maximum batch size: %s\n", formatSize(int(config.MaxBatchSize)))
	
	if err := http.ListenAndServe(":"+config.ServerPort, handler); err != nil {
		log.Fatal("Error starting server:", err)
	}
}
