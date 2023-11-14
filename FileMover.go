package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/rjeczalik/notify"
)

var debugLevel int

func init() {
	// Initialize the debug level from the DEBUG_LEVEL environment variable
	debugLevel = 0 // Default is 0 (no logs)

	// Check if the DEBUG_LEVEL environment variable is set
	level, exists := os.LookupEnv("DEBUG_LEVEL")
	if exists {
		// Try to parse the environment variable into an integer
		if _, err := fmt.Sscanf(level, "%d", &debugLevel); err != nil {
			log.Println("Invalid DEBUG_LEVEL, defaulting to 0")
		}
	}
}

// logDebug logs a message if the current debug level is high enough
func logDebug(level int, message string) {
	if debugLevel >= level {
		log.Println(message)
	}
}

// moveFiles moves files from the source directory to the destination directory
func moveFiles(source, destination string) {
	//logDebug(2, fmt.Sprintf("src to dst: %v to %v", source, destination))
	files, _ := os.ReadDir(source)
	for _, file := range files {
		// Ignore files or folders with ".syncthing" in their names
		if strings.Contains(file.Name(), ".syncthing") {
			logDebug(2, fmt.Sprintf("syncthingfolder found: %v", file.Name()))
			continue
		}
		destPath := filepath.Join(destination, file.Name())
		if file.IsDir() {
			logDebug(2, fmt.Sprintf("Found folder: %v", file.Name()))
			// Create the directory at the destination
			if err := os.MkdirAll(destPath, 0760); err != nil {
				logDebug(1, fmt.Sprintf("Error creating folder: %v", err))
			} else {
				logDebug(2, fmt.Sprintf("Created folder: %v", destPath))
				moveFiles(filepath.Join(source, file.Name()), destPath)
				filePath := filepath.Join(source, file.Name())
				// Cleanup empty paths
				emptyDir, _ := isDirEmpty(filePath)
				if file != nil && file.IsDir() && emptyDir {
					if err := os.Remove(filePath); err != nil {
						logDebug(1, fmt.Sprintf("Could not remove folder: %v", err))
					} else {
						logDebug(2, fmt.Sprintf("Deleted folder: %v", filePath))
					}
				}
			}
		} else {
			// Move each file from the source to the destination
			err := os.Rename(filepath.Join(source, file.Name()), destPath)
			if err != nil {
				logDebug(1, fmt.Sprintf("Error moving file: %v", err))
			} else {
				logDebug(2, fmt.Sprintf("Moved file: %v", destPath))
			}
		}
	}
}

// isDirEmpty checks if a directory is empty
func isDirEmpty(dirname string) (bool, error) {
	f, err := os.Open(dirname)
	if err != nil {
		return false, err
	}
	defer f.Close()

	_, err = f.Readdirnames(1)
	return err == io.EOF, err
}

// watchFolder sets up a watcher on the source directory
func watchFolder(source, destination string) {
	c := make(chan notify.EventInfo, 1)
	//var timer *time.Timer
	timer := time.NewTimer(10 * time.Second) // Create a timer for 10 seconds
	timer.Stop()                             // Stop the timer initially
	timerActive := false                     // Indicates whether the timer is active

	// Move all existing files/folders at launch
	moveFiles(source, destination)

	// Start watching the source folder for changes
	if err := notify.Watch(filepath.Join(source, "..."), c, notify.Create, notify.Write); err != nil {
		logDebug(1, fmt.Sprintf("Error setting up watcher: %v", err))
		return
	}
	defer notify.Stop(c)
	logDebug(1, fmt.Sprintf("Started watching: %s", source))

	for {
		select {
		case ei := <-c:
			logDebug(2, fmt.Sprintf("Got event: %v", ei))
			if !timerActive {
				timer = time.NewTimer(5 * time.Second)
				timerActive = true
			} else {
				// Reset the timer if it's already active
				if !timer.Stop() {
					<-timer.C
				}
				timer.Reset(10 * time.Second)
			}
		case <-timer.C:
			// Timer has elapsed, no events in the last 10 seconds
			logDebug(2, "Idle time exceeded. Moving files")
			moveFiles(source, destination)
			timerActive = false // Reset timerActive when the timer triggers
		}
	}
}

func main() {
	// Configure log format
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// Check if the config file exists
	configFile := "folder_pairs.conf"
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		// File does not exist, create it
		file, err := os.Create(configFile)
		if err != nil {
			fmt.Println("Error creating file:", err)
			return
		}
		file.Close()
		log.Println("Configuraton file created:", configFile)
		log.Fatal("Please add your configuration. Exiting...")
	} else if err != nil {
		// An error occurred while checking the file status
		log.Fatal("Error:", err)
	}

	// Open the configuration file
	file, err := os.Open(configFile)
	if err != nil {
		logDebug(1, fmt.Sprintf("Error opening config file: %v", err))
		return
	}
	defer file.Close()

	// Read folder pairs from the configuration file
	scanner := bufio.NewScanner(file)
	var validPair = false
	for scanner.Scan() {
		var source, destination string
		line := scanner.Text()
		if len(strings.TrimSpace(line)) == 0 || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "//") {
			continue
		} else if strings.Contains(line, ":") {
			validPair = true
			pair := strings.Split(line, ":")
			source, destination = pair[0], pair[1]
			// Start a goroutine to watch each folder pair
			go watchFolder(source, destination)
		}
	}

	// Prevent the program from exiting
	if validPair {
		select {}
	} else {
		log.Fatal("No folder pairs given in configuration...")
	}

}
