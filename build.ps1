# ppopcode build script
param(
    [switch]$Run,
    [switch]$Clean
)

$BinaryName = "ppopcode.exe"

function Clean-Build {
    Write-Host "Cleaning..." -ForegroundColor Yellow
    if (Test-Path $BinaryName) {
        Remove-Item -Force $BinaryName
    }
    go clean
    Write-Host "✓ Clean complete" -ForegroundColor Green
}

function Build-App {
    Write-Host "Building $BinaryName..." -ForegroundColor Cyan
    
    # Clean previous build
    Clean-Build
    
    # Tidy modules
    go mod tidy
    
    # Build
    go build -o $BinaryName ./cmd/ppopcode
    
    if ($LASTEXITCODE -eq 0) {
        Write-Host "✓ Build complete: $BinaryName" -ForegroundColor Green
        return $true
    } else {
        Write-Host "✗ Build failed" -ForegroundColor Red
        return $false
    }
}

# Main
if ($Clean) {
    Clean-Build
} else {
    $success = Build-App
    if ($Run -and $success) {
        Write-Host "`nRunning $BinaryName..." -ForegroundColor Cyan
        & ".\$BinaryName"
    }
}
