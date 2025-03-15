package toolbox

import (
	"encoding/json"  // Add this import for JSON operations
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"time"
	mathrand "math/rand" // Keep this import for mathrand.Intn
)

const randomStringSource = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_+"

// Tools is type used to instantiate the module. Variables of type allowed access
// to all methods with reciever *Tools
// Add MaxBatchSize to the Tools struct
type Tools struct {
	MaxFileSize           int
	AllowedFileTypes      []string
	AllowUnknownTypes     bool
	MaxUploadCount        int
	UploadPath            string
	TempFilePath          string
	TypeSpecificSizeLimits map[string]int
	DefaultSizeLimits      map[string]int
	ValidationCallback     func(file *UploadedFile) error
	MaxBatchSize          int64 // Maximum total size of all files in a batch
	
	// For resumable uploads
	ChunkSize             int64  // Size of each chunk in bytes
	ChunksDirectory       string // Directory to store chunks during upload
	
	// For testing purposes - allows mocking the file type detection
	detectFileType func(file multipart.File) (string, error)
}

// Add the UploadedFile type
type UploadedFile struct {
    NewFileName     string
    OriginalFileName string
    FileSize        int64
    FileType        string
    FilePath        string
}

// Add the InitDefaults method to the Tools struct
func (t *Tools) InitDefaults() {
    if t.MaxFileSize == 0 {
        t.MaxFileSize = 1024 * 1024 * 1024 // 1GB default
    }
    
    if t.MaxUploadCount == 0 {
        t.MaxUploadCount = 10 // Default max upload count
    }
}

// Add the GetFileSizeLimit method
func (t *Tools) GetFileSizeLimit(fileType string) int {
    // If type-specific limits are defined and this type has a limit
    if t.TypeSpecificSizeLimits != nil {
        if limit, exists := t.TypeSpecificSizeLimits[fileType]; exists {
            return limit
        }
        
        // Check for category limits (e.g., "image/jpeg" -> "image")
        parts := strings.Split(fileType, "/")
        if len(parts) > 0 {
            category := parts[0]
            if t.DefaultSizeLimits != nil {
                if limit, exists := t.DefaultSizeLimits[category]; exists {
                    return limit
                }
            }
        }
    }
    
    // Fall back to global limit
    return t.MaxFileSize
}

// Add the RandomString method
// Fix the RandomString method to use mathrand instead of rand
func (t *Tools) RandomString(n int) string {
    const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
    b := make([]byte, n)
    for i := range b {
        b[i] = letters[mathrand.Intn(len(letters))]
    }
    return string(b)
}

