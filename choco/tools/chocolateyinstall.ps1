$ErrorActionPreference = 'Stop';
$toolsDir       = "$(Split-Path -parent $MyInvocation.MyCommand.Definition)"
$packageName    = $env:ChocolateyPackageName
$file           = "$($toolsDir)\theme.exe"
$version        = "v1.0.2"
$url            = "https://shopify-themekit.s3.amazonaws.com/$($version)/windows-386/theme.exe"
$url64          = "https://shopify-themekit.s3.amazonaws.com/$($version)/windows-amd64/theme.exe"
$checksum       = '16e1073b67419268430110b3775c0463'
$checksum64     = 'cb4e86ee372916f8247974a6012a3896'
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
