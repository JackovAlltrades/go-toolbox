# h:\Projects\toolbox-project\remove_duplicate_tests.ps1

# Define the path to the test file
$testFilePath = "h:\Projects\toolbox-project\toolbox\tools_test.go"

# Create a backup of the original file
$backupPath = "$testFilePath.backup"
Copy-Item -Path $testFilePath -Destination $backupPath
Write-Host "Created backup at $backupPath"

# Read the content of the file
$content = Get-Content -Path $testFilePath -Raw

# Define the test functions to keep (these are unique to tools_test.go)
$functionsToKeep = @(
    "TestTools_RandomString",
    "TestTools_ConfigFields",
    "TestTools_UploadOneFile",
    "TestTools_CreateDirIfNotExist",
    "TestTools_Slugify"
)

# Define the test functions to remove (these are duplicated in other files)
$functionsToRemove = @(
    "TestTools_UploadFiles",
    "TestTools_BatchSizeLimits",
    "TestTools_TypeSpecificSizeLimits",
    "TestTools_FileTypeValidation",
    "TestTools_ContentVerification",
    "TestTools_MaxUploadCount",
    "TestTools_ResumableUploads"
)

# Create a new content string with proper imports
$newContent = @"
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

"@

# Add the uploadTests variable definition as it's used by other tests
$uploadTestsPattern = "(?s)// Define uploadTests.*?}\)"
if ($content -match $uploadTestsPattern) {
    $uploadTestsDefinition = $matches[0]
    $newContent += "`n$uploadTestsDefinition`n`n"
}

# Extract and add each function to keep
foreach ($funcName in $functionsToKeep) {
    $pattern = "(?s)func $funcName.*?^}$"
    if ($content -match $pattern) {
        $functionCode = $matches[0]
        $newContent += "$functionCode`n`n"
    }
    else {
        Write-Host "Warning: Could not find function $funcName" -ForegroundColor Yellow
    }
}

# Add the calculateFileMD5 helper function if it exists
$md5Pattern = "(?s)// Helper function to calculate MD5.*?^}$"
if ($content -match $md5Pattern) {
    $md5Function = $matches[0]
    $newContent += "$md5Function`n"
}

# Write the new content to the file
Set-Content -Path $testFilePath -Value $newContent
Write-Host "Successfully removed duplicate test functions from $testFilePath" -ForegroundColor Green
Write-Host "Kept functions: $($functionsToKeep -join ', ')"
Write-Host "Removed functions: $($functionsToRemove -join ', ')"
Write-Host "If you need to restore the original file, use the backup at $backupPath"