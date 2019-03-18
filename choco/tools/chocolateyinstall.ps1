$ErrorActionPreference = 'Stop';
$packageName = $env:ChocolateyPackageName
$toolsDir    = "$(Split-Path -parent $MyInvocation.MyCommand.Definition)"
$url         = 'https://shopify-themekit.s3.amazonaws.com/v1.0.0/windows-386/theme.exe'
$url64       = 'https://shopify-themekit.s3.amazonaws.com/v1.0.0/windows-amd64/theme.exe'
$checksum    = 'd005b0d4538257036b435b35029b4d07'
$checksum64  = '264ea64005bb8a3e56189430a7a166e3'

Get-ChocolateyWebFile -PackageName $packageName -FileFullPath "$toolsDir\themekit.exe" -Url $url --Url64bit $url64 -Checksum $checksum -ChecksumType "md5" -Checksum64 $checksum64 -ChecksumType64 "md5"
