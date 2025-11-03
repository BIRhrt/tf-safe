# Chocolatey uninstall script for tf-safe

$ErrorActionPreference = 'Stop'

$packageName = 'tf-safe'
$toolsDir = "$(Split-Path -parent $MyInvocation.MyCommand.Definition)"

# Remove the executable
$exePath = Join-Path $toolsDir 'tf-safe.exe'
if (Test-Path $exePath) {
    Remove-Item $exePath -Force
    Write-Host "Removed tf-safe executable"
}

Write-Host "tf-safe has been uninstalled successfully"