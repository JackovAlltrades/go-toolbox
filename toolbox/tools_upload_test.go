package toolbox_test_test

import (
	toolbox "github.com/JackovAlltrades/go-toolbox"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

// TestTools_UploadFiles tests the basic file upload functionality
func TestTools_UploadFilesExtended(t *testing.T) {
	// Create the uploads directory if it doesn't exist
	setupTestDir(t, "./testdata/uploads/")
	defer cleanupTestDir(t, "./testdata/uploads/")
	
	// Test cases
	tests := []struct {
		name          string
		allowedTypes  []string
		renameFile    bool
		errorExpected bool
	}{
		{
			name:          "allowed no rename",
			allowedTypes:  []string{"text/plain"},
			renameFile:    false,
			errorExpected: false,
		},
		{
			name:          "allowed rename",
			allowedTypes:  []string{"text/plain"},
			renameFile:    true,
			errorExpected: false,
		},
		{
			name:          "not allowed",
			allowedTypes:  []string{"image/jpeg"},
			renameFile:    false,
			errorExpected: true,
		},
	}
	
	// Run tests
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create a test request with a file
			pr, pw := io.Pipe()
			writer := multipart.NewWriter(pw)
			
			go func() {
				defer writer.Close()
				
				part, err := writer.CreateFormFile("file", "test.txt")
				if err != nil {
					t.Error(err)
					return
				}
				
				part.Write([]byte("This is a test file"))
			}()
			
			// Create the request
			request := httptest.NewRequest("POST", "/", pr)
			request.Header.Add("Content-Type", writer.FormDataContentType())
			
			// Configure the tools
			tools := tools.Tools{
				MaxFileSize:      10 * 1024 * 1024, // 10MB
				AllowedFileTypes: tc.allowedTypes,
			}
			
			// Attempt to upload the file
			_, err := tools.UploadFiles(request, "./testdata/uploads/", tc.renameFile)
			
			// Check error expectations
			if err != nil && !tc.errorExpected {
				t.Errorf("got unexpected error: %s", err.Error())
			}
			
			if err == nil && tc.errorExpected {
				t.Error("expected error but got none")
			}
		})
	}
}

// TestTools_UploadOneFile tests the UploadOneFile function
func TestTools_UploadOneFile(t *testing.T) {
	// Create the uploads directory if it doesn't exist
	setupTestDir(t, "./testdata/uploads/")
	defer cleanupTestDir(t, "./testdata/uploads/")
	
	// Create a test request with a file
	pr, pw := io.Pipe()
	writer := multipart.NewWriter(pw)
	
	go func() {
		defer writer.Close()
		
		part, err := writer.CreateFormFile("file", "test.txt")
		if err != nil {
			t.Error(err)
			return
		}
		
		part.Write([]byte("This is a test file"))
	}()
	
	// Create the request
	request := httptest.NewRequest("POST", "/", pr)
	request.Header.Add("Content-Type", writer.FormDataContentType())
	
	// Configure the tools
	tools := tools.Tools{
		MaxFileSize:      10 * 1024 * 1024, // 10MB
		AllowedFileTypes: []string{"text/plain"},
	}
	
	// Attempt to upload the file
	file, err := UploadOneFile(request, "./testdata/uploads/", true)
	
	// Check error expectations
	if err != nil {
		t.Errorf("got unexpected error: %s", err.Error())
	}
	
	if file == nil {
		t.Error("expected file but got nil")
	} else {
		// Clean up
		os.Remove(fmt.Sprintf("./testdata/uploads/%s", file.NewFileName))
	}
}




















