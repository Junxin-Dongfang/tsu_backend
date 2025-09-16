# Test register API PowerShell script

Write-Host "=== Test Register API ===" -ForegroundColor Green

# Test data
$testData = @{
    email = "test@example.com"
    username = "testuser123" 
    password = "password123"
}

$jsonBody = $testData | ConvertTo-Json
Write-Host "Request data:" -ForegroundColor Yellow
Write-Host $jsonBody

# Test direct login_shim service (port 8090)
Write-Host "`n--- Test direct login_shim service (localhost:8090) ---" -ForegroundColor Cyan

try {
    $response = Invoke-RestMethod -Uri "http://localhost:8090/auth/register" `
        -Method POST `
        -ContentType "application/json" `
        -Body $jsonBody `
        -TimeoutSec 30

    Write-Host "Response status: Success" -ForegroundColor Green
    Write-Host "Response content:" -ForegroundColor Green
    $response | ConvertTo-Json -Depth 10 | Write-Host
} catch {
    Write-Host "Request failed:" -ForegroundColor Red
    Write-Host "Status code: $($_.Exception.Response.StatusCode)" -ForegroundColor Red
    Write-Host "Error message: $($_.Exception.Message)" -ForegroundColor Red
    
    if ($_.Exception.Response) {
        $reader = New-Object System.IO.StreamReader($_.Exception.Response.GetResponseStream())
        $responseBody = $reader.ReadToEnd()
        Write-Host "Response body: $responseBody" -ForegroundColor Red
    }
}

# Test via nginx proxy (port 80)
Write-Host "`n--- Test via nginx proxy (localhost/auth/register) ---" -ForegroundColor Cyan

try {
    $response = Invoke-RestMethod -Uri "http://localhost/auth/register" `
        -Method POST `
        -ContentType "application/json" `
        -Body $jsonBody `
        -TimeoutSec 30

    Write-Host "Response status: Success" -ForegroundColor Green  
    Write-Host "Response content:" -ForegroundColor Green
    $response | ConvertTo-Json -Depth 10 | Write-Host
} catch {
    Write-Host "Request failed:" -ForegroundColor Red
    Write-Host "Status code: $($_.Exception.Response.StatusCode)" -ForegroundColor Red
    Write-Host "Error message: $($_.Exception.Message)" -ForegroundColor Red
    
    if ($_.Exception.Response) {
        $reader = New-Object System.IO.StreamReader($_.Exception.Response.GetResponseStream())
        $responseBody = $reader.ReadToEnd()
        Write-Host "Response body: $responseBody" -ForegroundColor Red
    }
}

Write-Host "`n=== Test Complete ===" -ForegroundColor Green