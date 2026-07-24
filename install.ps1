#!/usr/bin/env pwsh
# aiwf — install script (Windows)
# Usage: iwr -useb https://github.com/agaspardev/aiwf/releases/latest/download/install.ps1 | iex

$Repo = "agaspardev/aiwf"
$Binary = "aiwf.exe"

# Detect architecture
$Arch = switch ([Environment]::Is64BitOperatingSystem) {
    $true  { "amd64" }
    $false { "arm64" }
}

Write-Host "🔍 Detectando última versión..." -ForegroundColor Cyan
$ApiUrl = "https://api.github.com/repos/$Repo/releases/latest"
try {
    $Release = Invoke-RestMethod -Uri $ApiUrl -ErrorAction Stop
    $Version = $Release.tag_name
} catch {
    Write-Error "❌ No se pudo detectar la última versión: $_"
    exit 1
}
Write-Host "📦 Versión: $Version" -ForegroundColor Green

# Download zip
$Archive = "aiwf_${Version}_windows_${Arch}.zip"
$Url = "https://github.com/$Repo/releases/download/$Version/$Archive"
$TempZip = Join-Path $env:TEMP $Archive

Write-Host "⬇️ Descargando $Url..." -ForegroundColor Cyan
Invoke-WebRequest -Uri $Url -OutFile $TempZip -UseBasicParsing

# Extract
Write-Host "📂 Extrayendo..." -ForegroundColor Cyan
$TempDir = Join-Path $env:TEMP "aiwf_install"
Expand-Archive -Path $TempZip -DestinationPath $TempDir -Force

# Install directory
$InstallDir = if ($env:AIWF_INSTALL_DIR) {
    $env:AIWF_INSTALL_DIR
} else {
    $UserDir = [Environment]::GetFolderPath("UserProfile")
    Join-Path $UserDir ".aiwf\bin"
}

Write-Host "📁 Instalando en $InstallDir..." -ForegroundColor Cyan
New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
Move-Item -Path "$TempDir\$Binary" -Destination "$InstallDir\$Binary" -Force

# Add to PATH (user-level) if not already there
$Path = [Environment]::GetEnvironmentVariable("PATH", "User")
if ($Path -notlike "*$InstallDir*") {
    $NewPath = "$InstallDir;$Path"
    [Environment]::SetEnvironmentVariable("PATH", $NewPath, "User")
    Write-Host "🔧 $InstallDir añadido al PATH de usuario" -ForegroundColor Yellow
    Write-Host "   Reinicia la terminal o ejecuta:" -ForegroundColor Yellow
    Write-Host "   `$env:PATH = `"$InstallDir;`$env:PATH`"" -ForegroundColor Yellow
}

# Cleanup
Remove-Item -Path $TempZip -Force -ErrorAction SilentlyContinue
Remove-Item -Path $TempDir -Recurse -Force -ErrorAction SilentlyContinue

Write-Host "✅ aiwf instalado en $InstallDir\$Binary" -ForegroundColor Green
Write-Host "   Ejecuta 'aiwf doctor' para verificar" -ForegroundColor Green
