param([switch]$HiddenSmoke)
$ErrorActionPreference = 'Stop'
$projectRoot = Split-Path -Parent $PSScriptRoot
Push-Location $projectRoot
$previousSocket = $env:MYGO_AUTOMATION
$previousProbe = $env:NUI_AUTOMATION_SOCKET
$previousArtifacts = $env:NUI_NATIVE_ARTIFACTS
$previousCgo = $env:CGO_ENABLED
$previousHidden = $env:NUI_NATIVE_HIDDEN
$previousNarrow = $env:NUI_NATIVE_NARROW
$previousLocale = $env:MYGO_LOCALE
$galleryProcess = $null
try {
    $artifactDir = Join-Path $projectRoot '.work/native'
    New-Item -ItemType Directory -Force $artifactDir | Out-Null
    $env:CGO_ENABLED = '0'
    # Verify Gallery falls back to English even with a Chinese language preference.
    $env:MYGO_LOCALE = 'zh-Hans'
    go build -tags e2e -o .work/native/gallery-e2e.exe ./examples/gallery
    if ($LASTEXITCODE -ne 0) { throw 'Gallery build failed' }
    # Keep the socket on the checkout filesystem: redirected Windows TEMP
    # directories can create AF_UNIX endpoints that clients cannot connect to.
    $socketDir = Join-Path $artifactDir ('nui-' + [guid]::NewGuid().ToString('N').Substring(0, 8))
    New-Item -ItemType Directory $socketDir | Out-Null
    $env:MYGO_AUTOMATION = Join-Path $socketDir 'gallery.sock'
    $env:NUI_AUTOMATION_SOCKET = $env:MYGO_AUTOMATION
    $widths = @(1080, 720)
    if ($HiddenSmoke) { $widths = @(1080) }
    foreach ($width in $widths) {
        $runDir = $artifactDir
        $env:NUI_NATIVE_NARROW = '0'
        if ($width -eq 720) {
            $runDir = Join-Path $artifactDir 'narrow'
            $env:NUI_NATIVE_NARROW = '1'
            New-Item -ItemType Directory -Force $runDir | Out-Null
        }
        $env:NUI_NATIVE_ARTIFACTS = $runDir
        $launch = @{ FilePath = (Join-Path $projectRoot '.work/native/gallery-e2e.exe'); PassThru = $true; RedirectStandardOutput = (Join-Path $runDir 'stdout.log'); RedirectStandardError = (Join-Path $runDir 'stderr.log'); WindowStyle = 'Hidden' }
        $env:NUI_NATIVE_HIDDEN = '0'
        $launch.ArgumentList = '-static', '-width', $width
        if ($HiddenSmoke) { $launch.ArgumentList += '-hidden'; $env:NUI_NATIVE_HIDDEN = '1' }
        $galleryProcess = Start-Process @launch
        go test ./internal/nativeprobe -run TestNativeGallery -v -count=1
        if ($LASTEXITCODE -ne 0) { throw 'Native validation failed; see .work/native' }
        Stop-Process -Id $galleryProcess.Id
        $galleryProcess.WaitForExit(5000) | Out-Null
        $galleryProcess = $null
        if (Test-Path -LiteralPath $env:MYGO_AUTOMATION) { Remove-Item -LiteralPath $env:MYGO_AUTOMATION -Force }
    }
} finally {
    if ($galleryProcess -and -not $galleryProcess.HasExited) {
        Stop-Process -Id $galleryProcess.Id
        $galleryProcess.WaitForExit(5000) | Out-Null
    }
    if ($socketDir) {
        # Only these two exact paths, created by this invocation; no recursive deletion.
        $socketFile = Join-Path $socketDir 'gallery.sock'
        if (Test-Path -LiteralPath $socketFile) { Remove-Item -LiteralPath $socketFile -Force }
        if (Test-Path -LiteralPath $socketDir) { Remove-Item -LiteralPath $socketDir }
    }
    $env:MYGO_AUTOMATION = $previousSocket
    $env:NUI_AUTOMATION_SOCKET = $previousProbe
    $env:NUI_NATIVE_ARTIFACTS = $previousArtifacts
    $env:CGO_ENABLED = $previousCgo
    $env:NUI_NATIVE_HIDDEN = $previousHidden
    $env:NUI_NATIVE_NARROW = $previousNarrow
    $env:MYGO_LOCALE = $previousLocale
    Pop-Location
}
