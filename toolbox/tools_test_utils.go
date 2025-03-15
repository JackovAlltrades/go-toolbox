package toolbox

import (
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// setupTestDir creates a test directory if it doesn't exist
func setupTestDir(t *testing.T, path string) {
	err := os.MkdirAll(path, os.ModePerm)
	if err != nil {
		t.Fatal(err)
	}
}

// cleanupTestDir removes a test directory and its contents
func cleanupTestDir(t *testing.T, path string) {
	err := os.RemoveAll(path)
	if err != nil {
		t.Logf("Warning: failed to clean up test directory %s: %v", path, err)
	}
}

// calculateFileMD5 returns the MD5 hash of a file
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

// createTestFile creates a temporary file with random data of the specified size
func createTestFile(t *testing.T, dir string, prefix string, size int64) string {
	// Create a test file
	testFile, err := os.CreateTemp(dir, prefix)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	
	// Write data to the file
	data := make([]byte, size)
	_, err = testFile.Write(data)
	if err != nil {
		testFile.Close()
		os.Remove(testFile.Name())
		t.Fatalf("Failed to write test data: %v", err)
	}
	
	testFile.Close()
	return testFile.Name()
}

// ensureTestDataDir ensures the testdata directory exists
func ensureTestDataDir(t *testing.T) {
	setupTestDir(t, "./testdata")
}

// cleanupTestFiles removes a list of files
func cleanupTestFiles(t *testing.T, files []string) {
	for _, file := range files {
		if err := os.Remove(file); err != nil && !os.IsNotExist(err) {
			t.Logf("Warning: failed to clean up test file %s: %v", file, err)
		}
	}
}