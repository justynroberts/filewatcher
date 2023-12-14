"""MIT License

(c) 2023/2024 Justyn Roberts
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
SOFTWARE."""

import time
import os
import logging
import json
import fnmatch
import requests
from watchdog.observers import Observer
from watchdog.events import FileSystemEventHandler

class Watcher:
    def __init__(self, directories, event_handler):
        self.observer = Observer()
        self.directories = directories
        self.event_handler = event_handler

    def run(self):
        for directory in self.directories:
            self.observer.schedule(self.event_handler, directory, recursive=True)
        self.observer.start()
        try:
            while True:
                time.sleep(5)
        except KeyboardInterrupt:
            self.observer.stop()
            logging.info("Observer Stopped by User")
        except Exception as e:
            logging.error(f"Error occurred: {e}")
            self.observer.stop()

        self.observer.join()

class Handler(FileSystemEventHandler):
    def __init__(self, event_types, post_url, authentication_header, file_extension_pattern):
        self.event_types = event_types
        self.post_url = post_url
        self.authentication_header = authentication_header
        self.file_extension_pattern = file_extension_pattern

    def on_any_event(self, event):
        if event.is_directory or not fnmatch.fnmatch(event.src_path, self.file_extension_pattern):
            return

        if event.event_type in self.event_types:
            try:
                file_path = os.path.abspath(event.src_path)
                payload = {"path": file_path}
                payload_json = json.dumps(payload)
                headers = {
                    'Authorization': self.authentication_header,
                    'Content-Type': 'application/json',
                    'Accept': 'application/json'
                }
                response = requests.post(self.post_url, headers=headers, data=payload_json)

                if response.status_code != 200:
                    logging.error(f"Error in HTTP POST request: {response.status_code} - {response.text}")
                else:
                    logging.info(f"Posted file info successfully: {file_path}")
            except Exception as e:
                logging.error(f"Error occurred while sending POST request: {e}")

if __name__ == '__main__':
    logging.basicConfig(level=logging.INFO, format='%(asctime)s - %(levelname)s - %(message)s', datefmt='%Y-%m-%d %H:%M:%S')

    try:
        with open('config.json', 'r') as config_file:
            config = json.load(config_file)
    except FileNotFoundError:
        logging.error("Configuration file 'config.json' not found.")
        exit(1)
    except json.JSONDecodeError:
        logging.error("Error decoding 'config.json'. Please check its format.")
        exit(1)

    watch_directories = config['FileWatcher']['directories']
    print(f"Directories to Observe: {watch_directories}")
    event_types = config['FileWatcher']['event_types']
    print(f"Event Types: {watch_directories}")
    post_url = config['FileWatcher']['post_url']
    print (post_url)
    authentication_header = config['FileWatcher']['authentication_header']
    print (authentication_header)
    file_extension_pattern = config['FileWatcher'].get('file_extension_pattern', '*.*')  # Default to all files if not specified
    print(file_extension_pattern)
    event_handler = Handler(event_types, post_url, authentication_header, file_extension_pattern)
    w = Watcher(watch_directories, event_handler)
    w.run()
