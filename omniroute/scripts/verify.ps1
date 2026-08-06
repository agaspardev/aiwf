param(
    [string]$Uri = "http://localhost:20128",
    [string]$ApiKey = "sk-898af3cd7ede926e-a76154-6891fe35"
)

$passed = 0
$failed = 0

function Check {
    param($Name, $ScriptBlock)
    try {
        $result = & $ScriptBlock
        if ($result) {
            Write-Output "  [PASS] $Name"
            $script:passed++
        } else {
            Write-Output "  [FAIL] $Name"
            $script:failed++
        }
    } catch {
        Write-Output "  [FAIL] $Name -- $($_.Exception.Message)"
        $script:failed++
    }
}

function Invoke-OGET {
    param($Endpoint)
    $wr = [System.Net.WebRequest]::Create("${Uri}${Endpoint}")
    $wr.Method = "GET"
    $wr.Headers.Add("Authorization", "Bearer $ApiKey")
    $resp = $wr.GetResponse()
    $reader = New-Object System.IO.StreamReader($resp.GetResponseStream())
    $data = $reader.ReadToEnd() | ConvertFrom-Json
    $reader.Close()
    return $data
}

Write-Output "=== OmniRoute Verification ==="
Write-Output ""

# 1. Server running
Check "Server running on port 20128" { 
    $c = Get-NetTCPConnection -LocalPort 20128 -ErrorAction SilentlyContinue
    if ($c) { $true } else { $false }
}

# 2-4. Memory
$mem = Invoke-OGET "/api/settings/memory"
Check "Memory enabled == true" { $mem.enabled -eq $true }
Check "Memory strategy == hybrid" { $mem.strategy -eq "hybrid" }

$health = Invoke-OGET "/api/memory/health"
Check "Memory health working == true" { $health.working -eq $true }

$eng = Invoke-OGET "/api/memory/engine-status"
Check "FTS5 available" { $eng.keyword.available -eq $true }
Check "Embeddings available" { $eng.embedding.available -eq $true }
Check "VectorStore available" { $eng.vectorStore.available -eq $true }
Check "VectorStore backend == sqlite-vec" { $eng.vectorStore.backend -eq "sqlite-vec" }

# 5-7. Cache
$cache = Invoke-OGET "/api/settings/cache-config"
Check "Cache maxSize == 200" { $cache.semanticCacheMaxSize -eq 200 }
Check "Cache TTL == 3600000" { $cache.semanticCacheTTL -eq 3600000 }
Check "Prompt cache enabled == true" { $cache.promptCacheEnabled -eq $true }

# 8. Guardrails
$guards = Invoke-OGET "/api/guardrails"
Check "Guardrails prompt-injection registered" { ($guards.guardrails.name -contains "prompt-injection") }
Check "Guardrails pii-masker registered" { ($guards.guardrails.name -contains "pii-masker") }
Check "Guardrails vision-bridge registered" { ($guards.guardrails.name -contains "vision-bridge") }

# 9. Feature flags
$flagsData = Invoke-OGET "/api/settings/feature-flags"
$flagInj = $flagsData.flags | Where-Object { $_.key -eq "INJECTION_GUARD_MODE" }
$flagPii = $flagsData.flags | Where-Object { $_.key -eq "PII_REDACTION_ENABLED" }

Write-Output "  [NOTE] INJECTION_GUARD_MODE flag=$($flagInj.effectiveValue) (system not tracked; process.env reads .env directly)"
Write-Output "  [NOTE] PII_REDACTION_ENABLED flag=$($flagPii.effectiveValue) (system not tracked; process.env reads .env directly)"

# 13. Rate limits - count protected connections
$provsData = Invoke-OGET "/api/providers"
$protectedCount = ($provsData.connections | Where-Object { $_.rateLimitProtection -eq $true }).Count
$concurrentCount = ($provsData.connections | Where-Object { $_.maxConcurrent -gt 0 }).Count
Write-Output "  [INFO] rateLimitProtection=true: $protectedCount providers (of $($provsData.connections.Count))"
Write-Output "  [INFO] maxConcurrent>0: $concurrentCount providers"

Write-Output ""
Write-Output "Passed: $passed / $($passed+$failed)"
Write-Output "Failed: $failed / $($passed+$failed)"
if ($failed -gt 0) { exit 1 }
