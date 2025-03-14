package toolbox

import (
	"errors"
	"os"
	"testing"
)

// TestTools_ContentVerification tests the content verification functionality
func TestTools_ContentVerification(t *testing.T) {
	// Create the uploads directory if it doesn't exist
	setupTestDir(t, "./testdata/uploads/")
	defer cleanupTestDir(t, "./testdata/uploads/")

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
		})
	}
}