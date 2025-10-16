$baseUrl = "http://localhost:8072/api/v1/game"

Write-Host "=== Game Module Flow Test ===" -ForegroundColor Green

# 1. Register
Write-Host "`n[1] Register user..." -ForegroundColor Cyan
$registerBody = @{
    username = "testuser2"
    email = "test2@example.com"
    password = "Test123456"
} | ConvertTo-Json

try {
    $registerResult = Invoke-RestMethod -Uri "$baseUrl/auth/register" -Method POST -Body $registerBody -ContentType "application/json"
    Write-Host "Register Success" -ForegroundColor Green
    Write-Host "User ID: $($registerResult.data.user_id)"
} catch {
    Write-Host "Register Failed: $($_.Exception.Message)" -ForegroundColor Red
    if ($_.ErrorDetails) {
        Write-Host $_.ErrorDetails.Message
    }
}

# 2. Login
Write-Host "`n[2] Login user..." -ForegroundColor Cyan
$loginBody = @{
    username = "testuser2"
    password = "Test123456"
} | ConvertTo-Json

try {
    $loginResult = Invoke-RestMethod -Uri "$baseUrl/auth/login" -Method POST -Body $loginBody -ContentType "application/json"
    Write-Host "Login Success" -ForegroundColor Green
    $token = $loginResult.data.token
    Write-Host "Token: $token"
} catch {
    Write-Host "Login Failed: $($_.Exception.Message)" -ForegroundColor Red
    if ($_.ErrorDetails) {
        Write-Host $_.ErrorDetails.Message
    }
}

# 3. Get classes
Write-Host "`n[3] Get basic classes..." -ForegroundColor Cyan
try {
    $classesResult = Invoke-RestMethod -Uri "$baseUrl/classes/basic" -Method GET
    Write-Host "Get Classes Success" -ForegroundColor Green
    Write-Host "Found $($classesResult.data.Count) classes"
    $firstClass = $classesResult.data[0]
    Write-Host "First class: $($firstClass.class_name) (ID: $($firstClass.id))"
} catch {
    Write-Host "Get Classes Failed: $($_.Exception.Message)" -ForegroundColor Red
}

if ($token -and $firstClass) {
    # 4. Create hero
    Write-Host "`n[4] Create hero..." -ForegroundColor Cyan
    $createHeroBody = @{
        class_id = $firstClass.id
        hero_name = "TestHero"
        description = "Test hero"
    } | ConvertTo-Json
    
    $headers = @{
        Authorization = "Bearer $token"
    }
    
    try {
        $heroResult = Invoke-RestMethod -Uri "$baseUrl/heroes" -Method POST -Body $createHeroBody -ContentType "application/json" -Headers $headers
        Write-Host "Create Hero Success" -ForegroundColor Green
        $heroId = $heroResult.data.id
        Write-Host "Hero ID: $heroId"
        Write-Host "Hero Name: $($heroResult.data.hero_name)"
        Write-Host "Level: $($heroResult.data.current_level)"
    } catch {
        Write-Host "Create Hero Failed: $($_.Exception.Message)" -ForegroundColor Red
        if ($_.ErrorDetails) {
            Write-Host $_.ErrorDetails.Message
        }
    }
    
    if ($heroId) {
        # 5. Get hero full info
        Write-Host "`n[5] Get hero full info..." -ForegroundColor Cyan
        try {
            $heroFullResult = Invoke-RestMethod -Uri "$baseUrl/heroes/$heroId/full" -Method GET -Headers $headers
            Write-Host "Get Hero Full Info Success" -ForegroundColor Green
            Write-Host "Hero: $($heroFullResult.data.hero_name)"
            Write-Host "Class: $($heroFullResult.data.class.class_name)"
            Write-Host "Attributes: $($heroFullResult.data.attributes.Count)"
            Write-Host "Skills: $($heroFullResult.data.skills.Count)"
        } catch {
            Write-Host "Get Hero Full Info Failed: $($_.Exception.Message)" -ForegroundColor Red
        }
        
        # 6. Get available skills
        Write-Host "`n[6] Get available skills..." -ForegroundColor Cyan
        try {
            $skillsResult = Invoke-RestMethod -Uri "$baseUrl/heroes/$heroId/skills/available" -Method GET -Headers $headers
            Write-Host "Get Available Skills Success" -ForegroundColor Green
            Write-Host "Available skills: $($skillsResult.data.Count)"
        } catch {
            Write-Host "Get Available Skills Failed: $($_.Exception.Message)" -ForegroundColor Red
        }
    }
}

Write-Host "`n=== Test Complete ===" -ForegroundColor Green
