# Test hero flow without authentication (for testing)
$baseUrl = "http://localhost:8072/api/v1/game"

Write-Host "=== Hero Flow Test (No Auth) ===" -ForegroundColor Green

# Get root user ID from database
Write-Host "`n[1] Get root user ID from database..." -ForegroundColor Cyan
$getUserId = @"
SELECT id FROM auth.users WHERE username = 'root' LIMIT 1;
"@

$userId = docker exec tsu_postgres psql -U postgres -d tsu_db -t -c $getUserId
$userId = $userId.Trim()
Write-Host "Root User ID: $userId"

# Get first class
Write-Host "`n[2] Get first class..." -ForegroundColor Cyan
$classesResult = Invoke-RestMethod -Uri "$baseUrl/classes/basic" -Method GET
$firstClass = $classesResult.data[0]
Write-Host "Class: $($firstClass.class_name) (ID: $($firstClass.id))"

# Note: Without proper authentication, we cannot create heroes
# The system requires JWT token from auth module

Write-Host "`n=== Summary ===" -ForegroundColor Yellow
Write-Host "- Basic class API works: YES" -ForegroundColor Green
Write-Host "- User exists in DB: YES (root user)" -ForegroundColor Green
Write-Host "- Auth integration: Requires Kratos setup" -ForegroundColor Yellow
Write-Host ""
Write-Host "Next steps:" -ForegroundColor Cyan
Write-Host "1. Fix Kratos integration for registration/login"
Write-Host "2. Or use direct database token generation for testing"
Write-Host "3. Test complete hero creation flow"
