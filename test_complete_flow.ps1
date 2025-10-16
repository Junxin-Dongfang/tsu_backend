# Complete Game Flow Test
# Note: Using direct connection due to Oathkeeper routing issue
$baseUrl = "http://localhost:8072/api/v1/game"
$timestamp = Get-Date -Format "yyyyMMddHHmmss"
$testUsername = "testuser_$timestamp"
$testEmail = "test_${timestamp}@example.com"

Write-Host "========================================" -ForegroundColor Green
Write-Host "Complete Game Flow Test" -ForegroundColor Green
Write-Host "Direct connection to Game service" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Green
Write-Host "Test User: $testUsername" -ForegroundColor Cyan
Write-Host ""

# 1. Register new user
Write-Host "[1] Registering new user..." -ForegroundColor Cyan
$registerBody = @{
    username = $testUsername
    email = $testEmail
    password = "Test123456"
} | ConvertTo-Json

try {
    $registerResult = Invoke-RestMethod -Uri "$baseUrl/auth/register" -Method POST -Body $registerBody -ContentType "application/json"
    Write-Host "SUCCESS: User registered" -ForegroundColor Green
    Write-Host "User ID: $($registerResult.data.user_id)" -ForegroundColor Gray
    $userId = $registerResult.data.user_id
} catch {
    Write-Host "FAILED: Registration failed" -ForegroundColor Red
    Write-Host "Error: $($_.Exception.Message)" -ForegroundColor Red
    if ($_.ErrorDetails) {
        Write-Host $_.ErrorDetails.Message -ForegroundColor Yellow
    }
    exit 1
}

Start-Sleep -Seconds 2

# 2. Login
Write-Host "`n[2] Logging in..." -ForegroundColor Cyan
$loginBody = @{
    identifier = $testUsername
    password = "Test123456"
} | ConvertTo-Json

try {
    $loginResult = Invoke-RestMethod -Uri "$baseUrl/auth/login" -Method POST -Body $loginBody -ContentType "application/json"
    Write-Host "SUCCESS: User logged in" -ForegroundColor Green
    
    # Check response structure
    if ($loginResult.data.session_token) {
        $token = $loginResult.data.session_token
    } elseif ($loginResult.data.token) {
        $token = $loginResult.data.token
    } else {
        Write-Host "Response data:" -ForegroundColor Yellow
        Write-Host ($loginResult | ConvertTo-Json -Depth 5) -ForegroundColor Gray
        Write-Host "WARNING: No token found in response, using user_id for testing" -ForegroundColor Yellow
        $token = "test_token_$($loginResult.data.user_id)"
    }
    
    if ($token -and $token.Length -gt 50) {
        Write-Host "Token: $($token.Substring(0, 50))..." -ForegroundColor Gray
    } else {
        Write-Host "Token: $token" -ForegroundColor Gray
    }
} catch {
    Write-Host "FAILED: Login failed" -ForegroundColor Red
    Write-Host "Error: $($_.Exception.Message)" -ForegroundColor Red
    if ($_.ErrorDetails) {
        Write-Host $_.ErrorDetails.Message -ForegroundColor Yellow
    }
    exit 1
}

# For direct connection, we need to manually add X-User-ID
$headers = @{
    Authorization = "Bearer $token"
    "X-User-ID" = $loginResult.data.user_id
    "X-Session-Token" = $token
}

# 3. Get basic classes
Write-Host "`n[3] Getting basic classes..." -ForegroundColor Cyan
try {
    $classesResult = Invoke-RestMethod -Uri "$baseUrl/classes/basic" -Method GET
    Write-Host "SUCCESS: Got $($classesResult.data.Count) classes" -ForegroundColor Green
    $firstClass = $classesResult.data[0]
    Write-Host "Selected class: $($firstClass.class_name)" -ForegroundColor Gray
} catch {
    Write-Host "FAILED: Get classes failed" -ForegroundColor Red
    Write-Host "Error: $($_.Exception.Message)" -ForegroundColor Red
    if ($_.ErrorDetails) {
        Write-Host $_.ErrorDetails.Message -ForegroundColor Yellow
    }
    Write-Host "Trying direct connection to game service..." -ForegroundColor Yellow
    try {
        $directResult = Invoke-RestMethod -Uri "http://localhost:8072/api/v1/game/classes/basic" -Method GET
        Write-Host "Direct connection works! Issue is with Nginx/Oathkeeper routing" -ForegroundColor Yellow
        $classesResult = $directResult
        $firstClass = $classesResult.data[0]
    } catch {
        Write-Host "Direct connection also failed" -ForegroundColor Red
        exit 1
    }
}

