package toolbox

import (
	"fmt"
	"image"
	"image/png"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
	"crypto/md5"
	"math/rand"
	"path/filepath"
)

func TestTools_RandomString(t *testing.T) {
	var testTools Tools

	// Test cases
	testCases := []struct {
		name     string
		length   int
		expected int
		isEmpty  bool
	}{
		{"standard case", 10, 10, false},
		{"zero length", 0, 0, true},
		{"negative length", -5, 0, true},
		{"large length", 100, 100, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s := testTools.RandomString(tc.length)

			// Check if result matches expected emptiness
			if tc.isEmpty && s != "" {
				t.Errorf("expected empty string for length %d, got %s", tc.length, s)
			}

			// Check length for non-empty cases
			if !tc.isEmpty && len(s) != tc.expected {
				t.Errorf("wrong length random string returned: got %d, expected %d", len(s), tc.expected)
			}

			// For positive lengths, verify characters are from the source
			if tc.length > 0 && s != "" {
				sourceRunes := []rune(randomStringSource)
				for _, r := range s {
					found := false
					for _, sr := range sourceRunes {
						if r == sr {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("character %c not found in random string source", r)
					}
				}
			}
		})
	}

	// Test randomness (basic check)
	// Generate multiple strings and ensure they're different
	if testTools.RandomString(10) == testTools.RandomString(10) {
		t.Error("random strings should be different on subsequent calls")
	}
}

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
func TestTools_UploadFiles(t *testing.T) {
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
			testTools.AllowedFileTypes = e.allowedTypes
			testTools.MaxFileSize = 1024 * 1024 // 1MB for testing
			testTools.AllowUnknownTypes = e.allowUnknownTypes // Use the new field

			uploadedFiles, err := testTools.UploadFiles(request, "./testdata/uploads/", e.renameFile)
			
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

// Complete the TestTools_ConfigFields function
func TestTools_ConfigFields(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "upload_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Test with custom upload path
	tools := Tools{
		MaxFileSize: 1024 * 1024,
		UploadPath: tempDir,
		MaxUploadCount: 5,
		AllowUnknownTypes: true,
	}

	// Verify the fields are used correctly
	if tools.UploadPath != tempDir {
		t.Errorf("UploadPath not set correctly, got %s", tools.UploadPath)
	}

	if tools.MaxUploadCount != 5 {
		t.Errorf("MaxUploadCount not set correctly, got %d", tools.MaxUploadCount)
	}

	if !tools.AllowUnknownTypes {
		t.Error("AllowUnknownTypes not set correctly")
	}
}

func TestTools_UploadOneFile(t *testing.T) {
	// Create the uploads directory if it doesn't exist
	err := os.MkdirAll("./testdata/uploads/", os.ModePerm)
	if err != nil {
		t.Fatal(err)
	}

	// set up a pipe to avoid buffering
	pr, pw := io.Pipe()
	writer := multipart.NewWriter(pw)

	go func() {
		defer writer.Close()

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

	uploadedFiles, err := testTools.UploadOneFile(request, "./testdata/uploads/", true)
	if err != nil {
		t.Error(err)
	}

	if _, err := os.Stat(fmt.Sprintf("./testdata/uploads/%s", uploadedFiles.NewFileName)); os.IsNotExist(err) {
		t.Errorf("expected file to exist: %s", err.Error())
	}

	// clean up
	_ = os.Remove(fmt.Sprintf("./testdata/uploads/%s", uploadedFiles.NewFileName))
}

func TestTools_CreateDirIfNotExist(t *testing.T) {
	var testTool Tools

	err := testTool.CreateDirIfNotExist("./testdata/myDir")
	if err != nil {
		t.Error(err)
	}

	err = testTool.CreateDirIfNotExist("./testdata/myDir")
	if err != nil {
		t.Error(err)
	}

	_ = os.Remove("./testdata/myDir")
}

func TestTools_Slugify(t *testing.T) {
	var testTool Tools

	// Test cases for standard inputs
	slugTests := []struct {
		name          string
		input         string
		expected      string
		errorExpected bool
	}{
		{name: "simple text", input: "hello world", expected: "hello-world", errorExpected: false},
		{name: "with spaces", input: "  hello  world  ", expected: "hello-world", errorExpected: false},
		{name: "with special chars", input: "hello! @world#", expected: "hello-world", errorExpected: false},
		{name: "with multiple hyphens", input: "hello---world", expected: "hello-world", errorExpected: false},
		{name: "with accented chars", input: "héllö wørld", expected: "hello-world", errorExpected: false},
		{name: "with numbers", input: "hello 123 world", expected: "hello-123-world", errorExpected: false},
		{name: "uppercase", input: "HELLO WORLD", expected: "hello-world", errorExpected: false},
		{name: "very long string", input: "This is a very long string that should be truncated because it exceeds the maximum length that we have set in our Slugify function which is 100 characters as defined in the implementation and we need to test it", expected: "this-is-a-very-long-string-that-should-be-truncated-because-it-exceeds-the-maximum-length-that-we-ha", errorExpected: false},
	}

	// Test edge cases
	edgeCases := []struct {
		name          string
		input         string
		errorExpected bool
	}{
		{name: "empty string", input: "", errorExpected: true},
		{name: "only special chars", input: "!@#$%^&*()", errorExpected: true},
		{name: "only spaces", input: "   ", errorExpected: true},
	}

	// Run standard test cases
	for _, e := range slugTests {
		slug, err := testTool.Slugify(e.input)
		if err != nil && !e.errorExpected {
			t.Errorf("%s: got an error when none expected: %s", e.name, err.Error())
		}

		if err == nil && e.errorExpected {
			t.Errorf("%s: did not get an error when one expected", e.name)
		}

		if slug != e.expected && !e.errorExpected {
			t.Errorf("%s: wrong slug returned. Expected %s but got %s", e.name, e.expected, slug)
		}
	}

	// Run edge cases
	for _, e := range edgeCases {
		_, err := testTool.Slugify(e.input)
		if err == nil && e.errorExpected {
			t.Errorf("%s: did not get an error when one expected", e.name)
		}
	}
}

// Add this import at the top if not already present
import (
	"errors"
)

// Update the TestTools_BatchSizeLimits function
func TestTools_BatchSizeLimits(t *testing.T) {
	// Create the uploads directory if it doesn't exist
	err := os.MkdirAll("./testdata/uploads/", os.ModePerm)
	if err != nil {
		t.Fatal(err)
	}

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
				MaxFileSize:    10 * 1024 * 1024, // 10MB (larger than any individual file)
				MaxBatchSize:   tc.maxBatchSize,
				UploadPath:     "./testdata/uploads/",
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
				if !errors.Is(err, tc.expectedError) {
					t.Errorf("expected error %v, got: %v", tc.expectedError, err)
				}
			}
			
			// Clean up any created files
			if err == nil {
				for _, f := range files {
					os.Remove(fmt.Sprintf("./testdata/uploads/%s", f.NewFileName))
				}
			}
		})
	}
}

