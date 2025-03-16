package toolbox_test_test

import (
	toolbox "github.com/JackovAlltrades/go-toolbox"
	"fmt"
	"image"
	"image/png"
	"io"
	"mime/multipart"
	// // // // "net/http" // Unused import // Unused import // Unused import // Unused import
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
	// // // // "path/filepath" // Unused import // Unused import // Unused import // Unused import
)

// Define uploadTests with the allowUnknownTypes field
var uploadTests = []struct {
	name              string
	allowedTypes      []string
	renameFile        bool
	errorExpected     bool
	allowUnknownTypes bool
}{
	{name: "allowed no rename", allowedTypes: []string{"image/jpeg", "image/png"}, renameFile: false, errorExpected: false, allowUnknownTypes: false},
	{name: "allowed rename", allowedTypes: []string{"image/jpeg", "image/png"}, renameFile: true, errorExpected: false, allowUnknownTypes: false},
	{name: "not allowed", allowedTypes: []string{"image/jpeg"}, renameFile: false, errorExpected: true, allowUnknownTypes: false},
	{name: "unknown allowed", allowedTypes: []string{}, renameFile: true, errorExpected: false, allowUnknownTypes: true},
	{name: "all common types", allowedTypes: []string{
		"image/jpeg", "image/png", "image/gif", "application/pdf", "text/plain",
	}, renameFile: true, errorExpected: false, allowUnknownTypes: false},
}

// Fix the paths in the TestTools_UploadFiles function
func TestTools_tools.UploadFiles(t *testing.T) {
	// Create the uploads directory if it doesn't exist
	err := os.MkdirAll("./testdata/uploads/", os.ModePerm)
	if err != nil {
		t.Fatal(err)
	}

	for _, e := range uploadTests {
		t.Run(e.name, func(t *testing.T) {
			// set up a pipe to avoid buffering
			pr, pw := io.Pipe()
			writer := multipart.NewWriter(pw)
			wg := sync.WaitGroup{}
			wg.Add(1)

			go func() {
				defer writer.Close()
				defer wg.Done()

				/// create the form data field 'file'
				part, err := writer.CreateFormFile("file", "./testdata/img.png")
				if err != nil {
					t.Error(err)
				}

				f, err := os.Open("./testdata/img.png")
				if err != nil {
					t.Error(err)
				}
				defer f.Close()

				img, _, err := image.Decode(f)
				if err != nil {
					t.Error("error decoding image", err)
				}

				err = png.Encode(part, img)
				if err != nil {
					t.Error(err)
				}
			}()

			// read from the pipe which receives data
			request := httptest.NewRequest("POST", "/", pr)
			request.Header.Add("Content-Type", writer.FormDataContentType())

			var testTools Tools
			testAllowedFileTypes = e.allowedTypes
			testMaxFileSize = 1024 * 1024 // 1MB for testing
			testAllowUnknownTypes = e.allowUnknownTypes // Use the new field

			uploadedFiles, err := testtools.UploadFiles(request, "./testdata/uploads/", e.renameFile)
			
			// Check error expectations
			if err != nil && !e.errorExpected {
				t.Errorf("got error when none expected: %s", err.Error())
			}

			if err == nil && e.errorExpected {
				t.Error("did not get error when one expected")
			}

			// Check file creation for successful uploads
			if !e.errorExpected && err == nil {
				if len(uploadedFiles) == 0 {
					t.Error("no files were returned")
				} else {
					if _, err := os.Stat(fmt.Sprintf("./testdata/uploads/%s", uploadedFiles[0].NewFileName)); os.IsNotExist(err) {
						t.Errorf("%s: expected file to exist: %s", e.name, err.Error())
					}

					// Check file properties
					if uploadedFiles[0].FileSize == 0 {
						t.Error("returned file has zero size")
					}

					// Check FileType field
					if uploadedFiles[0].FileType == "" {
						t.Error("file type not detected")
					}

					// Verify correct file type detection
					if !strings.Contains(uploadedFiles[0].FileType, "image/") {
						t.Errorf("expected image file type, got %s", uploadedFiles[0].FileType)
					}

					// Check filename behavior
					if !e.renameFile && uploadedFiles[0].NewFileName != uploadedFiles[0].OriginalFileName {
						t.Error("filename should not have been changed")
					}

					if e.renameFile && uploadedFiles[0].NewFileName == uploadedFiles[0].OriginalFileName {
						t.Error("filename should have been changed")
					}

					// clean up
					_ = os.Remove(fmt.Sprintf("./testdata/uploads/%s", uploadedFiles[0].NewFileName))
				}
			}

			wg.Wait()
		})
	}
}










// Tests moved from tools_filetype_test.go

// TestTools_FileTypeValidation tests the file type validation functionality
func TestTools_FileTypeValidation(t *testing.T) {
	// Create test directory
	testDir := "./testdata/uploads"
	err := os.MkdirAll(testDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	// Clean up after test
	defer os.RemoveAll(testDir)

	tests := []struct {
		name              string
		fileType          string
		allowedTypes      []string
		allowUnknownTypes bool
		shouldBeAllowed   bool
		expectedError     error
	}{
		// Test cases remain the same
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
			tools := tools.Tools{
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
			_, err := tools.UploadFiles(request, testDir, true)
			
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




















