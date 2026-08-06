$ErrorActionPreference = "Stop"

$root = "C:\Users\anton\AppData\Roaming\npm\node_modules\omniroute\dist\.build\next\server\chunks"
$files = @('_0e8rngb._.js', '_0fpntkz._.js', '_0k-_ib1._.js', '_1zjt2j7._.js')

$old = 'i=(l=t.toolNameMap?.get(r)||r,o.REVERSE_MAP[l]??l)'
$new = 'i=(l=t.toolNameMap?.get(r)||r,Object.keys(o.REVERSE_MAP).find(q=>o.REVERSE_MAP[q]===l)??l)'
$marker = 'o.REVERSE_MAP[q]===l'

$utf8 = New-Object System.Text.UTF8Encoding($false)

foreach ($n in $files) {
  $f = Join-Path $root $n
  $raw = [System.IO.File]::ReadAllText($f)
  if ($raw.Contains($marker)) { "SKIP  $n  (already patched)"; continue }
  $count = ([regex]::Matches($raw, [regex]::Escape($old))).Count
  if ($count -ne 1) { "ERROR $n  count=$count"; continue }
  $bak = "$f.bak"
  if (-not (Test-Path -LiteralPath $bak)) { [System.IO.File]::WriteAllText($bak, $raw, $utf8); "BACKUP $n" }
  $patched = $raw.Replace($old, $new)
  [System.IO.File]::WriteAllText($f, $patched, $utf8)
  node --check $f
  if ($?) { $sha = (Get-FileHash -LiteralPath $f -Algorithm SHA256).Hash.Substring(0,16); "PATCH  $n  sha=$sha" }
  else { "SYNTAX ERROR in $n - REVERTING"; [System.IO.File]::WriteAllText($f, $raw, $utf8) }
}
