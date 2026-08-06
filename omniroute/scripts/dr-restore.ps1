# dr-restore.ps1 — Restauración completa de OmniRoute en UN comando (Disaster Recovery)
#
# ¿Qué hace?
#   1. Instala omniroute@<Version> pineada (3.8.49) via npm
#   2. Restaura storage.sqlite desde el snapshot DR más reciente
#   3. Restaura .env (con STORAGE_ENCRYPTION_KEY y OMNIROUTE_API_KEY)
#   4. Re-aplica los patches de tool names (6 chunks)
#   5. Copia + activa los 3 plugins locales
#   6. Reinicia el server (scheduled task OmniRoute)
#   7. Ejecuta verify.ps1 (cross-check contra la config restaurada)
#
# Pre-requisitos: Node.js, npm, scheduled task "OmniRoute".
#
# Uso:
#   .\scripts\dr-restore.ps1                          # restore completo (modo interactivo)
#   .\scripts\dr-restore.ps1 -Version 3.8.49          # pinear otra versión
#   .\scripts\dr-restore.ps1 -SkipInstall            # no tocar el paquete npm
#   .\scripts\dr-restore.ps1 -SnapshotDir "D:\bk\latest"   # snapshot específico
#   .\scripts\dr-restore.ps1 -DryRun                  # ensayo (no modifica nada)

param(
    [string]$Version = "3.8.49",
    [string]$SnapshotDir = "",
    [switch]$SkipInstall,
    [switch]$SkipRestart,
    [switch]$SkipPlugins,
    [switch]$DryRun,
    [switch]$Force
)

$ErrorActionPreference = "Stop"
$root = "$env:USERPROFILE\.omniroute"
$repo = Split-Path -Parent $PSScriptRoot
$apiKey = ""
if (Test-Path "$root\.env") {
    $apiKey = ((Get-Content "$root\.env" | Where-Object { $_ -match "^OMNIROUTE_API_KEY=" }) -replace "^OMNIROUTE_API_KEY=","").Trim()
}
$serverUrl = "http://localhost:20128"

# Resolver snapshot destino
if (-not $SnapshotDir) {
    $SnapshotDir = "$root\dr-snapshots\latest"
}
if (-not (Test-Path "$SnapshotDir\info.json")) {
    Write-Host "[ERROR] Snapshot no encontrado: $SnapshotDir (ejecutar .\scripts\dr-snapshot.ps1 primero)" -ForegroundColor Red
    exit 1
}
$snapInfo = Get-Content -Raw "$SnapshotDir\info.json" | ConvertFrom-Json
Write-Host "=== OmniRoute DR Restore ===" -ForegroundColor Cyan
Write-Host "Snapshot : $SnapshotDir (version $($snapInfo.version), $($snapInfo.exportedAt))"
Write-Host "Target   : omniroute@$Version en $root"
if ($DryRun) { Write-Host "MODO DRY-RUN: no se modificará nada`n" -ForegroundColor Yellow }

function Say-Exec {
    param([string]$Title, [scriptblock]$Block)
    Write-Host "[>] $Title" -ForegroundColor Yellow
    if (-not $DryRun) { & $Block }
}

# --- 0. Requisitos ---
Write-Host "`n[0/6] Pre-requisitos..." -ForegroundColor Yellow
if (-not (Get-Command node -ErrorAction SilentlyContinue)) { Write-Host "[ERROR] Node.js no encontrado" -ForegroundColor Red; exit 1 }
Write-Host "  Node OK. version destino: $Version"

# --- 1. Instalar / verificar paquete ---
Write-Host "`n[1/6] Paquete npm..." -ForegroundColor Yellow
if ($SkipInstall) {
    Write-Host "  SkipInstall - sin tocar el paquete npm"
} else {
    Say-Exec "Instalando omniroute@$Version..." {
        npm install -g "omniroute@$Version"
        if ($LASTEXITCODE -ne 0) { throw "npm install falló" }
        $installed = (Get-Content -Raw "C:\Users\anton\AppData\Roaming\npm\node_modules\omniroute\package.json" | ConvertFrom-Json).version
        Write-Host "  Instalada version: $installed"
        if ($installed -ne $Version) { Write-Host "  [WARN] version instalada ($installed) != destino ($Version)" -ForegroundColor Yellow }
    }
}

# --- 2. Detener server (para reemplazar DB) ---
Write-Host "`n[2/6] Deteniendo server..." -ForegroundColor Yellow
if ($DryRun) {
    Write-Host "  [dry] Stop-ScheduledTask OmniRoute"
} else {
    try {
        Stop-ScheduledTask -TaskName "OmniRoute" -ErrorAction Stop
        Start-Sleep -Seconds 3
        Write-Host "  Server detenido"
    } catch {
        Write-Host "  [WARN] No se pudo detener via scheduled task: $($_.Exception.Message)" -ForegroundColor Yellow
    }
}