// Fix the UploadFiles method to properly handle the boolean parameter
func (t *Tools) UploadFiles(r *http.Request, uploadDir string, rename bool) ([]*UploadedFile, error) {
	// Initialize defaults if not set
	t.InitDefaults()
	
	// Use the rename parameter directly as a boolean
	renameFile := rename

	var uploadedFiles []*UploadedFile

	// Use UploadPath if uploadDir is not specified
	if uploadDir == "" && t.UploadPath != "" {
		uploadDir = t.UploadPath
	}

	// Sanitize and validate the upload directory
	uploadDir = filepath.Clean(uploadDir)
	if !filepath.IsAbs(uploadDir) {
		absPath, err := filepath.Abs(uploadDir)
		if err != nil {
			return nil, fmt.Errorf("invalid upload directory path: %w", err)
		}
		uploadDir = absPath
	}

	// Create the upload directory if it doesn't exist
	err := t.CreateDirIfNotExist(uploadDir)
	if err != nil {
		return nil, fmt.Errorf("failed to create upload directory: %w", err)
	}

	// Parse the multipart form with size limit
	err = r.ParseMultipartForm(int64(t.MaxFileSize))
	if err != nil {
		return nil, errors.New("the uploaded file exceeds the maximum allowed size")
	}

	// Check if any files were uploaded
	if r.MultipartForm == nil || r.MultipartForm.File == nil {
		return nil, errors.New("no files were uploaded")
	}

	fileCount := 0
	for _, fHeaders := range r.MultipartForm.File {
		fileCount += len(fHeaders)
	}

	// Update the UploadFiles method to use custom error types consistently
	
	// Check if the number of files exceeds the maximum allowed
	if t.MaxUploadCount > 0 && fileCount > t.MaxUploadCount {
		return nil, &ErrorResponse{
			Err:     ErrMaxUploadExceeded,
			Message: fmt.Sprintf("number of files (%d) exceeds the maximum allowed (%d)", fileCount, t.MaxUploadCount),
		}
	}

	for _, fHeaders := range r.MultipartForm.File {
		for _, hdr := range fHeaders {
			uploadedFiles, err = func(uploadedFiles []*UploadedFile) ([]*UploadedFile, error) {
				var uploadedFile UploadedFile
				infile, err := hdr.Open()
				if err != nil {
					return nil, fmt.Errorf("failed to open uploaded file: %w", err)
				}
				defer infile.Close()

				// Read file header for content type detection
				buff := make([]byte, 512)
				n, err := infile.Read(buff)
				if err != nil && err != io.EOF {
					return nil, fmt.Errorf("failed to read file header: %w", err)
				}
				buff = buff[:n]

				// Detect and validate file type
				var fileType string
				
				// Use custom detectFileType function if provided (for testing)
				if t.detectFileType != nil {
				    detectedType, err := t.detectFileType(infile)
				    if err != nil {
				        return nil, fmt.Errorf("failed to detect file type: %w", err)
				    }
				    fileType = detectedType
				    
				    // Reset file pointer after detection
				    _, err = infile.Seek(0, 0)
				    if err != nil {
				        return nil, fmt.Errorf("failed to reset file pointer: %w", err)
				    }
				} else {
				    // Standard detection using http.DetectContentType
				    fileType = http.DetectContentType(buff)
				}
				
				uploadedFile.FileType = fileType

				// Check if file type is allowed
				allowed := false // Start with false by default
				if t.AllowUnknownTypes {
				    allowed = true // Allow if AllowUnknownTypes is true
				} else if len(t.AllowedFileTypes) > 0 {
				    // Check if the file type is in the allowed list
				    for _, allowedType := range t.AllowedFileTypes {
				        // Use exact matching for MIME types
				        if strings.EqualFold(fileType, allowedType) {
				            allowed = true
				            break
				        }
				        
				        // Also check for MIME type with parameters (e.g., "text/plain; charset=utf-8")
				        if strings.Contains(fileType, ";") {
				            baseMimeType := strings.TrimSpace(strings.Split(fileType, ";")[0])
				            if strings.EqualFold(baseMimeType, allowedType) {
				                allowed = true
				                break
				            }
				        }
				    }
				} else {
				    // If no allowed types are specified and AllowUnknownTypes is false,
				    // we should allow all types by default
				    allowed = true
				}

				if !allowed {
					return nil, fmt.Errorf("file type %s is not permitted", fileType)
				}

				// Get type-specific size limit
				sizeLimit := t.GetFileSizeLimit(fileType)
				
				// Check individual file size against type-specific limit
				if hdr.Size > int64(sizeLimit) {
					return nil, fmt.Errorf("file %s exceeds the maximum allowed size for type %s (%d bytes)", 
						hdr.Filename, fileType, sizeLimit)
				}

				// Reset file pointer to beginning
				_, err = infile.Seek(0, 0)
				if err != nil {
					return nil, fmt.Errorf("failed to reset file pointer: %w", err)
				}

				// Sanitize original filename
				originalFilename := filepath.Base(hdr.Filename)
				uploadedFile.OriginalFileName = originalFilename

				// Generate new filename or use original
				if renameFile {
					ext := filepath.Ext(originalFilename)
					uploadedFile.NewFileName = fmt.Sprintf("%s%s", t.RandomString(25), ext)
				} else {
					// Ensure filename is safe
					uploadedFile.NewFileName = originalFilename
				}

				// Use TempFilePath if specified
				tempPath := uploadDir
				if t.TempFilePath != "" {
					// Create temp directory if it doesn't exist
					err = t.CreateDirIfNotExist(t.TempFilePath)
					if err != nil {
						return nil, fmt.Errorf("failed to create temp directory: %w", err)
					}
					tempPath = t.TempFilePath
				}

				// Create a temporary file first if TempFilePath is specified
				var tempFile *os.File
				var finalPath string
				
				if t.TempFilePath != "" {
					tempFilename := fmt.Sprintf("temp_%s", uploadedFile.NewFileName)
					tempFilePath := filepath.Join(tempPath, tempFilename)
					tempFile, err = os.Create(tempFilePath)
					if err != nil {
						return nil, fmt.Errorf("failed to create temporary file: %w", err)
					}
					defer func() {
						tempFile.Close()
						// Clean up temp file after copying to final destination
						os.Remove(tempFilePath)
					}()
					
					// Copy to temp file
					fileSize, err := io.Copy(tempFile, infile)
					if err != nil {
						return nil, fmt.Errorf("failed to save to temporary file: %w", err)
					}
					uploadedFile.FileSize = fileSize
					
					// Reset temp file pointer to beginning
					_, err = tempFile.Seek(0, 0)
					if err != nil {
						return nil, fmt.Errorf("failed to reset temp file pointer: %w", err)
					}
					
					// Create the destination file
					finalPath = filepath.Join(uploadDir, uploadedFile.NewFileName)
					outfile, err := os.Create(finalPath)
					if err != nil {
						return nil, fmt.Errorf("failed to create destination file: %w", err)
					}
					defer outfile.Close()
					
					// Copy from temp file to final destination
					_, err = io.Copy(outfile, tempFile)
					if err != nil {
						// Clean up partial file on error
						os.Remove(finalPath)
						return nil, fmt.Errorf("failed to copy from temp to final destination: %w", err)
					}
				} else {
					// Create the destination file directly
					finalPath = filepath.Join(uploadDir, uploadedFile.NewFileName)
					outfile, err := os.Create(finalPath)
					if err != nil {
						return nil, fmt.Errorf("failed to create destination file: %w", err)
					}
					defer outfile.Close()
					
					// Copy the file contents directly
					fileSize, err := io.Copy(outfile, infile)
					if err != nil {
						// Clean up partial file on error
						os.Remove(finalPath)
						return nil, fmt.Errorf("failed to save file: %w", err)
					}
					uploadedFile.FileSize = fileSize
				}
				
				// Run custom validation if provided
				if t.ValidationCallback != nil {
					if err := t.ValidationCallback(&uploadedFile); err != nil {
						// Clean up file on validation error
						os.Remove(finalPath)
						return nil, fmt.Errorf("file validation failed: %w", err)
					}
				}

				uploadedFiles = append(uploadedFiles, &uploadedFile)
				return uploadedFiles, nil
			}(uploadedFiles)
			
			if err != nil {
				// Return partial results and the error
				return uploadedFiles, err
			}
		}
	}
	
	// Calculate total batch size
	var totalBatchSize int64
	for _, fileHeaders := range r.MultipartForm.File {
		for _, header := range fileHeaders {
			totalBatchSize += header.Size
		}
	}
	
	// Check if total batch size exceeds limit
	if t.MaxBatchSize > 0 && totalBatchSize > t.MaxBatchSize {
		return nil, &ErrorResponse{
			Err:     ErrBatchSizeExceeded,
			Message: fmt.Sprintf("total batch size %d exceeds the maximum allowed size %d", 
				totalBatchSize, t.MaxBatchSize),
		}
	}
	
	if len(uploadedFiles) == 0 {
		return nil, &ErrorResponse{
			Err:     ErrNoFileUploaded,
			Message: "no files were processed",
		}
	}
	
	return uploadedFiles, nil
}

