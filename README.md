<div align="center">
  
# 👀 FileWatcher

**Elegant file system monitoring with automated webhook triggers**

[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-1.18%2B-00ADD8.svg)](https://golang.org/)
[![Python Version](https://img.shields.io/badge/Python-3.6%2B-blue.svg)](https://www.python.org/)

*Watch directories, detect changes, trigger automation - Simple yet powerful*

</div>

## 📋 Table of Contents

- [Overview](#-overview)
- [Features](#-features)
- [Installation](#-installation)
  - [Go Version](#go-version)
  - [Python Version](#python-version)
- [Configuration](#-configuration)
- [Usage](#-usage)
- [Running as a Service](#-running-as-a-service)
  - [Linux (systemd)](#linux-systemd)
  - [Windows](#windows)
- [Webhook Integration](#-webhook-integration)
- [Scalability](#-scalability)
- [Contributing](#-contributing)
- [License](#-license)

## 🔍 Overview

FileWatcher is a lightweight, cross-platform utility designed to monitor file system events and trigger webhook-based automation workflows. Available in both Go and Python implementations, it provides a simple yet powerful solution for initiating automated processes in response to file creation, modification, deletion, or movement.

The Go version offers enhanced performance and cross-platform compatibility, while the Python version provides excellent flexibility and ease of use.

## ✨ Features

- **Real-time monitoring** of file system events
- **Configurable event filtering** (created, modified, deleted, moved)
- **Pattern matching** for specific file extensions
- **Webhook integration** with custom authentication
- **Cross-platform support** via Go implementation
- **Non-blocking webhook calls** for improved performance
- **Comprehensive logging**
- **Easy deployment** as a system service

## 📥 Installation

### Go Version

#### Prerequisites

- Go 1.18 or higher

#### Building from Source

Build for your current platform:

```bash
go build -o watcher
```

Or use the included build script for multi-platform builds:

```bash
go run tools/build.go
```

You can also use the Makefile for easier building:

```bash
# Build for current platform
make build

# Build for all platforms
make build-all

# Build only for current platform using build script
make build-current
```

This creates binaries for various platforms in the `dist` directory.

#### Build Script Options

```
-output string
    Output directory for binaries (default "dist")
-version string
    Version number for the build (default "1.0.0")
-current
    Build only for the current platform
```

Example with custom options:

```bash
go run build.go -output releases -version 1.2.0
```

### Python Version

#### Prerequisites

- Python 3.6 or higher
- Required packages:
  - `watchdog`: Core file system monitoring
  - `requests`: HTTP client for webhook integration

#### Installation Steps

1. Clone the repository
2. Install dependencies:

```bash
pip install -r requirements.txt
```

## ⚙️ Configuration

FileWatcher uses a JSON configuration file (`config.json` by default) with the following structure:

```json
{
    "FileWatcher": {
        "directories": ["/path/to/watch", "/another/path"],
        "event_types": ["created", "modified", "deleted", "moved"],
        "file_extension_pattern": "*.csv",
        "post_url": "https://your-webhook-url.com/endpoint",
        "authentication_header": "YOUR-AUTH-TOKEN"
    }
}
```

### Configuration Parameters

| Parameter | Description | Required | Default |
|-----------|-------------|----------|---------|
| `directories` | Array of directory paths to monitor (includes subdirectories) | Yes | - |
| `event_types` | Array of event types to monitor (`created`, `modified`, `deleted`, `moved`) | Yes | - |
| `file_extension_pattern` | File pattern to match (e.g., `*.csv`, `*.log`) | No | `*.*` |
| `post_url` | Webhook URL to receive event notifications | Yes | - |
| `authentication_header` | Authentication token for webhook security | Yes | - |

## 🚀 Usage

### Go Version

Run with default configuration:

```bash
./watcher
```

Specify a custom configuration file:

```bash
./watcher -config /path/to/custom-config.json
```

### Python Version

Run with default configuration:

```bash
python3 main.py
```

Specify a custom configuration file:

```bash
python3 main.py -c /path/to/custom-config.json
```

## 🔄 Running as a Service

### Linux (systemd)

1. Create a systemd service file:

```bash
sudo nano /etc/systemd/system/watcher.service
```

2. Add the following content (adjust paths as needed):

#### For Go Version

```ini
[Unit]
Description=File Watcher Service (Go)
After=network.target

[Service]
Type=simple
User=yourusername
WorkingDirectory=/path/to/watcher
ExecStart=/path/to/watcher/watcher
Restart=on-failure

[Install]
WantedBy=multi-user.target
```

#### For Python Version

```ini
[Unit]
Description=File Watcher Service (Python)
After=network.target

[Service]
Type=simple
User=yourusername
WorkingDirectory=/path/to/watcher
ExecStart=/usr/bin/python3 /path/to/watcher/main.py
Restart=on-failure

[Install]
WantedBy=multi-user.target
```

3. Enable and start the service:

```bash
sudo systemctl enable watcher.service
sudo systemctl start watcher.service
```

### Windows

For Windows environments, you can use NSSM (Non-Sucking Service Manager):

1. Download NSSM from [https://nssm.cc/](https://nssm.cc/)
2. Install the service:

#### For Go Version

```
nssm install FileWatcher-Go
nssm set FileWatcher-Go Application C:\path\to\watcher.exe
nssm set FileWatcher-Go AppDirectory C:\path\to\watcher
```

#### For Python Version

```
nssm install FileWatcher-Python
nssm set FileWatcher-Python Application C:\path\to\python.exe
nssm set FileWatcher-Python AppParameters C:\path\to\watcher\main.py
nssm set FileWatcher-Python AppDirectory C:\path\to\watcher
```

3. Start the service:

```
nssm start FileWatcher-Go
# or
nssm start FileWatcher-Python
```

## 🔗 Webhook Integration

When a matching file event occurs, FileWatcher sends a webhook POST request with the following payload:

```json
{
  "filepath": "/full/path/to/directory",
  "filename": "file.csv",
  "event_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

The webhook includes the authentication header specified in your configuration.

Both implementations use:
- Non-blocking webhook calls (Go uses goroutines, Python uses threads)
- Debouncing to prevent duplicate events
- Unique event IDs for tracking

For integration with automation platforms, use `$.filepath` and `$.filename` to reference the file information in your automation workflows.

## 📏 Scalability

FileWatcher is designed to efficiently monitor several hundred directories on a single instance. For larger deployments:

- Deploy multiple instances with different directory subsets
- Consider batching file events for high-volume scenarios
- Ensure your webhook endpoint can handle the expected request volume

**Performance Considerations:**
- The Go version offers significantly better performance for high-volume monitoring
- Webhook calls are non-blocking to prevent monitoring interruptions
- Consider both the parallelism and queuing capabilities of your automation system

**Limitations:**
- Events that occur while the service is stopped will not be detected upon restart
- Very high file change rates may require custom throttling or batching

## 👥 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

Please report issues via the GitHub issue tracker.

## 📜 License

This project is licensed under the MIT License - see below for details:

```
MIT License

Copyright (c) 2025 Justyn Roberts

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
