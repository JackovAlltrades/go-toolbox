package toolbox

import (
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const randomStringSource = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_+"

// Tools is type used to instantiate the module. Variables of type allowed access
// to all methods with reciever *Tools
type Tools struct {
	MaxFileSize      int
	AllowedFileTypes []string
}

// RandomString returns a random string of characters of length n, using randomStringSource
// as source of string. Returns an empty string if n <= 0.
func (t *Tools) RandomString(n int) string {
	// Handle invalid input
	if n <= 0 {
		return ""
	}
	
	s, r := make([]rune, n), []rune(randomStringSource)
	
	// Check if randomStringSource is empty
	if len(r) == 0 {
		return ""
	}
	
	for i := range s {
		// Handle potential errors from rand.Prime
		p, err := rand.Prime(rand.Reader, len(r))
		if err != nil {
			// Fallback to a simpler random method if Prime fails
			b := make([]byte, 1)
			_, err = rand.Read(b)
			if err != nil {
				// If all random methods fail, use a deterministic approach
				s[i] = r[i%len(r)]
				continue
			}
			s[i] = r[int(b[0])%len(r)]
			continue
		}
		
		x, y := p.Uint64(), uint64(len(r))
		s[i] = r[x%y]
	}

	return string(s)
}

// UploadedFile a struct used to save information on uploaded file
type UploadedFile struct {
	NewFileName      string
	OriginalFileName string
	FileSize         int64
}

// UploadOneFile a convenience method that calls UploadFiles, and expects one file
// // to be uploaded.
func (t *Tools) UploadOneFile(r *http.Request, uploadDir string, rename ...bool) (*UploadedFile, error) {
	renameFile := true
	if len(rename) > 0 {
		renameFile = rename[0]
	}

	files, err := t.UploadFiles(r, uploadDir, renameFile)
	if err != nil {
		return nil, err
	}

	return files[0], nil
}

// UploadFiles uploads one or more file to a specified directory, and gives each files a random name.
// Returns slice containing new file names, original file names,and total size of all files,
// and anyerrors. If optional last parameter, set to true, will not rename files, will
// use the original file names.
func (t *Tools) UploadFiles(r *http.Request, uploadDir string, rename ...bool) ([]*UploadedFile, error) {
	renameFile := true
	if len(rename) > 0 {
		renameFile = rename[0]
	}

	var uploadedFiles []*UploadedFile

	if t.MaxFileSize == 0 {
		t.MaxFileSize = 2048 * 2048 * 2048
	}

	err := t.CreateDirIfNotExist(uploadDir)
	if err != nil {
		return nil, err
	}

	err = r.ParseMultipartForm(int64(t.MaxFileSize))
	if err != nil {
		return nil, errors.New("the uploaded file is too big")
	}

	for _, fHeaders := range r.MultipartForm.File {
		for _, hdr := range fHeaders {
			uploadedFiles, err = func(uploadedFiles []*UploadedFile) ([]*UploadedFile, error) {
				var uploadedFile UploadedFile
				infile, err := hdr.Open()
				if err != nil {
					return nil, err
				}
				defer infile.Close()

				buff := make([]byte, 512)
				_, err = infile.Read(buff)
				if err != nil {
					return nil, err
				}

				// check to see if the file type is permitted
				allowed := false
				fileType := http.DetectContentType(buff)

				if len(t.AllowedFileTypes) > 0 {
					for _, x := range t.AllowedFileTypes {
						if strings.EqualFold(fileType, x) {
							allowed = true
						}
					}
				} else {
					allowed = true
				}

				if !allowed {
					return nil, errors.New("the uploaded file type is not permitted")
				}

				_, err = infile.Seek(0, 0)
				if err != nil {
					return nil, err
				}

				if renameFile {
					uploadedFile.NewFileName = fmt.Sprintf("%s%s", t.RandomString(25), filepath.Ext(hdr.Filename))
				} else {
					uploadedFile.NewFileName = hdr.Filename
				}

				uploadedFile.OriginalFileName = hdr.Filename

				var outfile *os.File
				defer outfile.Close()

				if outfile, err = os.Create(filepath.Join(uploadDir, uploadedFile.NewFileName)); err != nil {
					return nil, err
				} else {
					fileSize, err := io.Copy(outfile, infile)
					if err != nil {
						return nil, err
					}
					uploadedFile.FileSize = fileSize
				}

				uploadedFiles = append(uploadedFiles, &uploadedFile)

				return uploadedFiles, nil
			}(uploadedFiles)
			if err != nil {
				return uploadedFiles, err
			}
		}
	}
	return uploadedFiles, nil
}

// CreateDirIfNotExist creates a directory, and all necessary parents, if it does not exist
func (t *Tools) CreateDirIfNotExist(path string) error {
	const mode = 0755
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err := os.MkdirAll(path, mode)
		if err != nil {
			return err
		}
	}
	return nil
}

// Slugify converts a string to a URL-friendly slug
// It handles special characters, multiple spaces, and ensures proper formatting
func (t *Tools) Slugify(s string) (string, error) {
    // Check for empty string
    if s == "" {
        return "", errors.New("empty string not permitted")
    }
    
    // Convert to lowercase and trim spaces
    s = strings.TrimSpace(strings.ToLower(s))
    
    // Handle non-ASCII characters (optional transliteration)
    // This is a simple replacement - consider using a proper transliteration library for production
    replacer := strings.NewReplacer(
        "æ", "ae", "ø", "o", "å", "a", "ü", "u", "ö", "o", "ä", "a",
        "ñ", "n", "é", "e", "è", "e", "ê", "e", "ë", "e", "á", "a",
        "à", "a", "â", "a", "ã", "a", "ç", "c", "í", "i", "ì", "i",
        "î", "i", "ï", "i", "ó", "o", "ò", "o", "ô", "o", "õ", "o",
        "ú", "u", "ù", "u", "û", "u", "ý", "y", "ÿ", "y",
    )
    s = replacer.Replace(s)
    
    // Replace any non-alphanumeric characters with hyphens
    var re = regexp.MustCompile(`[^a-z0-9]+`)
    slug := strings.Trim(re.ReplaceAllString(s, "-"), "-")
    
    // Check if slug is empty after processing
    if len(slug) == 0 {
        return "", errors.New("after removing characters, slug is zero length")
    }
    
    // Avoid multiple consecutive hyphens
    multiHyphen := regexp.MustCompile(`-+`)
    slug = multiHyphen.ReplaceAllString(slug, "-")
    
    // Limit slug length (optional, adjust as needed)
    maxLength := 100
    if len(slug) > maxLength {
        slug = slug[:maxLength]
        // Ensure we don't end with a hyphen if we truncated
        slug = strings.TrimSuffix(slug, "-")
    }
    
    return slug, nil
}
