$ErrorActionPreference = "Stop"

$root = "C:\Users\anton\AppData\Roaming\npm\node_modules\omniroute\dist\.build\next\server\chunks"
$files = @('_0e8rngb._.js', '_0fpntkz._.js', '_0k-_ib1._.js')

$old1 = 'name:l.name||""'
$new1 = 'name:(Object.keys(c.REVERSE_MAP).find(q=>c.REVERSE_MAP[q]===l.name)??l.name)||""'
$old2 = 'name:e.name||""'
$new2 = 'name:(Object.keys(c.REVERSE_MAP).find(q=>c.REVERSE_MAP[q]===e.name)??e.name)||""'

$marker1 = 'c.REVERSE_MAP[q]===l.name'
$marker2 = 'c.REVERSE_MAP[q]===e.name'

$utf8 = New-Object System.Text.UTF8Encoding($false)

foreach ($n in $files) {
  $f = Join-Path $root $n
  $raw = [System.IO.File]::ReadAllText($f)
  $already = ($raw.Contains($marker1)) -and ($raw.Contains($marker2))
  $count1 = ([regex]::Matches($raw, [regex]::Escape($old1))).Count
  $count2 = ([regex]::Matches($raw, [regex]::Escape($old2))).Count

  if ($already) {
    "SKIP  $n  (already patched)"
    continue
  }
  if ($count1 -ne 1 -or $count2 -ne 1) {
    "ERROR $n  unexpected counts s1=$count1 s2=$count2 - aborting without changes"
    continue
  }

  $bak = "$f.bak"
  if (-not (Test-Path -LiteralPath $bak)) {
    [System.IO.File]::WriteAllText($bak, $raw, $utf8)
    "BACKUP $n -> .bak"
  }

  $patched = $raw.Replace($old1, $new1).Replace($old2, $new2)
  [System.IO.File]::WriteAllText($f, $patched, $utf8)
  $sha = (Get-FileHash -LiteralPath $f -Algorithm SHA256).Hash.Substring(0, 16)
  "PATCH  $n  s1=$count1 s2=$count2  sha=$sha"
}
