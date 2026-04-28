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

# Install directory
$InstallDir = Join-Path $env:LOCALAPPDATA "Programs\exchange"
New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null

# Download checksums once
$ChecksumsUrl = "https://github.com/$Repo/releases/download/$Tag/checksums.txt"
$Checksums = (Invoke-WebRequest -Uri $ChecksumsUrl -UseBasicParsing).Content

function Install-Binary {
    param([string]$Name)

    $Asset = "$Name-windows-$Arch.exe"
    $Url = "https://github.com/$Repo/releases/download/$Tag/$Asset"
    $Dest = Join-Path $InstallDir "$Name.exe"

    Write-Host "Installing $Name $Tag (windows/$Arch)..."

    $TmpFile = [System.IO.Path]::GetTempFileName()
    try {
        Invoke-WebRequest -Uri $Url -OutFile $TmpFile -UseBasicParsing

        $Expected = ($Checksums -split "`n" | Where-Object { $_ -match $Asset } | ForEach-Object { ($_ -split "\s+")[0] })
        if ($Expected) {
            $Actual = (Get-FileHash -Path $TmpFile -Algorithm SHA256).Hash.ToLower()
            if ($Actual -ne $Expected) {
                Write-Error "Checksum mismatch for $Asset`n  expected: $Expected`n  got:      $Actual"
                exit 1
            }
        }

        Move-Item -Force -Path $TmpFile -Destination $Dest
    } finally {
        if (Test-Path $TmpFile) { Remove-Item $TmpFile -Force }
    }

    Write-Host "  Installed to $Dest"
}

# Download both binaries
Install-Binary $Binary
Install-Binary "$Binary-mcp"

# Check PATH
$UserPath = [Environment]::GetEnvironmentVariable("PATH", "User")
if ($UserPath -notlike "*$InstallDir*") {
    [Environment]::SetEnvironmentVariable("PATH", "$InstallDir;$UserPath", "User")
    Write-Host "Added $InstallDir to your user PATH. Restart your terminal to apply."
} else {
    Write-Host "Run 'exchange --version' to verify."
}

# Offer Claude Code setup
$ClaudeConfig = Join-Path $env:USERPROFILE ".claude.json"
if (Test-Path $ClaudeConfig) {
    $Reply = Read-Host "Claude Code detected. Register exchange MCP server? [y/N]"
    if ($Reply -eq "y" -or $Reply -eq "Y") {
        & (Join-Path $InstallDir "exchange-mcp.exe") --setup
    }
}
