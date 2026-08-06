# dr-snapshot.ps1 — Snapshot de Disaster Recovery de OmniRoute
# Crea un snapshot autocontenido y consistente del estado ACTUAL:
#   - storage.sqlite (copia online vía better-sqlite3 backup, sin detener el server)
#   - .env (con secrets necesarios para descifrar credenciales de providers)
#   - export-json (settings + combos + apiKeys como referencia/fallback)
#   - version.txt (versión del paquete npm)
#
# Uso:
#   .\scripts\dr-snapshot.ps1                    # snapshot a ~/.omniroute/dr-snapshots/latest/
#   .\scripts\dr-snapshot.ps1 -OutDir "D:\backup\omniroute"   # snapshot a otra ruta

param(
    [string]$OutDir = ""
)

$ErrorActionPreference = "Stop"
$root = "$env:USERPROFILE\.omniroute"

# --- Resolver directorio destino ---
if (-not $OutDir) {
    $OutDir = "$root\dr-snapshots\latest"
} else {
    $OutDir = Join-Path $OutDir "latest"
}
New-Item -ItemType Directory -Path $OutDir -Force | Out-Null

$stamp = Get-Date -Format "yyyyMMdd-HHmmss"
Write-Host "=== OmniRoute DR Snapshot ===" -ForegroundColor Cyan
Write-Host "Destino: $OutDir"

# --- 1. storage.sqlite (backup online consistente) ---
$srcDb = "$root\storage.sqlite"
if (Test-Path $srcDb) {
    Write-Host "[1/4] Backup de storage.sqlite..." -ForegroundColor Yellow
    $helper = Join-Path $PSScriptRoot "dr-sqlite-backup.js"
    $outDb = Join-Path $OutDir "storage.sqlite"
    $result = node $helper $srcDb $outDb 2>&1
    Write-Host "       $result"
    if ($LASTEXITCODE -ne 0) { Write-Host "[ERROR] Backup de DB falló" -ForegroundColor Red; exit 1 }
    Remove-Item -LiteralPath "$outDb-wal" -Force -ErrorAction SilentlyContinue
    Remove-Item -LiteralPath "$outDb-shm" -Force -ErrorAction SilentlyContinue
} else {
    Write-Host "[WARN] storage.sqlite no encontrado en $srcDb" -ForegroundColor Yellow
}

# --- 2. .env (con secrets) ---
Write-Host "[2/4] Copiando .env..." -ForegroundColor Yellow
$srcEnv = "$root\.env"
if (Test-Path $srcEnv) {
    Copy-Item -LiteralPath $srcEnv -Destination (Join-Path $OutDir ".env") -Force
    Write-Host "       .env copiado"
} else {
    Write-Host "[WARN] .env no encontrado" -ForegroundColor Yellow
}

# --- 3. export-json (referencia/fallback) ---
Write-Host "[3/4] Export-json via API..." -ForegroundColor Yellow
$apiKey = ""
if (Test-Path "$root\.env") {
    $envApiKey = (Get-Content "$root\.env" | Where-Object { $_ -match "^OMNIROUTE_API_KEY=" }) -replace "^OMNIROUTE_API_KEY=",""
    if ($envApiKey) { $apiKey = $envApiKey.Trim() }
}
if (-not $apiKey) {
    Write-Host "[WARN] Sin OMNIROUTE_API_KEY en .env - omitiendo export-json" -ForegroundColor Yellow
} else {
try {
    $resp = Invoke-RestMethod -Uri "http://localhost:20128/api/settings/export-json" -Headers @{ Authorization = "Bearer $apiKey" } -TimeoutSec 30
    $json = $resp | ConvertTo-Json -Depth 20
    [System.IO.File]::WriteAllText((Join-Path $OutDir "export.json"), $json, (New-Object System.Text.UTF8Encoding($false)))
    Write-Host "       export.json guardado ($([math]::Round($json.Length/1KB)) KB)"
} catch {
    Write-Host "[WARN] No se pudo exportar via API: $($_.Exception.Message)" -ForegroundColor Yellow
}
}
Write-Host "[4/4] Metadata..." -ForegroundColor Yellow
$version = "desconocida"
try { $version = (Get-Content -Raw "C:\Users\anton\AppData\Roaming\npm\node_modules\omniroute\package.json" | ConvertFrom-Json).version } catch {}
$info = @{
    exportedAt = $stamp
    version    = $version
    dbBackup   = Test-Path (Join-Path $OutDir "storage.sqlite")
    envBackup  = Test-Path (Join-Path $OutDir ".env")
} | ConvertTo-Json
[System.IO.File]::WriteAllText((Join-Path $OutDir "info.json"), $info, (New-Object System.Text.UTF8Encoding($false)))
Write-Host "       version: $version"
Write-Host ""
Write-Host "=== Snapshot completo: $OutDir ===" -ForegroundColor Green
Get-ChildItem -LiteralPath $OutDir | Select-Object Name, Length
