# ppopcode Installation Script for Windows
# This script installs ppopcode globally so you can run it from anywhere

param(
    [switch]$Uninstall
)

$ErrorActionPreference = "Stop"

$INSTALL_DIR = "$env:USERPROFILE\bin"
$BINARY_NAME = "ppopcode.exe"
$PROJECT_ROOT = Split-Path -Parent $PSScriptRoot

function Write-Success {
    param([string]$Message)
    Write-Host "✓ $Message" -ForegroundColor Green
}

function Write-Info {
    param([string]$Message)
    Write-Host "→ $Message" -ForegroundColor Cyan
}

function Write-Error {
    param([string]$Message)
    Write-Host "✗ $Message" -ForegroundColor Red
}

function Test-InPath {
    param([string]$Directory)
    $userPath = [Environment]::GetEnvironmentVariable("Path", "User")
    return $userPath -like "*$Directory*"
}

function Add-ToPath {
    param([string]$Directory)
    
    if (Test-InPath $Directory) {
        Write-Info "Already in PATH: $Directory"
        return
    }
    
    Write-Info "Adding to PATH: $Directory"
    $userPath = [Environment]::GetEnvironmentVariable("Path", "User")
    $newPath = $userPath + ";$Directory"
    [Environment]::SetEnvironmentVariable("Path", $newPath, "User")
    Write-Success "Added to PATH"
}

function Remove-FromPath {
    param([string]$Directory)
    
    if (-not (Test-InPath $Directory)) {
        Write-Info "Not in PATH: $Directory"
        return
    }
    
    Write-Info "Removing from PATH: $Directory"
    $userPath = [Environment]::GetEnvironmentVariable("Path", "User")
    $newPath = ($userPath -split ';' | Where-Object { $_ -ne $Directory }) -join ';'
    [Environment]::SetEnvironmentVariable("Path", $newPath, "User")
    Write-Success "Removed from PATH"
}

function Install-Ppopcode {
    Write-Host "`n=== Installing ppopcode ===" -ForegroundColor Yellow
    
    # Check if binary exists
    $binaryPath = Join-Path $PROJECT_ROOT $BINARY_NAME
    if (-not (Test-Path $binaryPath)) {
        Write-Error "Binary not found: $binaryPath"
        Write-Info "Please build first: go build -o ppopcode.exe ./cmd/ppopcode"
        exit 1
    }
    
    # Create install directory
    if (-not (Test-Path $INSTALL_DIR)) {
        Write-Info "Creating directory: $INSTALL_DIR"
        New-Item -ItemType Directory -Path $INSTALL_DIR | Out-Null
        Write-Success "Created directory"
    }
    
    # Copy binary
    Write-Info "Copying $BINARY_NAME to $INSTALL_DIR"
    $targetPath = Join-Path $INSTALL_DIR $BINARY_NAME
    
    try {
        Copy-Item $binaryPath $targetPath -Force -ErrorAction Stop
        Write-Success "Copied binary"
    } catch {
        Write-Error "Failed to copy binary. The file might be in use."
        Write-Info "Please close any running ppopcode instances and try again."
        Write-Info "Or manually copy: Copy-Item '$binaryPath' '$targetPath' -Force"
        exit 1
    }
    
    # Add to PATH
    Add-ToPath $INSTALL_DIR
    
    Write-Host "`n=== Installation Complete! ===" -ForegroundColor Green
    Write-Host ""
    Write-Host "To use ppopcode in your current terminal, run:" -ForegroundColor Yellow
    Write-Host '  $env:Path = [System.Environment]::GetEnvironmentVariable("Path","Machine") + ";" + [System.Environment]::GetEnvironmentVariable("Path","User")' -ForegroundColor Cyan
    Write-Host ""
    Write-Host "Or simply open a new terminal and type:" -ForegroundColor Yellow
    Write-Host "  ppopcode" -ForegroundColor Cyan
    Write-Host ""
}

function Uninstall-Ppopcode {
    Write-Host "`n=== Uninstalling ppopcode ===" -ForegroundColor Yellow
    
    # Remove binary
    $installedBinary = Join-Path $INSTALL_DIR $BINARY_NAME
    if (Test-Path $installedBinary) {
        Write-Info "Removing: $installedBinary"
        Remove-Item $installedBinary -Force
        Write-Success "Removed binary"
    } else {
        Write-Info "Binary not found: $installedBinary"
    }
    
    # Remove from PATH if directory is empty
    if (Test-Path $INSTALL_DIR) {
        $items = Get-ChildItem $INSTALL_DIR
        if ($items.Count -eq 0) {
            Write-Info "Removing empty directory: $INSTALL_DIR"
            Remove-Item $INSTALL_DIR -Force
            Remove-FromPath $INSTALL_DIR
            Write-Success "Removed directory and PATH entry"
        } else {
            Write-Info "Directory not empty, keeping: $INSTALL_DIR"
        }
    }
    
    Write-Host "`n=== Uninstallation Complete! ===" -ForegroundColor Green
}

# Main
if ($Uninstall) {
    Uninstall-Ppopcode
} else {
    Install-Ppopcode
}