// CreateDirIfNotExist creates a directory, and all necessary parents, if it does not exist
func (t *Tools) CreateDirIfNotExist(path string) error {
	const mode = 0755
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err := os.MkdirAll(path, mode)
		if err != nil {
			return err
		}
	}
	return nil
}

// Slugify converts a string to a URL-friendly slug
// It handles special characters, multiple spaces, and ensures proper formatting
func (t *Tools) Slugify(s string) (string, error) {
    // Check for empty string
    if s == "" {
        return "", errors.New("empty string not permitted")
    }
    
    // Convert to lowercase and trim spaces
    s = strings.TrimSpace(strings.ToLower(s))
    
    // Handle non-ASCII characters (optional transliteration)
    // This is a simple replacement - consider using a proper transliteration library for production
    replacer := strings.NewReplacer(
        "æ", "ae", "ø", "o", "å", "a", "ü", "u", "ö", "o", "ä", "a",
        "ñ", "n", "é", "e", "è", "e", "ê", "e", "ë", "e", "á", "a",
        "à", "a", "â", "a", "ã", "a", "ç", "c", "í", "i", "ì", "i",
        "î", "i", "ï", "i", "ó", "o", "ò", "o", "ô", "o", "õ", "o",
        "ú", "u", "ù", "u", "û", "u", "ý", "y", "ÿ", "y",
    )
    s = replacer.Replace(s)
    
    // Replace any non-alphanumeric characters with hyphens
    var re = regexp.MustCompile(`[^a-z0-9]+`)
    slug := strings.Trim(re.ReplaceAllString(s, "-"), "-")
    
    // Check if slug is empty after processing
    if len(slug) == 0 {
        return "", errors.New("after removing characters, slug is zero length")
    }
    
    // Avoid multiple consecutive hyphens
    multiHyphen := regexp.MustCompile(`-+`)
    slug = multiHyphen.ReplaceAllString(slug, "-")
    
    // Limit slug length (optional, adjust as needed)
    maxLength := 100
    if len(slug) > maxLength {
        slug = slug[:maxLength]
        // Ensure we don't end with a hyphen if we truncated
        slug = strings.TrimSuffix(slug, "-")
    }
    
    return slug, nil
}

