param(
    [string]$Url = "http://localhost:8080/"
)

try {
    $r = Invoke-WebRequest -Uri $Url -UseBasicParsing -TimeoutSec 5
    Write-Host "STATUS $($r.StatusCode)"
    $preview = $r.Content
    if ($preview.Length -gt 800) { $preview = $preview.Substring(0,800) }
    Write-Host "--- CONTENT PREVIEW ---"
    Write-Host $preview
} catch {
    Write-Host "ERROR: $($_.Exception.Message)"
    exit 1
}