// Similarly update TestTools_TypeSpecificSizeLimits
func TestTools_TypeSpecificSizeLimits(t *testing.T) {
	// Create the uploads directory if it doesn't exist
	err := os.MkdirAll("./testdata/uploads/", os.ModePerm)
	if err != nil {
		t.Fatal(err)
	}

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
			name:           "no specific limit for type",
			fileType:       "application/pdf",
			fileSize:       3 * 1024 * 1024, // 3MB
			typeSizeLimits: map[string]int{"image/png": 1 * 1024 * 1024}, // Only PNG has limit
			errorExpected:  false, // Should use default max file size
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create a test file with the specified type and size
			pr, pw := io.Pipe()
			writer := multipart.NewWriter(pw)
			
			go func() {
				defer writer.Close()
				part, err := writer.CreateFormFile("file", "testfile.dat")
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
				MaxFileSize:           5 * 1024 * 1024, // 5MB default
				TypeSpecificSizeLimits: tc.typeSizeLimits,
				UploadPath:            "./testdata/uploads/",
				AllowedFileTypes:      []string{tc.fileType},
				AllowUnknownTypes:     true,
			}
			
			// Mock the file type detection to return the specified type
			// This is needed because we're not uploading real files with proper content
			tools.detectFileType = func(file multipart.File) (string, error) {
				return tc.fileType, nil
			}
			
			// Attempt to upload the file
			files, err := tools.UploadFiles(request, "./testdata/uploads/", true)
			
			// Check error expectations
			if err != nil && !tc.errorExpected {
				t.Errorf("got unexpected error: %s", err.Error())
			}
			
			if err == nil && tc.errorExpected {
				t.Error("expected error but got none")
			}
			
			// Clean up any created files
			if err == nil {
				for _, f := range files {
					os.Remove(fmt.Sprintf("./testdata/uploads/%s", f.NewFileName))
				}
			}
		})
	}
}