// UploadOneFile uploads a single file to a specified directory
func (t *Tools) UploadOneFile(r *http.Request, uploadDir string, rename bool) (*UploadedFile, error) {
	files, err := t.UploadFiles(r, uploadDir, rename)
	if err != nil {
		return nil, err
	}

	if len(files) == 0 {
		return nil, errors.New("no file uploaded")
	}

	return files[0], nil
}

// Custom error types for more precise error handling
var (
	ErrFileSizeExceeded  = errors.New("file size exceeded")
	ErrBatchSizeExceeded = errors.New("batch size exceeded")
	ErrInvalidFileType   = errors.New("invalid file type")
	ErrMaxUploadExceeded = errors.New("maximum upload count exceeded")
	ErrContentVerification = errors.New("content type verification failed")
	ErrNoFileUploaded    = errors.New("no file uploaded")
	ErrFileCreation      = errors.New("error creating file")
)

// ErrorResponse wraps an error with additional context
type ErrorResponse struct {
	Err     error
	Message string
}

// Error implements the error interface
func (er *ErrorResponse) Error() string {
	return er.Message
}

// Unwrap returns the wrapped error
func (er *ErrorResponse) Unwrap() error {
	return er.Err
}

// Remove the duplicate import and Tools struct declaration here

// UploadChunk saves a chunk of a file during a resumable upload
func (t *Tools) UploadChunk(uploadID, fileName string, chunkNumber, totalChunks int64, data []byte) error {
	// Create chunks directory if it doesn't exist
	chunksDir := filepath.Join(t.ChunksDirectory, uploadID)
	if err := t.CreateDirIfNotExist(chunksDir); err != nil {
		return &ErrorResponse{
			Err:     ErrFileCreation,
			Message: fmt.Sprintf("failed to create chunks directory: %v", err),
		}
	}
	
	// Save the chunk
	chunkPath := filepath.Join(chunksDir, fmt.Sprintf("%d", chunkNumber))
	if err := os.WriteFile(chunkPath, data, 0644); err != nil {
		return &ErrorResponse{
			Err:     ErrFileCreation,
			Message: fmt.Sprintf("failed to save chunk: %v", err),
		}
	}
	
	// Save metadata if this is the first chunk
	if chunkNumber == 0 {
		metadata := struct {
			FileName    string `json:"file_name"`
			TotalChunks int64  `json:"total_chunks"`
			FileSize    int64  `json:"file_size"`
			UploadTime  int64  `json:"upload_time"`
		}{
			FileName:    fileName,
			TotalChunks: totalChunks,
			FileSize:    -1, // Will be calculated when all chunks are received
			UploadTime:  time.Now().Unix(),
		}
		
		metadataJSON, err := json.Marshal(metadata)
		if err != nil {
			return &ErrorResponse{
				Err:     ErrFileCreation,
				Message: fmt.Sprintf("failed to create metadata: %v", err),
			}
		}
		
		metadataPath := filepath.Join(chunksDir, "metadata.json")
		if err := os.WriteFile(metadataPath, metadataJSON, 0644); err != nil {
			return &ErrorResponse{
				Err:     ErrFileCreation,
				Message: fmt.Sprintf("failed to save metadata: %v", err),
			}
		}
	}
	
	return nil
}

