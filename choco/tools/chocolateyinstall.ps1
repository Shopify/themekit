$ErrorActionPreference = 'Stop';
$toolsDir       = "$(Split-Path -parent $MyInvocation.MyCommand.Definition)"
$packageName    = $env:ChocolateyPackageName
$file           = "$($toolsDir)\theme.exe"
$version        = "v1.1.2"
$url            = "https://shopify-themekit.s3.amazonaws.com/$($version)/windows-386/theme.exe"
$url64          = "https://shopify-themekit.s3.amazonaws.com/$($version)/windows-amd64/theme.exe"
$checksum       = 'fc1f491247d2d6e5f3f9974f12e6b896'
$checksum64     = '6c10c267e799dd0a3bb06c2b6c29f66c'
$validExitCodes = @(0)

Get-ChocolateyWebFile `
  -PackageName $packageName `
  -FileFullPath $file `
  -Url "$url" `
  -Checksum $checksum `
  -ChecksumType "md5" `
  -Url64bit "$url64" `
  -Checksum64 $checksum64 `
  -ChecksumType64 "md5"
