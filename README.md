# Ding - Terminal Notification Tool

A simple CLI tool written in Go that notifies you when your long-running processes complete.

## Features

- Play custom MP3 sound (or terminal bell as fallback)
- Customizable notification message
- Cross-platform support (Linux, macOS, Windows)
- Silent mode (no text output)

## Installation

```bash
# Clone the repository
git clone https://github.com/yevhenorlov/ding.git
cd ding

# Build the application
go build -o ding
```

### Prerequisites

- **Build**: Go 1.24+
- **Sound support**:
  - **Linux**: mpg123, mpg321, mplayer, or ffplay
  - **macOS**: No additional requirements (uses built-in afplay)
  - **Windows**: No additional requirements (uses PowerShell)

## Usage

Basic usage:

```bash
# Notify when a command completes
./long-running-process; ./ding
```

### Command-line options

| Option | Description                               | Default Value          |
| ------ | ----------------------------------------- | ---------------------- |
| `-m`   | Message to display                        | "Process completed"    |
| `-f`   | Path to custom sound file                 | "~/code/ding/ding.mp3" |
| `-s`   | Silent mode (no text output)              |                        |
| `-b`   | Use terminal bell instead of custom sound |                        |

### Examples

```bash
# Custom message
./task; ./ding -m "Backup completed!"

# Use terminal bell
./calculation; ./ding -b

# Just sound, no message
./process; ./ding -s

# Custom sound file
./import; ./ding -f /path/to/success.mp3
```

## Sound File Configuration

By default, Ding looks for a sound file at `~/code/ding/ding.mp3`. You can:

1. Place your sound file at this default location
2. Specify a different file with the `-f` flag
3. Modify the default path in the source code (`getSoundFilePath()` function in `ding.go`)

If Ding can't find or play the sound file, it automatically falls back to the terminal bell.

## How it works

Ding automatically selects the appropriate audio player based on your operating system:

- **Linux**: Uses mpg123, mpg321, mplayer, or ffplay (whichever is available)
- **macOS**: Uses afplay (built-in)
- **Windows**: Uses PowerShell Media.SoundPlayer

## License

MIT License
