# ==============================================================================
# JO4 DEV - ZERO-DOWNTIME RED/BLUE DEPLOYMENT SCRIPT (POWERSHELL / WINDOWS)
# ==============================================================================

$ErrorActionPreference = "Stop"

$upstreamFile = "nginx/upstream.inc"
if (-not (Test-Path $upstreamFile)) {
    New-Item -ItemType Directory -Force -Path "nginx" | Out-Null
    Set-Content -Path $upstreamFile -Value "server web-blue:5000;"
}

if (-not (Test-Path ".env")) {
    Write-Host "⚠️ .env file not found! Copying from .env.example..." -ForegroundColor Yellow
    Copy-Item ".env.example" ".env"
}

$currentContent = Get-Content $upstreamFile -Raw
if ($currentContent -match "web-blue") {
    $active = "BLUE"
    $target = "GREEN"
    $targetService = "web-green"
    $targetPort = 5002
    $oldService = "web-blue"
} else {
    $active = "GREEN"
    $target = "BLUE"
    $targetService = "web-blue"
    $targetPort = 5001
    $oldService = "web-green"
}

Write-Host "==========================================================" -ForegroundColor Cyan
Write-Host "🚀 JO4 Dev Zero-Downtime Deployment" -ForegroundColor Cyan
Write-Host "   Active Environment : $active"
Write-Host "   Deploy Target      : $target ($targetService)"
Write-Host "==========================================================" -ForegroundColor Cyan

Write-Host "📦 Ensuring Nginx reverse proxy is running..." -ForegroundColor Gray
docker compose up -d nginx

Write-Host "🔨 Building and starting $targetService..." -ForegroundColor Gray
docker compose build $targetService
docker compose up -d $targetService

Write-Host "🔍 Waiting for $targetService health check on port $targetPort..." -ForegroundColor Gray
$maxAttempts = 20
$attempt = 0
$healthy = $false

while ($attempt -lt $maxAttempts) {
    $attempt++
    try {
        $resp = Invoke-WebRequest -Uri "http://127.0.0.1:$targetPort/health" -UseBasicParsing -TimeoutSec 2
        if ($resp.StatusCode -eq 200) {
            $healthy = $true
            break
        }
    } catch {
        # Waiting
    }
    Write-Host "   Attempt $attempt/$maxAttempts: Waiting for server response..." -ForegroundColor DarkGray
    Start-Sleep -Seconds 2
}

if (-not $healthy) {
    Write-Host "❌ Health check FAILED on $targetService!" -ForegroundColor Red
    Write-Host "   Aborting deployment. Active environment remains $active." -ForegroundColor Red
    docker compose logs --tail=50 $targetService
    docker compose stop $targetService
    exit 1
}

Write-Host "✅ Health check PASSED for $targetService!" -ForegroundColor Green

Write-Host "🔄 Switching Nginx upstream to $targetService..." -ForegroundColor Gray
Set-Content -Path $upstreamFile -Value "server $targetService:5000;"

docker compose exec -T nginx nginx -s reload

Write-Host "⏳ Draining connections from $oldService (5s)..." -ForegroundColor Gray
Start-Sleep -Seconds 5

Write-Host "🛑 Stopping idle environment ($oldService)..." -ForegroundColor Gray
docker compose stop $oldService

Write-Host "==========================================================" -ForegroundColor Green
Write-Host "🎉 DEPLOYMENT COMPLETE! Live traffic on: $target" -ForegroundColor Green
Write-Host "==========================================================" -ForegroundColor Green
