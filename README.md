# 👀 FileWatcher 👀

> Watch a directory, and trigger Automation. Its not complex :) 

👀 Watcher gives an example of how to run automation jobs as the result of a file activity.  
This project is available in both Python and Go versions, with the Go version supporting cross-platform builds.

## Configuration

The main configuration (`config.json`) is based on the json file format:  
Example:  

```json
{
    "FileWatcher": {
        "directories": ["/Users/jroberts/work/filewatcher/test", "directory2"],
        "event_types": ["created", "modified"],
        "file_extension_pattern": "*.csv",
        "post_url": "https://YOURSERVERWEBHOOKURL",
        "authentication_header": "YOURHEADER"
    }
}
```

File Sections:

`directories` 
Full directory path to be monitored (sub directories will automatically be monitored)


`event_types`
Self explanatory - one or more of the following types:  

 - created 	
 - modified 	
 - deleted 	
 - moved

`file_extension_pattern`

Extension to monitor. If omitted `*.*` is assumed

`post_url`  
 
Full webhook location eg 

> https://myhost.domain.com/api/webhook

`authentication_header`  
A good security practice is to use the additional authentication header for each webhook. This is generated at runtime when the webhook is initially configured

## 🔧 Running the watcher

### Go Version

#### Prerequisites

- Go 1.18 or higher

#### Building

You can build the application for your current platform:

```bash
go build -o watcher
```

Or use the included build script to build for multiple platforms:

```bash
go run build.go
```

This will create binaries for various platforms in the `dist` directory.

Options for the build script:

```
-output string
    Output directory for binaries (default "dist")
-version string
    Version number for the build (default "1.0.0")
-current
    Build only for the current platform
```

Example:

```bash
go run build.go -output releases -version 1.2.0
```

#### Running

```bash
./watcher
```

Or specify a custom config file:

```bash
./watcher -config /path/to/config.json
```

### Python Version

Clone the repo to a directory.

The watcher will need Python3 and the following modules with pip:
- `watchdog` - the core watcher module
- `requests` - the library used for initiating the webhook

A requirements.txt is also supplied for purists.
   
>  pip install -r requirements.txt

#### Running

`python3 main.py`  

## Running as a Service

### Linux (systemd)

Create a systemd service file:

```bash
sudo nano /etc/systemd/system/watcher.service
```

Add the following content (adjust paths as needed):

```
[Unit]
Description=File Watcher Service
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

Enable and start the service:

```bash
sudo systemctl enable watcher.service
sudo systemctl start watcher.service
```

### Windows

For Windows, you can use NSSM (Non-Sucking Service Manager): [https://nssm.cc/](https://nssm.cc/)

## Automation Webhook
This will pass filename and filepath as a payload.
Add this as an advanced webhook option as `$.path` - this will be passed as an option across to use within your automation jobs.

## 📏 Scale:  

Estimated a few hundred directories, but could be split over different executables/runners to scale to more.

**📝 Notes:**  
 - Presently 1 file = 1 webhook invocation, but could be modified to batch files.
 - Some nice features such as POST being non blocking and Log integration. 
 - Consider both the parallelism and the queuing capability of the job for this.
 - The Go version offers improved performance and cross-platform support.
 - If service stops, files could be missed, and wouldn't be picked up on restart
   
**⚠️ Issues?**  
Please post to the repo, not the author.

**📜 License**  

MIT License  

> Permission is hereby granted, free of charge, to any person obtaining
> a copy of this software and associated documentation files (the
> "Software"), to deal in the Software without restriction, including
> without limitation the rights to use, copy, modify, merge, publish,
> distribute, sublicense, and/or sell copies of the Software, and to
> permit persons to whom the Software is furnished to do so, subject to
> the following conditions:
> 
>  The above copyright notice and this permission notice shall be
> included in all copies or substantial portions of the Software.
> 
> THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
> EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF
> MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
> IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY
> CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT,
> TORT OR OTHERWISE, ARISING FROM,OUT OF OR IN CONNECTION WITH THE
> SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
