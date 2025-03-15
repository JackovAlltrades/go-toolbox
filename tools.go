// Add these imports at the top of your file if they're not already there
import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

// VerifyFileContent checks if the file content matches the claimed content type
// Returns true if content matches or verification is not required, false otherwise
func (t *Tools) VerifyFileContent(file *os.File, claimedType string) (bool, error) {
	// Read the first 512 bytes to detect content type
	buffer := make([]byte, 512)
	_, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		return false, fmt.Errorf("error reading file for content verification: %w", err)
	}
	
	// Reset file pointer to beginning
	_, err = file.Seek(0, 0)
	if err != nil {
		return false, fmt.Errorf("error resetting file pointer: %w", err)
	}
	
	// Detect content type
	detectedType := http.DetectContentType(buffer)
	
	// Special case for text files
	if claimedType == "text/plain" {
		// Check if file contains binary data that's not typical for text files
		if strings.HasPrefix(detectedType, "application/octet-stream") {
			// Further check for executable headers
			if len(buffer) >= 2 && buffer[0] == 0x4D && buffer[1] == 0x5A { // MZ header (Windows executable)
				return false, nil
			}
			if len(buffer) >= 4 && buffer[0] == 0x7F && buffer[1] == 0x45 && buffer[2] == 0x4C && buffer[3] == 0x46 { // ELF header (Linux executable)
				return false, nil
			}
		}
	}
	
	// For images, PDFs, etc., check if detected type matches claimed type
	if !strings.HasPrefix(detectedType, claimedType) && detectedType != "application/octet-stream" {
		// Some flexibility for text types
		if claimedType == "text/plain" && (strings.HasPrefix(detectedType, "text/") || detectedType == "application/xml") {
			return true, nil
		}
		return false, nil
	}
	
	return true, nil
}

// In your UploadFiles method, after validating file type but before saving the file
// Validate file type
if !t.isAllowedFileType(fileType) {
    return nil, errors.New("the uploaded file type is not permitted")
}

// Save the file to disk temporarily for content verification
tempFile, err := os.CreateTemp(t.TempFilePath, "verify-*")
if err != nil {
    return nil, fmt.Errorf("error creating temp file: %w", err)
}
defer os.Remove(tempFile.Name())
defer tempFile.Close()

_, err = io.Copy(tempFile, file)
if err != nil {
    return nil, fmt.Errorf("error saving temp file: %w", err)
}

// Verify file content matches claimed type
isValid, err := t.VerifyFileContent(tempFile, fileType)
if err != nil {
    return nil, fmt.Errorf("error during content verification: %w", err)
}
if !isValid {
    return nil, errors.New("content type verification failed")
}

// Reset file pointer for the multipart file
file.Seek(0, 0)

// Continue with the rest of your upload logic