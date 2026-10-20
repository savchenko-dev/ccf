$ErrorActionPreference = 'Stop'

$repo = 'savchenko-dev/ccf'
$dest = if ($env:CCF_INSTALL_DIR) { $env:CCF_INSTALL_DIR } else { Join-Path $env:LOCALAPPDATA 'Programs\ccf' }

$arch = if ($env:PROCESSOR_ARCHITECTURE -eq 'ARM64') { 'arm64' } else { 'amd64' }
$url = "https://github.com/$repo/releases/latest/download/ccf_windows_$arch.zip"

$tmp = Join-Path ([System.IO.Path]::GetTempPath()) ([System.Guid]::NewGuid())
New-Item -ItemType Directory -Path $tmp | Out-Null
try {
    Write-Host "Downloading $url"
    Invoke-WebRequest -Uri $url -OutFile (Join-Path $tmp 'ccf.zip') -UseBasicParsing
    Expand-Archive -Path (Join-Path $tmp 'ccf.zip') -DestinationPath $tmp -Force
    New-Item -ItemType Directory -Path $dest -Force | Out-Null
    Copy-Item (Join-Path $tmp 'ccf.exe') (Join-Path $dest 'ccf.exe') -Force
} finally {
    Remove-Item $tmp -Recurse -Force
}
Write-Host "Installed ccf to $dest\ccf.exe"

$userPath = [Environment]::GetEnvironmentVariable('Path', 'User')
if (($userPath -split ';') -notcontains $dest) {
    [Environment]::SetEnvironmentVariable('Path', "$userPath;$dest", 'User')
    Write-Host "Added $dest to your user PATH. Restart the terminal to use ccf."
}
