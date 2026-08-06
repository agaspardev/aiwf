# setup-plugins.ps1 — Escanea y activa plugins locales en OmniRoute
# Ejecutar después de copiar plugins a ~/.omniroute/plugins/

param(
    [string]$ApiKey = "",
    [string]$BaseUrl = "http://localhost:20128"
)

$ErrorActionPreference = "Stop"

if (-not $ApiKey) {
    $envFile = "$env:USERPROFILE\.omniroute\.env"
    if (Test-Path $envFile) {
        $ApiKey = (Get-Content $envFile | Where-Object { $_ -match "^OMNIROUTE_API_KEY=" }) -replace "^OMNIROUTE_API_KEY=",""
    }
}
if (-not $ApiKey) {
    Write-Host "[ERROR] OMNIROUTE_API_KEY no encontrada. Pasar -ApiKey o configurar .env" -ForegroundColor Red
    exit 1
}

$headers = @{ Authorization = "Bearer $ApiKey"; "Content-Type" = "application/json" }

# 1. Verificar que OmniRoute está corriendo
try {
    $null = Invoke-RestMethod -Uri "$BaseUrl/api/plugins/marketplace" -Method GET -Headers $headers -ErrorAction Stop
} catch {
    Write-Host "[ERROR] OmniRoute no responde en $BaseUrl" -ForegroundColor Red
    exit 1
}

# 2. Scan
Write-Host "[1/3] Escaneando plugins en ~/.omniroute/plugins/..." -ForegroundColor Yellow
$scan = Invoke-RestMethod -Uri "$BaseUrl/api/plugins/scan" -Method POST -Headers $headers -ErrorAction Stop
Write-Host "       Descubiertos: $($scan.discovered), Errores: $($scan.errors.Count)" -ForegroundColor $(if ($scan.errors.Count -eq 0) { "Green" } else { "Red" })

if ($scan.errors.Count -gt 0) {
    foreach ($e in $scan.errors) {
        Write-Host "       [ERROR] $($e.name): $($e.error)" -ForegroundColor Red
    }
}

Start-Sleep -Seconds 1

# 3. Activar cada plugin
Write-Host "[2/3] Activando plugins..." -ForegroundColor Yellow
$plugins = Invoke-RestMethod -Uri "$BaseUrl/api/plugins" -Method GET -Headers $headers -ErrorAction SilentlyContinue
if (-not $plugins) {
    # Fallback: leer DB directamente
    $dbPlugins = node -e "
        const Database = require('C:\\Users\\anton\\AppData\\Roaming\\npm\\node_modules\\omniroute\\node_modules\\better-sqlite3');
        const db = new Database('$env:USERPROFILE\\.omniroute\\storage.sqlite', {readonly:true});
        const rows = db.prepare('SELECT name FROM plugins WHERE status = \'installed\'').all();
        console.log(JSON.stringify(rows));
        db.close();
    " 2>$null | ConvertFrom-Json
    $pluginNames = $dbPlugins | ForEach-Object { $_.name }
} else {
    $pluginNames = $plugins | ForEach-Object { $_.name }
}

foreach ($name in $pluginNames) {
    try {
        $uri = "$BaseUrl/api/plugins/$name/activate"
        $wr = [System.Net.WebRequest]::Create($uri)
        $wr.Method = "POST"
        $wr.ContentType = "application/json"
        $wr.Headers.Add("Authorization", "Bearer $ApiKey")
        $bodyBytes = [System.Text.Encoding]::UTF8.GetBytes("{}")
        $wr.ContentLength = $bodyBytes.Length
        $ws = $wr.GetRequestStream()
        $ws.Write($bodyBytes, 0, $bodyBytes.Length)
        $ws.Close()
        $resp = $wr.GetResponse()
        $reader = New-Object System.IO.StreamReader($resp.GetResponseStream())
        $result = $reader.ReadToEnd()
        $reader.Close()
        Write-Host "       [OK] $name activado" -ForegroundColor Green
    } catch {
        $sc = $_.Exception.Response.StatusCode.value__
        Write-Host "       [WARN] $name no se activó (HTTP $sc) — se cargará en próximo restart" -ForegroundColor Yellow
    }
    Start-Sleep -Milliseconds 500
}

# 4. Verificar estado final
Write-Host "[3/3] Estado final de plugins..." -ForegroundColor Yellow
$final = node -e "
    const Database = require('C:\\Users\\anton\\AppData\\Roaming\\npm\\node_modules\\omniroute\\node_modules\\better-sqlite3');
    const db = new Database('$env:USERPROFILE\\.omniroute\\storage.sqlite', {readonly:true});
    const rows = db.prepare('SELECT name, status, enabled FROM plugins ORDER BY name').all();
    console.log(JSON.stringify(rows, null, 2));
    db.close();
" 2>&1
Write-Output $final
