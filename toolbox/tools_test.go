package toolbox

import (
	"fmt"
	"image"
	"image/png"
	"io"
	"mime/multipart"
	"net/http/httptest"
	"os"
	"sync"
	"testing"
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

var uploadTests = []struct {
	name          string
	allowedTypes  []string
	renameFile    bool
	errorExpected bool
}{
	{name: "allowed no rename", allowedTypes: []string{"image/jpeg", "image/png"}, renameFile: false, errorExpected: false},
	{name: "allowed rename", allowedTypes: []string{"image/jpeg", "image/png"}, renameFile: true, errorExpected: false},
	{name: "not allowed", allowedTypes: []string{"image/jpeg"}, renameFile: false, errorExpected: true},
}

func TestTools_UploadFiles(t *testing.T) {
	// Create the uploads directory if it doesn't exist
	err := os.MkdirAll("./testdata/uploads/", os.ModePerm)
	if err != nil {
		t.Fatal(err)
	}

	for _, e := range uploadTests {
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

		uploadedFiles, err := testTools.UploadFiles(request, "./testdata/uploads/", e.renameFile)
		if err != nil && !e.errorExpected {
			t.Error(err)
		}

		if !e.errorExpected {
			if _, err := os.Stat(fmt.Sprintf("./testdata/uploads/%s", uploadedFiles[0].NewFileName)); os.IsNotExist(err) {
				t.Errorf("%s: expected file to exist: %s", e.name, err.Error())
			}

			// clean up
			_ = os.Remove(fmt.Sprintf("./testdata/uploads/%s", uploadedFiles[0].NewFileName))
		}

		if !e.errorExpected && err != nil {
			t.Errorf("%s: error expected but none received", e.name)
		}

		wg.Wait()
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
