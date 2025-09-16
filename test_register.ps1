# 测试注册接口的 PowerShell 脚本

Write-Host "=== Test Register API ===" -ForegroundColor Green

# 测试数据
$testData = @{
    email = "test@example.com"
    username = "testuser123" 
    password = "password123"
}

$jsonBody = $testData | ConvertTo-Json
Write-Host "Request data:" -ForegroundColor Yellow
Write-Host $jsonBody

# 测试直连 login_shim 服务 (8090端口)
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
    Write-Host "请求失败:" -ForegroundColor Red
    Write-Host "状态码: $($_.Exception.Response.StatusCode)" -ForegroundColor Red
    Write-Host "错误信息: $($_.Exception.Message)" -ForegroundColor Red
    
    if ($_.Exception.Response) {
        $reader = New-Object System.IO.StreamReader($_.Exception.Response.GetResponseStream())
        $responseBody = $reader.ReadToEnd()
        Write-Host "响应体: $responseBody" -ForegroundColor Red
    }
}

# 测试通过 nginx 代理 (80端口)
Write-Host "`n--- Test via nginx proxy (localhost/auth/register) ---" -ForegroundColor Cyan

try {
    $response = Invoke-RestMethod -Uri "http://localhost/auth/register" `
        -Method POST `
        -ContentType "application/json" `
        -Body $jsonBody `
        -TimeoutSec 30

    Write-Host "响应状态: 成功" -ForegroundColor Green  
    Write-Host "响应内容:" -ForegroundColor Green
    $response | ConvertTo-Json -Depth 10 | Write-Host
} catch {
    Write-Host "请求失败:" -ForegroundColor Red
    Write-Host "状态码: $($_.Exception.Response.StatusCode)" -ForegroundColor Red
    Write-Host "错误信息: $($_.Exception.Message)" -ForegroundColor Red
    
    if ($_.Exception.Response) {
        $reader = New-Object System.IO.StreamReader($_.Exception.Response.GetResponseStream())
        $responseBody = $reader.ReadToEnd()
        Write-Host "响应体: $responseBody" -ForegroundColor Red
    }
}

Write-Host "`n=== Test Complete ===" -ForegroundColor Green