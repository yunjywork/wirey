# Wirey Windows Build & Package Script
# Usage: .\scripts\build-windows.ps1 [version]
# Example: .\scripts\build-windows.ps1 v1.0.1

param(
    [string]$Version
)

$ErrorActionPreference = "Stop"

# Project root
$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$ProjectRoot = Split-Path -Parent $ScriptDir
$BuildDir = Join-Path $ProjectRoot "build\bin"
$DistDir = Join-Path $ProjectRoot "dist"

# Version from argument or git describe
if (-not $Version) {
    $Version = git describe --tags --always 2>$null
    if (-not $Version) { $Version = "dev" }
}

Write-Host "=== Wirey Windows Build Script ===" -ForegroundColor Yellow
Write-Host "Version: $Version" -ForegroundColor Green
Write-Host "Project: $ProjectRoot"
Write-Host ""

# Checkout version if explicitly provided as argument
if ($PSBoundParameters.ContainsKey('Version') -and $Version -match "^v\d") {
    Write-Host "Fetching tags..." -ForegroundColor Yellow
    git fetch --tags

    Write-Host "Checking out $Version..." -ForegroundColor Yellow
    git checkout $Version
    Write-Host ""
}

# Build
Write-Host "Building for windows/amd64..." -ForegroundColor Yellow
Push-Location $ProjectRoot
try {
    wails build -platform windows/amd64
    if ($LASTEXITCODE -ne 0) {
        throw "Build failed!"
    }
} finally {
    Pop-Location
}

Write-Host "Build complete!" -ForegroundColor Green
Write-Host ""

# Create dist directory
if (-not (Test-Path $DistDir)) {
    New-Item -ItemType Directory -Path $DistDir | Out-Null
}

# Create ZIP
Write-Host "Creating ZIP..." -ForegroundColor Yellow
$ZipName = "wirey-x64.zip"
$ZipPath = Join-Path $DistDir $ZipName
$ExePath = Join-Path $BuildDir "Wirey.exe"

# Remove old ZIP if exists
if (Test-Path $ZipPath) {
    Remove-Item $ZipPath -Force
}

Compress-Archive -Path $ExePath -DestinationPath $ZipPath -Force

Write-Host ""
Write-Host "=== Build Complete ===" -ForegroundColor Green
Write-Host "Output: $ZipPath"
Get-Item $ZipPath | Select-Object Name, Length
