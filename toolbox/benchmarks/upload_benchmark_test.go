package benchmarks

import (
	"bytes"
	"io"
	"math/rand"
	"mime/multipart"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"h:/Projects/toolbox-project/toolbox"
)

// setupBenchDir creates a directory for benchmark files
func setupBenchDir(b *testing.B, path string) {
	err := os.MkdirAll(path, os.ModePerm)
	if err != nil {
		b.Fatal(err)
	}
}

// cleanupBenchDir removes a benchmark directory
func cleanupBenchDir(b *testing.B, path string) {
	err := os.RemoveAll(path)
	if err != nil {
		b.Logf("Warning: failed to clean up benchmark directory %s: %v", path, err)
	}
}

// BenchmarkUploadSingleFile benchmarks uploading a single file of various sizes
func BenchmarkUploadSingleFile(b *testing.B) {
	// Create the uploads directory if it doesn't exist
	uploadPath := "./benchmark_data/uploads/"
	setupBenchDir(b, uploadPath)
	defer cleanupBenchDir(b, uploadPath)

	// Test with different file sizes
	fileSizes := []struct {
		name string
		size int64
	}{
		{"Small", 10 * 1024},        // 10KB
		{"Medium", 1 * 1024 * 1024}, // 1MB
		{"Large", 10 * 1024 * 1024}, // 10MB
	}

	for _, fs := range fileSizes {
		b.Run(fs.name, func(b *testing.B) {
			// Create a test file with the specified size
			data := make([]byte, fs.size)
			rand.Read(data)

			// Reset the timer for each iteration
			b.ResetTimer()

			// Run the benchmark
			for i := 0; i < b.N; i++ {
				b.StopTimer() // Stop timer while preparing the request

				// Create a multipart request
				pr, pw := io.Pipe()
				writer := multipart.NewWriter(pw)

				go func() {
					defer writer.Close()
					part, err := writer.CreateFormFile("file", "benchmark.dat")
					if err != nil {
						b.Error(err)
						return
					}
					part.Write(data)
				}()

				request := httptest.NewRequest("POST", "/", pr)
				request.Header.Add("Content-Type", writer.FormDataContentType())

				// Configure the tools
				tools := toolbox.Tools{
					MaxFileSize:      100 * 1024 * 1024, // 100MB
					AllowedFileTypes: []string{"application/octet-stream"},
					AllowUnknownTypes: true,
					UploadPath:       uploadPath,
				}

				b.StartTimer() // Resume timer for the actual upload

				// Perform the upload
				files, err := tools.UploadFiles(request, uploadPath, true)
				if err != nil {
					b.Fatalf("Upload failed: %v", err)
				}

				b.StopTimer() // Stop timer for cleanup

				// Clean up
				for _, f := range files {
					os.Remove(filepath.Join(uploadPath, f.NewFileName))
				}
			}
		})
	}
}

