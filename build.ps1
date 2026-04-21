# Cross-build yandex-speed-cli (requires Go 1.21+).
# x64 = amd64, x86 = 386. macOS: amd64 + arm64 only (no darwin/386 in Go).
$ErrorActionPreference = "Stop"
Set-Location $PSScriptRoot
New-Item -ItemType Directory -Force dist | Out-Null
$env:CGO_ENABLED = "0"
$v = if ($env:VERSION) { $env:VERSION } else { "dev" }
$ldx = "-X main.version=$v"
@(
    @{ GOOS = "linux";   GOARCH = "amd64";  Ext = "" },
    @{ GOOS = "linux";   GOARCH = "386";    Ext = "" },
    @{ GOOS = "linux";   GOARCH = "arm64";  Ext = "" },
    @{ GOOS = "windows"; GOARCH = "amd64";  Ext = ".exe" },
    @{ GOOS = "windows"; GOARCH = "386";    Ext = ".exe" },
    @{ GOOS = "darwin";  GOARCH = "amd64";  Ext = "" },
    @{ GOOS = "darwin";  GOARCH = "arm64";  Ext = "" }
) | ForEach-Object {
    $env:GOOS = $_.GOOS
    $env:GOARCH = $_.GOARCH
    $out = "dist/yandex-speed-cli-$($_.GOOS)-$($_.GOARCH)$($_.Ext)"
    Write-Host "==> $out"
    go build -trimpath -ldflags="-s -w $ldx" -o $out .
}
Write-Host "Done: $PSScriptRoot/dist/"
