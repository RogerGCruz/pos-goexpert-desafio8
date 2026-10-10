param(
    [int]$MaxWait = 60
)

$Root = Split-Path -Parent $MyInvocation.MyCommand.Definition
Set-Location $Root
Set-Location ..

Write-Host "Bringing up docker-compose (build)"
docker-compose up -d --build

Write-Host "Waiting for app to become available on http://localhost:8080/auction?status=0"
for ($i=1; $i -le $MaxWait; $i++) {
    try {
        $resp = Invoke-WebRequest -Uri 'http://localhost:8080/auction?status=0' -UseBasicParsing -TimeoutSec 5 -ErrorAction Stop
        if ($resp.StatusCode -eq 200) {
            Write-Host "App is reachable (HTTP 200)"
            break
        }
    } catch {
        Write-Host "Waiting... ($i/$MaxWait) - $(($_.Exception).Message)"
    }
    Start-Sleep -Seconds 1
}

try {
    $resp = Invoke-WebRequest -Uri 'http://localhost:8080/auction?status=0' -UseBasicParsing -TimeoutSec 5 -ErrorAction Stop
    if ($resp.StatusCode -ne 200) {
        throw "App did not become healthy in time, final status=$($resp.StatusCode)"
    }
} catch {
    Write-Host "App did not become healthy in time: $_"
    Write-Host "Last 200 lines of app logs:"
    docker-compose logs --no-color --tail=200 app
    docker-compose down
    exit 2
}

if (-not $env:AUCTION_DURATION) { $env:AUCTION_DURATION = '2' }

Set-Location tests_bdd
Write-Host "Running BDD tests (go test -v)"
go mod tidy
if (go test -v) {
    Write-Host "BDD tests passed"
    docker-compose down
    exit 0
} else {
    Write-Host "BDD tests failed"
    docker-compose logs --no-color --timestamps --tail=200
    docker-compose down
    exit 1
}
