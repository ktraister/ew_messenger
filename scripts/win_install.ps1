# REQUIRED -- https:/go.microsoft.com/fwlink/?LinkID=135170

# Check if running as administrator
if (-not ([Security.Principal.WindowsPrincipal] [Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole] "Administrator")) {
    Write-Host "Script must be run as administrator for correct installation permissions"
    exit 1
}

# Check if file exists
if (Test-Path "$env:TEMP\ew_messenger.zip") {
     Remove-Item -Recurse -Force "$env:TEMP\ew_messenger.zip"
}
if (Test-Path "$env:TEMP\ew_messenger.exe") {
     Remove-Item -Recurse -Force "$env:TEMP\ew_messenger.exe"
}
if (Test-Path "$env:TEMP\shortcuts") {
     Remove-Item -Recurse -Force "$env:TEMP\shortcuts"
}
# Download the remote installer zip
Invoke-WebRequest -Uri "https://endless-waltz-xyz-downloads.s3.us-east-2.amazonaws.com/ew_messenger_win.zip" -OutFile "$env:TEMP\ew_messenger.zip"
# Unzip it to $env:TEMP
Expand-Archive -Path "$env:TEMP\ew_messenger.zip" -DestinationPath "$env:TEMP\\"

# Create the correct location
if (-Not ( Test-Path "C:\Program Files\EndlessWaltz" ) ) {
New-Item -Path "C:\Program Files" -Name "EndlessWaltz" -ItemType "directory"
}
if (-Not ( Test-Path "C:\Program Files\EndlessWaltz\Icon" ) ) {
New-Item -Path "C:\Program Files\EndlessWaltz" -Name "Icon" -ItemType "directory"
}

# check if current files exist and remove
if (Test-Path "C:\Program Files\EndlessWaltz\ew_messenger.exe" ) {
     Remove-Item -Recurse -Force "C:\Program Files\EndlessWaltz\ew_messenger.exe"
}
if (Test-Path "C:\Program Files\EndlessWaltz\Icon\Icon.ico" ) {
     Remove-Item -Recurse -Force "C:\Program Files\EndlessWaltz\Icon\Icon.ico"
}

# Copy files into the correct locations
Move-Item -Path "$env:TEMP\ew_messenger.exe" -Destination "C:\Program Files\EndlessWaltz\ew_messenger.exe"
Move-Item -Path "$env:TEMP\shortcuts\Icon.ico" -Destination "C:\Program Files\EndlessWaltz\Icon\Icon.ico"

# Loop through user profiles and copy desktop shortcut
foreach ($userProfile in Get-WmiObject Win32_UserProfile | Where-Object { $_.Special -eq $false }) {
    $desktopPath = Join-Path $userProfile.LocalPath "Desktop"
    $desktopShortcutPath = Join-Path $desktopPath "Endless Waltz Messenger.lnk"
    if ((Test-Path $desktopPath) -and (-not (Test-Path $desktopShortcutPath) )) {
        Copy-Item -Path "$env:TEMP\shortcuts\Endless Waltz Messenger.lnk" -Destination $desktopPath
        Set-ItemProperty -Path $desktopShortcutPath -Name IsReadOnly -Value $false
    }
}

# Modify file permissions
Set-ExecutionPolicy -Scope Process -ExecutionPolicy Bypass
Set-ItemProperty -Path "C:\Program Files\EndlessWaltz\ew_messenger.exe" -Name IsReadOnly -Value $false
Set-ItemProperty -Path "C:\Program Files\EndlessWaltz\Icon\Icon.ico" -Name IsReadOnly -Value $false

#let the end user know all is well
Write-Host
Write-Host "Endless Waltz messenger has been installed at the latest available version!"
Write-Host