// BenchmarkChunkedUpload benchmarks the chunked upload process
func BenchmarkChunkedUpload(b *testing.B) {
	// Create the uploads directory if it doesn't exist
	uploadPath := "./benchmark_data/uploads/"
	chunksPath := "./benchmark_data/chunks/"
	setupBenchDir(b, uploadPath)
	setupBenchDir(b, chunksPath)
	defer cleanupBenchDir(b, uploadPath)
	defer cleanupBenchDir(b, chunksPath)

	// Test with different file sizes and chunk sizes
	testCases := []struct {
		name      string
		fileSize  int64
		chunkSize int64
	}{
		{"Small_SmallChunks", 1 * 1024 * 1024, 256 * 1024},      // 1MB with 256KB chunks
		{"Medium_MediumChunks", 5 * 1024 * 1024, 1 * 1024 * 1024}, // 5MB with 1MB chunks
		{"Large_LargeChunks", 20 * 1024 * 1024, 5 * 1024 * 1024},  // 20MB with 5MB chunks
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			// Create a test file with the specified size
			data := make([]byte, tc.fileSize)
			rand.Read(data)

			// Reset the timer for each iteration
			b.ResetTimer()

			// Run the benchmark
			for i := 0; i < b.N; i++ {
				b.StopTimer() // Stop timer while preparing

				// Create a Tools instance
				tools := toolbox.Tools{
					MaxFileSize:      100 * 1024 * 1024, // 100MB
					ChunkSize:        tc.chunkSize,
					ChunksDirectory:  chunksPath,
					UploadPath:       uploadPath,
					AllowUnknownTypes: true,
				}

				// Generate a unique upload ID
				uploadID := tools.RandomString(20)
				fileName := "benchmark.dat"

				// Calculate total chunks
				totalChunks := (tc.fileSize + tc.chunkSize - 1) / tc.chunkSize

				b.StartTimer() // Start timing the actual upload process

				// Upload chunks
				for j := int64(0); j < totalChunks; j++ {
					// Determine chunk size (last chunk may be smaller)
					currentChunkSize := tc.chunkSize
					if j == totalChunks-1 {
						currentChunkSize = tc.fileSize - (j * tc.chunkSize)
					}

					// Get the chunk data
					start := j * tc.chunkSize
					end := start + currentChunkSize
					if end > tc.fileSize {
						end = tc.fileSize
					}
					chunkData := data[start:end]

					// Upload the chunk
					err := tools.UploadChunk(uploadID, fileName, j, totalChunks, chunkData)
					if err != nil {
						b.Fatalf("Failed to upload chunk %d: %v", j, err)
					}
				}

				// Complete the upload
				uploadedFile, err := tools.CompleteChunkedUpload(uploadID, fileName)
				if err != nil {
					b.Fatalf("Failed to complete upload: %v", err)
				}

				b.StopTimer() // Stop timer for cleanup

				// Clean up
				os.Remove(filepath.Join(uploadPath, uploadedFile.NewFileName))
			}
		})
	}
}

// BenchmarkConcurrentUploads benchmarks multiple concurrent uploads
func BenchmarkConcurrentUploads(b *testing.B) {
	// Create the uploads directory if it doesn't exist
	uploadPath := "./benchmark_data/uploads/"
	setupBenchDir(b, uploadPath)
	defer cleanupBenchDir(b, uploadPath)

	// Test with different concurrency levels
	concurrencyLevels := []int{1, 5, 10, 20}
	fileSize := int64(1 * 1024 * 1024) // 1MB files

	for _, concurrency := range concurrencyLevels {
		b.Run(fmt.Sprintf("Concurrency_%d", concurrency), func(b *testing.B) {
			// Create test data
			data := make([]byte, fileSize)
			rand.Read(data)

			b.ResetTimer()

			// Run the benchmark
			for i := 0; i < b.N; i++ {
				b.StopTimer()

				// Create a wait group to synchronize goroutines
				var wg sync.WaitGroup
				wg.Add(concurrency)

				// Create an error channel
				errCh := make(chan error, concurrency)

				b.StartTimer()

				// Start concurrent uploads
				for j := 0; j < concurrency; j++ {
					go func(id int) {
						defer wg.Done()

						// Create a multipart request
						pr, pw := io.Pipe()
						writer := multipart.NewWriter(pw)

						go func() {
							defer writer.Close()
							part, err := writer.CreateFormFile("file", fmt.Sprintf("benchmark_%d.dat", id))
							if err != nil {
								errCh <- err
								return
							}
							part.Write(data)
						}()

						request := httptest.NewRequest("POST", "/", pr)
						request.Header.Add("Content-Type", writer.FormDataContentType())

						// Configure the tools
						tools := toolbox.Tools{
							MaxFileSize:      100 * 1024 * 1024, // 100MB
							AllowedFileTypes: []string{"application/octet-stream"},
							AllowUnknownTypes: true,
							UploadPath:       uploadPath,
						}

						// Perform the upload
						files, err := tools.UploadFiles(request, uploadPath, true)
						if err != nil {
							errCh <- err
							return
						}

						// Clean up files (don't count this in the timing)
						for _, f := range files {
							os.Remove(filepath.Join(uploadPath, f.NewFileName))
						}
					}(j)
				}

				// Wait for all uploads to complete
				wg.Wait()
				close(errCh)

				b.StopTimer()

				// Check for errors
				for err := range errCh {
					if err != nil {
						b.Fatalf("Upload failed: %v", err)
					}
				}
			}
		})
	}
}