# install.ps1 - Install ostrakon from GitHub releases
# Usage: iwr -useb https://raw.githubusercontent.com/PapaDanielVi/ostrakon/main/install.ps1 | iex

param(
    [string]$InstallDir = "$env:PROGRAMDATA\ostrakon",
    [string]$Repo = "PapaDanielVi/ostrakon"
)

$BinaryName = "ostrakon"

# Detect architecture
$Arch = $env:PROCESSOR_ARCHITECTURE
switch ($Arch) {
    "AMD64" { $ReleaseArch = "x86_64" }
    "ARM64" { $ReleaseArch = "arm64" }
    default {
        Write-Error "Unsupported architecture: $Arch"
        exit 1
    }
}

# Detect latest version
Write-Host "Detecting latest version..."
$LatestVersion = (Invoke-RestMethod -Uri "https://api.github.com/repos/$Repo/releases/latest").tag_name

if (-not $LatestVersion) {
    Write-Error "Failed to detect latest version"
    exit 1
}
Write-Host "Latest version: $LatestVersion"

# Check if binary already exists
$ExistingBinary = Get-Command $BinaryName -ErrorAction SilentlyContinue
if ($ExistingBinary) {
    $CurrentVersion = & $BinaryName --version 2>$null
    Write-Host "Current installed version: $CurrentVersion"
}
Write-Host "Installing ostrakon..."

# Create temp directory
$TmpDir = [System.IO.Path]::GetTempPath() + [System.Guid]::NewGuid()
New-Item -ItemType Directory -Path $TmpDir -Force | Out-Null

try {
    # Download
    $DownloadUrl = "https://github.com/$Repo/releases/download/$LatestVersion/${BinaryName}_Windows_${ReleaseArch}.tar.gz"
    $TmpFile = "$TmpDir\ostrakon.tar.gz"
    Write-Host "Downloading from: $DownloadUrl"
    Invoke-WebRequest -Uri $DownloadUrl -OutFile $TmpFile

    # Extract using tar (Windows 10+)
    tar -xzf $TmpFile -C $TmpDir

    # Install directory
    if (-not (Test-Path $InstallDir)) {
        New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
    }

    # Copy binary
    Copy-Item "$TmpDir\$BinaryName.exe" "$InstallDir\$BinaryName.exe" -Force

    # Add to PATH if not already there
    $CurrentPath = [Environment]::GetEnvironmentVariable("Path", "Machine")
    if ($CurrentPath -notlike "*$InstallDir*") {
        [Environment]::SetEnvironmentVariable("Path", "$CurrentPath;$InstallDir", "Machine")
        Write-Host "Added $InstallDir to system PATH. Restart your terminal for changes to take effect."
    }

    # Verify installation
    Write-Host "Verifying installation..."
    & "$InstallDir\$BinaryName.exe" --version

    Write-Host "Installation complete! ostrakon is installed to $InstallDir\$BinaryName.exe"
    Write-Host "Run 'ostrakon --help' to get started."
}
finally {
    # Cleanup
    Remove-Item -Path $TmpDir -Recurse -Force -ErrorAction SilentlyContinue
}