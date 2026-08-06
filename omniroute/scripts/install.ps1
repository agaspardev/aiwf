# install.ps1 — Instalación limpia de OmniRoute en Windows
# Ejecutar como administrador si es necesario

Write-Host "=== OmniRoute Installation ===" -ForegroundColor Cyan

# 1. Verificar Node.js
$nodeVersion = node --version 2>$null
if (-not $nodeVersion) {
    Write-Host "[ERROR] Node.js no encontrado. Instalar desde https://nodejs.org" -ForegroundColor Red
    exit 1
}
Write-Host "[OK] Node.js $nodeVersion" -ForegroundColor Green

# 2. Instalar OmniRoute globalmente
Write-Host "Instalando OmniRoute..." -ForegroundColor Yellow
npm install -g omniroute
if ($LASTEXITCODE -ne 0) {
    Write-Host "[ERROR] Fallo al instalar OmniRoute" -ForegroundColor Red
    exit 1
}
Write-Host "[OK] OmniRoute instalado" -ForegroundColor Green

# 3. Verificar versión
$version = omniroute --version 2>$null
Write-Host "[OK] OmniRoute version: $version" -ForegroundColor Green

# 4. Crear directorio de configuración
$omnirouteDir = "$env:USERPROFILE\.omniroute"
if (-not (Test-Path $omnirouteDir)) {
    New-Item -ItemType Directory -Path $omnirouteDir -Force | Out-Null
    Write-Host "[OK] Directorio creado: $omnirouteDir" -ForegroundColor Green
} else {
    Write-Host "[OK] Directorio existe: $omnirouteDir" -ForegroundColor Green
}

# 5. Verificar si .env existe
$envFile = "$omnirouteDir\.env"
if (-not (Test-Path $envFile)) {
    Write-Host "[WARN] Archivo .env no encontrado en $envFile" -ForegroundColor Yellow
    Write-Host "       Copiar desde config/env.example y llenar con credenciales reales" -ForegroundColor Yellow
} else {
    Write-Host "[OK] Archivo .env encontrado" -ForegroundColor Green
}

# 6. Verificar si storage.sqlite existe
$dbFile = "$omnirouteDir\storage.sqlite"
if (-not (Test-Path $dbFile)) {
    Write-Host "[WARN] storage.sqlite no encontrado (se crea al iniciar OmniRoute)" -ForegroundColor Yellow
} else {
    $dbSize = (Get-Item $dbFile).Length / 1MB
    Write-Host "[OK] storage.sqlite encontrado ($([math]::Round($dbSize, 2)) MB)" -ForegroundColor Green
}

Write-Host ""
Write-Host "=== Siguiente paso ===" -ForegroundColor Cyan
Write-Host "1. Copiar config/env.example a $envFile" -ForegroundColor White
Write-Host "2. Llenar las credenciales reales en .env" -ForegroundColor White
Write-Host "3. Ejecutar: omniroute restart" -ForegroundColor White
Write-Host "4. Ejecutar: .\scripts\setup-compression-api.ps1" -ForegroundColor White
Write-Host "5. Ejecutar: node scripts\setup-safety-gates.mjs" -ForegroundColor White
Write-Host "6. Ejecutar: omniroute restart" -ForegroundColor White
Write-Host "7. Ejecutar: .\scripts\verify.ps1" -ForegroundColor White
