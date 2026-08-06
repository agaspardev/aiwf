$ErrorActionPreference = "Stop"
$base = "C:\Users\anton\AppData\Roaming\npm\node_modules\omniroute\dist\.build\next\server\chunks"

$files = @("open-sse_0rryb_x._.js", "open-sse_1n23zok._.js")

# PATCH 1: ANTIGRAVITY/GEMINI->CLAUDE converter (module 762111)
# forces REVERSE_MAP (PascalCase->lowercase) unconditionally -> returns "read".
# Fix: reverse-lookup so lowercase "read" maps back to "Read".
$p1Old = 'i=(a=t.toolNameMap?.get(o)||o,s.REVERSE_MAP[a]??a)'
$p1New = 'i=(a=t.toolNameMap?.get(o)||o,Object.keys(s.REVERSE_MAP).find(q=>s.REVERSE_MAP[q]===a)??a)'
$p1Marker = 'Object.keys(s.REVERSE_MAP).find(q=>s.REVERSE_MAP[q]===a)'

# PATCH 2: OPENAI->CLAUDE streaming converter (module 75659, var u=e.i(786790)=REVERSE_MAP)
# emits name:a.name / name:e.name raw -> lowercase would reach Claude Code.
# Fix: reverse-lookup so lowercase "read" maps back to "Read".
$p2Old1 = 'name:a.name||""'
$p2New1 = 'name:(Object.keys(u.REVERSE_MAP).find(q=>u.REVERSE_MAP[q]===a.name)??a.name)||""'
$p2Old2 = 'name:e.name||""'
$p2New2 = 'name:(Object.keys(u.REVERSE_MAP).find(q=>u.REVERSE_MAP[q]===e.name)??e.name)||""'
$p2Marker1 = 'find(q=>u.REVERSE_MAP[q]===a.name)'
$p2Marker2 = 'find(q=>u.REVERSE_MAP[q]===e.name)'

foreach ($f in $files) {
  $p = Join-Path $base $f
  $raw = [System.IO.File]::ReadAllText($p)
  $changed = $false

  # --- Patch 1 ---
  if ($raw.Contains($p1Marker)) {
    Write-Host "PATCH1 ALREADY DONE: $f"
  } else {
    $count = ([regex]::Matches($raw, [regex]::Escape($p1Old))).Count
    if ($count -ne 1) {
      Write-Host "SKIP PATCH1 $f : pattern count = $count (expected 1)"
    } else {
      $raw = $raw.Replace($p1Old, $p1New)
      $changed = $true
      Write-Host "PATCH1 APPLIED: $f"
    }
  }

  # --- Patch 2 (two sites) ---
  $sites = @(@($p2Old1, $p2New1, $p2Marker1), @($p2Old2, $p2New2, $p2Marker2))
  foreach ($s in $sites) {
    if ($raw.Contains($s[2])) {
      Write-Host "PATCH2 site '$($s[0])' ALREADY DONE: $f"
    } else {
      $count = ([regex]::Matches($raw, [regex]::Escape($s[0]))).Count
      if ($count -ne 1) {
        Write-Host "SKIP PATCH2 site '$($s[0])' $f : pattern count = $count (expected 1)"
      } else {
        $raw = $raw.Replace($s[0], $s[1])
        $changed = $true
        Write-Host "PATCH2 site '$($s[0])' APPLIED: $f"
      }
    }
  }

  if ($changed) {
    $bak = "$p.bak-opensse-2"
    if (-not (Test-Path -LiteralPath $bak)) {
      [System.IO.File]::WriteAllText($bak, $raw)
      Write-Host "backup (current state) written: $bak"
    }
    [System.IO.File]::WriteAllText($p, $raw)
    Write-Host "FILE WRITTEN: $f"
  } else {
    Write-Host "NO CHANGES NEEDED: $f"
  }
}

Write-Host "`nDone. Validate with: node --check <file>"
