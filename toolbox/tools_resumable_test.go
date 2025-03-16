package toolbox_test_test_test_test_test_test_test

import (
	tools toolbox toolbox toolbox toolbox toolbox "github.com/JackovAlltrades/go-toolbox"
	"."
	"os"
	"path/filepath"
	"testing"

	"."
)

// TestTools_ResumableUploads tests the chunked upload functionality
func TestTools_ResumableUploads(t *testing.T) {
	// Create the uploads directory if it doesn't exist
	setupTestDir(t, "./testdata/uploads/")
	defer cleanupTestDir(t, "./testdata/uploads/")
	
	// Create a temp directory for chunks
	setupTestDir(t, "./testdata/chunks/")
	defer cleanupTestDir(t, "./testdata/chunks/")
	
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
			tools := tools.Tools{
				MaxFileSize:      10 * 1024 * 1024,
				ChunkSize:        tc.chunkSize,
				ChunksDirectory:  "./testdata/chunks/",
				UploadPath:       "./testdata/uploads/",
				AllowUnknownTypes: true,
			}
			
			// Generate a unique upload ID
			uploadID := RandomString(20)
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
				err = UploadChunk(uploadID, fileName, i, totalChunks, chunkData)
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
					err = UploadChunk(uploadID, fileName, i, totalChunks, chunkData)
					if err != nil {
						t.Fatalf("Failed to upload chunk %d: %v", i, err)
					}
				}
			}
			
			// Complete the upload by assembling chunks
			uploadedFile, err := CompleteChunkedUpload(uploadID, fileName)
			
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

// TestTools_UploadProgress tests the upload progress tracking functionality
func TestTools_UploadProgress(t *testing.T) {
	// Create the uploads directory if it doesn't exist
	setupTestDir(t, "./testdata/uploads/")
	defer cleanupTestDir(t, "./testdata/uploads/")
	
	// Create a temp directory for chunks
	setupTestDir(t, "./testdata/chunks/")
	defer cleanupTestDir(t, "./testdata/chunks/")
	
	// Create a test file
	testFile, err := os.CreateTemp("./testdata", "progress-*")
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	defer os.Remove(testFile.Name())
	
	// Write some data to the file
	fileSize := int64(3 * 1024 * 1024) // 3MB
	data := make([]byte, fileSize)
	rand.Read(data) // Fill with random data
	_, err = testFile.Write(data)
	if err != nil {
		t.Fatalf("Failed to write test data: %v", err)
	}
	testFile.Close()
	
	// Create a Tools instance
	tools := tools.Tools{
		ChunkSize:       1 * 1024 * 1024, // 1MB chunks
		ChunksDirectory: "./testdata/chunks/",
		UploadPath:      "./testdata/uploads/",
	}
	
	// Generate a unique upload ID
	uploadID := RandomString(20)
	fileName := filepath.Base(testFile.Name())
	
	// Calculate total chunks
	totalChunks := (fileSize + ChunkSize - 1) / ChunkSize
	
	// Upload chunks one by one and check progress
	for i := int64(0); i < totalChunks; i++ {
		// Open the file for reading the chunk
		f, err := os.Open(testFile.Name())
		if err != nil {
			t.Fatalf("Failed to open test file: %v", err)
		}
		
		// Seek to the chunk position
		_, err = f.Seek(i*ChunkSize, 0)
		if err != nil {
			f.Close()
			t.Fatalf("Failed to seek in file: %v", err)
		}
		
		// Determine chunk size (last chunk may be smaller)
		currentChunkSize := ChunkSize
		if i == totalChunks-1 {
			currentChunkSize = fileSize - (i * ChunkSize)
		}
		
		// Read the chunk
		chunkData := make([]byte, currentChunkSize)
		_, err = io.ReadFull(f, chunkData)
		f.Close()
		if err != nil {
			t.Fatalf("Failed to read chunk: %v", err)
		}
		
		// Upload the chunk
		err = UploadChunk(uploadID, fileName, i, totalChunks, chunkData)
		if err != nil {
			t.Fatalf("Failed to upload chunk %d: %v", i, err)
		}
		
		// Check progress
		progress, err := GetUploadProgress(uploadID)
		if err != nil {
			t.Fatalf("Failed to get upload progress: %v", err)
		}
		
		// Calculate expected progress
		expectedProgress := float64(i+1) / float64(totalChunks) * 100.0
		
		// Allow for small floating point differences
		if progress < expectedProgress-0.1 || progress > expectedProgress+0.1 {
			t.Errorf("Expected progress around %.2f%%, got %.2f%%", expectedProgress, progress)
		}
	}
	
	// Complete the upload
	uploadedFile, err := CompleteChunkedUpload(uploadID, fileName)
	if err != nil {
		t.Fatalf("Failed to complete upload: %v", err)
	}
	
	// Clean up
	os.Remove(filepath.Join("./testdata/uploads/", uploadedFile.NewFileName))
}

// TestTools_CancelUpload tests the upload cancellation functionality
func TestTools_CancelUpload(t *testing.T) {
	// Create the uploads directory if it doesn't exist
	setupTestDir(t, "./testdata/uploads/")
	defer cleanupTestDir(t, "./testdata/uploads/")
	
	// Create a temp directory for chunks
	setupTestDir(t, "./testdata/chunks/")
	defer cleanupTestDir(t, "./testdata/chunks/")
	
	// Create a test file
	testFile, err := os.CreateTemp("./testdata", "cancel-*")
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	defer os.Remove(testFile.Name())
	
	// Write some data to the file
	fileSize := int64(2 * 1024 * 1024) // 2MB
	data := make([]byte, fileSize)
	rand.Read(data) // Fill with random data
	_, err = testFile.Write(data)
	if err != nil {
		t.Fatalf("Failed to write test data: %v", err)
	}
	testFile.Close()
	
	// Create a Tools instance
	tools := tools.Tools{
		ChunkSize:       1 * 1024 * 1024, // 1MB chunks
		ChunksDirectory: "./testdata/chunks/",
		UploadPath:      "./testdata/uploads/",
	}
	
	// Generate a unique upload ID
	uploadID := RandomString(20)
	fileName := filepath.Base(testFile.Name())
	
	// Upload first chunk
	f, err := os.Open(testFile.Name())
	if err != nil {
		t.Fatalf("Failed to open test file: %v", err)
	}
	
	chunkData := make([]byte, ChunkSize)
	_, err = io.ReadFull(f, chunkData)
	f.Close()
	if err != nil {
		t.Fatalf("Failed to read chunk: %v", err)
	}
	
	err = UploadChunk(uploadID, fileName, 0, 2, chunkData)
	if err != nil {
		t.Fatalf("Failed to upload chunk: %v", err)
	}
	
	// Cancel the upload
	err = CancelChunkedUpload(uploadID)
	if err != nil {
		t.Fatalf("Failed to cancel upload: %v", err)
	}
	
	// Verify the upload was cancelled
	_, err = GetUploadProgress(uploadID)
	if err == nil {
		t.Error("Expected error after cancellation, but got none")
	}
	
	// Verify the chunks directory was removed
	chunksDir := filepath.Join(ChunksDirectory, uploadID)
	if _, err := os.Stat(chunksDir); !os.IsNotExist(err) {
		t.Errorf("Chunks directory still exists after cancellation")
	}
}




