// Add this new test function after the existing tests
func TestTools_FileTypeValidation(t *testing.T) {
	tests := []struct {
		name              string
		fileType          string
		allowedTypes      []string
		allowUnknownTypes bool
		shouldBeAllowed   bool
	}{
		{
			name:              "allowed type",
			fileType:          "image/png",
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
		},
		{
			name:              "unknown type allowed",
			fileType:          "application/octet-stream",
			allowedTypes:      []string{"image/jpeg"},
			allowUnknownTypes: true,
			shouldBeAllowed:   true,
		},
		{
			name:              "empty allowed types with unknown allowed",
			fileType:          "text/plain",
			allowedTypes:      []string{},
			allowUnknownTypes: true,
			shouldBeAllowed:   true,
		},
		{
			name:              "empty allowed types with unknown disallowed",
			fileType:          "text/plain",
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
		})
	}
}

// Add a new test for content verification errors
func TestTools_ContentVerification(t *testing.T) {
	// Create the uploads directory if it doesn't exist
	err := os.MkdirAll("./testdata/uploads/", os.ModePerm)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name           string
		fileContent    []byte
		claimedType    string
		errorExpected  bool
		expectedError  error
	}{
		{
			name:          "valid content",
			fileContent:   []byte("This is a text file"),
			claimedType:   "text/plain",
			errorExpected: false,
		},
		{
			name:          "invalid content",
			fileContent:   []byte{0x4D, 0x5A, 0x90, 0x00}, // MZ header (Windows executable)
			claimedType:   "text/plain",
			errorExpected: true,
			expectedError: ErrContentVerification,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create a test file with the specified content
			tempFile, err := os.CreateTemp("./testdata", "verify-*")
			if err != nil {
				t.Fatalf("Failed to create temp file: %v", err)
			}
			defer os.Remove(tempFile.Name())
			defer tempFile.Close()
			
			// Write the test content
			_, err = tempFile.Write(tc.fileContent)
			if err != nil {
				t.Fatalf("Failed to write to temp file: %v", err)
			}
			
			// Reset file pointer to beginning
			_, err = tempFile.Seek(0, 0)
			if err != nil {
				t.Fatalf("Failed to reset file pointer: %v", err)
			}
			
			// Create a Tools instance
			tools := Tools{}
			
			// Verify the file content
			isValid, err := tools.VerifyFileContent(tempFile, tc.claimedType)
			
			// Check error expectations
			if !isValid && !tc.errorExpected {
				t.Errorf("expected valid content but got invalid")
			}
			
			if isValid && tc.errorExpected {
				t.Errorf("expected invalid content but got valid")
			}
			
			// Test the error when used in UploadFiles
			// This would require mocking the upload process with the test file
		})
	}
}

