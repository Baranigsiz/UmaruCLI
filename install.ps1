# Umaru CLI - Windows PowerShell Installer
# Usage:
#   irm https://raw.githubusercontent.com/Baranigsiz/UmaruCLI/main/install.ps1 | iex

$ErrorActionPreference = "Stop"

function Write-Banner {
    Write-Host ""
    Write-Host "  ================================================================" -ForegroundColor Magenta
    Write-Host "     UMARU CLI - Automatic Windows Installer" -ForegroundColor Cyan
    Write-Host "     Production Scaffolding in Milliseconds" -ForegroundColor DarkGray
    Write-Host "  ================================================================" -ForegroundColor Magenta
    Write-Host ""
}

function Get-Arch {
    if ([System.Environment]::Is64BitOperatingSystem) {
        if ($env:PROCESSOR_ARCHITECTURE -eq "ARM64") {
            return "arm64"
        }
        return "amd64"
    }
    throw "Unsupported Windows architecture: 32-bit is not supported by Umaru CLI."
}

function Get-LatestTag {
    $repo = "Baranigsiz/UmaruCLI"
    $url = "https://api.github.com/repos/$repo/releases/latest"
    try {
        $headers = @{ "User-Agent" = "Umaru-Installer" }
        $release = Invoke-RestMethod -Uri $url -Headers $headers -TimeoutSec 10
        if ($release.tag_name) {
            return $release.tag_name
        }
    } catch {
        Write-Host "  [!] Could not query GitHub API for latest release, falling back to v1.8.0" -ForegroundColor Yellow
    }
    return "v1.8.0"
}

try {
    Write-Banner

    $arch = Get-Arch
    Write-Host "  [>] Detected Architecture: Windows / $arch" -ForegroundColor Cyan

    Write-Host "  [>] Fetching latest release information..." -ForegroundColor DarkGray
    $tag = Get-LatestTag
    $cleanVer = $tag.TrimStart("v")

    Write-Host "  [>] Target Version: $tag" -ForegroundColor Green

    $assetName = "umaru_" + $cleanVer + "_windows_" + $arch + ".zip"
    $downloadUrl = "https://github.com/Baranigsiz/UmaruCLI/releases/download/$tag/$assetName"

    $tempZip = Join-Path $env:TEMP $assetName
    $tempExtractDir = Join-Path $env:TEMP ("umaru-extract-" + $cleanVer)

    Write-Host "  [>] Downloading $assetName..." -ForegroundColor Cyan
    [Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12
    Invoke-WebRequest -Uri $downloadUrl -OutFile $tempZip -UseBasicParsing

    Write-Host "  [>] Extracting binary archive..." -ForegroundColor DarkGray
    if (Test-Path $tempExtractDir) {
        Remove-Item -Recurse -Force $tempExtractDir | Out-Null
    }
    Expand-Archive -Path $tempZip -DestinationPath $tempExtractDir -Force

    $sourceExe = Join-Path $tempExtractDir "umaru.exe"
    if (-not (Test-Path $sourceExe)) {
        throw "Extraction failed: umaru.exe was not found inside the downloaded archive."
    }

    $installDir = Join-Path $env:LOCALAPPDATA "Programs\umaru"
    if (-not (Test-Path $installDir)) {
        New-Item -ItemType Directory -Path $installDir -Force | Out-Null
    }

    $targetExe = Join-Path $installDir "umaru.exe"
    Copy-Item -Path $sourceExe -Destination $targetExe -Force

    # Clean temporary files
    Remove-Item -Path $tempZip -Force -ErrorAction SilentlyContinue
    Remove-Item -Path $tempExtractDir -Recurse -Force -ErrorAction SilentlyContinue

    # Update User PATH if needed
    $userPath = [Environment]::GetEnvironmentVariable("Path", "User")
    if ($userPath -notlike "*$installDir*") {
        Write-Host "  [+] Adding $installDir to User PATH..." -ForegroundColor DarkGray
        [Environment]::SetEnvironmentVariable("Path", "$userPath;$installDir", "User")
    }

    # Add to current process PATH so it works immediately in current shell
    if ($env:PATH -notlike "*$installDir*") {
        $env:PATH = "$env:PATH;$installDir"
    }

    # Verify installation
    Write-Host ""
    Write-Host "  [SUCCESS] Umaru CLI has been installed successfully!" -ForegroundColor Green
    Write-Host "  Binary location: $targetExe" -ForegroundColor DarkGray
    Write-Host ""
    Write-Host "  Verification output:" -ForegroundColor Cyan
    & $targetExe version

    Write-Host ""
    Write-Host "  ================================================================" -ForegroundColor Magenta
    Write-Host "  Get started: Run 'umaru init' to create your first project." -ForegroundColor Cyan
    Write-Host "  Check system readiness: 'umaru doctor'" -ForegroundColor Yellow
    Write-Host "  ================================================================" -ForegroundColor Magenta
    Write-Host ""

} catch {
    Write-Host ""
    Write-Host "  [ERROR] Installation failed: $_" -ForegroundColor Red
    Write-Host ""
    exit 1
}
