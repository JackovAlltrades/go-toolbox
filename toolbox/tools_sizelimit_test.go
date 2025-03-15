package toolbox

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http/httptest"
	"os"
	"testing"
)

// TestTools_BatchSizeLimits tests the batch size limit functionality
func TestTools_BatchSizeLimits(t *testing.T) {
	// Create the uploads directory if it doesn't exist
	setupTestDir(t, "./testdata/uploads/")
	defer cleanupTestDir(t, "./testdata/uploads/")

	tests := []struct {
		name           string
		maxBatchSize   int64
		fileSizes      []int64
		errorExpected  bool
		expectedError  error
	}{
		{
			name:          "within batch limit",
			maxBatchSize:  5 * 1024 * 1024, // 5MB
			fileSizes:     []int64{1 * 1024 * 1024, 2 * 1024 * 1024}, // 3MB total
			errorExpected: false,
		},
		{
			name:          "exceeds batch limit",
			maxBatchSize:  3 * 1024 * 1024, // 3MB
			fileSizes:     []int64{2 * 1024 * 1024, 2 * 1024 * 1024}, // 4MB total
			errorExpected: true,
			expectedError: ErrBatchSizeExceeded,
		},
		{
			name:          "at batch limit",
			maxBatchSize:  4 * 1024 * 1024, // 4MB
			fileSizes:     []int64{2 * 1024 * 1024, 2 * 1024 * 1024}, // 4MB total
			errorExpected: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create a test request with multiple files
			pr, pw := io.Pipe()
			writer := multipart.NewWriter(pw)
			
			go func() {
				defer writer.Close()
				
				// Create multiple files with specified sizes
				for i, size := range tc.fileSizes {
					part, err := writer.CreateFormFile("file", fmt.Sprintf("testfile%d.txt", i))
					if err != nil {
						t.Error(err)
						return
					}
					
					// Write dummy data of specified size
					data := make([]byte, size)
					if _, err := part.Write(data); err != nil {
						t.Error(err)
						return
					}
				}
			}()
			
			// Create the request
			request := httptest.NewRequest("POST", "/", pr)
			request.Header.Add("Content-Type", writer.FormDataContentType())
			
			// Configure the tools with batch size limit
			tools := Tools{
				MaxFileSize:      10 * 1024 * 1024, // 10MB (larger than any individual file)
				MaxBatchSize:     tc.maxBatchSize,
				UploadPath:       "./testdata/uploads/",
				AllowedFileTypes: []string{"text/plain"},
				AllowUnknownTypes: true,
			}
			
			// Attempt to upload the files
			files, err := tools.UploadFiles(request, "./testdata/uploads/", true)
			
			// Check error expectations
			if err != nil && !tc.errorExpected {
				t.Errorf("got unexpected error: %s", err.Error())
			}
			
			if err == nil && tc.errorExpected {
				t.Error("expected error but got none")
			}
			
			// Check for specific error type
			if err != nil && tc.errorExpected && tc.expectedError != nil {
				var errResp *ErrorResponse
				if errors.As(err, &errResp) {
					if !errors.Is(errResp.Err, tc.expectedError) {
						t.Errorf("expected error %v, got: %v", tc.expectedError, errResp.Err)
					}
				} else {
					t.Errorf("expected ErrorResponse type, got: %T", err)
				}
			}
			
			// Verify file count if no error
			if err == nil {
				if len(files) != len(tc.fileSizes) {
					t.Errorf("expected %d files, got %d", len(tc.fileSizes), len(files))
				}
				
				// Clean up any created files
				for _, f := range files {
					os.Remove(fmt.Sprintf("./testdata/uploads/%s", f.NewFileName))
				}
			}
		})
	}
}

// TestTools_TypeSpecificSizeLimits tests the type-specific size limit functionality
func TestTools_TypeSpecificSizeLimits(t *testing.T) {
	// Create the uploads directory if it doesn't exist
	setupTestDir(t, "./testdata/uploads/")
	defer cleanupTestDir(t, "./testdata/uploads/")
	
	tests := []struct {
		name              string
		fileType          string
		fileSize          int64
		typeSizeLimits    map[string]int
		errorExpected     bool
		expectedError     error
	}{
		{
			name:           "within type limit",
			fileType:       "image/png",
			fileSize:       500 * 1024, // 500KB
			typeSizeLimits: map[string]int{"image/png": 1 * 1024 * 1024}, // 1MB limit
			errorExpected:  false,
		},
		{
			name:           "exceeds type limit",
			fileType:       "image/jpeg",
			fileSize:       2 * 1024 * 1024, // 2MB
			typeSizeLimits: map[string]int{"image/jpeg": 1 * 1024 * 1024}, // 1MB limit
			errorExpected:  true,
			expectedError:  ErrFileSizeExceeded,
		},
		{
			name:           "no type limit specified",
			fileType:       "application/pdf",
			fileSize:       3 * 1024 * 1024, // 3MB
			typeSizeLimits: map[string]int{"image/jpeg": 1 * 1024 * 1024}, // No limit for PDF
			errorExpected:  false,
		},
	}
	
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create a unique test directory for this test case
			testDir := fmt.Sprintf("./testdata/uploads/%s", tc.name)
			setupTestDir(t, testDir)
			defer cleanupTestDir(t, testDir)
			
			// Create a test request with a file
			pr, pw := io.Pipe()
			writer := multipart.NewWriter(pw)
			
			go func() {
				defer writer.Close()
				
				part, err := writer.CreateFormFile("file", "test.dat")
				if err != nil {
					t.Error(err)
					return
				}
				
				// Write dummy data of specified size
				data := make([]byte, tc.fileSize)
				if _, err := part.Write(data); err != nil {
					t.Error(err)
					return
				}
			}()
			
			// Create the request
			request := httptest.NewRequest("POST", "/", pr)
			request.Header.Add("Content-Type", writer.FormDataContentType())
			
			// Configure the tools with type-specific size limits
			tools := Tools{
				MaxFileSize:          10 * 1024 * 1024, // 10MB (larger than any individual file)
				TypeSpecificSizeLimits: tc.typeSizeLimits,
				UploadPath:           "./testdata/uploads/",
				AllowedFileTypes:     []string{tc.fileType},
				AllowUnknownTypes:    true,
			}
			
			// Mock the file type detection
			tools.detectFileType = func(file multipart.File) (string, error) {
				return tc.fileType, nil
			}
			
			// Attempt to upload the file
			files, err := tools.UploadFiles(request, testDir, true)
			
			// Check error expectations
			if err != nil && !tc.errorExpected {
				t.Errorf("got unexpected error: %s", err.Error())
			}
			
			if err == nil && tc.errorExpected {
				t.Error("expected error but got none")
			}
			
			// Check for specific error type
			if err != nil && tc.errorExpected && tc.expectedError != nil {
				var errResp *ErrorResponse
				if errors.As(err, &errResp) {
					if !errors.Is(errResp.Err, tc.expectedError) {
						t.Errorf("expected error %v, got: %v", tc.expectedError, errResp.Err)
					}
				} else {
					t.Errorf("expected ErrorResponse type, got: %T", err)
				}
			}
			
			// Add cleanup for any created files
			if err == nil && files != nil {
				for _, f := range files {
					os.Remove(fmt.Sprintf("%s/%s", testDir, f.NewFileName))
				}
			}
		})
	}
}