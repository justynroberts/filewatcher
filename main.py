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
    def __init__(self, event_types, post_url, authentication_header, file_extension_pattern):
        self.event_types = event_types
        self.post_url = post_url
        self.authentication_header = authentication_header
        self.file_extension_pattern = file_extension_pattern

    def on_any_event(self, event):
        if event.is_directory or not fnmatch.fnmatch(event.src_path, self.file_extension_pattern):
            return

        if event.event_type in self.event_types:
            event_id = str(uuid.uuid4())  # Generate a unique ID for the event
            logging.info(f"Detected event {event_id} for file: {event.src_path}")
            thread = threading.Thread(target=self.post_event, args=(event, event_id))
            thread.start()

    def post_event(self, event, event_id):
        try:
            file_path = os.path.abspath(event.src_path)
            payload = {"path": file_path, "event_id": event_id}
            payload_json = json.dumps(payload)
            headers = {
                'Authorization': self.authentication_header,
                'Content-Type': 'application/json',
                'Accept': 'application/json'
            }
            logging.info(f"Sending HTTP POST request for event {event_id}: {file_path}")
            response = requests.post(self.post_url, headers=headers, data=payload_json)

            if response.status_code != 200:
                logging.error(f"Error in HTTP POST request for event {event_id}: {response.status_code} - {response.text}")
            else:
                logging.info(f"Successfully posted file info for event {event_id}: {file_path}")
        except Exception as e:
            logging.error(f"Error occurred while sending POST request for event {event_id}: {e}")

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

    logging.info("🟢 File Watcher Config 🟢")
    watch_directories = config['FileWatcher']['directories']
    logging.info(f"🟢 Directories to Observe: {watch_directories}")
    event_types = config['FileWatcher']['event_types']
    file_extension_pattern = config['FileWatcher'].get('file_extension_pattern', '*.*')  # Default to all files if not specified
    logging.info(f"🟢 File Extension Pattern: {file_extension_pattern}")
    post_url = config['FileWatcher']['post_url']
    logging.info(f"🟢 Runbook Automation URL: {post_url}")
    authentication_header = config['FileWatcher']['authentication_header']
    logging.info(f"🟢 Authentication Header: REDACTED (see config.json)")

    event_handler = Handler(event_types, post_url, authentication_header, file_extension_pattern)
    w = Watcher(watch_directories, event_handler)
    w.run()

