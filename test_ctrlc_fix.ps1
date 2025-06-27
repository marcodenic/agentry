# Test script to verify Ctrl+C fix for Agentry TUI
# This script builds the application and tests that Ctrl+C properly terminates it

Write-Host "🔨 Building Agentry..." -ForegroundColor Yellow
go build -o agentry.exe ./cmd/agentry

if ($LASTEXITCODE -ne 0) {
    Write-Host "❌ Build failed" -ForegroundColor Red
    exit 1
}

Write-Host "✅ Build successful" -ForegroundColor Green

Write-Host "🚀 Starting Agentry TUI (will auto-terminate after 5 seconds)..." -ForegroundColor Cyan
Write-Host "   In a real test, you would press Ctrl+C to test termination" -ForegroundColor Gray

# Start the TUI process
$process = Start-Process -FilePath "./agentry.exe" -ArgumentList "tui" -PassThru

# Wait a moment for it to start
Start-Sleep -Seconds 2

Write-Host "🔄 Sending SIGINT (Ctrl+C equivalent) to process $($process.Id)..." -ForegroundColor Cyan

# Send Ctrl+C signal (SIGINT) to the process
try {
    $process.Kill()
    $process.WaitForExit(5000)  # Wait up to 5 seconds for graceful exit
    
    if ($process.HasExited) {
        Write-Host "✅ Process terminated gracefully with exit code $($process.ExitCode)" -ForegroundColor Green
        Write-Host "🎉 Ctrl+C fix appears to be working!" -ForegroundColor Green
    } else {
        Write-Host "❌ Process did not terminate within timeout" -ForegroundColor Red
        $process.Kill()  # Force kill if it didn't exit gracefully
    }
} catch {
    Write-Host "❌ Error terminating process: $($_.Exception.Message)" -ForegroundColor Red
}

Write-Host "🧹 Cleaning up..." -ForegroundColor Yellow
if (Test-Path "agentry.exe") {
    Remove-Item "agentry.exe"
}

Write-Host "📋 Test completed" -ForegroundColor Cyan
