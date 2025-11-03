# Chocolatey install script for tf-safe

$ErrorActionPreference = 'Stop'

$packageName = 'tf-safe'
$toolsDir = "$(Split-Path -parent $MyInvocation.MyCommand.Definition)"
$version = '1.0.0'
$url64 = "https://github.com/BIRhrt/tf-safe/releases/download/v$version/tf-safe-windows-amd64.zip"

$packageArgs = @{
  packageName   = $packageName
  unzipLocation = $toolsDir
  url64bit      = $url64
  softwareName  = 'tf-safe*'
  checksum64    = 'REPLACE_WITH_ACTUAL_CHECKSUM'
  checksumType64= 'sha256'
  validExitCodes= @(0)
}

Install-ChocolateyZipPackage @packageArgs

# Create shim for the executable
$exePath = Join-Path $toolsDir 'tf-safe.exe'
if (Test-Path $exePath) {
    Write-Host "tf-safe installed successfully to $exePath"
    
    # Test the installation
    try {
        & $exePath --version
        Write-Host "tf-safe is working correctly"
    } catch {
        Write-Warning "tf-safe installation may have issues: $_"
    }
} else {
    throw "tf-safe executable not found at $exePath"
}