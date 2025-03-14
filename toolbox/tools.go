package toolbox

import (
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
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
	ChunkSize              int
	ValidationCallback     func(file *UploadedFile) error
	MaxBatchSize          int64 // Maximum total size of all files in a batch
	
	// For testing purposes - allows mocking the file type detection
	detectFileType func(file multipart.File) (string, error)
}

// Modify the UploadFiles method to check batch size
func (t *Tools) UploadFiles(r *http.Request, uploadDir string, rename bool) ([]*UploadedFile, error) {
	// Initialize defaults if not set
	t.InitDefaults()
	
	renameFile := true
	if len(rename) > 0 {
		renameFile = rename[0]
	}

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

	// Check if the number of files exceeds the maximum allowed
	if t.MaxUploadCount > 0 && fileCount > t.MaxUploadCount {
		return nil, fmt.Errorf("number of files (%d) exceeds the maximum allowed (%d)", fileCount, t.MaxUploadCount)
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
				fileType := http.DetectContentType(buff)
				uploadedFile.FileType = fileType
				
				// Check if file type is allowed
				allowed := t.AllowUnknownTypes // Allow if AllowUnknownTypes is true
				if !allowed && len(t.AllowedFileTypes) > 0 {
					for _, x := range t.AllowedFileTypes {
						if strings.EqualFold(fileType, x) {
							allowed = true
							break
						}
					}
				} else if len(t.AllowedFileTypes) == 0 {
					allowed = true // If no restrictions specified, allow all
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
		return nil, fmt.Errorf("total batch size %d exceeds the maximum allowed size %d", 
			totalBatchSize, t.MaxBatchSize)
	}
	
	if len(uploadedFiles) == 0 {
		return nil, errors.New("no files were processed")
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
