# Windows IPv6 Monitor Service

A Windows service that monitors your global IPv6 address and sends notifications through Telegram when changes are detected.

## Features

- Runs as a Windows service
- Monitors global IPv6 address at configurable intervals
- Sends notifications through Telegram when IPv6 changes
- Configurable logging levels with file output
- PowerShell installation script included

## Prerequisites

- Windows operating system
- Go 1.16 or later
- Telegram bot token and chat ID
- PowerShell (for service installation)
- Administrator privileges for installation

## Building

1. Clone the repository:
```bash
git clone https://github.com/fauzanelka/windows-ipv6-monitor-service.git
cd windows-ipv6-monitor-service
```

2. Install dependencies:
```bash
go mod tidy
```

3. Build the binary:
```bash
go build -o ipv6-monitor.exe ./cmd/service
```

## Installation

1. Create a Telegram bot and get your bot token from [@BotFather](https://t.me/botfather)
2. Get your chat ID by sending a message to [@userinfobot](https://t.me/userinfobot)
3. Run the installation script with administrator privileges:

```powershell
.\scripts\install-service.ps1 -BotToken "your-bot-token" -ChatID "your-chat-id"
```

The installation script will:
- Copy the binary to `C:\ProgramData\IPv6MonitorService\ipv6-monitor.exe`
- Create necessary directories for logs and binary
- Install and start the Windows service

Optional parameters:
- `-CheckInterval`: IPv6 check interval in minutes (default: 5)
- `-LogLevel`: Logging level (debug, info, warn, error) (default: info)
- `-LogFile`: Log file path (default: C:\ProgramData\IPv6MonitorService\ipv6-monitor.log)

## Manual Usage

You can also run the service directly from the command line:

```bash
./ipv6-monitor.exe --bot-token "your-bot-token" --chat-id "your-chat-id" --check-interval 5 --log-level info --log-file "ipv6-monitor.log"
```

## Service Management

- Check service status:
```powershell
Get-Service IPv6MonitorService
```

- Stop the service:
```powershell
Stop-Service IPv6MonitorService
```

- Start the service:
```powershell
Start-Service IPv6MonitorService
```

- Remove the service:
```powershell
# Stop the service first
Stop-Service IPv6MonitorService

# Remove the service
sc.exe delete IPv6MonitorService

# Remove the installation directory (optional)
Remove-Item -Path "C:\ProgramData\IPv6MonitorService" -Recurse -Force
```

## Directory Structure

When installed as a service:
- Binary location: `C:\ProgramData\IPv6MonitorService\ipv6-monitor.exe`
- Default log file: `C:\ProgramData\IPv6MonitorService\ipv6-monitor.log`

## Logging

The service uses structured logging with different levels:
- `debug`: Detailed information for debugging
- `info`: General operational information
- `warn`: Warning messages for potential issues
- `error`: Error messages for actual problems

Logs are written to both the console and a log file. The log file location can be configured using the `--log-file` flag or `-LogFile` parameter in the installation script. By default, when installed as a service, logs are written to `C:\ProgramData\IPv6MonitorService\ipv6-monitor.log`.

The log file uses JSON format for better parsing and analysis, while console output uses a more human-readable format. Log files are automatically rotated to prevent excessive disk usage.

You can also view service logs in the Windows Event Viewer under the Application logs.

## License

MIT License 