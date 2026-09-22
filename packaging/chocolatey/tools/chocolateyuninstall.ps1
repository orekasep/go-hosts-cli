$ErrorActionPreference = 'Stop'

$toolsDir = "$(Split-Path -parent $MyInvocation.MyCommand.Definition)"
Remove-Item -Path "$toolsDir\hostcli.exe" -Force -ErrorAction SilentlyContinue
