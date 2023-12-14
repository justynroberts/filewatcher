


# 👀 FileWatcher 👀

> Watch a directory, and trigger Automation. Its not complex :) 


👀 Watcher gives an example of how to run automation jobs as the result of a file activity.  
It centers around the pip module watcher ([watcher](https://pypi.org/project/watcher/)) and is configurable to run an existing webhook.

Webhooks are preferable as they will obey queuing and can easily be modified. The file/path is passed as an option should you need to modify it. 
The main configuration (`config.json`) is based on the json file format:  
Example:  

    "FileWatcher": {
    "directories": ["/Users/jroberts/work/filewatcher/test", "directory2"],
    "event_types": ["created", "modified"],
    "file_extension_pattern": "*.csv",
    "post_url": "https://YOURSERVERWEBHOOKURL",
    "authentication_header": "YOURHEADER"
    }

File Sections:-

    directories 
Full directory path to be monitored (sub directories will automatically be monitored)


   `event_types`
Self explanatory - one or more of the following types:  

 - created 	
 - modified 	
 - deleted 	
 - moved.


`file_extension_pattern`

Extension to monitor. If ommited `*.*` is assumed

 `post_url`  
 
Full webhook location eg 

> https://myhost.domain.com/api/webhook

 `authentication_header`  
 A good security practice is to use the additional authentication header for each webhook. This is generated at runtime when the webhook is initially configured

**🔧 Running the watcher.**

Clone the repo to a directory.

The watcher will need Python3 and the following modules with pip:
- `watcher` - the core watcher module
- `requests` - the library used for initiating the webhook

A requirements.txt is also supplied for purists.
   
>  pip install -r requirements.txt

## Running

`python3 main.py`  

The can be installed something like [here](https://medium.com/codex/setup-a-python-script-as-a-service-through-systemctl-systemd-f0cc55a42267)

Or could work with windows using the nssm ( [https://nssm.cc/](https://nssm.cc/))

## Automation Webhook
This will pass filepath as a payload.
add this as an advanced webhook option as `$.path` - this will be passed as an option across to use within your automation jobs.

## 📏 Scale:  

Estimated a few hundred directories, but could be split over different executables/runners to scale to more.

**📝 Notes:**  
Presently 1 file = 1 webhook invocation, but could be modified to batch files.
Some nice features such as POST being non blocking and Log integration. 

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