# 4. Create hero
Write-Host "`n[4] Creating hero..." -ForegroundColor Cyan
$createHeroBody = @{
    class_id = $firstClass.id
    hero_name = "TestHero"
    description = "Test hero"
} | ConvertTo-Json

try {
    $heroResult = Invoke-RestMethod -Uri "$baseUrl/heroes" -Method POST -Body $createHeroBody -ContentType "application/json" -Headers $headers
    Write-Host "SUCCESS: Hero created" -ForegroundColor Green
    $heroId = $heroResult.data.id
    Write-Host "Hero ID: $heroId" -ForegroundColor Gray
    Write-Host "Hero Name: $($heroResult.data.hero_name)" -ForegroundColor Gray
    Write-Host "Level: $($heroResult.data.current_level)" -ForegroundColor Gray
    Write-Host "Class: $($firstClass.class_name)" -ForegroundColor Gray
} catch {
    Write-Host "FAILED: Create hero failed" -ForegroundColor Red
    Write-Host "Error: $($_.Exception.Message)" -ForegroundColor Red
    if ($_.ErrorDetails) {
        Write-Host $_.ErrorDetails.Message -ForegroundColor Yellow
    }
    exit 1
}

# 5. Get hero full info (NEW API)
Write-Host "`n[5] Getting hero full info (NEW API)..." -ForegroundColor Cyan
try {
    $heroFullResult = Invoke-RestMethod -Uri "$baseUrl/heroes/$heroId/full" -Method GET -Headers $headers
    Write-Host "SUCCESS: Got hero full info" -ForegroundColor Green
    Write-Host "Hero: $($heroFullResult.data.hero_name)" -ForegroundColor Gray
    Write-Host "Class: $($heroFullResult.data.class.class_name) ($($heroFullResult.data.class.tier))" -ForegroundColor Gray
    Write-Host "Attributes: $($heroFullResult.data.attributes.Count)" -ForegroundColor Gray
    Write-Host "Skills: $($heroFullResult.data.skills.Count)" -ForegroundColor Gray
} catch {
    Write-Host "FAILED: Get hero full info failed" -ForegroundColor Red
    Write-Host "Error: $($_.Exception.Message)" -ForegroundColor Red
}

# 6. Get hero attributes
Write-Host "`n[6] Getting hero attributes..." -ForegroundColor Cyan
try {
    $attrsResult = Invoke-RestMethod -Uri "$baseUrl/heroes/$heroId/attributes" -Method GET -Headers $headers
    Write-Host "SUCCESS: Got $($attrsResult.data.Count) attributes" -ForegroundColor Green
    $attrsResult.data | ForEach-Object {
        Write-Host "  - $($_.attribute_name): $($_.final_value) (base: $($_.base_value), bonus: $($_.class_bonus))" -ForegroundColor Gray
    }
} catch {
    Write-Host "FAILED: Get attributes failed" -ForegroundColor Red
}

# 7. Get available skills
Write-Host "`n[7] Getting available skills..." -ForegroundColor Cyan
try {
    $skillsResult = Invoke-RestMethod -Uri "$baseUrl/heroes/$heroId/skills/available" -Method GET -Headers $headers
    Write-Host "SUCCESS: Got $($skillsResult.data.Count) available skills" -ForegroundColor Green
    if ($skillsResult.data.Count -gt 0) {
        $firstSkill = $skillsResult.data[0]
        Write-Host "First skill: $($firstSkill.skill_name)" -ForegroundColor Gray
    }
} catch {
    Write-Host "FAILED: Get available skills failed" -ForegroundColor Red
}

