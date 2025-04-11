# CodeSpinneret MCP Server Build Script

param (
    [string]$Target = "build",
    [int]$Port = 8080,
    [int]$Timeout = 60,
    [string]$BaseURL = ""
)

# Configuration
$ServerBin = "bin/mcp-server.exe"
$ClientBin = "bin/mcp-client.exe"
$ServerCmd = "./cmd/mcp-server"
$ClientCmd = "./cmd/mcp-client"
$DemoTimeout = 300 # 5 minutes

# Set colors for status messages
$ErrorColor = "Red"
$SuccessColor = "Green"
$InfoColor = "Cyan"
$WarningColor = "Yellow"

# Helper functions
function Write-Status {
    param (
        [string]$Message,
        [string]$Color = $InfoColor
    )
    Write-Host $Message -ForegroundColor $Color
}

function Test-CommandExists {
    param (
        [string]$Command
    )
    return [bool](Get-Command -Name $Command -ErrorAction SilentlyContinue)
}

function Build-Server {
    Write-Status "Building MCP server..."
    try {
        go build -o $ServerBin $ServerCmd
        if ($LASTEXITCODE -ne 0) {
            Write-Status "Failed to build server" $ErrorColor
            return $false
        }
        Write-Status "Server built successfully" $SuccessColor
        return $true
    }
    catch {
        Write-Status "Error building server: $_" $ErrorColor
        return $false
    }
}

function Build-Client {
    Write-Status "Building MCP client..."
    try {
        go build -o $ClientBin $ClientCmd
        if ($LASTEXITCODE -ne 0) {
            Write-Status "Failed to build client" $ErrorColor
            return $false
        }
        Write-Status "Client built successfully" $SuccessColor
        return $true
    }
    catch {
        Write-Status "Error building client: $_" $ErrorColor
        return $false
    }
}

function Run-Tests {
    Write-Status "Running tests..."
    try {
        go test ./...
        if ($LASTEXITCODE -ne 0) {
            Write-Status "Tests failed" $ErrorColor
            return $false
        }
        Write-Status "Tests passed" $SuccessColor
        return $true
    }
    catch {
        Write-Status "Error running tests: $_" $ErrorColor
        return $false
    }
}

function Clean-Build {
    Write-Status "Cleaning build artifacts..."
    try {
        if (Test-Path $ServerBin) {
            Remove-Item $ServerBin -Force
        }
        if (Test-Path $ClientBin) {
            Remove-Item $ClientBin -Force
        }
        Write-Status "Clean completed" $SuccessColor
        return $true
    }
    catch {
        Write-Status "Error cleaning build: $_" $ErrorColor
        return $false
    }
}

function Run-Server {
    param (
        [int]$Port = 8080,
        [string]$BaseURL = "",
        [int]$Timeout = 300
    )
    
    Write-Status "Starting MCP server on port $Port..."
    
    $serverArgs = @(
        "-port", $Port
    )
    
    if ($BaseURL -ne "") {
        $serverArgs += "-baseurl", $BaseURL
    }
    
    if ($Timeout -gt 0) {
        $serverArgs += "-timeout", $Timeout
    }
    
    try {
        & $ServerBin $serverArgs
        return $true
    }
    catch {
        Write-Status "Error running server: $_" $ErrorColor
        return $false
    }
}

function Run-Client {
    param (
        [string]$ServerURL = "http://localhost:8080",
        [int]$Timeout = 60
    )
    
    Write-Status "Starting MCP client connecting to $ServerURL..."
    
    $clientArgs = @(
        "-server", $ServerURL
    )
    
    if ($Timeout -gt 0) {
        $clientArgs += "-timeout", $Timeout
    }
    
    try {
        & $ClientBin $clientArgs
        return $true
    }
    catch {
        Write-Status "Error running client: $_" $ErrorColor
        return $false
    }
}

function Run-Demo {
    param (
        [int]$Port = 8080,
        [int]$Timeout = $DemoTimeout
    )
    
    Write-Status "Starting demo mode (server and client)..." $InfoColor
    
    # Check if port is in use
    try {
        $portCheck = Get-NetTCPConnection -LocalPort $Port -ErrorAction SilentlyContinue
        if ($portCheck) {
            Write-Status "Port $Port is already in use. Please choose a different port." $ErrorColor
            return $false
        }
    }
    catch {
        # Port is likely free if we get an error
    }
    
    # Start server in background
    $serverURL = "http://localhost:$Port"
    $serverProcess = Start-Process -FilePath $ServerBin -ArgumentList "-port", $Port, "-timeout", $Timeout -PassThru -NoNewWindow
    
    # Give the server time to start
    Write-Status "Waiting for server to start..." $InfoColor
    Start-Sleep -Seconds 2
    
    # Run client
    Write-Status "Starting client..." $InfoColor
    & $ClientBin -server $serverURL -timeout 30
    
    # Stop server
    Write-Status "Stopping server..." $InfoColor
    Stop-Process -Id $serverProcess.Id -Force
    
    Write-Status "Demo completed" $SuccessColor
    return $true
}

function Update-Dependencies {
    Write-Status "Updating dependencies..."
    try {
        go get -u ./...
        if ($LASTEXITCODE -ne 0) {
            Write-Status "Failed to update dependencies" $ErrorColor
            return $false
        }
        
        go mod tidy
        if ($LASTEXITCODE -ne 0) {
            Write-Status "Failed to tidy go.mod" $ErrorColor
            return $false
        }
        
        go mod vendor
        if ($LASTEXITCODE -ne 0) {
            Write-Status "Failed to vendor dependencies" $ErrorColor
            return $false
        }
        
        Write-Status "Dependencies updated successfully" $SuccessColor
        return $true
    }
    catch {
        Write-Status "Error updating dependencies: $_" $ErrorColor
        return $false
    }
}

# Check if Go is installed
if (-not (Test-CommandExists "go")) {
    Write-Status "Go is not installed or not in PATH. Please install Go and try again." $ErrorColor
    exit 1
}

# Create bin directory if it doesn't exist
if (-not (Test-Path "bin")) {
    New-Item -ItemType Directory -Path "bin" | Out-Null
}

# Execute the specified target
switch ($Target) {
    "build" {
        Build-Server
        Build-Client
    }
    "build-server" {
        Build-Server
    }
    "build-client" {
        Build-Client
    }
    "test" {
        Run-Tests
    }
    "clean" {
        Clean-Build
    }
    "run-server" {
        if (Build-Server) {
            Run-Server -Port $Port -BaseURL $BaseURL -Timeout $Timeout
        }
    }
    "run-client" {
        if (Build-Client) {
            $serverURL = if ($BaseURL -ne "") { $BaseURL } else { "http://localhost:$Port" }
            Run-Client -ServerURL $serverURL -Timeout $Timeout
        }
    }
    "demo" {
        if (Build-Server -and Build-Client) {
            Run-Demo -Port $Port -Timeout $Timeout
        }
    }
    "update-deps" {
        Update-Dependencies
    }
    default {
        Write-Status "Unknown target: $Target" $ErrorColor
        Write-Status "Available targets: build, build-server, build-client, test, clean, run-server, run-client, demo, update-deps" $InfoColor
    }
}