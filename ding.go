package main

import (
	"flag"
	"fmt"
	"os"
	"time"
)

const (
	// ANSI escape code for bell
	bellChar = "\a"
	// Default message
	defaultMessage = "Process completed"
)

func main() {
	// Define command line flags
	message := flag.String("m", defaultMessage, "Message to display")
	repeat := flag.Int("r", 1, "Number of times to repeat the sound")
	interval := flag.Duration("i", 500*time.Millisecond, "Interval between sounds")
	noMessage := flag.Bool("s", false, "Silent mode (no text output)")

	flag.Parse()

	// Play the ding sound the specified number of times
	for i := 0; i < *repeat; i++ {
		// Print the bell character to trigger the terminal bell
		fmt.Fprint(os.Stdout, bellChar)

		// If it's not the last iteration and repeat > 1, wait for the specified interval
		if i < *repeat-1 {
			time.Sleep(*interval)
		}
	}

	// Print the message unless silent mode is enabled
	if !*noMessage {
		fmt.Println(*message)
	}
}
