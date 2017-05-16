$web_client = New-Object System.Net.WebClient
$latest_release_url = "https://shopify-themekit.s3.amazonaws.com/releases/latest.json"

Write-Output "Fetching release data";
$release = $web_client.DownloadString($latest_release_url) | ConvertFrom-Json

$destFolder = "C:\Program Files (x86)\Theme Kit"
if ([System.Environment]::Is64BitOperatingSystem) {
  $destFolder = "C:\Program Files\Theme Kit"
}
$dest = "$($destFolder)\theme.exe"

New-Item -ItemType Directory -Force -Path $destFolder | Out-Null

foreach($platform in $release.platforms) {
  if (($platform.name -eq "windows-amd64" -And [System.Environment]::Is64BitOperatingSystem) -Or
    ($platform.name -eq "windows-386" -And ![System.Environment]::Is64BitOperatingSystem)) {
    Write-Output "Downloading version $($release.version) of Shopify Themekit.";
    $web_client.DownloadFile($platform.url, $dest)
  }
}

Write-Output "Setting Environment Variable";
[Environment]::SetEnvironmentVariable(
  "Path",
  "$($env:Path);$($destFolder)",
  [EnvironmentVariableTarget]::Machine
)

Write-Output "Install Complete. Please restart your Powershell.";
