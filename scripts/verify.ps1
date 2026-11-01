$ErrorActionPreference = 'Stop'
$projectRoot = Split-Path -Parent $PSScriptRoot
$previousCgo = $env:CGO_ENABLED
$previousOS = $env:GOOS
$previousArch = $env:GOARCH
Push-Location $projectRoot
try {
    $env:CGO_ENABLED = '0'
    Remove-Item Env:GOOS -ErrorAction SilentlyContinue
    Remove-Item Env:GOARCH -ErrorAction SilentlyContinue
    $files = @(Get-ChildItem -LiteralPath $projectRoot -Filter '*.go' -File) + @(Get-ChildItem -Path icons,internal,examples -Filter '*.go' -File -Recurse)
    # The full Lucide set exceeds Windows' command-line length limit in one call.
    for ($fileOffset = 0; $fileOffset -lt $files.Count; $fileOffset += 100) {
        $fileEnd = [Math]::Min($fileOffset + 99, $files.Count - 1)
        $fileBatch = $files[$fileOffset..$fileEnd]
        $unformatted = gofmt -l @($fileBatch.FullName)
        if ($LASTEXITCODE -ne 0 -or $unformatted) { throw "gofmt: $unformatted" }
    }
    go test ./... -count=1
    if ($LASTEXITCODE -ne 0) { throw 'Tests failed' }
    go vet ./...
    if ($LASTEXITCODE -ne 0) { throw 'Go vet failed' }
    go tool mygo vet .
    if ($LASTEXITCODE -ne 0) { throw 'MyGo lifecycle check failed' }
    $generated = @(Get-ChildItem -LiteralPath icons/lucide -Filter '*.go' -File) + @(Get-Item -LiteralPath palette_gen.go)
    $before = @{}
    foreach ($file in $generated) { $before[$file.FullName] = (Get-FileHash -LiteralPath $file.FullName).Hash }
    go generate ./...
    if ($LASTEXITCODE -ne 0) { throw 'Generation failed' }
    foreach ($file in $generated) { if ((Get-FileHash -LiteralPath $file.FullName).Hash -ne $before[$file.FullName]) { throw "Generated file drift: $($file.Name)" } }
    foreach ($target in @(@('windows','amd64'),@('linux','amd64'),@('darwin','arm64'))) {
        $env:GOOS, $env:GOARCH = $target
        go build ./...
        if ($LASTEXITCODE -ne 0) { throw "Build failed: $target" }
        Write-Output "Built $($target -join '/') with CGO_ENABLED=0"
    }
} finally {
    $env:CGO_ENABLED = $previousCgo
    $env:GOOS = $previousOS
    $env:GOARCH = $previousArch
    Pop-Location
}
