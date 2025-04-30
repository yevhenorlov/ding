# Ding - Terminal Notification Tool

A simple CLI tool written in Go that notifies you when your long-running processes complete.

## Features

- Play custom MP3 sound (or terminal bell as fallback)
- Customizable notification message
- Repeat sounds with configurable intervals
- Cross-platform support (Linux, macOS, Windows)
- Silent mode (no text output)
- Flexible configuration via command-line flags or config file
- Portable default sound path using user's home directory

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

### Configuration

Ding can be configured in two ways (in order of precedence):

1. **Command-line flags** (highest priority)
2. **Local config file** (used if no flag is specified)

If neither option is provided, the terminal bell will be used as a fallback.

#### Config File

Ding looks for a configuration file called `.env` in the current directory where you run the command.

The config file uses a simple key=value format:

```
# Configuration for ding
SOUND_FILE=./audio/ding.mp3
```

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

## Configuration Examples

### Creating a Config File

Create a file named `.env` in your project directory:

```
# Ding configuration
SOUND_FILE=./audio/notification.mp3
```

### Sample Config File

```
# .env - Project-specific settings
# Relative paths are resolved relative to the .env file location
SOUND_FILE=./sounds/build-complete.mp3
```

**Note:** If you specify a relative path (starting with `./` or without a leading `/`), it will be resolved relative to the location of the `.env` file, not where you run the command from. This makes it easier to keep your audio files with your project.

## License

MIT License
