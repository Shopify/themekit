$ErrorActionPreference = 'Stop';
$toolsDir       = "$(Split-Path -parent $MyInvocation.MyCommand.Definition)"
$packageName    = $env:ChocolateyPackageName
$file           = "$($toolsDir)\theme.exe"
$version        = "v1.1.1"
$url            = "https://shopify-themekit.s3.amazonaws.com/$($version)/windows-386/theme.exe"
$url64          = "https://shopify-themekit.s3.amazonaws.com/$($version)/windows-amd64/theme.exe"
$checksum       = 'cee0ece1e248ce7e9d4f340045dcdbb3'
$checksum64     = '3a1fb85c06922a842d6cec841c302e3a'
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
