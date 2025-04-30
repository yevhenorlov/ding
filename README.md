# Ding - Terminal Notification Tool

A simple CLI tool written in Go that notifies you when your long-running processes complete.

## Features

- Play custom MP3 sound (or terminal bell as fallback)
- Customizable notification message
- Repeat sounds with configurable intervals
- Cross-platform support (Linux, macOS, Windows)
- Silent mode (no text output)

## Installation

### Prerequisites

To build from source:

- Go 1.16 or higher

For MP3 sound support:

- **Linux**: mpg123, mpg321, mplayer, or ffplay
- **macOS**: Already includes afplay
- **Windows**: No additional requirements (uses PowerShell)

### Building

```bash
# Clone the repository
git clone https://github.com/yourusername/ding.git
cd ding

# Build the application
go build -o ding
```

### Default Sound File Location

By default, the application looks for a sound file at:

```
~/code/ding/audio/ding.mp3
```

You can customize this default path by modifying the `getSoundFilePath()` function in `ding.go`. Alternatively, you can specify a different sound file with the `-f` flag when running the command.

## Usage

Basic usage:

```bash
# Notify when a long-running command completes (with default sound)
./long-running-process; ./ding
```

### Command-line options

```
-m string    Message to display (default "Process completed")
-r int       Number of times to repeat the sound (default 1)
-i duration  Interval between sounds (default 500ms)
-s           Silent mode (no text output)
-b           Use terminal bell instead of custom sound
-f string    Path to custom sound file (overrides default)
```

### Examples

```bash
# Custom message
./time-consuming-task; ./ding -m "Backup completed!"

# Play sound 3 times with 1 second interval
./build-project; ./ding -r 3 -i 1s

# Use terminal bell instead of custom sound
./long-calculation; ./ding -b

# Just sound, no message
./lengthy-process; ./ding -s

# Use a specific sound file for this notification
./data-import; ./ding -f /path/to/success.mp3

# Combine options
./database-backup; ./ding -m "Backup finished!" -r 2 -i 750ms -f ~/sounds/tada.mp3
```

## How it works

- By default: Uses the sound file located at `~/code/ding/audio/ding.mp3`
- With `-f` flag: Uses the custom sound file specified in the command line
- With `-b` flag: Uses the terminal bell character `\a` instead
- Sound player is automatically selected based on your OS:
  - Linux: mpg123, mpg321, mplayer, or ffplay (whichever is available)
  - macOS: afplay (built-in)
  - Windows: PowerShell Media.SoundPlayer
- If the sound file can't be found or played, falls back to terminal bell

## Customizing the Default Path

If you want to modify the default sound file path for all users, you can edit the `getSoundFilePath()` function:

```go
func getSoundFilePath() string {
    homeDir, err := os.UserHomeDir()
    if err != nil {
        return ""
    }

    // Modify this line to change the default location
    return filepath.Join(homeDir, "code", "ding", "ding.mp3")
}
```

## License

MIT License
