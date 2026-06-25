# =====================================================
# Casbin Demo API Test Script
# Usage: powershell -ExecutionPolicy Bypass -File ./scripts/test-api.ps1
# =====================================================

$ErrorActionPreference = "Continue"
$BaseUrl = "http://localhost:8080"
$Passed = 0
$Failed = 0
$Results = @()

function Write-Color($text, $color = "White") {
    Write-Host $text -ForegroundColor $color
}

function Test-Endpoint {
    param(
        [string]$Name,
        [string]$Method,
        [string]$Url,
        [object]$Body = $null,
        [hashtable]$Headers = @{},
        [int]$ExpectedStatus = 200,
        [scriptblock]$ValidateResponse = $null
    )

    $fullUrl = "$BaseUrl$Url"
    $result = @{ Name = $Name; Passed = $false; Error = $null; StatusCode = 0; Response = $null }

    try {
        $params = @{
            Method = $Method
            Uri = $fullUrl
            Headers = @{ "Content-Type" = "application/json" } + $Headers
            UseBasicParsing = $true
            ErrorAction = "Stop"
        }

        if ($Body -ne $null) {
            $params.Body = ($Body | ConvertTo-Json -Depth 10)
        }

        $response = Invoke-WebRequest @params
        $result.StatusCode = [int]$response.StatusCode
        $result.Response = $response.Content | ConvertFrom-Json -Depth 10

        if ($result.StatusCode -ne $ExpectedStatus) {
            $result.Error = "Expected status $ExpectedStatus, got $($result.StatusCode)"
        } elseif ($result.Response.code -ne 0) {
            $result.Error = "Response code: $($result.Response.code), message: $($result.Response.message)"
        } elseif ($ValidateResponse -ne $null) {
            $validateResult = & $ValidateResponse $result.Response
            if ($validateResult -ne $null) {
                $result.Error = $validateResult
            } else {
                $result.Passed = $true
            }
        } else {
            $result.Passed = $true
        }
    }
    catch {
        $result.Error = $_.Exception.Message
        if ($_.Exception.Response) {
            $reader = New-Object System.IO.StreamReader($_.Exception.Response.GetResponseStream())
            $errorBody = $reader.ReadToEnd()
            try {
                $result.Response = $errorBody | ConvertFrom-Json
                $result.StatusCode = [int]$_.Exception.Response.StatusCode
            } catch {}
        }
    }

    if ($result.Passed) {
        $script:Passed++
        Write-Color "[PASS] $Name" "Green"
    } else {
        $script:Failed++
        Write-Color "[FAIL] $Name - $($result.Error)" "Red"
        if ($result.Response) {
            Write-Color "  Response: $($result.Response | ConvertTo-Json -Depth 5 -Compress)" "DarkGray"
        }
    }

    $script:Results += $result
    return $result
}

Write-Color ""
Write-Color "========================================" "Cyan"
Write-Color "  Casbin Demo API Test Suite" "Cyan"
Write-Color "========================================" "Cyan"
Write-Color ""

Write-Color "[1] Testing Health Check..." "Yellow"
Test-Endpoint -Name "Health Endpoint" -Method "GET" -Url "/health" -ValidateResponse {
    param($r)
    if ($r.status -ne "ok") { return "Expected status ok" }
    if ($r.service -ne "casbin-demo") { return "Expected service casbin-demo" }
    return $null
}

Write-Color ""
Write-Color "[2] Testing Authentication..." "Yellow"

$adminToken = $null
$loginResult = Test-Endpoint -Name "Admin Login" -Method "POST" -Url "/api/login" `
    -Body @{ username = "admin"; password = "123456" } `
    -ValidateResponse {
        param($r)
        if (-not $r.data.token) { return "Missing token" }
        if ($r.data.user.username -ne "admin") { return "Expected admin user" }
        return $null
    }

if ($loginResult.Passed) {
    $adminToken = $loginResult.Response.data.token
    Write-Color "  Token obtained: $($adminToken.Substring(0, [Math]::Min(30, $adminToken.Length)))..." "DarkGray"
}

$userToken = $null
$userLoginResult = Test-Endpoint -Name "User Login" -Method "POST" -Url "/api/login" `
    -Body @{ username = "user"; password = "123456" } `
    -ValidateResponse {
        param($r)
        if (-not $r.data.token) { return "Missing token" }
        return $null
    }

if ($userLoginResult.Passed) {
    $userToken = $userLoginResult.Response.data.token
}

Test-Endpoint -Name "Login with wrong password" -Method "POST" -Url "/api/login" `
    -Body @{ username = "admin"; password = "wrongpass" } `
    -ExpectedStatus 200 `
    -ValidateResponse {
        param($r)
        if ($r.code -eq 0) { return "Should fail with wrong password" }
        return $null
    }

Test-Endpoint -Name "Login with empty fields" -Method "POST" -Url "/api/login" `
    -Body @{ username = ""; password = "" } `
    -ExpectedStatus 400

Write-Color ""
Write-Color "[3] Testing Authorized Endpoints..." "Yellow"

if ($adminToken) {
    $authHeaders = @{ Authorization = "Bearer $adminToken" }

    Test-Endpoint -Name "Get User Info" -Method "GET" -Url "/api/userinfo" -Headers $authHeaders `
        -ValidateResponse {
            param($r)
            if ($r.data.username -ne "admin") { return "Expected admin" }
            return $null
        }

    Test-Endpoint -Name "Get Dashboard (Admin)" -Method "GET" -Url "/api/dashboard" -Headers $authHeaders `
        -ValidateResponse {
            param($r)
            if (-not $r.data.stats) { return "Missing stats" }
            return $null
        }

    Test-Endpoint -Name "Get User List (Admin)" -Method "GET" -Url "/api/users" -Headers $authHeaders `
        -ValidateResponse {
            param($r)
            if (-not $r.data.list) { return "Missing user list" }
            return $null
        }

    Test-Endpoint -Name "Get Role List (Admin)" -Method "GET" -Url "/api/roles" -Headers $authHeaders `
        -ValidateResponse {
            param($r)
            if (-not $r.data) { return "Missing role list" }
            return $null
        }

    Test-Endpoint -Name "Get Permission List (Admin)" -Method "GET" -Url "/api/permissions" -Headers $authHeaders `
        -ValidateResponse {
            param($r)
            if (-not $r.data) { return "Missing permission list" }
            return $null
        }
} else {
    Write-Color "[SKIP] Authorized tests - no admin token" "DarkYellow"
}

Test-Endpoint -Name "Access without token" -Method "GET" -Url "/api/userinfo" `
    -ExpectedStatus 401

Write-Color ""
Write-Color "[4] Testing Permission Control..." "Yellow"

if ($userToken) {
    $userHeaders = @{ Authorization = "Bearer $userToken" }

    Test-Endpoint -Name "Get Dashboard (User)" -Method "GET" -Url "/api/dashboard" -Headers $userHeaders

    Test-Endpoint -Name "Get User List (User - should be denied)" -Method "GET" -Url "/api/users" -Headers $userHeaders `
        -ExpectedStatus 403
}

Write-Color ""
Write-Color "========================================" "Cyan"
Write-Color "  Test Results: $Passed passed, $Failed failed" $(if ($Failed -eq 0) { "Green" } else { "Red" })
Write-Color "========================================" "Cyan"
Write-Color ""

if ($Failed -gt 0) {
    exit 1
}
exit 0
