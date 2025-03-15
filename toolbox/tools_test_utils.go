package toolbox

import (
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"testing"
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