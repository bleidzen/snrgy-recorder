$ErrorActionPreference = "Stop"

Set-Location $PSScriptRoot

$mingwPath = "C:\Users\bjorn\AppData\Local\Microsoft\WinGet\Packages\BrechtSanders.WinLibs.POSIX.UCRT_Microsoft.Winget.Source_8wekyb3d8bbwe\mingw64\bin"

$env:CGO_ENABLED = "1"
$env:CC = "$mingwPath\gcc.exe"
$env:PATH = "$mingwPath;$env:PATH"

Write-Host ""
Write-Host "  SNRGY Recorder - Build Script" -ForegroundColor Cyan
Write-Host "  ==============================" -ForegroundColor Cyan
Write-Host ""

# Determine output filename (use alternate if main is locked)
$outputFile = "snrgy-recorder.exe"
try {
    if (Test-Path $outputFile) {
        Remove-Item -Force $outputFile -ErrorAction Stop
    }
} catch {
    $outputFile = "snrgy-recorder-new.exe"
    Write-Host "Using alternate filename: $outputFile" -ForegroundColor Yellow
}

if (Test-Path "go.sum") {
    Remove-Item -Force "go.sum" -ErrorAction SilentlyContinue
}

Write-Host "Downloading dependencies..." -ForegroundColor Yellow
& "C:\Program Files\Go\bin\go.exe" mod tidy
if ($LASTEXITCODE -ne 0) {
    Write-Host "Failed to download dependencies" -ForegroundColor Red
    exit 1
}

Write-Host "Building..." -ForegroundColor Yellow
& "C:\Program Files\Go\bin\go.exe" build -ldflags="-H windowsgui -s -w" -o $outputFile .

if ($LASTEXITCODE -eq 0) {
    $size = (Get-Item $outputFile).Length / 1MB
    Write-Host ""
    Write-Host "  Build successful!" -ForegroundColor Green
    Write-Host "  Output: $outputFile ($([math]::Round($size, 1)) MB)" -ForegroundColor Green
    Write-Host ""
} else {
    Write-Host ""
    Write-Host "  Build failed!" -ForegroundColor Red
    Write-Host ""
    exit $LASTEXITCODE
}
