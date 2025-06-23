param(
  [string]$Version
)

$binary = "dist/agentry-windows-amd64.exe"
$sha = (Get-FileHash $binary -Algorithm SHA256).Hash
$manifestDir = "dist/scoop"
New-Item -ItemType Directory -Path $manifestDir -Force | Out-Null
$manifestPath = Join-Path $manifestDir "agentry.json"

@{
    version = $Version
    architecture = @{ 
        "64bit" = @{ url = "https://github.com/marcodenic/agentry/releases/download/v$Version/agentry-windows-amd64.exe"; hash = $sha }
    }
    bin = "agentry-windows-amd64.exe"
    description = "Minimal, performant AI-Agent runtime"
    homepage = "https://github.com/marcodenic/agentry"
} | ConvertTo-Json -Depth 3 | Out-File -Encoding ASCII $manifestPath

Write-Output "Scoop manifest written to $manifestPath"
