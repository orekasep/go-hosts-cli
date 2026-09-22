$ErrorActionPreference = 'Stop'

$toolsDir   = "$(Split-Path -parent $MyInvocation.MyCommand.Definition)"
$version    = '1.0.0'
$url64      = "https://github.com/orekasep/go-hosts-cli/releases/download/v$version/hostcli_${version}_windows_amd64.zip"

$packageArgs = @{
  packageName   = 'hostcli'
  unzipLocation = $toolsDir
  url64bit      = $url64
  softwareName  = 'hostcli*'
  checksum64    = 'CHECKSUM_PLACEHOLDER'
  checksumType64= 'sha256'
}

Install-ChocolateyZipPackage @packageArgs
