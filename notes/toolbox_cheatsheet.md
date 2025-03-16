# Go Toolbox Project Cheat Sheet

## Table of Contents
1. [Project Structure](#project-structure)
2. [PowerShell Commands](#powershell-commands)
   - [Directory Operations](#directory-operations)
   - [View Directory Structure](#view-directory-structure)
3. [Git Commands](#git-commands)
   - [Basic Git Workflow](#basic-git-workflow)
   - [Git Configuration](#git-configuration)
4. [Go Commands](#go-commands)
   - [Module Management](#module-management)
   - [Testing](#testing)
   - [Building](#building)
5. [Project Scripts](#project-scripts)
   - [reorganize_project.ps1](#reorganize_projectps1)
   - [cleanup_duplicates.ps1](#cleanup_duplicatesps1)
   - [fix_imports.ps1](#fix_importsps1)
6. [Common Tasks](#common-tasks)
   - [Update README](#update-readme)
   - [Run Example Code](#run-example-code)
   - [Check for Module Issues](#check-for-module-issues)

## Project Structure

The project follows standard Go project layout:

```
go-toolbox/
├── cmd/api/         # Main application code
├── toolbox/         # Core library code
├── docs/            # Documentation
├── examples/        # Example usage code
├── testdata/        # Test data files
└── temp_backup/     # Backup of original files
```

## PowerShell Commands

### Directory Operations
```powershell
# Create a directory
New-Item -ItemType Directory -Path "path\to\directory"

# Copy files/directories
Copy-Item -Path "source" -Destination "destination" -Recurse

# Remove files/directories
Remove-Item -Path "path\to\remove" -Recurse -Force

# Check if path exists
Test-Path "path\to\check"
```

### View Directory Structure
```powershell
# Show directory structure with files
tree /F "H:\Projects\toolbox-project"

# Show only directory structure
tree "H:\Projects\toolbox-project"

# Save structure to a file
tree /F "H:\Projects\toolbox-project" > "H:\Projects\toolbox-project\project_structure.txt"
```

## Git Commands

### Basic Git Workflow
```bash
# Add all changes
git add .

# Commit changes
git commit -m "Descriptive message about changes"

# Push to remote repository
git push origin main
```

### Git Configuration
```powershell
# Create .gitignore file
Set-Content -Path ".gitignore" -Value "notes/`ntemp_backup/`n*.exe`n*.dll`n*.so`n*.dylib`n"
```

## Go Commands

### Module Management
```bash
# Initialize a new module
go mod init github.com/JackovAlltrades/go-toolbox

# Update dependencies
go mod tidy

# Sync workspace
go work sync
```

### Testing
```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific test
go test ./toolbox -run TestUploadFiles
```

### Building
```bash
# Build the application
go build -o toolbox.exe ./cmd/api

# Install the application
go install ./cmd/api
```

## Project Scripts

### reorganize_project.ps1
```powershell
# Used to reorganize the project structure to follow standard Go layout
.\reorganize_project.ps1
```

### cleanup_duplicates.ps1
```powershell
# Used to clean up duplicate directories and ensure proper project structure
.\cleanup_duplicates.ps1
```

### fix_imports.ps1
```powershell
# Used to fix import paths in Go files to match the module structure
.\fix_imports.ps1
```

## Common Tasks

### Update README
```powershell
# Update README from content file
Copy-Item -Path "readme_content.txt" -Destination "README.md" -Force
```

### Run Example Code
```bash
# Run an example
go run ./examples/upload_example.go
```

### Check for Module Issues
If you encounter module path issues:

1. Run the fix_imports.ps1 script
2. Run `go mod tidy`
3. Check that imports use `github.com/JackovAlltrades/go-toolbox`