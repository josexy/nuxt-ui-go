param([switch]$HiddenSmoke)
$ErrorActionPreference = 'Stop'
$projectRoot = Split-Path -Parent $PSScriptRoot
Push-Location $projectRoot
$previousSocket = $env:MYGO_AUTOMATION
$previousProbe = $env:NUI_AUTOMATION_SOCKET
$previousArtifacts = $env:NUI_NATIVE_ARTIFACTS
$previousCgo = $env:CGO_ENABLED
$previousHidden = $env:NUI_NATIVE_HIDDEN
$previousLocale = $env:MYGO_LOCALE
$galleryProcess = $null
try {
    $artifactDir = Join-Path $projectRoot '.work/native'
    New-Item -ItemType Directory -Force $artifactDir | Out-Null
    $env:CGO_ENABLED = '0'
    # Native probes use English labels; SetLocale can still switch after launch.
    # The override also prevents test language changes from being persisted.
    $env:MYGO_LOCALE = 'en'
    go build -tags e2e -o .work/native/gallery-e2e.exe ./examples/gallery
    if ($LASTEXITCODE -ne 0) { throw 'Gallery build failed' }
    $socketDir = Join-Path ([System.IO.Path]::GetTempPath()) ('nui-native-' + [guid]::NewGuid().ToString('N'))
    New-Item -ItemType Directory $socketDir | Out-Null
    $env:MYGO_AUTOMATION = Join-Path $socketDir 'gallery.sock'
    $env:NUI_AUTOMATION_SOCKET = $env:MYGO_AUTOMATION
    $env:NUI_NATIVE_ARTIFACTS = $artifactDir
    $launch = @{ FilePath = (Join-Path $projectRoot '.work/native/gallery-e2e.exe'); PassThru = $true; RedirectStandardOutput = (Join-Path $artifactDir 'stdout.log'); RedirectStandardError = (Join-Path $artifactDir 'stderr.log'); WindowStyle = 'Hidden' }
    $env:NUI_NATIVE_HIDDEN = '0'
    $launch.ArgumentList = '-static'
    if ($HiddenSmoke) { $launch.ArgumentList = '-static', '-hidden'; $env:NUI_NATIVE_HIDDEN = '1' }
    $galleryProcess = Start-Process @launch
    go test ./internal/nativeprobe -run TestNativeGallery -v -count=1
    if ($LASTEXITCODE -ne 0) { throw 'Native validation failed; see .work/native' }
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
    $env:MYGO_LOCALE = $previousLocale
    Pop-Location
}
