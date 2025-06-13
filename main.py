import time
import os
import logging
import json
import fnmatch
import requests
import threading
import uuid
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
    def __init__(self, event_types, post_url, authentication_header, file_extension_pattern,
                 file_stability_check_enabled=False, file_stability_wait_seconds=5):
        self.event_types = event_types
        self.post_url = post_url
        self.authentication_header = authentication_header
        self.file_extension_pattern = file_extension_pattern
        self.last_event_time = {}
        self.debounce_seconds = 1.0  # Adjust as needed
        self.file_stability_check_enabled = file_stability_check_enabled
        self.file_stability_wait_seconds = file_stability_wait_seconds
        self.file_size_trackers = {}  # Track file sizes for stability check

    def on_any_event(self, event):
        if event.is_directory or not fnmatch.fnmatch(event.src_path, self.file_extension_pattern):
            return

        current_time = time.time()
        last_time = self.last_event_time.get(event.src_path, 0)
        if current_time - last_time < self.debounce_seconds:
            return  # Skip this event because it's too soon since the last one

        if event.event_type in self.event_types:
            self.last_event_time[event.src_path] = current_time
            event_id = str(uuid.uuid4())
            logging.info(f"🚥 Detected event {event_id} ({event.event_type}) for file: {event.src_path}")
            thread = threading.Thread(target=self.post_event, args=(event, event_id))
            thread.start()

    def check_file_stability(self, file_path, event_id):
        """Check if file size remains stable for the configured duration"""
        try:
            stable_count = 0
            last_size = None
            
            while stable_count < self.file_stability_wait_seconds:
                if not os.path.exists(file_path):
                    logging.warning(f"File {file_path} no longer exists during stability check for event {event_id}")
                    return False
                
                current_size = os.path.getsize(file_path)
                
                if last_size is None:
                    last_size = current_size
                elif current_size == last_size:
                    stable_count += 1
                else:
                    # File size changed, reset counter
                    logging.info(f"File size changed for {file_path} (event {event_id}): {last_size} -> {current_size}, resetting stability timer")
                    stable_count = 0
                    last_size = current_size
                
                time.sleep(1)  # Check every second
            
            logging.info(f"File {file_path} is stable for {self.file_stability_wait_seconds} seconds (event {event_id})")
            return True
            
        except Exception as e:
            logging.error(f"Error during file stability check for event {event_id}: {e}")
            return False

    def post_event(self, event, event_id):
        try:
            file_path = os.path.abspath(event.src_path)
            
            # Perform file stability check if enabled
            if self.file_stability_check_enabled:
                logging.info(f"🕐 Starting file stability check for event {event_id}: {file_path}")
                if not self.check_file_stability(file_path, event_id):
                    logging.warning(f"⚠️ File stability check failed for event {event_id}: {file_path}")
                    return
            
            directory, file_name = os.path.split(file_path)  # Split the path into directory and filename
            payload = {
                "filepath": directory,
                "filename": file_name,
                "event_id": event_id
            }
            payload_json = json.dumps(payload)
            # Expand environment variables in authentication header
            auth_header = os.path.expandvars(self.authentication_header)
            
            headers = {
                'Authorization': auth_header,
                'Content-Type': 'application/json',
                'Accept': 'application/json'
            }
            logging.info(f"↔️ Sending HTTP POST request for event {event_id}: {file_path}")
            response = requests.post(self.post_url, headers=headers, data=payload_json)

            if response.status_code != 200:
                logging.error(f"🚩 Error in HTTP POST request for event {event_id}: {response.status_code} - {response.text}")
            else:
                logging.info(f"✅ Successfully posted file info for event {event_id}: {directory}, {file_name}")
        except Exception as e:
            logging.error(f"🚩 Error occurred while sending POST request for event {event_id}: {e}")

if __name__ == '__main__':
    logging.basicConfig(level=logging.INFO, format='%(asctime)s - %(levelname)s - %(message)s', datefmt='%Y-%m-%d %H:%M:%S')

    try:
        with open('config.json', 'r') as config_file:
            config = json.load(config_file)
    except FileNotFoundError:
        logging.error("Configuration file 'config.json' not found.")
        exit(1)
    except json.JSONDecodeError:
        logging.error("🚩 Error decoding 'config.json'. Please check its format.")
        exit(1)

    logging.info("🟢 File Watcher Config 🟢")
    watch_directories = config['FileWatcher']['directories']
    logging.info(f"🟢 Directories to Observe: {watch_directories}")
    event_types = config['FileWatcher']['event_types']
    file_extension_pattern = config['FileWatcher'].get('file_extension_pattern', '*.*')  # Default to all files if not specified
    logging.info(f"🟢 File Extension Pattern: {file_extension_pattern}")
    file_stability_check_enabled = config['FileWatcher'].get('file_stability_check_enabled', False)
    file_stability_wait_seconds = config['FileWatcher'].get('file_stability_wait_seconds', 5)
    logging.info(f"🟢 File Stability Check: {'Enabled' if file_stability_check_enabled else 'Disabled'}")
    if file_stability_check_enabled:
        logging.info(f"🟢 File Stability Wait Time: {file_stability_wait_seconds} seconds")
    post_url = config['FileWatcher']['post_url']
    logging.info(f"🟢 Runbook Automation URL: {post_url}")
    authentication_header = config['FileWatcher']['authentication_header']
    logging.info(f"🟢 Authentication Header: REDACTED (see config.json)")

    event_handler = Handler(event_types, post_url, authentication_header, file_extension_pattern,
                          file_stability_check_enabled, file_stability_wait_seconds)
    w = Watcher(watch_directories, event_handler)
    w.run()