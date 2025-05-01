package main

import (
	"flag"
	"io"
	"os"
	"os/exec"
	"testing"
	"time"
)

// TestGetSoundFilePath tests the default sound file path resolution
func TestGetSoundFilePath(t *testing.T) {
	// Save the original home directory and restore it after the test
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)

	// Table of test cases
	tests := []struct {
		name     string
		homeDir  string
		expected string
	}{
		{
			name:     "normal case",
			homeDir:  "/Users/testuser",
			expected: "/Users/testuser/code/ding/ding.mp3",
		},
		{
			name:     "empty home",
			homeDir:  "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up the mock home directory
			os.Setenv("HOME", tt.homeDir)

			// Call the function
			result := getSoundFilePath()

			// Check the result
			if result != tt.expected {
				t.Errorf("getSoundFilePath() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// Mock the os.Stat function for testing file existence checks
type mockFileInfo struct{}

func (m mockFileInfo) Name() string       { return "ding.mp3" }
func (m mockFileInfo) Size() int64        { return 1024 }
func (m mockFileInfo) Mode() os.FileMode  { return 0644 }
func (m mockFileInfo) ModTime() time.Time { return time.Now() }
func (m mockFileInfo) IsDir() bool        { return false }
func (m mockFileInfo) Sys() interface{}   { return nil }

// TestPlayWithExternalPlayer tests the sound player selection and execution
func TestPlayWithExternalPlayer(t *testing.T) {
	// We'll use a mock exec.Command function
	// Save the original and restore after test
	originalExecCommand := execCommand
	defer func() { execCommand = originalExecCommand }()

	tests := []struct {
		name           string
		filePath       string
		mockCmdSuccess bool
		expectError    bool
		expectedCmd    string
		expectedArgs   []string
	}{
		{
			name:           "successful play on macOS",
			filePath:       "/path/to/sound.mp3",
			mockCmdSuccess: true,
			expectError:    false,
			expectedCmd:    "afplay",
			expectedArgs:   []string{"/path/to/sound.mp3"},
		},
		{
			name:           "failed play on macOS",
			filePath:       "/path/to/sound.mp3",
			mockCmdSuccess: false,
			expectError:    true,
			expectedCmd:    "afplay",
			expectedArgs:   []string{"/path/to/sound.mp3"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up the mock command
			execCommand = func(cmd string, args ...string) *exec.Cmd {
				if cmd != tt.expectedCmd {
					t.Errorf("execCommand called with cmd = %v, want %v", cmd, tt.expectedCmd)
				}

				for i, arg := range args {
					if i >= len(tt.expectedArgs) || arg != tt.expectedArgs[i] {
						t.Errorf("execCommand called with unexpected args: got %v, want %v", args, tt.expectedArgs)
						break
					}
				}

				// Return a mock command that either succeeds or fails as specified
				if tt.mockCmdSuccess {
					return exec.Command("echo", "success") // Will always succeed
				} else {
					return exec.Command("false") // Will always fail
				}
			}

			// Call the function
			err := playWithExternalPlayer(tt.filePath)

			// Check the result
			if (err != nil) != tt.expectError {
				t.Errorf("playWithExternalPlayer() error = %v, expectError %v", err, tt.expectError)
			}
		})
	}
}

// TestFlagParsing tests the command line flag parsing
func TestFlagParsing(t *testing.T) {
	// Save original os.Args and restore after test
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	tests := []struct {
		name                    string
		args                    []string
		expectedMessage         string
		expectedNoMessage       bool
		expectedUseTerminalBell bool
		expectedCustomSoundFile string
	}{
		{
			name:                    "defaults",
			args:                    []string{"ding"},
			expectedMessage:         defaultMessage,
			expectedNoMessage:       false,
			expectedUseTerminalBell: false,
			expectedCustomSoundFile: "",
		},
		{
			name:                    "custom message",
			args:                    []string{"ding", "-m", "Test completed"},
			expectedMessage:         "Test completed",
			expectedNoMessage:       false,
			expectedUseTerminalBell: false,
			expectedCustomSoundFile: "",
		},
		{
			name:                    "silent mode",
			args:                    []string{"ding", "-s"},
			expectedMessage:         defaultMessage,
			expectedNoMessage:       true,
			expectedUseTerminalBell: false,
			expectedCustomSoundFile: "",
		},
		{
			name:                    "terminal bell",
			args:                    []string{"ding", "-b"},
			expectedMessage:         defaultMessage,
			expectedNoMessage:       false,
			expectedUseTerminalBell: true,
			expectedCustomSoundFile: "",
		},
		{
			name:                    "custom sound file",
			args:                    []string{"ding", "-f", "/custom/sound.mp3"},
			expectedMessage:         defaultMessage,
			expectedNoMessage:       false,
			expectedUseTerminalBell: false,
			expectedCustomSoundFile: "/custom/sound.mp3",
		},
		{
			name:                    "multiple flags",
			args:                    []string{"ding", "-m", "All done", "-f", "/custom/sound.mp3", "-s"},
			expectedMessage:         "All done",
			expectedNoMessage:       true,
			expectedUseTerminalBell: false,
			expectedCustomSoundFile: "/custom/sound.mp3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset flags for each test
			flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

			// Set up the flags
			message := flag.String("m", defaultMessage, "Message to display")
			noMessage := flag.Bool("s", false, "Silent mode (no text output)")
			useTerminalBell := flag.Bool("b", false, "Use terminal bell instead of custom sound")
			customSoundFile := flag.String("f", "", "Path to custom sound file (overrides default)")

			// Parse with the test args
			os.Args = tt.args
			flag.Parse()

			// Check results
			if *message != tt.expectedMessage {
				t.Errorf("message flag = %v, want %v", *message, tt.expectedMessage)
			}
			if *noMessage != tt.expectedNoMessage {
				t.Errorf("noMessage flag = %v, want %v", *noMessage, tt.expectedNoMessage)
			}
			if *useTerminalBell != tt.expectedUseTerminalBell {
				t.Errorf("useTerminalBell flag = %v, want %v", *useTerminalBell, tt.expectedUseTerminalBell)
			}
			if *customSoundFile != tt.expectedCustomSoundFile {
				t.Errorf("customSoundFile flag = %v, want %v", *customSoundFile, tt.expectedCustomSoundFile)
			}
		})
	}
}

// TestMainIntegration to test the core logic
func TestMainIntegration(t *testing.T) {
	// This will be a more integrated test that mocks several components
	// Save originals
	originalExecCommand := execCommand
	originalOsStat := osStat
	originalFmtPrint := fmtPrint
	originalFmtPrintln := fmtPrintln
	originalArgs := os.Args

	defer func() {
		execCommand = originalExecCommand
		osStat = originalOsStat
		fmtPrint = originalFmtPrint
		fmtPrintln = originalFmtPrintln
		os.Args = originalArgs
	}()

	tests := []struct {
		name               string
		args               []string
		fileExists         bool
		shouldPlaySound    bool
		shouldPrintBell    bool
		shouldPrintMessage bool
		expectedMessage    string
	}{
		{
			name:               "default behavior - sound file exists",
			args:               []string{"ding"},
			fileExists:         true,
			shouldPlaySound:    true,
			shouldPrintBell:    false,
			shouldPrintMessage: true,
			expectedMessage:    defaultMessage,
		},
		{
			name:               "sound file doesn't exist - fallback to bell",
			args:               []string{"ding"},
			fileExists:         false,
			shouldPlaySound:    false,
			shouldPrintBell:    true,
			shouldPrintMessage: true,
			expectedMessage:    defaultMessage,
		},
		{
			name:               "terminal bell flag",
			args:               []string{"ding", "-b"},
			fileExists:         true, // Doesn't matter
			shouldPlaySound:    false,
			shouldPrintBell:    true,
			shouldPrintMessage: true,
			expectedMessage:    defaultMessage,
		},
		{
			name:               "silent mode",
			args:               []string{"ding", "-s"},
			fileExists:         true,
			shouldPlaySound:    true,
			shouldPrintBell:    false,
			shouldPrintMessage: false,
			expectedMessage:    "", // Not used in silent mode
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mocks
			playedSound := false
			printedBell := false
			printedMessage := false
			var actualMessage string

			// Mock exec.Command
			execCommand = func(cmd string, args ...string) *exec.Cmd {
				playedSound = true
				return exec.Command("echo", "success") // Always succeeds
			}

			// Mock os.Stat
			osStat = func(name string) (os.FileInfo, error) {
				if tt.fileExists {
					return mockFileInfo{}, nil
				}
				return nil, os.ErrNotExist
			}

			// Mock fmt.Fprint - note the correct io.Writer type
			fmtPrint = func(w io.Writer, a ...any) (n int, err error) {
				if len(a) > 0 && a[0] == bellChar {
					printedBell = true
				}
				return 0, nil
			}

			// Mock fmt.Println
			fmtPrintln = func(a ...any) (n int, err error) {
				printedMessage = true
				if len(a) > 0 {
					actualMessage = a[0].(string)
				}
				return 0, nil
			}

			// Set up and run main with test args
			os.Args = tt.args
			flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
			main()

			// Verify results
			if playedSound != tt.shouldPlaySound {
				t.Errorf("playedSound = %v, want %v", playedSound, tt.shouldPlaySound)
			}
			if printedBell != tt.shouldPrintBell {
				t.Errorf("printedBell = %v, want %v", printedBell, tt.shouldPrintBell)
			}
			if printedMessage != tt.shouldPrintMessage {
				t.Errorf("printedMessage = %v, want %v", printedMessage, tt.shouldPrintMessage)
			}
			if printedMessage && actualMessage != tt.expectedMessage {
				t.Errorf("message = %v, want %v", actualMessage, tt.expectedMessage)
			}
		})
	}
}
