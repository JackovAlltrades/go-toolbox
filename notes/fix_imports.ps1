# Script to fix import paths in Go files

# Define the main project path
$projectRoot = "H:\Projects\toolbox-project"

# Find all Go files in the project
$goFiles = Get-ChildItem -Path $projectRoot -Filter "*.go" -Recurse

Write-Host "Checking import paths in Go files..." -ForegroundColor Cyan

foreach ($file in $goFiles) {
    $content = Get-Content -Path $file.FullName -Raw
    $modified = $false
    
    # Check for incorrect import paths
    if ($content -match 'import\s+["\(]github\.com/JackovAlltrades/toolbox') {
        $content = $content -replace 'github\.com/JackovAlltrades/toolbox', 'github.com/JackovAlltrades/go-toolbox'
        $modified = $true
        Write-Host "  Fixed import path in $($file.FullName)" -ForegroundColor Yellow
    }
    
    # Fix the specific issue with toolbox module imports
    if ($content -match 'github\.com/JackovAlltrades/go-toolbox/toolbox') {
        $content = $content -replace 'github\.com/JackovAlltrades/go-toolbox/toolbox', 'github\.com/JackovAlltrades/go-toolbox'
        $modified = $true
        Write-Host "  Fixed toolbox module import in $($file.FullName)" -ForegroundColor Yellow
    }
    
    # Save the file if modified
    if ($modified) {
        Set-Content -Path $file.FullName -Value $content
        Write-Host "  Saved changes to $($file.FullName)" -ForegroundColor Green
    }
}

# Check and fix go.mod files
$goModFiles = Get-ChildItem -Path $projectRoot -Filter "go.mod" -Recurse
foreach ($file in $goModFiles) {
    $content = Get-Content -Path $file.FullName -Raw
    $modified = $false
    
    # Make sure the module path is correct
    if ($file.DirectoryName -eq $projectRoot) {
        # Main module should be github.com/JackovAlltrades/go-toolbox
        if ($content -match 'module\s+github\.com/JackovAlltrades/go-toolbox/toolbox') {
            $content = $content -replace 'module\s+github\.com/JackovAlltrades/go-toolbox/toolbox', 'module github.com/JackovAlltrades/go-toolbox'
            $modified = $true
            Write-Host "  Fixed main module path in $($file.FullName)" -ForegroundColor Yellow
        }
    } elseif ($file.DirectoryName -like "*\toolbox") {
        # Toolbox module should be github.com/JackovAlltrades/go-toolbox
        if ($content -match 'module\s+github\.com/JackovAlltrades/go-toolbox/toolbox') {
            $content = $content -replace 'module\s+github\.com/JackovAlltrades/go-toolbox/toolbox', 'module github.com/JackovAlltrades/go-toolbox'
            $modified = $true
            Write-Host "  Fixed toolbox module path in $($file.FullName)" -ForegroundColor Yellow
        }
    }
    
    # Save the file if modified
    if ($modified) {
        Set-Content -Path $file.FullName -Value $content
        Write-Host "  Saved changes to $($file.FullName)" -ForegroundColor Green
    }
}

Write-Host "Import path check completed." -ForegroundColor Green
Write-Host "Run 'go mod tidy' to update dependencies." -ForegroundColor Cyan