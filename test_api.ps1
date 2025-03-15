# Script to update project and test API

# Navigate to the project directory
Set-Location -Path "h:\Projects\toolbox-project"

# Run tests to ensure everything is working
Write-Host "Running tests..." -ForegroundColor Cyan
go test ./toolbox/...

# Build the project
Write-Host "Building project..." -ForegroundColor Cyan
go build -o toolbox.exe ./cmd/api

# Start the API server
Write-Host "Starting API server..." -ForegroundColor Cyan
Write-Host "Press Ctrl+C to stop the server" -ForegroundColor Yellow
.\toolbox.exe