# Run this using
# (New-Object System.Net.WebClient).DownloadString("https://raw.githubusercontent.com/Shopify/themekit/master/scripts/install.ps1") | powershell -command -
$web_client = New-Object System.Net.WebClient
$latest_release_url = "https://shopify-themekit.s3.amazonaws.com/releases/latest.json"
$release = $web_client.DownloadString($latest_release_url) | ConvertFrom-Json

$destFolder = "C:\Program Files (x86)\Theme Kit"
if ([System.Environment]::Is64BitOperatingSystem) {
  $destFolder = "C:\Program Files\Theme Kit"
}
$dest = "$($destFolder)\themekit.exe"

New-Item -ItemType Directory -Force -Path $destFolder

foreach($platform in $release.platforms) {
  if ((platform.name -eq "windows-amd64" -And [System.Environment]::Is64BitOperatingSystem) -Or
    (platform.name -eq "windows-386" -And ![System.Environment]::Is64BitOperatingSystem)) {
    $web_client.DownloadFile(platform.url, $dest)
  }
}

[Environment]::SetEnvironmentVariable(
  "Path",
  "$($env:Path);$($destFolder)",
  [EnvironmentVariableTarget]::Machine
)
