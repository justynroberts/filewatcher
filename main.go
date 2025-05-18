package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/google/uuid"
)

// Config represents the application configuration
type Config struct {
	FileWatcher struct {
		Directories        []string `json:"directories"`
		EventTypes         []string `json:"event_types"`
		FileExtensionPattern string `json:"file_extension_pattern"`
		PostURL           string   `json:"post_url"`
		AuthenticationHeader string `json:"authentication_header"`
	} `json:"FileWatcher"`
}

// EventPayload represents the payload sent to the webhook
type EventPayload struct {
	Filepath string `json:"filepath"`
	Filename string `json:"filename"`
	EventID  string `json:"event_id"`
}

// Watcher handles file system monitoring
type Watcher struct {
	watcher       *fsnotify.Watcher
	config        *Config
	lastEventTime map[string]time.Time
	mutex         sync.Mutex
	debounceTime  time.Duration
}

// NewWatcher creates a new file watcher
func NewWatcher(config *Config) (*Watcher, error) {
	fsWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	return &Watcher{
		watcher:       fsWatcher,
		config:        config,
		lastEventTime: make(map[string]time.Time),
		mutex:         sync.Mutex{},
		debounceTime:  1 * time.Second,
	}, nil
}

// matchesPattern checks if a file matches the configured pattern
func (w *Watcher) matchesPattern(path string) bool {
	pattern := w.config.FileWatcher.FileExtensionPattern
	if pattern == "" {
		pattern = "*.*" // Default to all files
	}
	
	matched, err := filepath.Match(pattern, filepath.Base(path))
	if err != nil {
		log.Printf("Error matching pattern: %v", err)
		return false
	}
	return matched
}

// isEventTypeWatched checks if an event type is in the configured list
func (w *Watcher) isEventTypeWatched(eventType string) bool {
	for _, t := range w.config.FileWatcher.EventTypes {
		if t == eventType {
			return true
		}
	}
	return false
}

// shouldProcessEvent determines if an event should be processed based on debounce time
func (w *Watcher) shouldProcessEvent(path string) bool {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	
	now := time.Now()
	lastTime, exists := w.lastEventTime[path]
	
	if !exists || now.Sub(lastTime) >= w.debounceTime {
		w.lastEventTime[path] = now
		return true
	}
	
	return false
}

// postEvent sends the event information to the configured webhook
func (w *Watcher) postEvent(path string, eventID string) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		log.Printf("🚩 Error getting absolute path: %v", err)
		return
	}
	
	dir, file := filepath.Split(absPath)
	
	payload := EventPayload{
		Filepath: dir,
		Filename: file,
		EventID:  eventID,
	}
	
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		log.Printf("🚩 Error marshaling JSON: %v", err)
		return
	}
	
	req, err := http.NewRequest("POST", w.config.FileWatcher.PostURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		log.Printf("🚩 Error creating request: %v", err)
		return
	}
	
	req.Header.Set("Authorization", w.config.FileWatcher.AuthenticationHeader)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	
	log.Printf("↔️ Sending HTTP POST request for event %s: %s", eventID, absPath)
	
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("🚩 Error sending POST request for event %s: %v", eventID, err)
		return
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("🚩 Error in HTTP POST request for event %s: %d - %s", eventID, resp.StatusCode, string(body))
		return
	}
	
	log.Printf("✅ Successfully posted file info for event %s: %s, %s", eventID, dir, file)
}

