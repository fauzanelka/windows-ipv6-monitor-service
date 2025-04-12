# Requires -RunAsAdministrator

param (
    [Parameter(Mandatory=$true)]
    [string]$BotToken,
    
    [Parameter(Mandatory=$true)]
    [string]$ChatID,
    
    [Parameter(Mandatory=$false)]
    [int]$CheckInterval = 5,
    
    [Parameter(Mandatory=$false)]
    [string]$LogLevel = "info",

    [Parameter(Mandatory=$false)]
    [string]$LogFile = "C:\ProgramData\IPv6MonitorService\ipv6-monitor.log"
)

$ServiceName = "IPv6MonitorService"
$InstallDir = "C:\ProgramData\IPv6MonitorService"
$SourceBinary = Join-Path $PSScriptRoot "..\ipv6-monitor.exe"
$BinaryPath = Join-Path $InstallDir "ipv6-monitor.exe"
$Description = "IPv6 address monitoring service with Telegram notifications"

# Check if source binary exists
if (-not (Test-Path $SourceBinary)) {
    Write-Error "Service binary not found at: $SourceBinary"
    exit 1
}

# Create installation directory if it doesn't exist
if (-not (Test-Path $InstallDir)) {
    New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
    Write-Host "Created installation directory: $InstallDir"
}

# Copy binary to installation directory
try {
    Copy-Item -Path $SourceBinary -Destination $BinaryPath -Force
    Write-Host "Copied binary to: $BinaryPath"
} catch {
    Write-Error "Failed to copy binary: $_"
    exit 1
}

# Create log directory if it doesn't exist
$LogDir = Split-Path -Parent $LogFile
if (-not (Test-Path $LogDir)) {
    New-Item -ItemType Directory -Path $LogDir -Force | Out-Null
    Write-Host "Created log directory: $LogDir"
}

# Check if service already exists
$service = Get-Service -Name $ServiceName -ErrorAction SilentlyContinue
if ($service) {
    Write-Host "Service already exists. Stopping and removing..."
    Stop-Service -Name $ServiceName -Force
    Start-Sleep -Seconds 2
    sc.exe delete $ServiceName
    Start-Sleep -Seconds 2
}

# Construct service arguments
$ServiceArgs = @(
    "--bot-token=$BotToken",
    "--chat-id=$ChatID",
    "--check-interval=$CheckInterval",
    "--log-level=$LogLevel",
    "--log-file=`"$LogFile`""
)

# Create the service
$BinPath = "`"$BinaryPath`" $($ServiceArgs -join ' ')"
Write-Host "Creating service with path: $BinPath"

sc.exe create $ServiceName binPath= $BinPath start= auto DisplayName= "IPv6 Monitor Service"
sc.exe description $ServiceName $Description

# Start the service
Start-Service -Name $ServiceName

Write-Host "Service installed and started successfully!"
Write-Host "Service Name: $ServiceName"
Write-Host "Binary Path: $BinaryPath"
Write-Host "Check Interval: $CheckInterval minutes"
Write-Host "Log Level: $LogLevel"
Write-Host "Log File: $LogFile"

Write-Host "`nTo uninstall the service:"
Write-Host "1. Stop-Service $ServiceName"
Write-Host "2. sc.exe delete $ServiceName"
Write-Host "3. Remove-Item -Path `"$InstallDir`" -Recurse -Force" 