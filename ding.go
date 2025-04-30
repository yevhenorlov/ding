package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const (
	// ANSI escape code for bell
	bellChar = "\a"
	// Default message
	defaultMessage = "Process completed"
	// Config file name
	configFileName = ".env"
)

// loadConfigFile loads configuration from .env files in multiple locations
func loadConfigFile() map[string]string {
	config := make(map[string]string)

	// Check multiple locations in order of priority
	// 1. Current directory
	// 2. Binary directory
	// 3. Home directory

	configPaths := []string{
		filepath.Join(".", configFileName),
	}

	// Add executable directory if we can determine it
	execPath, err := os.Executable()
	if err == nil {
		execDir := filepath.Dir(execPath)
		configPaths = append(configPaths, filepath.Join(execDir, configFileName))
	}

	// Add home directory if we can determine it
	homeDir, err := os.UserHomeDir()
	if err == nil {
		configPaths = append(configPaths, filepath.Join(homeDir, configFileName))
	}

	// Try each location in priority order
	for _, configPath := range configPaths {
		file, err := os.Open(configPath)
		if err != nil {
			continue // Try next location if file doesn't exist
		}
		defer file.Close()

		// Read file line by line
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()
			line = strings.TrimSpace(line)

			// Skip comments and empty lines
			if strings.HasPrefix(line, "#") || line == "" {
				continue
			}

			// Parse key=value pairs
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])

				// Remove quotes if present
				value = strings.Trim(value, "\"'")

				// Convert to absolute path if it's a file path and not already absolute
				if key == "SOUND_FILE" && !filepath.IsAbs(value) {
					// Make path absolute relative to the config file's location
					configDir := filepath.Dir(configPath)
					value = filepath.Join(configDir, value)
				}

				config[key] = value
			}
		}

		// If we found and successfully read a config file, return the config
		return config
	}

	// No config found in any location
	return config
}

func getSoundFilePath() string {
	config := loadConfigFile()
	if soundFile, exists := config["SOUND_FILE"]; exists && soundFile != "" {
		return soundFile
	}

	return ""
}

func playBellSound(repeat int, interval time.Duration) {
	for i := 0; i < repeat; i++ {
		// Print the bell character to trigger the terminal bell
		fmt.Fprint(os.Stdout, bellChar)

		// If it's not the last iteration and repeat > 1, wait for the specified interval
		if i < repeat-1 {
			time.Sleep(interval)
		}
	}
}

func playWithExternalPlayer(filePath string, repeat int, interval time.Duration) error {
	var cmd *exec.Cmd
	var playerName string

	switch runtime.GOOS {
	case "darwin":
		playerName = "afplay"
		for i := 0; i < repeat; i++ {
			cmd = exec.Command(playerName, filePath)
			err := cmd.Run()
			if err != nil {
				return err
			}
			if i < repeat-1 {
				time.Sleep(interval)
			}
		}
	case "linux":
		// Try several players in order of preference
		players := []string{"mpg123", "mpg321", "mplayer", "ffplay"}
		found := false

		for _, p := range players {
			if _, err := exec.LookPath(p); err == nil {
				playerName = p
				found = true
				break
			}
		}

		if !found {
			return fmt.Errorf("no suitable MP3 player found, falling back to terminal bell")
		}

		for i := 0; i < repeat; i++ {
			var args []string
			switch playerName {
			case "mpg123", "mpg321":
				args = []string{"-q", filePath} // quiet mode
			case "mplayer":
				args = []string{"-really-quiet", filePath}
			case "ffplay":
				args = []string{"-nodisp", "-autoexit", "-loglevel", "quiet", filePath}
			}

			cmd = exec.Command(playerName, args...)
			err := cmd.Run()
			if err != nil {
				return err
			}
			if i < repeat-1 {
				time.Sleep(interval)
			}
		}
	case "windows":
		for i := 0; i < repeat; i++ {
			cmd = exec.Command("powershell", "-c", "(New-Object Media.SoundPlayer '"+filePath+"').PlaySync();")
			err := cmd.Run()
			if err != nil {
				return err
			}
			if i < repeat-1 {
				time.Sleep(interval)
			}
		}
	default:
		return fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}

	return nil
}

func main() {
	// Define command line flags
	message := flag.String("m", defaultMessage, "Message to display")
	repeat := flag.Int("r", 1, "Number of times to repeat the sound")
	interval := flag.Duration("i", 500*time.Millisecond, "Interval between sounds")
	noMessage := flag.Bool("s", false, "Silent mode (no text output)")
	useTerminalBell := flag.Bool("b", false, "Use terminal bell instead of custom sound")
	customSoundFile := flag.String("f", "", "Path to custom sound file (overrides default)")

	flag.Parse()

	// Determine which sound file to use
	soundFilePath := ""
	if *customSoundFile != "" {
		// Use the sound file specified via command line
		soundFilePath = *customSoundFile
	} else if !*useTerminalBell {
		// Use the sound file from config
		soundFilePath = getSoundFilePath()
	}

	// Play sound based on the provided options
	if soundFilePath != "" {
		// Try to use the sound file
		absPath, err := filepath.Abs(soundFilePath)
		if err != nil {
			log.Printf("Error with sound file path: %v. Falling back to terminal bell.", err)
			playBellSound(*repeat, *interval)
		} else {
			// Check if the file exists
			if _, err := os.Stat(absPath); os.IsNotExist(err) {
				log.Printf("Sound file not found: %s. Falling back to terminal bell.", absPath)
				playBellSound(*repeat, *interval)
			} else {
				err := playWithExternalPlayer(absPath, *repeat, *interval)
				if err != nil {
					log.Printf("Error playing sound: %v. Falling back to terminal bell.", err)
					playBellSound(*repeat, *interval)
				}
			}
		}
	} else {
		// Use the terminal bell
		playBellSound(*repeat, *interval)
	}

	// Print the message unless silent mode is enabled
	if !*noMessage {
		fmt.Println(*message)
	}
}
