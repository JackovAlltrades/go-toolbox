package toolbox

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestTools_DownloadStaticFile(t *testing.T) {
	// Create test directory and file if they don't exist
	testDir := "./testdata"
	err := os.MkdirAll(testDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	// Create a test file or verify it exists
	testFile := filepath.Join(testDir, "pic.jpg")
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		// Create a dummy file for testing
		f, err := os.Create(testFile)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
		// Write some dummy content
		f.WriteString("This is a test file content for download testing")
		f.Close()
	}

	rr := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)

	var testTool Tools

	testTool.DownloadStaticFile(rr, req, testDir, "pic.jpg", "puppy.jpg")

	res := rr.Result()
	defer res.Body.Close()

	// Check content disposition header
	if res.Header.Get("Content-Disposition") != "attachment; filename=\"puppy.jpg\"" {
		t.Error("wrong content disposition:", res.Header.Get("Content-Disposition"))
	}

	// Read the response body
	_, err = io.ReadAll(res.Body)
	if err != nil {
		t.Error(err)
	}
}
