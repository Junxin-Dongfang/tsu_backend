# Game Module API 测试脚本
$baseUrl = "http://localhost:8072/api/v1/game"

Write-Host "========================================" -ForegroundColor Green
Write-Host "Game Module API 测试" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Green
Write-Host ""

# 1. 测试注册
Write-Host "[1] 测试用户注册..." -ForegroundColor Cyan
$registerBody = @{
    username = "testuser"
    email = "test@example.com"
    password = "Test123456"
} | ConvertTo-Json

try {
    $registerResponse = Invoke-RestMethod -Uri "$baseUrl/auth/register" -Method Post -Body $registerBody -ContentType "application/json"
    Write-Host "✅ 注册成功" -ForegroundColor Green
    Write-Host ($registerResponse | ConvertTo-Json -Depth 10)
} catch {
    Write-Host "❌ 注册失败: $($_.Exception.Message)" -ForegroundColor Red
    Write-Host $_.Exception.Response
}

Write-Host ""

# 2. 测试登录
Write-Host "[2] 测试用户登录..." -ForegroundColor Cyan
$loginBody = @{
    username = "testuser"
    password = "Test123456"
} | ConvertTo-Json

try {
    $loginResponse = Invoke-RestMethod -Uri "$baseUrl/auth/login" -Method Post -Body $loginBody -ContentType "application/json"
    Write-Host "✅ 登录成功" -ForegroundColor Green
    $token = $loginResponse.data.token
    Write-Host "Token: $token"
    Write-Host ($loginResponse | ConvertTo-Json -Depth 10)
} catch {
    Write-Host "❌ 登录失败: $($_.Exception.Message)" -ForegroundColor Red
}

Write-Host ""

# 3. 测试获取基础职业列表
Write-Host "[3] 测试获取基础职业列表..." -ForegroundColor Cyan
try {
    $classesResponse = Invoke-RestMethod -Uri "$baseUrl/classes/basic" -Method Get
    Write-Host "✅ 获取基础职业成功" -ForegroundColor Green
    Write-Host "职业数量: $($classesResponse.data.Count)"
    $classesResponse.data | ForEach-Object {
        Write-Host "  - $($_.class_name) ($($_.class_code))"
    }
} catch {
    Write-Host "❌ 获取职业失败: $($_.Exception.Message)" -ForegroundColor Red
}

Write-Host ""

if ($token) {
    # 4. 测试创建英雄
    Write-Host "[4] 测试创建英雄..." -ForegroundColor Cyan
    
    # 先获取一个职业ID
    $firstClass = $classesResponse.data[0]
    
    $createHeroBody = @{
        class_id = $firstClass.id
        hero_name = "TestHero"
        description = "测试英雄"
    } | ConvertTo-Json
    
    $headers = @{
        "Authorization" = "Bearer $token"
        "Content-Type" = "application/json"
    }
    
    try {
        $heroResponse = Invoke-RestMethod -Uri "$baseUrl/heroes" -Method Post -Body $createHeroBody -Headers $headers
        Write-Host "✅ 创建英雄成功" -ForegroundColor Green
        $heroId = $heroResponse.data.id
        Write-Host "英雄ID: $heroId"
        Write-Host ($heroResponse | ConvertTo-Json -Depth 10)
    } catch {
        Write-Host "❌ 创建英雄失败: $($_.Exception.Message)" -ForegroundColor Red
        Write-Host $_.ErrorDetails.Message
    }
    
    Write-Host ""
    
    if ($heroId) {
        # 5. 测试获取英雄完整信息
        Write-Host "[5] 测试获取英雄完整信息..." -ForegroundColor Cyan
        try {
            $heroFullResponse = Invoke-RestMethod -Uri "$baseUrl/heroes/$heroId/full" -Method Get -Headers $headers
            Write-Host "✅ 获取英雄完整信息成功" -ForegroundColor Green
            Write-Host ($heroFullResponse | ConvertTo-Json -Depth 10)
        } catch {
            Write-Host "❌ 获取英雄完整信息失败: $($_.Exception.Message)" -ForegroundColor Red
        }
        
        Write-Host ""
        
        # 6. 测试获取可学习技能
        Write-Host "[6] 测试获取可学习技能..." -ForegroundColor Cyan
        try {
            $skillsResponse = Invoke-RestMethod -Uri "$baseUrl/heroes/$heroId/skills/available" -Method Get -Headers $headers
            Write-Host "✅ 获取可学习技能成功" -ForegroundColor Green
            Write-Host "可学习技能数量: $($skillsResponse.data.Count)"
        } catch {
            Write-Host "❌ 获取可学习技能失败: $($_.Exception.Message)" -ForegroundColor Red
        }
    }
}

Write-Host ""
Write-Host "========================================" -ForegroundColor Green
Write-Host "测试完成" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Green
