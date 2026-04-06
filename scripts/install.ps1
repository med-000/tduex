param(
    [string]$Version = "latest",
    [string]$Repo = "med-000/tduex",
    [string]$InstallDir = "$env:LOCALAPPDATA\Programs\tduex\bin",
    [string]$SourceExe,
    [switch]$Force
)

$ErrorActionPreference = "Stop"

function Get-AssetArch {
    switch ($env:PROCESSOR_ARCHITECTURE.ToLowerInvariant()) {
        "arm64" { return "arm64" }
        default { return "x86_64" }
    }
}

function Get-ReleaseUrl {
    param(
        [string]$RepoName,
        [string]$ReleaseVersion,
        [string]$AssetName
    )

    if ($ReleaseVersion -eq "latest") {
        return "https://github.com/$RepoName/releases/latest/download/$AssetName"
    }

    return "https://github.com/$RepoName/releases/download/$ReleaseVersion/$AssetName"
}

function Find-LocalExe {
    param([string]$PreferredPath)

    if ($PreferredPath) {
        if (Test-Path $PreferredPath) {
            return (Resolve-Path $PreferredPath).Path
        }
        throw "Specified -SourceExe was not found: $PreferredPath"
    }

    $repoRoot = Split-Path -Parent $PSScriptRoot
    $candidates = @(
        (Join-Path $repoRoot "tduex.exe"),
        (Join-Path $repoRoot "dist\tduex.exe"),
        (Join-Path $repoRoot "build\tduex.exe")
    )

    foreach ($candidate in $candidates) {
        if (Test-Path $candidate) {
            return (Resolve-Path $candidate).Path
        }
    }

    return $null
}

function Install-Exe {
    param(
        [string]$ExePath,
        [string]$TargetDir,
        [bool]$Overwrite
    )

    New-Item -ItemType Directory -Path $TargetDir -Force | Out-Null
    $destination = Join-Path $TargetDir "tduex.exe"
    Copy-Item -Path $ExePath -Destination $destination -Force:$Overwrite

    $pathUpdated = Add-UserPath -PathToAdd $TargetDir

    Write-Host "Installed: $destination"
    if ($pathUpdated) {
        Write-Host "Added to user PATH: $TargetDir"
        Write-Host "Open a new PowerShell window to use 'tduex'."
    } else {
        Write-Host "You can now run 'tduex' from a new PowerShell window."
    }
}

function Build-FromSource {
    $goCmd = Get-Command go -ErrorAction SilentlyContinue
    if (-not $goCmd) {
        return $null
    }

    $repoRoot = Split-Path -Parent $PSScriptRoot
    $tempBuildDir = Join-Path ([System.IO.Path]::GetTempPath()) ("tduex-build-" + [System.Guid]::NewGuid().ToString("N"))
    $outputPath = Join-Path $tempBuildDir "tduex.exe"

    New-Item -ItemType Directory -Path $tempBuildDir -Force | Out-Null

    Push-Location $repoRoot
    try {
        & $goCmd.Source build -o $outputPath ./cmd/tduex
        if ($LASTEXITCODE -ne 0) {
            throw "go build failed."
        }
        return $outputPath
    }
    finally {
        Pop-Location
    }
}

function Add-UserPath {
    param([string]$PathToAdd)

    $current = [Environment]::GetEnvironmentVariable("Path", "User")
    $entries = @()
    if ($current) {
        $entries = $current.Split(';', [System.StringSplitOptions]::RemoveEmptyEntries)
    }

    if ($entries -contains $PathToAdd) {
        return $false
    }

    $updated = if ($current) { "$current;$PathToAdd" } else { $PathToAdd }
    [Environment]::SetEnvironmentVariable("Path", $updated, "User")
    return $true
}

$arch = Get-AssetArch
$assetName = "tduex_Windows_$arch.zip"
$downloadUrl = Get-ReleaseUrl -RepoName $Repo -ReleaseVersion $Version -AssetName $assetName

$tempRoot = Join-Path ([System.IO.Path]::GetTempPath()) ("tduex-install-" + [System.Guid]::NewGuid().ToString("N"))
$zipPath = Join-Path $tempRoot $assetName
$extractDir = Join-Path $tempRoot "extract"
$exeName = "tduex.exe"
$downloaded = $false
$builtExePath = $null

try {
    $localExe = Find-LocalExe -PreferredPath $SourceExe
    if ($localExe) {
        Write-Host "Using local executable: $localExe"
        Install-Exe -ExePath $localExe -TargetDir $InstallDir -Overwrite $Force.IsPresent
        return
    }

    New-Item -ItemType Directory -Path $tempRoot -Force | Out-Null
    New-Item -ItemType Directory -Path $extractDir -Force | Out-Null

    Write-Host "Downloading $downloadUrl"
    Invoke-WebRequest -Uri $downloadUrl -OutFile $zipPath
    $downloaded = $true

    Expand-Archive -Path $zipPath -DestinationPath $extractDir -Force

    $exePath = Get-ChildItem -Path $extractDir -Filter $exeName -Recurse | Select-Object -First 1 -ExpandProperty FullName
    if (-not $exePath) {
        throw "tduex.exe was not found in $assetName."
    }

    Install-Exe -ExePath $exePath -TargetDir $InstallDir -Overwrite $Force.IsPresent
}
catch {
    if (-not $downloaded) {
        $builtExePath = Build-FromSource
        if ($builtExePath) {
            Write-Host "Release asset was not found. Built from local source instead."
            Install-Exe -ExePath $builtExePath -TargetDir $InstallDir -Overwrite $Force.IsPresent
            return
        }

        Write-Error "Failed to download release asset '$assetName'. Put tduex.exe in the repo root and rerun, use -SourceExe, or publish a GitHub release asset with that name."
    }

    throw
}
finally {
    if (Test-Path $tempRoot) {
        Remove-Item -Path $tempRoot -Recurse -Force
    }
    if ($builtExePath) {
        $builtDir = Split-Path -Parent $builtExePath
        if (Test-Path $builtDir) {
            Remove-Item -Path $builtDir -Recurse -Force
        }
    }
}