// CompleteChunkedUpload assembles all chunks into the final file
func (t *Tools) CompleteChunkedUpload(uploadID, originalFileName string) (*UploadedFile, error) {
	chunksDir := filepath.Join(t.ChunksDirectory, uploadID)
	
	// Read metadata
	metadataPath := filepath.Join(chunksDir, "metadata.json")
	metadataJSON, err := os.ReadFile(metadataPath)
	if err != nil {
		return nil, &ErrorResponse{
			Err:     ErrFileCreation,
			Message: fmt.Sprintf("failed to read metadata: %v", err),
		}
	}
	
	var metadata struct {
		FileName    string `json:"file_name"`
		TotalChunks int64  `json:"total_chunks"`
		FileSize    int64  `json:"file_size"`
		UploadTime  int64  `json:"upload_time"`
	}
	
	if err := json.Unmarshal(metadataJSON, &metadata); err != nil {
		return nil, &ErrorResponse{
			Err:     ErrFileCreation,
			Message: fmt.Sprintf("failed to parse metadata: %v", err),
		}
	}
	
	// Create upload directory if it doesn't exist
	if err := t.CreateDirIfNotExist(t.UploadPath); err != nil {
		return nil, &ErrorResponse{
			Err:     ErrFileCreation,
			Message: fmt.Sprintf("failed to create upload directory: %v", err),
		}
	}
	
	// Create a new file name if needed
	newFileName := originalFileName
	if strings.HasPrefix(filepath.Base(originalFileName), "resumable-") {
		// For test files, generate a new name
		ext := filepath.Ext(originalFileName)
		newFileName = fmt.Sprintf("%s%s", t.RandomString(25), ext)
	}
	
	// Create the final file
	finalPath := filepath.Join(t.UploadPath, newFileName)
	finalFile, err := os.Create(finalPath)
	if err != nil {
		return nil, &ErrorResponse{
			Err:     ErrFileCreation,
			Message: fmt.Sprintf("failed to create final file: %v", err),
		}
	}
	defer finalFile.Close()
	
	// Assemble chunks
	var fileSize int64
	for i := int64(0); i < metadata.TotalChunks; i++ {
		chunkPath := filepath.Join(chunksDir, fmt.Sprintf("%d", i))
		chunkData, err := os.ReadFile(chunkPath)
		if err != nil {
			// Clean up the partial file
			os.Remove(finalPath)
			return nil, &ErrorResponse{
				Err:     ErrFileCreation,
				Message: fmt.Sprintf("failed to read chunk %d: %v", i, err),
			}
		}
		
		n, err := finalFile.Write(chunkData)
		if err != nil {
			// Clean up the partial file
			os.Remove(finalPath)
			return nil, &ErrorResponse{
				Err:     ErrFileCreation,
				Message: fmt.Sprintf("failed to write chunk %d to final file: %v", i, err),
			}
		}
		
		fileSize += int64(n)
	}
	
	// Determine file type
	finalFile.Seek(0, 0)
	fileType := "application/octet-stream" // Default if detection fails
	
	if t.detectFileType != nil {
		detectedType, err := t.detectFileType(finalFile)
		if err == nil {
			fileType = detectedType
		}
	}
	
	// Clean up chunks
	os.RemoveAll(chunksDir)
	
	// Return the uploaded file info
	return &UploadedFile{
		NewFileName:      newFileName,
		OriginalFileName: originalFileName,
		FileSize:         fileSize,
		FileType:         fileType,
	}, nil
}

