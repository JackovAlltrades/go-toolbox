package toolbox

import (
	"errors"
	"io"
	"mime/multipart"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

// TestTools_FileTypeValidation tests the file type validation functionality
func TestTools_FileTypeValidation(t *testing.T) {
	tests := []struct {
		name              string
		fileType          string
		allowedTypes      []string
		allowUnknownTypes bool
		shouldBeAllowed   bool
		expectedError     error
	}{
		{
			name:              "allowed type",
			fileType:          "image/jpeg",
			allowedTypes:      []string{"image/jpeg", "image/png"},
			allowUnknownTypes: false,
			shouldBeAllowed:   true,
		},
		{
			name:              "disallowed type",
			fileType:          "application/pdf",
			allowedTypes:      []string{"image/jpeg", "image/png"},
			allowUnknownTypes: false,
			shouldBeAllowed:   false,
			expectedError:     ErrInvalidFileType,
		},
		{
			name:              "unknown type allowed",
			fileType:          "application/octet-stream",
			allowedTypes:      []string{"image/jpeg", "image/png"},
			allowUnknownTypes: true,
			shouldBeAllowed:   true,
		},
		{
			name:              "no types specified",
			fileType:          "application/pdf",
			allowedTypes:      []string{},
			allowUnknownTypes: false,
			shouldBeAllowed:   true, // Default behavior is to allow all if no types specified
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create a Tools instance with the test configuration
			tools := Tools{
				AllowedFileTypes:  tc.allowedTypes,
				AllowUnknownTypes: tc.allowUnknownTypes,
			}

			// Create a mock file for testing
			pr, pw := io.Pipe()
			writer := multipart.NewWriter(pw)
			
			go func() {
				defer writer.Close()
				part, _ := writer.CreateFormFile("file", "test.dat")
				part.Write([]byte("test data"))
			}()
			
			request := httptest.NewRequest("POST", "/", pr)
			request.Header.Add("Content-Type", writer.FormDataContentType())
			
			// Mock the file type detection
			tools.detectFileType = func(file multipart.File) (string, error) {
				return tc.fileType, nil
			}
			
			// Try to upload the file
			_, err := tools.UploadFiles(request, "./testdata/uploads/", true)
			
			// Check if the result matches expectations
			if tc.shouldBeAllowed && err != nil && strings.Contains(err.Error(), "not permitted") {
				t.Errorf("expected file type to be allowed, but got error: %s", err.Error())
			}
			
			if !tc.shouldBeAllowed && (err == nil || !strings.Contains(err.Error(), "not permitted")) {
				t.Errorf("expected file type to be rejected, but it was allowed")
			}
			
			// Check for specific error type if applicable
			if err != nil && tc.expectedError != nil {
				var errResp *ErrorResponse
				if errors.As(err, &errResp) {
					if !errors.Is(errResp.Err, tc.expectedError) {
						t.Errorf("expected error %v, got: %v", tc.expectedError, errResp.Err)
					}
				}
			}
		})
	}
}