# --- 3. Restaurar storage.sqlite + .env ---
Write-Host "`n[3/6] Restaurando storage.sqlite y .env..." -ForegroundColor Yellow
$snapDb = "$SnapshotDir\storage.sqlite"
if (Test-Path $snapDb) {
    Say-Exec "Reemplazando storage.sqlite..." {
        Copy-Item -LiteralPath $snapDb -Destination "$root\storage.sqlite" -Force
        Remove-Item -LiteralPath "$root\storage.sqlite-wal" -Force -ErrorAction SilentlyContinue
        Remove-Item -LiteralPath "$root\storage.sqlite-shm" -Force -ErrorAction SilentlyContinue
        Write-Host "  storage.sqlite restaurado ($([math]::Round((Get-Item "$root\storage.sqlite").Length/1MB)) MB)"
    }
} else {
    Write-Host "  [WARN] No hay storage.sqlite en el snapshot - mantengo el existente" -ForegroundColor Yellow
}
$snapEnv = "$SnapshotDir\.env"
if (Test-Path $snapEnv) {
    Say-Exec "Restaurando .env..." {
        Copy-Item -LiteralPath $snapEnv -Destination "$root\.env" -Force
        Write-Host "  .env restaurado"
    }
} else {
    Write-Host "  [WARN] No hay .env en el snapshot" -ForegroundColor Yellow
}

# --- 3b. Validar que la clave descifra las credenciales de la DB restaurada ---
if (Test-Path "$root\storage.sqlite") {
    Say-Exec "Validando descifrado de credenciales (DB + clave juntas)..." {
        $credCheck = node (Join-Path $PSScriptRoot "dr-check-credentials.js") "$root\storage.sqlite" "$root\.env" 2>&1
        Write-Host "  $credCheck"
        if ($LASTEXITCODE -ne 0) {
            Write-Host "  [ERROR] Credenciales NO descifran con esta .env. Abortando (no arrancar con DB corrupta)." -ForegroundColor Red
            exit 1
        }
        Write-Host "  Credenciales OK: la DB restaurada es compatible con la clave restaurada"
    }
}

# --- 4. Patches de tool names ---
Write-Host "`n[4/6] Re-aplicando patches de tool names..." -ForegroundColor Yellow
$patchDir = Join-Path $repo "scripts\patches"
if (Test-Path $patchDir) {
    $scripts = @(
        "patch-toolcloak-v1.ps1",
        "patch-toolcloak-1zjt2j7.ps1",
        "patch-toolcloak-antigravity.ps1",
        "patch-toolcloak-opensse.ps1"
    )
    foreach ($s in $scripts) {
        $p = Join-Path $patchDir $s
        if (Test-Path $p) {
            Say-Exec "  Aplicando $s..." { & $p }
        } else {
            Write-Host "  [WARN] no existe $s" -ForegroundColor Yellow
        }
    }
} else {
    Write-Host "  [WARN] carpeta de patches no encontrada: $patchDir" -ForegroundColor Yellow
}

# --- 5. Plugins ---
if (-not $SkipPlugins) {
    Write-Host "`n[5/6] Plugins..." -ForegroundColor Yellow
    $repoPlugins = Join-Path $repo "plugins"
    if (Test-Path $repoPlugins) {
        Say-Exec "Copiando plugins del repo a ~/.omniroute/plugins/..." {
            New-Item -ItemType Directory -Path "$root\plugins" -Force | Out-Null
            Copy-Item -Path "$repoPlugins\*" -Destination "$root\plugins\" -Recurse -Force
            Write-Host "  plugins copiados"
        }
    } else {
        Write-Host "  [WARN] plugins del repo no encontrados: $repoPlugins" -ForegroundColor Yellow
    }
}

# --- 6. Reiniciar + verificar ---
Write-Host "`n[6/6] Arranque y verificación..." -ForegroundColor Yellow
if ($SkipRestart) {
    Write-Host "  SkipRestart - server no reiniciado"
} else {
    Say-Exec "Arrancando server..." {
        Start-ScheduledTask -TaskName "OmniRoute"
        Write-Host "  Scheduled task OmniRoute iniciada"
        Start-Sleep -Seconds 8
        $up = $false
        for ($i = 0; $i -lt 10; $i++) {
            if (Test-NetConnection -ComputerName localhost -Port 20128 -InformationLevel Quiet -WarningAction SilentlyContinue) { $up = $true; break }
            Start-Sleep -Seconds 2
        }
        if (-not $up) { Write-Host "  [WARN] server no responde en puerto 20128 tras $($i*2+8)s" -ForegroundColor Yellow }
        else { Write-Host "  Server UP en http://localhost:20128" }
    }
}

if (-not $SkipPlugins -and -not $SkipRestart -and -not $DryRun) {
    Say-Exec "Activando plugins (scan + activate)..." {
        & (Join-Path $repo "scripts\setup-plugins.ps1") -ApiKey $apiKey -BaseUrl $serverUrl
    }
}

Write-Host "`n=== Verificación final ===" -ForegroundColor Cyan
if ($DryRun) {
    Write-Host "  (saltada en dry-run)"
} else {
    & (Join-Path $repo "scripts\verify.ps1")
}

Write-Host "`n=== DR Restore completo ===" -ForegroundColor Green
