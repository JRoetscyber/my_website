# ==============================================================================
# JO4 DEV - ZERO-DOWNTIME RED/BLUE DEPLOYMENT SCRIPT (POWERSHELL / WINDOWS)
# ==============================================================================

$ErrorActionPreference = "Stop"

$upstreamFile = "nginx/upstream.inc"
if (-not (Test-Path "nginx/conf.d")) {
    New-Item -ItemType Directory -Force -Path "nginx/conf.d" | Out-Null
}

if (-not (Test-Path $upstreamFile)) {
    Set-Content -Path $upstreamFile -Value "server web-blue:5000;"
}

if (-not (Test-Path ".env")) {
    if (Test-Path ".env.example") {
        Write-Host "⚠️ .env file not found! Copying from .env.example..." -ForegroundColor Yellow
        Copy-Item ".env.example" ".env"
    }
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
Write-Host "   Deploy Target      : $target ($targetService on port $targetPort)"
Write-Host "==========================================================" -ForegroundColor Cyan

# 1. Build Target Container
Write-Host "🔨 Step 1/5: Building $targetService..." -ForegroundColor Gray
docker compose build $targetService

# 2. Launch Target Container
Write-Host "🚀 Step 2/5: Starting $targetService..." -ForegroundColor Gray
docker compose up -d $targetService

# 3. Health Check
Write-Host "🔍 Step 3/5: Checking http://127.0.0.1:$targetPort/health..." -ForegroundColor Gray
$maxAttempts = 25
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

# 4. Ensure Nginx is running
Write-Host "🌐 Step 4/5: Ensuring Nginx reverse proxy is active..." -ForegroundColor Gray
docker compose up -d nginx

# 5. Switch Upstream and Reload
Write-Host "🔄 Step 5/5: Switching Nginx upstream to $targetService..." -ForegroundColor Gray
Set-Content -Path $upstreamFile -Value "server $targetService:5000;"

docker compose exec -T nginx nginx -s reload

Write-Host "⏳ Draining connections from $oldService (5s)..." -ForegroundColor Gray
Start-Sleep -Seconds 5

Write-Host "🛑 Stopping idle environment ($oldService)..." -ForegroundColor Gray
docker compose stop $oldService

Write-Host "==========================================================" -ForegroundColor Green
Write-Host "🎉 DEPLOYMENT COMPLETE! Live traffic on: $target" -ForegroundColor Green
Write-Host "==========================================================" -ForegroundColor Green
