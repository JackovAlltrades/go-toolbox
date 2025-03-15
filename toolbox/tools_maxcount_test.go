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

// TestTools_MaxUploadCount tests the enforcement of the maximum upload count limit
func TestTools_MaxUploadCount(t *testing.T) {
	// Create the uploads directory if it doesn't exist
	setupTestDir(t, "./testdata/uploads/")
	defer cleanupTestDir(t, "./testdata/uploads/")

	tests := []struct {
		name           string
		maxUploadCount int
		fileCount      int
		errorExpected  bool
		expectedError  error
	}{
		{
			name:           "within upload count limit",
			maxUploadCount: 5,
			fileCount:      3,
			errorExpected:  false,
		},
		{
			name:           "at upload count limit",
			maxUploadCount: 3,
			fileCount:      3,
			errorExpected:  false,
		},
		{
			name:           "exceeds upload count limit",
			maxUploadCount: 2,
			fileCount:      4,
			errorExpected:  true,
			expectedError:  ErrMaxUploadExceeded,
		},
		{
			name:           "zero upload count limit",
			maxUploadCount: 0, // Should allow any number of files
			fileCount:      5,
			errorExpected:  false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create a test request with multiple files
			pr, pw := io.Pipe()
			writer := multipart.NewWriter(pw)
			
			go func() {
				defer writer.Close()
				
				// Create multiple files
				for i := 0; i < tc.fileCount; i++ {
					part, err := writer.CreateFormFile("file", fmt.Sprintf("testfile%d.txt", i))
					if err != nil {
						t.Error(err)
						return
					}
					
					// Write some dummy data
					part.Write([]byte(fmt.Sprintf("test content for file %d", i)))
				}
			}()
			
			// Create the request
			request := httptest.NewRequest("POST", "/", pr)
			request.Header.Add("Content-Type", writer.FormDataContentType())
			
			// Configure the tools with max upload count limit
			tools := Tools{
				MaxFileSize:      10 * 1024 * 1024, // 10MB
				MaxUploadCount:   tc.maxUploadCount,
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
				if len(files) != tc.fileCount {
					t.Errorf("expected %d files, got %d", tc.fileCount, len(files))
				}
				
				// Clean up any created files
				for _, f := range files {
					os.Remove(fmt.Sprintf("./testdata/uploads/%s", f.NewFileName))
				}
			}
		})
	}
}