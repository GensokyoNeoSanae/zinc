# 编译安装向导与卸载向导（C# WinForms，用系统自带的 .NET Framework 编译器，不需要联网）
# 产物：
#   installer\build\uninstall.exe   卸载向导（安装时复制进程序目录）
#   installer\build\setup.exe       安装向导（内嵌 payload.zip / uninstall.exe / LICENSE）
param(
    [string]$Version = '0.1.0',
    [string]$Go = 'go',
    [string]$Csc = 'C:\Windows\Microsoft.NET\Framework64\v4.0.30319\csc.exe'
)

$ErrorActionPreference = 'Stop'
$root = Split-Path -Parent $PSScriptRoot
$build = Join-Path $PSScriptRoot 'build'
$ico = Join-Path $root 'zinc.ico'
$license = Join-Path $root 'LICENSE'

if (-not (Test-Path $Csc)) { throw "找不到 C# 编译器: $Csc" }

# 1. 先出主程序的四个 exe
& (Join-Path $root 'build.ps1') -Go $Go

# 2. 准备 payload（安装时原样释放到程序目录）
if (Test-Path $build) { Remove-Item $build -Recurse -Force }
New-Item -ItemType Directory -Force -Path $build | Out-Null
$stage = Join-Path $build 'payload'
New-Item -ItemType Directory -Force -Path $stage | Out-Null
foreach ($f in 'zinc.exe','zincw.exe','zinc-install.exe','zinc-uninstall.exe','zinc.ico','README.md','README_zh.md','LICENSE') {
    Copy-Item (Join-Path $root $f) $stage
}
$payload = Join-Path $build 'payload.zip'
Compress-Archive -Path (Join-Path $stage '*') -DestinationPath $payload -CompressionLevel Optimal

# 3. 版本与程序集信息。0.1.0 -> 0.1.0.0
$parts = $Version.Split('.')
while ($parts.Length -lt 4) { $parts += '0' }
$asmVersion = ($parts[0..3] -join '.') -replace '[^0-9.]', ''  
$versionCs = Join-Path $build 'Version.cs'
@"
using System.Reflection;

[assembly: AssemblyTitle("zinc Setup")]
[assembly: AssemblyProduct("zinc")]
[assembly: AssemblyCompany("GensokyoNeoSanae")]
[assembly: AssemblyCopyright("MIT License")]
[assembly: AssemblyVersion("$asmVersion")]
[assembly: AssemblyFileVersion("$asmVersion")]

namespace ZincSetup
{
    internal static class Build
    {
        public const string Version = "$Version";
    }
}
"@ | Set-Content -Path $versionCs -Encoding utf8

# 4. 卸载向导（先编译，再嵌进安装程序）
$uninstaller = Join-Path $build 'uninstall.exe'
$common = @(
    '/nologo', '/target:winexe', '/platform:anycpu', '/optimize+',
    "/win32icon:$ico",
    '/r:System.dll', '/r:System.Core.dll', '/r:System.Drawing.dll', '/r:System.Windows.Forms.dll'
)
& $Csc @common "/out:$uninstaller" (Join-Path $PSScriptRoot 'Uninstall.cs') $versionCs
if ($LASTEXITCODE -ne 0) { throw 'uninstaller build failed' }

# 5. 安装向导（内嵌 payload、卸载器和许可证文本）
$setup = Join-Path $build 'setup.exe'
$licenseTxt = Join-Path $build 'license.txt'
Copy-Item $license $licenseTxt
& $Csc @common '/r:System.IO.Compression.dll' '/r:System.IO.Compression.FileSystem.dll' `
    "/out:$setup" `
    "/resource:$payload,payload.zip" `
    "/resource:$uninstaller,uninstall.exe" `
    "/resource:$licenseTxt,license.txt" `
    (Join-Path $PSScriptRoot 'Setup.cs') $versionCs
if ($LASTEXITCODE -ne 0) { throw 'setup build failed' }

Write-Host "built installer\build\setup.exe ($((Get-Item $setup).Length) bytes)"
Write-Host "built installer\build\uninstall.exe ($((Get-Item $uninstaller).Length) bytes)"
