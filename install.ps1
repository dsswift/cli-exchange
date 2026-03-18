$ErrorActionPreference = "Stop"

$Repo = "dsswift/cli-exchange"
$Binary = "exchange"

# Detect architecture
$Arch = if ([Environment]::Is64BitOperatingSystem) { "amd64" } else {
    Write-Error "Unsupported architecture: 32-bit Windows is not supported"
    exit 1
}

# Get latest release tag
$Release = Invoke-RestMethod -Uri "https://api.github.com/repos/$Repo/releases/latest"
$Tag = $Release.tag_name

if (-not $Tag) {
    Write-Error "Failed to resolve latest release"
    exit 1
}

$Asset = "$Binary-windows-$Arch.exe"
$Url = "https://github.com/$Repo/releases/download/$Tag/$Asset"

# Install directory
$InstallDir = Join-Path $env:LOCALAPPDATA "Programs\exchange"
New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null

$Dest = Join-Path $InstallDir "$Binary.exe"

Write-Host "Installing $Binary $Tag (windows/$Arch)..."

# Download binary
$TmpFile = [System.IO.Path]::GetTempFileName()
try {
    Invoke-WebRequest -Uri $Url -OutFile $TmpFile -UseBasicParsing

    # Verify checksum
    $ChecksumsUrl = "https://github.com/$Repo/releases/download/$Tag/checksums.txt"
    $Checksums = (Invoke-WebRequest -Uri $ChecksumsUrl -UseBasicParsing).Content
    $Expected = ($Checksums -split "`n" | Where-Object { $_ -match $Asset } | ForEach-Object { ($_ -split "\s+")[0] })

    if ($Expected) {
        $Actual = (Get-FileHash -Path $TmpFile -Algorithm SHA256).Hash.ToLower()
        if ($Actual -ne $Expected) {
            Write-Error "Checksum mismatch`n  expected: $Expected`n  got:      $Actual"
            exit 1
        }
    }

    Move-Item -Force -Path $TmpFile -Destination $Dest
} finally {
    if (Test-Path $TmpFile) { Remove-Item $TmpFile -Force }
}

Write-Host "Installed to $Dest"

# Check PATH
$UserPath = [Environment]::GetEnvironmentVariable("PATH", "User")
if ($UserPath -notlike "*$InstallDir*") {
    [Environment]::SetEnvironmentVariable("PATH", "$InstallDir;$UserPath", "User")
    Write-Host "Added $InstallDir to your user PATH. Restart your terminal to use 'exchange'."
} else {
    Write-Host "Run 'exchange --version' to verify."
}