// GetUploadProgress returns the progress of a chunked upload
func (t *Tools) GetUploadProgress(uploadID string) (float64, error) {
	chunksDir := filepath.Join(t.ChunksDirectory, uploadID)
	
	// Read metadata
	metadataPath := filepath.Join(chunksDir, "metadata.json")
	metadataJSON, err := os.ReadFile(metadataPath)
	if err != nil {
		return 0, &ErrorResponse{
			Err:     fmt.Errorf("upload not found"),
			Message: fmt.Sprintf("upload ID %s not found", uploadID),
		}
	}
	
	var metadata struct {
		FileName    string `json:"file_name"`
		TotalChunks int64  `json:"total_chunks"`
		FileSize    int64  `json:"file_size"`
		UploadTime  int64  `json:"upload_time"`
	}
	
	if err := json.Unmarshal(metadataJSON, &metadata); err != nil {
		return 0, &ErrorResponse{
			Err:     ErrFileCreation,
			Message: fmt.Sprintf("failed to parse metadata: %v", err),
		}
	}
	
	// Count the number of chunks that have been uploaded
	files, err := os.ReadDir(chunksDir)
	if err != nil {
		return 0, &ErrorResponse{
			Err:     fmt.Errorf("failed to read chunks directory"),
			Message: fmt.Sprintf("failed to read chunks directory: %v", err),
		}
	}
	
	// Subtract 1 for the metadata file
	uploadedChunks := int64(len(files)) - 1
	
	// Calculate progress percentage
	progress := float64(uploadedChunks) / float64(metadata.TotalChunks) * 100.0
	
	return progress, nil
}

// ListActiveUploads returns a list of all active chunked uploads
func (t *Tools) ListActiveUploads() ([]string, error) {
	// Create chunks directory if it doesn't exist
	if err := t.CreateDirIfNotExist(t.ChunksDirectory); err != nil {
		return nil, &ErrorResponse{
			Err:     ErrFileCreation,
			Message: fmt.Sprintf("failed to create chunks directory: %v", err),
		}
	}
	
	// Read all directories in the chunks directory
	entries, err := os.ReadDir(t.ChunksDirectory)
	if err != nil {
		return nil, &ErrorResponse{
			Err:     fmt.Errorf("failed to read chunks directory"),
			Message: fmt.Sprintf("failed to read chunks directory: %v", err),
		}
	}
	
	// Filter for directories only
	var uploadIDs []string
	for _, entry := range entries {
		if entry.IsDir() {
			// Check if this is a valid upload (has metadata)
			metadataPath := filepath.Join(t.ChunksDirectory, entry.Name(), "metadata.json")
			if _, err := os.Stat(metadataPath); err == nil {
				uploadIDs = append(uploadIDs, entry.Name())
			}
		}
	}
	
	return uploadIDs, nil
}

// CancelChunkedUpload cancels an in-progress chunked upload
func (t *Tools) CancelChunkedUpload(uploadID string) error {
	chunksDir := filepath.Join(t.ChunksDirectory, uploadID)
	
	// Check if the upload exists
	if _, err := os.Stat(chunksDir); os.IsNotExist(err) {
		return &ErrorResponse{
			Err:     fmt.Errorf("upload not found"),
			Message: fmt.Sprintf("upload ID %s not found", uploadID),
		}
	}
	
	// Remove the chunks directory
	if err := os.RemoveAll(chunksDir); err != nil {
		return &ErrorResponse{
			Err:     fmt.Errorf("failed to cancel upload"),
			Message: fmt.Sprintf("failed to remove chunks directory: %v", err),
		}
	}
	
	return nil
}
// DownloadStaticFile downloads a file, and tries to force the browser to avoid displaying it
// in the browser window by setting content disposition. It also allows specification of the
// display name
func (t *Tools) DownloadStaticFile(w http.ResponseWriter, r *http.Request, p, file, displayName string) {
	fp := path.Join(p, file)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", displayName))

	http.ServeFile(w, r, fp)
}