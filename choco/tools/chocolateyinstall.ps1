$ErrorActionPreference = 'Stop';
$toolsDir       = "$(Split-Path -parent $MyInvocation.MyCommand.Definition)"
$packageName    = $env:ChocolateyPackageName
$file           = "$($toolsDir)\theme.exe"
$version        = "v1.3.1"
$url            = "https://shopify-themekit.s3.amazonaws.com/$($version)/windows-386/theme.exe"
$url64          = "https://shopify-themekit.s3.amazonaws.com/$($version)/windows-amd64/theme.exe"
$checksum       = '3A75C7D3C25DD09F9701EB25522429B3'
$checksum64     = '38C357C696C560758CE9FA35C0447D8F'
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
