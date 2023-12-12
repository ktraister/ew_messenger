# Check if running as administrator
if (-not ([Security.Principal.WindowsPrincipal] [Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole] "Administrator")) {
    Write-Host "Script must be run as administrator for correct installation permissions"
    exit 1
}

# Download the remote installer zip
Invoke-WebRequest -Uri "https://endless-waltz-xyz-downloads.s3.us-east-2.amazonaws.com/ew_messenger_win.zip" -OutFile "$env:TEMP\ew_messenger.zip"

# Unzip it to $env:TEMP
Expand-Archive -Path "$env:TEMP\ew_messenger.zip" -DestinationPath "$env:TEMP\\"

# Create the correct location

# Copy files into the correct locations
Move-Item -Path "$env:TEMP\ew_messenger.exe" -Destination "C:\Program Files\endlesswaltz"
Copy-Item -Path "$env:TEMP\shortcuts\Icon.ico" -Destination "C:\Program Files\endlesswaltz\Icon.ico"

# Loop through user profiles and copy desktop shortcut
foreach ($userProfile in Get-WmiObject Win32_UserProfile | Where-Object { $_.Special -eq $false }) {
    $desktopPath = Join-Path $userProfile.LocalPath "Desktop"
    
    if (Test-Path $desktopPath) {
        Copy-Item -Path "$env:TEMP\shortcuts\Endless Waltz Messenger.lnk" -Destination $desktopPath
        $desktopShortcutPath = Join-Path $desktopPath "Endless Waltz Messenger.lnk"
        Set-ItemProperty -Path $desktopShortcutPath -Name IsReadOnly -Value $false
    }
}

# Modify file permissions
Set-ExecutionPolicy -Scope Process -ExecutionPolicy Bypass
Set-ItemProperty -Path "C:\Program Files\endlesswaltz\ew_messenger.exe" -Name IsReadOnly -Value $false
Set-ItemProperty -Path "C:\Program Files\endlesswaltz\Icon.ico" -Name IsReadOnly -Value $false

