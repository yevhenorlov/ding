# Installation

1. Install Go, then build it:

```sh
go build ding.go
```

2. Set it to an alias you want in your shell config

```sh
alias ding="~/path/to/ding"
```

# Basic usage after a long-running command

./long-running-process; ./ding

# Custom message

./long-running-process; ./ding -m "Backup completed!"

# Multiple dings with 1 second interval

./time-consuming-task; ./ding -r 3 -i 1s

# Just the sound, no message

./build-project; ./ding -s
