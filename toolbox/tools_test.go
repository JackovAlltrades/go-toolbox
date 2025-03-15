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
	"path/filepath"
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


