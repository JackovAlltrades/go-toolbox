# Script to push latest changes to GitHub

# Navigate to the project directory
Set-Location -Path "h:\Projects\toolbox-project"

# Add all changes to git
git add .

# Commit the changes with a descriptive message
git commit -m "Refactor test files to remove duplicates and improve test organization"

# Push to GitHub
git push origin main

Write-Host "Successfully pushed changes to GitHub" -ForegroundColor Green