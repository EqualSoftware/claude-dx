# Install claude-dx from GitHub Releases on Windows.
# Usage: install.ps1 [version]     Example: install.ps1 v0.1.0
#        irm https://raw.githubusercontent.com/EqualSoftware/claude-dx/main/scripts/install.ps1 | iex
# Environment:
#   CLAUDE_DX_VERSION      Version to install, if no argument is given. Default: latest
#   CLAUDE_DX_INSTALL_DIR  Install directory. Default: %LocalAppData%\Programs\claude-dx
#   CLAUDE_DX_BASE_URL     Release download URL, for tests. Default: GitHub Releases

# The whole script runs in a script block, so its variables and preferences do not stay in the session.
& {
    param([string]$FirstArgument)

    $ErrorActionPreference = "Stop"
    $ProgressPreference = "SilentlyContinue"

    $Repo = "EqualSoftware/claude-dx"

    # Use throw, not exit: with "irm | iex", exit closes the PowerShell window of the user.
    function Fail([string]$Message) {
        throw "claude-dx install: $Message"
    }

    function Get-Arch {
        $arch = $env:PROCESSOR_ARCHITEW6432
        if (-not $arch) { $arch = $env:PROCESSOR_ARCHITECTURE }
        switch ($arch) {
            "AMD64" { return "amd64" }
            "ARM64" { return "arm64" }
            default { Fail "this architecture is not supported: $arch" }
        }
    }

    $Version = if ($FirstArgument) { $FirstArgument } elseif ($env:CLAUDE_DX_VERSION) { $env:CLAUDE_DX_VERSION } else { "" }
    $InstallDir = if ($env:CLAUDE_DX_INSTALL_DIR) { $env:CLAUDE_DX_INSTALL_DIR } else { Join-Path $env:LOCALAPPDATA "Programs\claude-dx" }

    if ($Version -and $Version -notmatch '^v') { $Version = "v$Version" }
    if (-not $Version) {
        $Version = (Invoke-RestMethod "https://api.github.com/repos/$Repo/releases/latest").tag_name
        if (-not $Version) { Fail "cannot find the latest release" }
    }

    $Arch = Get-Arch
    $Archive = "claude-dx_$($Version.TrimStart('v'))_windows_$Arch.zip"
    $Base = if ($env:CLAUDE_DX_BASE_URL) { $env:CLAUDE_DX_BASE_URL } else { "https://github.com/$Repo/releases/download/$Version" }

    $Tmp = Join-Path ([System.IO.Path]::GetTempPath()) ("claude-dx-" + [System.Guid]::NewGuid())
    New-Item -ItemType Directory -Path $Tmp | Out-Null
    try {
        Invoke-WebRequest "$Base/$Archive" -OutFile (Join-Path $Tmp $Archive) -UseBasicParsing
        Invoke-WebRequest "$Base/checksums.txt" -OutFile (Join-Path $Tmp "checksums.txt") -UseBasicParsing

        $Line = Get-Content (Join-Path $Tmp "checksums.txt") | Where-Object { ($_ -split '\s+')[1] -eq $Archive }
        if (-not $Line) { Fail "checksums.txt has no entry for $Archive" }
        $Expected = ($Line -split '\s+')[0]
        $Actual = (Get-FileHash (Join-Path $Tmp $Archive) -Algorithm SHA256).Hash
        if ($Expected -ne $Actual) { Fail "checksum mismatch for $Archive. Nothing is installed." }

        Expand-Archive (Join-Path $Tmp $Archive) -DestinationPath $Tmp -Force
        New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null

        # Windows locks a running .exe, but a rename is permitted.
        # claude-dx deletes the .old file at its next start.
        $Target = Join-Path $InstallDir "claude-dx.exe"
        if (Test-Path $Target) {
            $Old = "$Target.old"
            $n = 0
            while ($true) {
                try { Move-Item $Target $Old -Force -ErrorAction Stop; break }
                catch { $n++; if ($n -gt 100) { throw }; $Old = "$Target.old.$n" }
            }
        }
        Copy-Item (Join-Path $Tmp "claude-dx.exe") $Target

        $UserPath = [Environment]::GetEnvironmentVariable("Path", "User")
        if (-not (($UserPath -split ';') -contains $InstallDir)) {
            $NewPath = if ($UserPath) { "$UserPath;$InstallDir" } else { $InstallDir }
            [Environment]::SetEnvironmentVariable("Path", $NewPath, "User")
            Write-Host "Added $InstallDir to your user PATH. Open a new terminal."
        }
        Write-Host "claude-dx $Version installed to $Target"
    }
    finally {
        Remove-Item $Tmp -Recurse -Force -ErrorAction SilentlyContinue
    }
} @args
