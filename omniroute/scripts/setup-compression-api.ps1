# setup-compression-api.ps1 — Aplica configuración de compresión via PUT API
# Requiere que OmniRoute esté corriendo en localhost:20128

$ErrorActionPreference = "Stop"

Write-Host "=== Configuración de Compresión via API ===" -ForegroundColor Cyan

# 1. Leer API key del .env
$envFile = "$env:USERPROFILE\.omniroute\.env"
if (-not (Test-Path $envFile)) {
    Write-Host "[ERROR] No se encontró $envFile" -ForegroundColor Red
    exit 1
}

$apiKey = (Get-Content $envFile | Where-Object { $_ -match "^OMNIROUTE_API_KEY=" }) -replace "^OMNIROUTE_API_KEY=",""
if (-not $apiKey) {
    Write-Host "[ERROR] OMNIROUTE_API_KEY no encontrada en .env" -ForegroundColor Red
    exit 1
}
Write-Host "[OK] API key cargada" -ForegroundColor Green

# 2. Verificar que OmniRoute está corriendo
try {
    $test = Invoke-RestMethod -Uri "http://localhost:20128/v1/models" -Headers @{ "Authorization" = "Bearer $apiKey" } -TimeoutSec 5
    Write-Host "[OK] OmniRoute está corriendo" -ForegroundColor Green
} catch {
    Write-Host "[ERROR] OmniRoute no está corriendo en localhost:20128" -ForegroundColor Red
    Write-Host "       Ejecutar: omniroute restart" -ForegroundColor Yellow
    exit 1
}

# 3. Aplicar configuración completa de compresión
$compressionBody = Get-Content -Path "$PSScriptRoot\..\config\compression.json" -Raw
$headers = @{
    "Authorization" = "Bearer $apiKey"
    "Content-Type" = "application/json"
}

try {
    $resp = Invoke-RestMethod -Uri "http://localhost:20128/api/settings/compression" -Method PUT -Headers $headers -Body $compressionBody
    Write-Host "[OK] Compresión configurada" -ForegroundColor Green
    Write-Host "     Pipeline: $($resp.stackedPipeline -join ' -> ')" -ForegroundColor Gray
} catch {
    Write-Host "[ERROR] Fallo al configurar compresión: $($_.Exception.Message)" -ForegroundColor Red
    exit 1
}

# 4. Aplicar settings adicionales via API (cavemanOutputMode, rtkConfig completo, preserveSystemPromptMode)
$apiBody = Get-Content -Path "$PSScriptRoot\..\config\compression-api.json" -Raw
try {
    $resp2 = Invoke-RestMethod -Uri "http://localhost:20128/api/settings/compression" -Method PUT -Headers $headers -Body $apiBody
    Write-Host "[OK] Settings adicionales aplicados" -ForegroundColor Green
    Write-Host "     cavemanOutputMode: enabled=$($resp2.cavemanOutputMode.enabled)" -ForegroundColor Gray
    Write-Host "     applyToCodeBlocks: $($resp2.rtkConfig.applyToCodeBlocks)" -ForegroundColor Gray
    Write-Host "     deduplicateThreshold: $($resp2.rtkConfig.deduplicateThreshold)" -ForegroundColor Gray
    Write-Host "     rawOutputRetention: $($resp2.rtkConfig.rawOutputRetention)" -ForegroundColor Gray
    Write-Host "     preserveSystemPromptMode: $($resp2.preserveSystemPromptMode)" -ForegroundColor Gray
} catch {
    Write-Host "[ERROR] Fallo al aplicar settings adicionales: $($_.Exception.Message)" -ForegroundColor Red
    exit 1
}

Write-Host ""
Write-Host "=== Compresión configurada exitosamente ===" -ForegroundColor Green
Write-Host "NOTA: Los safety gates (fidelityGate, riskGate, circuitBreaker) se configuran" -ForegroundColor Yellow
Write-Host "     por separado via setup-safety-gates.mjs" -ForegroundColor Yellow
