# Go Toolbox Project Cheat Sheet

## Table of Contents
1. [Project Structure](#project-structure)
2. [PowerShell Commands](#powershell-commands)
3. [Git Commands](#git-commands)
4. [Go Commands](#go-commands)
5. [Project Scripts](#project-scripts)
6. [Common Tasks](#common-tasks)

## Project Structure
The project follows standard Go project layout:
- `/cmd/api` - Main application code
- `/toolbox` - Core library code
- `/docs` - Documentation
- `/examples` - Example usage code
- `/testdata` - Test data files
- `/temp_backup` - Backup of original files

## PowerShell Commands

### View Directory Structure
```powershell
# Show directory structure with files
tree /F "H:\Projects\toolbox-project"

# Show only directory structure
tree "H:\Projects\toolbox-project"

# Save structure to a file
tree /F "H:\Projects\toolbox-project" > "H:\Projects\toolbox-project\project_structure.txt"
