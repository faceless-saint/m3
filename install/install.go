package main

import (
	"bufio"
	"fmt"
	"github.com/faceless-saint/m3/lib/config"
	"github.com/faceless-saint/m3/lib/output"
	"golang.org/x/crypto/ssh/terminal"
	"os"
)

// Set the download tracker interval
const pb_timer = 200

func main() {
	// Pause at program completion, but only if session is a terminal.
	if terminal.IsTerminal(int(os.Stdout.Fd())) {
		defer bufio.NewReader(os.Stdin).ReadBytes([]byte("\n")[0])
		defer fmt.Print("Press [enter] to exit...\n")
	}

	// Parse configuration options
	conf := config.Parse()

	s, err := conf.GetSpec()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	// Return to original working directory before exiting
	if original_path, err := os.Getwd(); err == nil {
		defer os.Chdir(original_path)
	} else {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
	// Create (if needed) and navigate to target directory
	os.MkdirAll(conf.Env.TargetDir, 0755)
	os.Chdir(conf.Env.TargetDir)

	// Start mod downloads
	respch, count, err := s.Mods.Fetch(conf.Env.Concurrency, conf.Verbose)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	// Track mod download progress
	modTracker := output.DownloadTracker{"mods", respch, nil, pb_timer, count, len(s.Mods.Items)}
	modTracker.Log()

	// Start config downloads
	respch, count, err = s.Config.Fetch(conf.Env.Concurrency, conf.Verbose)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	// Track config download progress
	configTracker := output.DownloadTracker{"configs", respch, nil, pb_timer, count, len(s.Config.Items)}
	configTracker.Log()

	if conf.Install.Client || conf.Install.Server {
		// Download the Forge intaller
		fmt.Print("Downloading Forge installer... ")
		_, err := s.Forge.Fetch(conf.Verbose)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}
		fmt.Print("Done.\n")

		if conf.Install.Server {
			// Install Forge server files
			fmt.Print("Installing Forge server files... ")
			if conf.Verbose {
				fmt.Print("\n")
			}
			if err := s.Forge.Install(conf.Verbose); err != nil {
				fmt.Fprintf(os.Stderr, "\n%v\n", err)
				os.Exit(1)
			} else {
				fmt.Print("Done.\n")
			}
		}
	}
	fmt.Print("Installation complete!\n")
}