// processEvent handles a file system event
func (w *Watcher) processEvent(event fsnotify.Event) {
	// Skip directories and non-matching files
	fileInfo, err := os.Stat(event.Name)
	if err != nil {
		// File might have been deleted
		if !os.IsNotExist(err) {
			log.Printf("Error getting file info: %v", err)
		}
		return
	}
	
	if fileInfo.IsDir() || !w.matchesPattern(event.Name) {
		return
	}
	
	// Convert fsnotify event to our event type
	var eventType string
	switch {
	case event.Op&fsnotify.Create == fsnotify.Create:
		eventType = "created"
	case event.Op&fsnotify.Write == fsnotify.Write:
		eventType = "modified"
	case event.Op&fsnotify.Remove == fsnotify.Remove:
		eventType = "deleted"
	case event.Op&fsnotify.Rename == fsnotify.Rename:
		eventType = "moved"
	default:
		return
	}
	
	// Check if we should process this event type
	if !w.isEventTypeWatched(eventType) {
		return
	}
	
	// Apply debouncing
	if !w.shouldProcessEvent(event.Name) {
		return
	}
	
	// Generate event ID
	eventID := uuid.New().String()
	log.Printf("🚥 Detected event %s (%s) for file: %s", eventID, eventType, event.Name)
	
	// Process event in a goroutine
	go w.postEvent(event.Name, eventID)
}

// addDirectoryRecursively adds a directory and all its subdirectories to the watcher
func (w *Watcher) addDirectoryRecursively(path string) error {
	return filepath.Walk(path, func(walkPath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		if info.IsDir() {
			if err := w.watcher.Add(walkPath); err != nil {
				log.Printf("Error watching directory %s: %v", walkPath, err)
				return err
			}
			log.Printf("Watching directory: %s", walkPath)
		}
		
		return nil
	})
}

// Start begins watching the configured directories
func (w *Watcher) Start() error {
	// Add all directories to the watcher
	for _, dir := range w.config.FileWatcher.Directories {
		expandedDir := os.ExpandEnv(dir) // Expand environment variables in path
		
		// Check if directory exists
		if _, err := os.Stat(expandedDir); os.IsNotExist(err) {
			log.Printf("Warning: Directory does not exist: %s", expandedDir)
			continue
		}
		
		log.Printf("Adding directory to watch: %s", expandedDir)
		if err := w.addDirectoryRecursively(expandedDir); err != nil {
			return fmt.Errorf("error adding directory %s: %w", expandedDir, err)
		}
	}
	
	// Process events
	go func() {
		for {
			select {
			case event, ok := <-w.watcher.Events:
				if !ok {
					return
				}
				w.processEvent(event)
				
			case err, ok := <-w.watcher.Errors:
				if !ok {
					return
				}
				log.Printf("Error: %v", err)
			}
		}
	}()
	
	return nil
}

// Stop closes the watcher
func (w *Watcher) Stop() {
	w.watcher.Close()
}

// loadConfig loads the configuration from a file
func loadConfig(configPath string) (*Config, error) {
	file, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	
	var config Config
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		return nil, err
	}
	
	return &config, nil
}

func main() {
	// Parse command line flags
	configPath := flag.String("config", "config.json", "Path to configuration file")
	flag.Parse()
	
	// Configure logging
	log.SetFlags(log.Ldate | log.Ltime | log.Lmsgprefix)
	log.SetPrefix("INFO - ")
	
	// Load configuration
	log.Printf("Loading configuration from %s", *configPath)
	config, err := loadConfig(*configPath)
	if err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}
	
	// Log configuration
	log.Println("🟢 File Watcher Config 🟢")
	log.Printf("🟢 Directories to Observe: %v", config.FileWatcher.Directories)
	log.Printf("🟢 Event Types: %v", config.FileWatcher.EventTypes)
	pattern := config.FileWatcher.FileExtensionPattern
	if pattern == "" {
		pattern = "*.*"
	}
	log.Printf("🟢 File Extension Pattern: %s", pattern)
	log.Printf("🟢 Runbook Automation URL: %s", config.FileWatcher.PostURL)
	log.Printf("🟢 Authentication Header: REDACTED (see config.json)")
	
	// Create and start watcher
	watcher, err := NewWatcher(config)
	if err != nil {
		log.Fatalf("Error creating watcher: %v", err)
	}
	defer watcher.Stop()
	
	if err := watcher.Start(); err != nil {
		log.Fatalf("Error starting watcher: %v", err)
	}
	
	// Wait for interrupt signal
	log.Println("Watcher started. Press Ctrl+C to stop.")
	select {}
}