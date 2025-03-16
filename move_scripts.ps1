# Get the directory where the script is located
$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$projectDir = $scriptDir  # Assume script is in the project root

# Create scripts directory if it doesn't exist
$scriptsDir = Join-Path $projectDir "scripts"
if (-not (Test-Path $scriptsDir)) {
    New-Item -ItemType Directory -Path $scriptsDir -Force
    Write-Host "Created scripts directory at $scriptsDir"
}

# Define patterns for script files to search for
$scriptPatterns = @(
    "run_*.ps1",
    "run_*.bat",
    "test_*.ps1",
    "test_*.bat"
)

# Define directories to search
$searchDirs = @(
    $projectDir,
    (Join-Path $projectDir "notes")
)

# Find and move script files
$movedFiles = @()

foreach ($dir in $searchDirs) {
    if (Test-Path $dir) {
        foreach ($pattern in $scriptPatterns) {
            $files = Get-ChildItem -Path $dir -Filter $pattern -File
            
            foreach ($file in $files) {
                # Skip files that are already in the scripts directory
                if ($file.DirectoryName -eq $scriptsDir) {
                    continue
                }
                
                # Skip the current script itself
                if ($file.FullName -eq $MyInvocation.MyCommand.Path) {
                    continue
                }
                
                $destPath = Join-Path $scriptsDir $file.Name
                
                # Check if file already exists in destination
                if (Test-Path $destPath) {
                    Write-Host "File already exists in scripts directory: $($file.Name)" -ForegroundColor Yellow
                } else {
                    # Move the file
                    Move-Item -Path $file.FullName -Destination $destPath
                    $movedFiles += $file.FullName
                    Write-Host "Moved: $($file.FullName) -> $destPath" -ForegroundColor Green
                }
            }
        }
    }
}

if ($movedFiles.Count -eq 0) {
    Write-Host "No script files found to move." -ForegroundColor Yellow
} else {
    Write-Host "`nSuccessfully moved $($movedFiles.Count) script files to $scriptsDir" -ForegroundColor Green
}

# Ask if scripts should be kept private
$keepPrivate = Read-Host "Do you want to keep the scripts private (not committed to the repository)? (y/n)"

if ($keepPrivate -eq "y" -or $keepPrivate -eq "Y") {
    # Add scripts directory to .gitignore
    $gitignorePath = Join-Path $projectDir ".gitignore"
    
    # Check if .gitignore exists
    if (-not (Test-Path $gitignorePath)) {
        # Create .gitignore file
        New-Item -ItemType File -Path $gitignorePath -Force | Out-Null
        Write-Host "Created .gitignore file" -ForegroundColor Green
    }
    
    # Check if scripts/ is already in .gitignore
    $gitignoreContent = Get-Content $gitignorePath -ErrorAction SilentlyContinue
    $scriptsPattern = "scripts/"
    
    if ($gitignoreContent -notcontains $scriptsPattern) {
        # Add scripts/ to .gitignore
        Add-Content -Path $gitignorePath -Value "`n# Private scripts`n$scriptsPattern"
        Write-Host "Added scripts/ to .gitignore - scripts will not be committed to the repository" -ForegroundColor Green
    } else {
        Write-Host "scripts/ is already in .gitignore" -ForegroundColor Yellow
    }
} else {
    Write-Host "Scripts will be included in the repository" -ForegroundColor Green
}

# Handle Git operations
Write-Host "`nUpdating Git repository..." -ForegroundColor Cyan

# Stage deleted files (original script locations)
git add --all

# Stage the new scripts directory
git add scripts/

# Stage the move_scripts.ps1 file itself
git add move_scripts.ps1

Write-Host "`nGit Status after staging changes:" -ForegroundColor Cyan
git status

Write-Host "`nTo commit these changes, run:" -ForegroundColor Yellow
Write-Host "git commit -m 'Reorganized scripts into scripts directory'" -ForegroundColor Yellow