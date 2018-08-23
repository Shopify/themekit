$web_client = New-Object System.Net.WebClient
$latest_release_url = "https://shopify-themekit.s3.amazonaws.com/releases/latest.json"

Write-Output "Fetching release data";
Try {
  $release = $web_client.DownloadString($latest_release_url) | ConvertFrom-Json
} Catch {
  Write-Output "Couldn't fetch release data";
  Write-Host -foreground Yellow -background Black "Couldn't fetch release data. Check your internet connection.";
  Write-Host -foreground Red    -background Black "Error: $($PSItem.Exception.Message)";
  Exit 1
}

$destFolder = "C:\Program Files (x86)\Theme Kit"
if ([System.Environment]::Is64BitOperatingSystem) {
  $destFolder = "C:\Program Files\Theme Kit"
}
$dest = "$($destFolder)\theme.exe"

Try {
  New-Item -ItemType Directory -Force -Path $destFolder | Out-Null
} Catch {
  Write-Output "Couldn't create directory ""$($destFolder)""";
  Write-Host -foreground Yellow -background Black "Couldn't create directory ""$($destFolder)"". Make sure you have Administrator access.";
  Write-Host -foreground Red    -background Black "Error: $($PSItem.Exception.Message)";
  Exit 1
}

foreach($platform in $release.platforms) {
  if (($platform.name -eq "windows-amd64" -And [System.Environment]::Is64BitOperatingSystem) -Or
    ($platform.name -eq "windows-386" -And ![System.Environment]::Is64BitOperatingSystem)) {
    Write-Output "Downloading version $($release.version) of Shopify Themekit.";
    Try {
      $web_client.DownloadFile($platform.url, $dest)

      $hashFromFile = Get-FileHash $dest -Algorithm MD5
      if ($hashFromFile.Hash -eq $platform.digest) {
        Write-Host -ForegroundColor Green 'Validated binary checksum'
      } else {
        Write-Host -ForegroundColor Red 'Downloaded binary did not match checksum.'
      }
    } Catch {
      Write-Output "Couldn't download Shopify Themekit";
      Write-Host -foreground Yellow -background Black "Couldn't download Shopify Themekit. Check your internet connection.";
      Write-Host -foreground Red    -background Black "Error: $($PSItem.Exception.Message)";
      Exit 1
    }
  }
}

Write-Output "Setting Environment Variable";
Try {
  [Environment]::SetEnvironmentVariable(
    "Path",
    "$($env:Path);$($destFolder)",
    [EnvironmentVariableTarget]::Machine
  )
} Catch {
  Write-Output "Couldn't set environment variable";
  Write-Host -foreground Yellow -background Black "Couldn't set environment variable. Make sure you have Administrator access.";
  Write-Host -foreground Red    -background Black "Error: $($PSItem.Exception.Message)";
  Exit 1
}

Write-Output "Install Complete. Please restart your Powershell.";
