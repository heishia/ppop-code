#!/usr/bin/env pwsh
# ppopcode test runner - clean output

Write-Host ""
Write-Host "Running tests..." -ForegroundColor Cyan
Write-Host ""

$results = go test ./... 2>&1
$passed = 0
$failed = 0
$skipped = 0

foreach ($line in $results) {
    if ($line -match "^ok\s+(\S+)") {
        $pkg = $matches[1] -replace "github.com/ppopcode/ppopcode/", ""
        Write-Host "  PASS: $pkg" -ForegroundColor Green
        $passed++
    }
    elseif ($line -match "^FAIL\s+(\S+)") {
        $pkg = $matches[1] -replace "github.com/ppopcode/ppopcode/", ""
        Write-Host "  FAIL: $pkg" -ForegroundColor Red
        $failed++
    }
    elseif ($line -match "^\?\s+(\S+).*\[no test files\]") {
        $pkg = $matches[1] -replace "github.com/ppopcode/ppopcode/", ""
        Write-Host "  SKIP: $pkg (no tests)" -ForegroundColor Yellow
        $skipped++
    }
    elseif ($line -match "--- FAIL:") {
        Write-Host "    $line" -ForegroundColor Red
    }
}

Write-Host ""
Write-Host "----------------------------------------" -ForegroundColor Gray
Write-Host "Summary: " -NoNewline
Write-Host "$passed PASSED" -ForegroundColor Green -NoNewline
Write-Host ", " -NoNewline
Write-Host "$failed FAILED" -ForegroundColor $(if ($failed -gt 0) { "Red" } else { "Green" }) -NoNewline
Write-Host ", " -NoNewline
Write-Host "$skipped SKIPPED" -ForegroundColor Yellow
Write-Host "----------------------------------------" -ForegroundColor Gray

if ($failed -gt 0) {
    exit 1
}