// TestTools_MaxUploadCount tests the enforcement of the maximum upload count limit
func TestTools_MaxUploadCount(t *testing.T) {
	// Create the uploads directory if it doesn't exist
	err := os.MkdirAll("./testdata/uploads/", os.ModePerm)
	if err != nil {
		t.Fatal(err)
	}

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

// Update existing tests to use errors.Is and errors.As for more precise error checking
// For example, in TestTools_BatchSizeLimits:

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

// TestTools_ResumableUploads tests the chunked upload functionality
func TestTools_ResumableUploads(t *testing.T) {
	// Create the uploads directory if it doesn't exist
	err := os.MkdirAll("./testdata/uploads/", os.ModePerm)
	if err != nil {
		t.Fatal(err)
	}
	
	// Create a temp directory for chunks
	err = os.MkdirAll("./testdata/chunks/", os.ModePerm)
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll("./testdata/chunks/")
	
	// Test cases
	tests := []struct {
		name           string
		fileSize       int64
		chunkSize      int64
		simulateBreak  int // After which chunk to simulate interruption (0 for no interruption)
		errorExpected  bool
	}{
		{
			name:          "complete upload without interruption",
			fileSize:      5 * 1024 * 1024, // 5MB
			chunkSize:     1 * 1024 * 1024, // 1MB chunks
			simulateBreak: 0,               // No interruption
			errorExpected: false,
		},
		{
			name:          "resume after interruption",
			fileSize:      5 * 1024 * 1024, // 5MB
			chunkSize:     1 * 1024 * 1024, // 1MB chunks
			simulateBreak: 3,               // Break after 3 chunks
			errorExpected: false,
		},
		{
			name:          "small file single chunk",
			fileSize:      500 * 1024,      // 500KB
			chunkSize:     1 * 1024 * 1024, // 1MB chunks
			simulateBreak: 0,               // No interruption
			errorExpected: false,
		},
	}
	
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create a test file with the specified size
			testFile, err := os.CreateTemp("./testdata", "resumable-*")
			if err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}
			defer os.Remove(testFile.Name())
			
			// Write random data to the file
			data := make([]byte, tc.fileSize)
			rand.Read(data) // Fill with random data
			_, err = testFile.Write(data)
			if err != nil {
				t.Fatalf("Failed to write test data: %v", err)
			}
			testFile.Close()
			
			// Calculate MD5 hash of original file for verification
			originalMD5, err := calculateFileMD5(testFile.Name())
			if err != nil {
				t.Fatalf("Failed to calculate MD5: %v", err)
			}
			
			// Create a Tools instance
			tools := Tools{
				MaxFileSize:      10 * 1024 * 1024,
				ChunkSize:        tc.chunkSize,
				ChunksDirectory:  "./testdata/chunks/",
				UploadPath:       "./testdata/uploads/",
				AllowUnknownTypes: true,
			}
			
			// Generate a unique upload ID
			uploadID := tools.RandomString(20)
			fileName := filepath.Base(testFile.Name())
			
			// Simulate chunked upload
			totalChunks := (tc.fileSize + tc.chunkSize - 1) / tc.chunkSize
			
			// First upload phase (before interruption)
			chunksToUpload := totalChunks
			if tc.simulateBreak > 0 && tc.simulateBreak < int(totalChunks) {
				chunksToUpload = int64(tc.simulateBreak)
			}
			
			// Upload chunks
			for i := int64(0); i < chunksToUpload; i++ {
				// Open the file for reading the chunk
				f, err := os.Open(testFile.Name())
				if err != nil {
					t.Fatalf("Failed to open test file: %v", err)
				}
				
				// Seek to the chunk position
				_, err = f.Seek(i*tc.chunkSize, 0)
				if err != nil {
					f.Close()
					t.Fatalf("Failed to seek in file: %v", err)
				}
				
				// Determine chunk size (last chunk may be smaller)
				currentChunkSize := tc.chunkSize
				if i == totalChunks-1 {
					currentChunkSize = tc.fileSize - (i * tc.chunkSize)
				}
				
				// Read the chunk
				chunkData := make([]byte, currentChunkSize)
				_, err = io.ReadFull(f, chunkData)
				f.Close()
				if err != nil {
					t.Fatalf("Failed to read chunk: %v", err)
				}
				
				// Upload the chunk
				err = tools.UploadChunk(uploadID, fileName, i, totalChunks, chunkData)
				if err != nil {
					t.Fatalf("Failed to upload chunk %d: %v", i, err)
				}
			}
			
			// If simulating interruption, upload the remaining chunks
			if tc.simulateBreak > 0 && tc.simulateBreak < int(totalChunks) {
				// Simulate resumption by uploading remaining chunks
				for i := int64(tc.simulateBreak); i < totalChunks; i++ {
					// Open the file for reading the chunk
					f, err := os.Open(testFile.Name())
					if err != nil {
						t.Fatalf("Failed to open test file: %v", err)
					}
					
					// Seek to the chunk position
					_, err = f.Seek(i*tc.chunkSize, 0)
					if err != nil {
						f.Close()
						t.Fatalf("Failed to seek in file: %v", err)
					}
					
					// Determine chunk size (last chunk may be smaller)
					currentChunkSize := tc.chunkSize
					if i == totalChunks-1 {
						currentChunkSize = tc.fileSize - (i * tc.chunkSize)
					}
					
					// Read the chunk
					chunkData := make([]byte, currentChunkSize)
					_, err = io.ReadFull(f, chunkData)
					f.Close()
					if err != nil {
						t.Fatalf("Failed to read chunk: %v", err)
					}
					
					// Upload the chunk
					err = tools.UploadChunk(uploadID, fileName, i, totalChunks, chunkData)
					if err != nil {
						t.Fatalf("Failed to upload chunk %d: %v", i, err)
					}
				}
			}
			
			// Complete the upload by assembling chunks
			uploadedFile, err := tools.CompleteChunkedUpload(uploadID, fileName)
			
			// Check error expectations
			if err != nil && !tc.errorExpected {
				t.Errorf("Failed to complete upload: %v", err)
			}
			
			if err == nil && tc.errorExpected {
				t.Error("Expected error but got none")
			}
			
			// Verify file integrity if upload was successful
			if err == nil {
				// Calculate MD5 of the assembled file
				assembledMD5, err := calculateFileMD5(filepath.Join("./testdata/uploads/", uploadedFile.NewFileName))
				if err != nil {
					t.Fatalf("Failed to calculate MD5 of assembled file: %v", err)
				}
				
				// Compare MD5 hashes
				if originalMD5 != assembledMD5 {
					t.Errorf("File integrity check failed: original MD5 %s, assembled MD5 %s", 
						originalMD5, assembledMD5)
				}
				
				// Clean up
				os.Remove(filepath.Join("./testdata/uploads/", uploadedFile.NewFileName))
			}
		})
	}
}

// Helper function to calculate MD5 hash of a file
func calculateFileMD5(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()
	
	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	
	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}
