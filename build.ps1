# 编译两个可执行文件：zinc.exe（控制台版）与 zincw.exe（GUI 子系统，计划任务用）
# 用法：.\build.ps1                用 PATH 里的 go
#       .\build.ps1 -Go C:\Go\bin\go.exe
param(
    [string]$Go = 'go'
)

$ErrorActionPreference = 'Stop'
$env:GOCACHE = Join-Path $PSScriptRoot '.gocache'
New-Item -ItemType Directory -Force -Path $env:GOCACHE | Out-Null

& $Go build -C $PSScriptRoot -ldflags "-s -w" -o zinc.exe .
if ($LASTEXITCODE -ne 0) { throw 'console build failed' }

& $Go build -C $PSScriptRoot -ldflags "-s -w -H=windowsgui" -o zincw.exe .
if ($LASTEXITCODE -ne 0) { throw 'GUI build failed' }

Write-Host 'built zinc.exe and zincw.exe'
