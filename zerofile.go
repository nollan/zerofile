package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// Arghss!
var (
	quiet     = flag.Bool("quiet", false, "No progress report during zeroing, a bit faster")
	blockSize = flag.Int("blocksize", 4096, "Amount of zeroes to write at each pass")
)

const (
	minAmount = 256
)

func main() {
	// Keep score of bytes written
	var bytesWritten = 0
	// Keep score of when started
	var startTime = time.Now()

	// Well, parse the flags! :)
	flag.Parse()

	// Temp-file selection
	tmpFilename := "0slask0.zro"
	if len(flag.Args()) > 0 {
		tmpFilename = flag.Arg(0)
	}

	// Handle ctrl-c and kills
	ctrlC := make(chan os.Signal, 2)
	signal.Notify(ctrlC, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-ctrlC
		if !*quiet {
			fmt.Println()
		}
		fmt.Println("Cleaning up")
		cleanup(tmpFilename)
		os.Exit(1)
	}()

	// Make some zeroes
	bunchOfZeroes := make([]byte, *blockSize)

	// Open tmpFilename, and handle closing and removing of it
	zeroFile, err := os.OpenFile(tmpFilename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		log.Fatal(err)
	}
	defer zeroFile.Close()
	defer cleanup(tmpFilename)
	fmt.Printf("Using tempfile: '%v'\n", tmpFilename)

	// Amount to write at each pass
	amount := *blockSize
	for {
		// Write zeroes up to amount (which is equal or less than blockSize)
		n, err := zeroFile.Write(bunchOfZeroes[:amount])
		if err != nil {
			// Check for no more space left on device
			if e, ok := err.(*os.PathError); ok && e.Err == syscall.ENOSPC {
				// if not all of the amount is written, halve it and try again
				// in case of very large blockSize.
				if n != amount {
					amount /= 2
				}
				// Limit of amount size, disk buffer size is normally larger
				if amount < minAmount {
					break
				}
			} else {
				// Other error
				cleanup(tmpFilename)
				log.Fatal(err)
				break
			}
		}
		// Update stats
		bytesWritten += n
		// Keep quiet if wanted
		if !*quiet {
			fmt.Printf("Written: %d bytes        \r", bytesWritten)
		}
	}
	// Print some stats
	if !*quiet {
		fmt.Println()
	}
	fmt.Printf("Duration: %v ; Performance: %.3f bytes/sec\n",
		time.Since(startTime),
		float64(bytesWritten)/time.Since(startTime).Seconds())
}

// Cleanup, remove tempfile
func cleanup(filename string) {
	fmt.Printf("Removing tempfile...")
	err := os.Remove(filename)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Done")
}
