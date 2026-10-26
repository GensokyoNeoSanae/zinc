# 打发布包：编译四个 exe，连同图标和文档装进 dist\zinc-<版本>-windows-amd64.zip
# 用法：.\package.ps1                      版本号取自 git describe
#       .\package.ps1 -Version 0.1.0
param(
    [string]$Version = '',
    [string]$Go = 'go'
)

$ErrorActionPreference = 'Stop'
$root = $PSScriptRoot

if (-not $Version) {
    $Version = (& git -C $root describe --tags --always --dirty 2>$null)
    if (-not $Version) { $Version = 'dev' }
}

& (Join-Path $root 'build.ps1') -Go $Go

$name = "zinc-$Version-windows-amd64"
$dist = Join-Path $root 'dist'
$stage = Join-Path $dist $name
if (Test-Path $stage) { Remove-Item $stage -Recurse -Force }
New-Item -ItemType Directory -Force -Path $stage | Out-Null

$files = @(
    'zinc.exe'
    'zincw.exe'
    'zinc-install.exe'
    'zinc-uninstall.exe'
    'zinc.ico'
    'README.md'
    'README_zh.md'
    'LICENSE'
)
foreach ($f in $files) {
    Copy-Item -Path (Join-Path $root $f) -Destination $stage
}

$zip = Join-Path $dist "$name.zip"
if (Test-Path $zip) { Remove-Item $zip -Force }
Compress-Archive -Path (Join-Path $stage '*') -DestinationPath $zip -CompressionLevel Optimal

Write-Host "wrote dist\$name.zip ($((Get-Item $zip).Length) bytes)"