# 8. Learn a skill (if available)
if ($skillsResult.data.Count -gt 0 -and $firstSkill) {
    Write-Host "`n[8] Learning skill..." -ForegroundColor Cyan
    $learnSkillBody = @{
        skill_id = $firstSkill.skill_id
    } | ConvertTo-Json
    
    try {
        $learnResult = Invoke-RestMethod -Uri "$baseUrl/heroes/$heroId/skills/learn" -Method POST -Body $learnSkillBody -ContentType "application/json" -Headers $headers
        Write-Host "SUCCESS: Skill learned" -ForegroundColor Green
        Write-Host "Skill: $($firstSkill.skill_name)" -ForegroundColor Gray
    } catch {
        Write-Host "FAILED: Learn skill failed" -ForegroundColor Red
        Write-Host "Error: $($_.Exception.Message)" -ForegroundColor Red
        if ($_.ErrorDetails) {
            Write-Host $_.ErrorDetails.Message -ForegroundColor Yellow
        }
    }
}

# 9. Get learned skills
Write-Host "`n[9] Getting learned skills..." -ForegroundColor Cyan
try {
    $learnedSkillsResult = Invoke-RestMethod -Uri "$baseUrl/heroes/$heroId/skills/learned" -Method GET -Headers $headers
    Write-Host "SUCCESS: Got $($learnedSkillsResult.data.Count) learned skills" -ForegroundColor Green
} catch {
    Write-Host "FAILED: Get learned skills failed" -ForegroundColor Red
}

# 10. Add experience to hero
Write-Host "`n[10] Adding experience to hero..." -ForegroundColor Cyan
$addExpBody = @{
    amount = 1000
} | ConvertTo-Json

try {
    $expResult = Invoke-RestMethod -Uri "$baseUrl/heroes/$heroId/experience" -Method POST -Body $addExpBody -ContentType "application/json" -Headers $headers
    Write-Host "SUCCESS: Experience added" -ForegroundColor Green
    Write-Host "New level: $($expResult.data.hero.current_level)" -ForegroundColor Gray
    Write-Host "Available XP: $($expResult.data.hero.experience_available)" -ForegroundColor Gray
} catch {
    Write-Host "FAILED: Add experience failed" -ForegroundColor Red
    Write-Host "Error: $($_.Exception.Message)" -ForegroundColor Red
}

# 11. Allocate attribute point
Write-Host "`n[11] Allocating attribute point..." -ForegroundColor Cyan
if ($attrsResult.data.Count -gt 0) {
    $firstAttr = $attrsResult.data[0]
    $allocateBody = @{
        attribute_code = $firstAttr.attribute_code
        points_to_add = 1
    } | ConvertTo-Json
    
    try {
        $allocateResult = Invoke-RestMethod -Uri "$baseUrl/heroes/$heroId/attributes/allocate" -Method POST -Body $allocateBody -ContentType "application/json" -Headers $headers
        Write-Host "SUCCESS: Attribute point allocated" -ForegroundColor Green
        Write-Host "Attribute: $($firstAttr.attribute_name)" -ForegroundColor Gray
        Write-Host "New value: $($allocateResult.data.new_value)" -ForegroundColor Gray
    } catch {
        Write-Host "FAILED: Allocate attribute failed" -ForegroundColor Red
        Write-Host "Error: $($_.Exception.Message)" -ForegroundColor Red
        if ($_.ErrorDetails) {
            Write-Host $_.ErrorDetails.Message -ForegroundColor Yellow
        }
    }
}

Write-Host "`n========================================" -ForegroundColor Green
Write-Host "Test Complete!" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Green
Write-Host "`nSummary:" -ForegroundColor Cyan
Write-Host "- User registered and logged in: YES" -ForegroundColor Green
Write-Host "- Hero created: YES" -ForegroundColor Green
Write-Host "- Hero full info retrieved: CHECK ABOVE" -ForegroundColor Yellow
Write-Host "- Attributes retrieved: CHECK ABOVE" -ForegroundColor Yellow
Write-Host "- Skills system working: CHECK ABOVE" -ForegroundColor Yellow
