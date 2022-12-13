$ErrorActionPreference = 'Stop';
$toolsDir       = "$(Split-Path -parent $MyInvocation.MyCommand.Definition)"
$packageName    = $env:ChocolateyPackageName
$file           = "$($toolsDir)\theme.exe"
$version        = "v1.3.1"
$url            = "https://shopify-themekit.s3.amazonaws.com/$($version)/windows-386/theme.exe"
$url64          = "https://shopify-themekit.s3.amazonaws.com/$($version)/windows-amd64/theme.exe"
$checksum       = '20292407e05210e43dd2a4ee0e660321'
$checksum64     = '8541545ada519c5a9f780b7eb05b5c02'
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
