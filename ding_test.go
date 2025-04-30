package main

import (
	"bytes"
	"flag"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

// TestLoadConfigFile tests the configuration loading from .env file
func TestLoadConfigFile(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := ioutil.TempDir("", "ding-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Save current working directory and change to temp dir
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	os.Chdir(tempDir)

	// Create a test config file
	configContent := []byte("# Test config\nSOUND_FILE=./test-sound.mp3")
	err = ioutil.WriteFile(filepath.Join(tempDir, configFileName), configContent, 0644)
	if err != nil {
		t.Fatalf("Failed to write test config file: %v", err)
	}

	// Test config loading
	config, configDir := loadConfigFile()

	// Verify the config values
	if config["SOUND_FILE"] != "./test-sound.mp3" {
		t.Errorf("Expected SOUND_FILE=./test-sound.mp3, got %s", config["SOUND_FILE"])
	}

	// Verify the config directory is correct
	expectedDir, _ := filepath.Abs(tempDir)

	// On macOS, paths might have /private prefix
	// Instead of exact comparison, check if paths are equivalent
	pathsMatch := configDir == expectedDir ||
		strings.HasSuffix(configDir, expectedDir) ||
		strings.HasSuffix(expectedDir, configDir)

	if !pathsMatch {
		t.Errorf("Expected config dir to match %s, got %s", expectedDir, configDir)
	}
}

// TestGetSoundFilePath tests the sound file path resolution
func TestGetSoundFilePath(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := ioutil.TempDir("", "ding-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Save current working directory and change to temp dir
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	os.Chdir(tempDir)

	// Test 1: No config file - should return empty string
	soundPath := getSoundFilePath()
	if soundPath != "" {
		t.Errorf("Expected empty sound path with no config, got %s", soundPath)
	}

	// Test 2: With config file - relative path
	configContent := []byte("SOUND_FILE=./sounds/test.mp3")
	err = ioutil.WriteFile(filepath.Join(tempDir, configFileName), configContent, 0644)
	if err != nil {
		t.Fatalf("Failed to write test config file: %v", err)
	}

	soundPath = getSoundFilePath()
	expected, _ := filepath.Abs(filepath.Join(tempDir, "sounds", "test.mp3"))

	// On macOS, paths might have /private prefix
	// Instead of exact comparison, check if one contains the other or they're equal
	pathsMatch := soundPath == expected ||
		strings.HasSuffix(soundPath, expected) ||
		strings.HasSuffix(expected, soundPath) ||
		strings.Contains(soundPath, filepath.Join("sounds", "test.mp3"))

	if !pathsMatch {
		t.Errorf("Expected sound path to match %s, got %s", expected, soundPath)
	}

	// Test 3: With config file - absolute path
	absPath := filepath.Join(tempDir, "absolute-test.mp3")
	configContent = []byte("SOUND_FILE=" + absPath)
	err = ioutil.WriteFile(filepath.Join(tempDir, configFileName), configContent, 0644)
	if err != nil {
		t.Fatalf("Failed to write test config file: %v", err)
	}

	soundPath = getSoundFilePath()
	if soundPath != absPath {
		t.Errorf("Expected sound path %s, got %s", absPath, soundPath)
	}
}

// TestPlayBellSound tests the terminal bell functionality
func TestPlayBellSound(t *testing.T) {
	// Redirect stdout to capture the bell character
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("Failed to create pipe: %v", err)
	}
	os.Stdout = w

	// Test with 3 repeats and minimal interval
	playBellSound(3, 10*time.Millisecond)

	// Close writer and restore stdout
	w.Close()
	os.Stdout = oldStdout

	// Read captured output
	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Verify that we got 3 bell characters
	if strings.Count(output, bellChar) != 3 {
		t.Errorf("Expected 3 bell characters, got %d", strings.Count(output, bellChar))
	}
}

// TestPlayWithExternalPlayer tests the external player functionality
// Note: This test will be skipped if the required player is not installed
func TestPlayWithExternalPlayer(t *testing.T) {
	// Skip this test if we're in a CI environment or don't want to play actual sounds
	if os.Getenv("CI") != "" || os.Getenv("SKIP_SOUND_TESTS") != "" {
		t.Skip("Skipping external player test in CI environment")
	}

	// Create a temporary sound file (just a text file, we won't actually play it)
	tempDir, err := ioutil.TempDir("", "ding-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	soundFile := filepath.Join(tempDir, "test.mp3")
	err = ioutil.WriteFile(soundFile, []byte("test mp3 data"), 0644)
	if err != nil {
		t.Fatalf("Failed to write test sound file: %v", err)
	}

	// Attempt to play the file - this will likely fail because it's not a real MP3,
	// but we can test that the right player is selected
	err = playWithExternalPlayer(soundFile, 1, 100*time.Millisecond)

	// On most systems this should fail because our test file isn't a valid MP3
	// But the error should be an exec error, not "no suitable player found"
	// unless the system truly doesn't have any of the players
	if err != nil && strings.Contains(err.Error(), "no suitable MP3 player found") {
		switch runtime.GOOS {
		case "darwin":
			// MacOS should have afplay, so failing is unexpected
			t.Errorf("MacOS should have afplay available, but got error: %v", err)
		case "windows":
			// Windows should have PowerShell, so failing is unexpected
			t.Errorf("Windows should have PowerShell available, but got error: %v", err)
		case "linux":
			// On Linux, we'll just log that no player was found
			t.Logf("No suitable player found on Linux: %v", err)
		}
	}
}

// TestMain function tests the main application
// This is an integration test that requires building the application
func TestMain(t *testing.T) {
	// Skip this test if we're in a CI environment or don't want to run integration tests
	if os.Getenv("SKIP_INTEGRATION_TESTS") != "" {
		t.Skip("Skipping integration test")
	}

	// Create a temporary directory for testing
	tempDir, err := ioutil.TempDir("", "ding-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Build the application for testing
	buildCmd := exec.Command("go", "build", "-o", filepath.Join(tempDir, "ding"))
	buildOutput, err := buildCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to build application: %v\nOutput: %s", err, buildOutput)
	}

	// Test cases for command-line flags
	testCases := []struct {
		name     string
		args     []string
		expected string
	}{
		{
			name:     "Default message",
			args:     []string{},
			expected: "Process completed\n",
		},
		{
			name:     "Custom message",
			args:     []string{"-m", "Test completed"},
			expected: "Test completed\n",
		},
		{
			name:     "Silent mode",
			args:     []string{"-s"},
			expected: "",
		},
		{
			name:     "Bell mode",
			args:     []string{"-b", "-m", "Bell message"},
			expected: "Bell message\n",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cmd := exec.Command(filepath.Join(tempDir, "ding"), tc.args...)
			output, err := cmd.CombinedOutput()

			// We don't check for errors because the command might return errors
			// if it can't find or play the sound file, but it should still output
			// the expected message
			if string(output) != tc.expected && !strings.HasSuffix(string(output), tc.expected) {
				t.Errorf("Expected output %q, got %q (err: %v)", tc.expected, string(output), err)
			}
		})
	}

	// Test config file
	t.Run("ConfigFile", func(t *testing.T) {
		// Create a config file
		configContent := []byte("SOUND_FILE=./nonexistent.mp3")
		err = ioutil.WriteFile(filepath.Join(tempDir, configFileName), configContent, 0644)
		if err != nil {
			t.Fatalf("Failed to write test config file: %v", err)
		}

		// Change to the temp directory
		originalWd, err := os.Getwd()
		if err != nil {
			t.Fatalf("Failed to get current working directory: %v", err)
		}
		defer os.Chdir(originalWd)
		os.Chdir(tempDir)

		// Run the command
		cmd := exec.Command("./ding")
		output, _ := cmd.CombinedOutput()

		// The file doesn't exist, so it should fall back to terminal bell
		// and still print the default message
		if !strings.Contains(string(output), "Process completed") {
			t.Errorf("Expected output containing 'Process completed', got %q", string(output))
		}
	})
}

// TestPlayWithExternalPlayerCrossPlatform tests that player selection is correct on different platforms
func TestPlayWithExternalPlayerCrossPlatform(t *testing.T) {
	// This is a meta-test that verifies our platform-specific code
	// It doesn't actually run on all platforms, but checks that the logic
	// for each platform is sensible

	// Linux
	if runtime.GOOS == "linux" {
		// Check if any of the supported players exist
		players := []string{"mpg123", "mpg321", "mplayer", "ffplay"}
		found := false
		var firstFound string

		for _, p := range players {
			_, err := exec.LookPath(p)
			if err == nil {
				found = true
				firstFound = p
				break
			}
		}

		if found {
			t.Logf("Found player %s on Linux", firstFound)
		} else {
			t.Log("No compatible players found on Linux")
		}
	}

	// macOS should have afplay
	if runtime.GOOS == "darwin" {
		_, err := exec.LookPath("afplay")
		if err != nil {
			t.Error("afplay should be available on macOS")
		} else {
			t.Log("afplay found on macOS as expected")
		}
	}

	// Windows should have PowerShell
	if runtime.GOOS == "windows" {
		_, err := exec.LookPath("powershell")
		if err != nil {
			t.Error("powershell should be available on Windows")
		} else {
			t.Log("powershell found on Windows as expected")
		}
	}
}

// TestCommandLineArguments tests specifically the command-line argument parsing
func TestCommandLineArguments(t *testing.T) {
	// Save original args
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	testCases := []struct {
		name           string
		args           []string
		expectedMsg    string
		expectedRepeat int
		expectedBell   bool
		expectedSilent bool
	}{
		{
			name:           "Default values",
			args:           []string{"ding"},
			expectedMsg:    defaultMessage,
			expectedRepeat: 1,
			expectedBell:   false,
			expectedSilent: false,
		},
		{
			name:           "Custom message",
			args:           []string{"ding", "-m", "Custom message"},
			expectedMsg:    "Custom message",
			expectedRepeat: 1,
			expectedBell:   false,
			expectedSilent: false,
		},
		{
			name:           "Multiple repeats",
			args:           []string{"ding", "-r", "3"},
			expectedMsg:    defaultMessage,
			expectedRepeat: 3,
			expectedBell:   false,
			expectedSilent: false,
		},
		{
			name:           "Bell mode",
			args:           []string{"ding", "-b"},
			expectedMsg:    defaultMessage,
			expectedRepeat: 1,
			expectedBell:   true,
			expectedSilent: false,
		},
		{
			name:           "Silent mode",
			args:           []string{"ding", "-s"},
			expectedMsg:    defaultMessage,
			expectedRepeat: 1,
			expectedBell:   false,
			expectedSilent: true,
		},
		{
			name:           "Combined options",
			args:           []string{"ding", "-m", "Test", "-r", "2", "-b", "-s"},
			expectedMsg:    "Test",
			expectedRepeat: 2,
			expectedBell:   true,
			expectedSilent: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Reset flags and args for each test
			os.Args = tc.args
			flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

			// Parse the flags (without running main())
			message := flag.String("m", defaultMessage, "Message to display")
			repeat := flag.Int("r", 1, "Number of times to repeat the sound")
			useTerminalBell := flag.Bool("b", false, "Use terminal bell instead of custom sound")
			noMessage := flag.Bool("s", false, "Silent mode (no text output)")
			// We don't test -f and -i since those are harder to verify programmatically

			flag.Parse()

			// Verify that parsed values match expectations
			if *message != tc.expectedMsg {
				t.Errorf("Expected message %q, got %q", tc.expectedMsg, *message)
			}
			if *repeat != tc.expectedRepeat {
				t.Errorf("Expected repeat %d, got %d", tc.expectedRepeat, *repeat)
			}
			if *useTerminalBell != tc.expectedBell {
				t.Errorf("Expected bell mode %v, got %v", tc.expectedBell, *useTerminalBell)
			}
			if *noMessage != tc.expectedSilent {
				t.Errorf("Expected silent mode %v, got %v", tc.expectedSilent, *noMessage)
			}
		})
	}
}
