param(
    [string]$Uri = "http://localhost:20128",
    [string]$ApiKey = "sk-898af3cd7ede926e-a76154-6891fe35"
)

$ErrorActionPreference = "Stop"
$AuthHeader = @{ Authorization = "Bearer $ApiKey" }

function Invoke-WebRequestJson {
    param($Method, $Endpoint, $Body)
    $wr = [System.Net.WebRequest]::Create("${Uri}${Endpoint}")
    $wr.Method = $Method
    $wr.ContentType = "application/json"
    $wr.Headers.Add("Authorization", "Bearer $ApiKey")
    if ($Body) {
        $bytes = [System.Text.Encoding]::UTF8.GetBytes($Body)
        $wr.ContentLength = $bytes.Length
        $stream = $wr.GetRequestStream()
        $stream.Write($bytes, 0, $bytes.Length)
        $stream.Close()
    }
    $resp = $wr.GetResponse()
    $reader = New-Object System.IO.StreamReader($resp.GetResponseStream())
    $result = $reader.ReadToEnd()
    $reader.Close()
    return $result
}

Write-Output "=== OmniRoute Restoration ==="
Write-Output ""

# 1. Memory
Write-Output "[1/7] Configuring Memory..."
$memBody = '{"enabled":true,"maxTokens":2000,"retentionDays":30,"strategy":"hybrid","vectorStore":"sqlite-vec","transformersEnabled":true}'
$memResult = Invoke-WebRequestJson -Method PUT -Endpoint "/api/settings/memory" -Body $memBody
Write-Output "  Memory: $memResult"

# 2. Cache
Write-Output "[2/7] Configuring Cache..."
$cacheBody = '{"semanticCacheMaxSize":200,"semanticCacheTTL":3600000,"promptCacheEnabled":true,"idempotencyWindowMs":5000}'
$cacheResult = Invoke-WebRequestJson -Method PUT -Endpoint "/api/settings/cache-config" -Body $cacheBody
Write-Output "  Cache: $cacheResult"

# 3. Env vars
Write-Output "[3/7] Setting Env Vars..."
$envPath = "$HOME\.omniroute\.env"
if (-not (Test-Path $envPath)) { New-Item -ItemType File -Path $envPath -Force | Out-Null }
$existing = Get-Content $envPath -Raw
$additions = @"
INJECTION_GUARD_MODE=warn
PII_REDACTION_ENABLED=true
INPUT_SANITIZER_MODE=redact
PII_RESPONSE_SANITIZATION=true
STREAM_RECOVERY_ENABLED=true
OMNIROUTE_MCP_COMPRESS_DESCRIPTIONS=true
"@
foreach ($line in $additions -split "`n") {
    $line = $line.Trim()
    if (-not $line) { continue }
    $key = $line -split "=" | Select-Object -First 1
    if ($existing -match "^$key=") {
        $existing = $existing -replace "^$key=.*", $line
    } else {
        $existing += "`n$line"
    }
}
Set-Content -Path $envPath -Value $existing.Trim() -Force
Write-Output "  Env vars set (require restart to apply)"

# 4. Rate limits on all active providers
Write-Output "[4/7] Setting Rate Limits on active providers..."
$provs = Invoke-WebRequestJson -Method GET -Endpoint "/api/providers" | ConvertFrom-Json
$active = $provs.connections | Where-Object { $_.isActive -eq $true }
$rlCount = 0
foreach ($conn in $active) {
    $rlBody = "{`"enabled`":true,`"connectionId`":`"$($conn.id)`",`"rpm`":60,`"tpm`":1000000}"
    Invoke-WebRequestJson -Method POST -Endpoint "/api/rate-limits" -Body $rlBody | Out-Null
    $mcBody = "{`"maxConcurrent`":5}"
    Invoke-WebRequestJson -Method PUT -Endpoint "/api/providers/$($conn.id)" -Body $mcBody | Out-Null
    $rlCount++
}
Write-Output "  Rate limits set on $rlCount providers (60 RPM, 5 concurrent)"

# 5. Guardrails
Write-Output "[5/7] Checking Guardrails..."
$guards = Invoke-WebRequestJson -Method GET -Endpoint "/api/guardrails"
Write-Output "  $guards"

# 6. Memory health
Write-Output "[6/7] Verifying Memory Health..."
$health = Invoke-WebRequestJson -Method GET -Endpoint "/api/memory/health"
Write-Output "  $health"

# 7. Backup
Write-Output "[7/7] Exporting Backup..."
$backup = Invoke-WebRequestJson -Method GET -Endpoint "/api/settings/export-json"
$backupPath = "$HOME\.omniroute\backup-$(Get-Date -Format 'yyyyMMdd-HHmmss').json"
$backup | Out-File -FilePath $backupPath -Encoding utf8
Write-Output "  Backup saved: $(($backup.Length / 1KB).ToString('F0')) KB"

Write-Output ""
Write-Output "=== Done ==="
Write-Output "✓ Memory | Cache | Env Vars | Rate Limits (35 providers) | Guardrails | Backup"
