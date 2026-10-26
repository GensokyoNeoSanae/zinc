# 打发布包：
#   dist\zinc-<版本>-setup.exe          Inno Setup 安装包（推荐给普通用户）
#   dist\zinc-<版本>-windows-amd64.zip  绿色版（解压即用，zinc init 轻量注册）
param(
    [string]$Version = '',
    [string]$Go = 'go',
    [string]$Iscc = ''
)

$ErrorActionPreference = 'Stop'
$root = $PSScriptRoot

if (-not $Version) {
    $Version = (& git -C $root describe --tags --always --dirty 2>$null)
    if (-not $Version) { $Version = 'dev' }
}
# 打 tag 时传进来的是 v0.1.0，文件名和程序集版本都用不带 v 的 0.1.0
if ($Version.StartsWith('v') -or $Version.StartsWith('V')) { $Version = $Version.Substring(1) }

$dist = Join-Path $root 'dist'
New-Item -ItemType Directory -Force -Path $dist | Out-Null

# 1. Inno Setup 安装包（内部会先编译四个 exe）
& (Join-Path $root 'installer\build.ps1') -Version $Version -Go $Go -Iscc $Iscc
$setup = Join-Path $dist "zinc-$Version-setup.exe"
if (-not (Test-Path $setup)) { throw "Inno Setup 没有产出 $setup" }

# 2. 绿色版 zip
$name = "zinc-$Version-windows-amd64"
$stage = Join-Path $dist $name
if (Test-Path $stage) { Remove-Item $stage -Recurse -Force }
New-Item -ItemType Directory -Force -Path $stage | Out-Null
foreach ($f in 'zinc.exe','zincw.exe','zinc-install.exe','zinc-uninstall.exe','zinc.ico','README.md','README_zh.md','LICENSE') {
    Copy-Item (Join-Path $root $f) $stage
}
$zip = Join-Path $dist "$name.zip"
if (Test-Path $zip) { Remove-Item $zip -Force }
Compress-Archive -Path (Join-Path $stage '*') -DestinationPath $zip -CompressionLevel Optimal
Remove-Item $stage -Recurse -Force

Write-Host "wrote dist\zinc-$Version-setup.exe ($((Get-Item $setup).Length) bytes)"
Write-Host "wrote dist\$name.zip ($((Get-Item $zip).Length) bytes)"
