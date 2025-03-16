# Script to identify and clean up duplicate directories outside the main project

# Define the main project path
$mainProjectPath = "H:\Projects\toolbox-project"
$parentDir = "H:\Projects"

# First, clean up the original app directories that should have been consolidated
$oldAppDirs = @(
    "app",
    "app-dir",
    "app-download",
    "app-json",
    "app-slug",
    "app-upload",
    "temp",
    "templates",
    "uploads"
)

# Add diagnostic information
Write-Host "Running diagnostic check on directories..." -ForegroundColor Cyan
foreach ($dir in $oldAppDirs) {
    $dirPath = Join-Path -Path $mainProjectPath -ChildPath $dir
    if (Test-Path $dirPath) {
        $dirInfo = Get-Item -Path $dirPath
        Write-Host "Directory exists: $($dirInfo.FullName)" -ForegroundColor Yellow
        Write-Host "  Created: $($dirInfo.CreationTime)" -ForegroundColor Yellow
        Write-Host "  Last Modified: $($dirInfo.LastWriteTime)" -ForegroundColor Yellow
        Write-Host "  Attributes: $($dirInfo.Attributes)" -ForegroundColor Yellow
        
        # Count files in the directory
        $fileCount = (Get-ChildItem -Path $dirPath -Recurse -File).Count
        Write-Host "  Contains $fileCount files" -ForegroundColor Yellow
    } else {
        Write-Host "Directory does not exist: $dirPath" -ForegroundColor Green
    }
}

Write-Host "Cleaning up original app directories that should have been consolidated..." -ForegroundColor Cyan
foreach ($dir in $oldAppDirs) {
    $dirPath = Join-Path -Path $mainProjectPath -ChildPath $dir
    if (Test-Path $dirPath) {
        try {
            # Make sure we have a backup before removing
            $backupPath = Join-Path -Path $mainProjectPath -ChildPath "temp_backup\$dir"
            if (-not (Test-Path $backupPath)) {
                Write-Host "  Warning: No backup found for $dir, creating one now..." -ForegroundColor Yellow
                if (-not (Test-Path "$mainProjectPath\temp_backup")) {
                    New-Item -ItemType Directory -Path "$mainProjectPath\temp_backup" -Force | Out-Null
                }
                Copy-Item -Path $dirPath -Destination $backupPath -Recurse -Force
            }
            
            # Remove the original directory
            Remove-Item -Path $dirPath -Recurse -Force
            Write-Host "  Removed original directory: $dir" -ForegroundColor Green
        } catch {
            Write-Host "  Failed to remove ${dir}: $($_.Exception.Message)" -ForegroundColor Red
        }
    }
}

# Ensure the standard Go project structure exists
$standardDirs = @(
    "cmd\api",
    "toolbox",
    "docs",
    "examples",
    "testdata"
)

foreach ($dir in $standardDirs) {
    $dirPath = Join-Path -Path $mainProjectPath -ChildPath $dir
    if (-not (Test-Path $dirPath)) {
        New-Item -ItemType Directory -Path $dirPath -Force | Out-Null
        Write-Host "Created missing standard directory: $dir" -ForegroundColor Green
    }
}

# Now look for duplicate directories outside the main project
Write-Host "Scanning for potential duplicate directories outside the main project..." -ForegroundColor Cyan
$potentialDuplicates = Get-ChildItem -Path $parentDir -Directory | Where-Object {
    # Skip the main project directory
    $_.FullName -ne $mainProjectPath -and
    # Look for directories with names that are specifically related to toolbox
    ($_.Name -like "toolbox*" -or 
     $_.Name -like "go-toolbox*")
}

# Display the potential duplicates
if ($potentialDuplicates.Count -eq 0) {
    Write-Host "No potential duplicate directories found outside the project." -ForegroundColor Green
} else {
    Write-Host "Found $($potentialDuplicates.Count) potential duplicate directories outside the project:" -ForegroundColor Yellow
    $potentialDuplicates | ForEach-Object {
        Write-Host "  - $($_.FullName)" -ForegroundColor Yellow
    }
    
    # Ask for confirmation before removing
    $confirmation = Read-Host "Do you want to remove these directories? (y/n)"
    if ($confirmation -eq 'y') {
        $potentialDuplicates | ForEach-Object {
            try {
                Remove-Item -Path $_.FullName -Recurse -Force
                Write-Host "Removed: $($_.FullName)" -ForegroundColor Green
            } catch {
                Write-Host "Failed to remove: $($_.FullName) - $($_.Exception.Message)" -ForegroundColor Red
            }
        }
        Write-Host "External cleanup completed." -ForegroundColor Green
    } else {
        Write-Host "External cleanup cancelled." -ForegroundColor Cyan
    }
}

Write-Host "Project structure cleanup completed. Run 'tree /F' to verify the new structure." -ForegroundColor Green