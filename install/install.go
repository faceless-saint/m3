package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/faceless-saint/m3/lib/output"
	"github.com/faceless-saint/m3/lib/spec"
	"golang.org/x/crypto/ssh/terminal"
	"io/ioutil"
	"os"
)

// Set the length of the progress bar
const pb_len = 50
const pb_timer = 200

type Config struct {
	ConfigFile      string
	File            string
	Remote          string
	TargetDirectory string
	Concurrency     int
	Server          bool
	Client          bool
	Verbose         bool
	VeryVerbose     bool
}

type stringFlag struct {
	set   bool
	value string
}

func (this *stringFlag) Set(str string) error {
	this.value = str
	this.set = true
	return nil
}

func (this *stringFlag) String() string {
	return this.value
}

func ParseConfig() *Config {
	c_File := stringFlag{value: "modpack.json"}
	c_Remote := stringFlag{}

	// Configure input files
	c_ConfigFile := flag.String("f", "m3.conf", "m3 configuration file")
	flag.Var(&c_File, "c", "local modpack definition JSON file")
	flag.Var(&c_Remote, "r", "remote modpack definition JSON")

	// Configure target directory
	c_TargetDirectory := flag.String("t", ".", "target directory")

	// Configure download concurrency
	c_Concurrency := flag.Int("n", 3, "number of simultaneous downloads")

	// Configure client/server mode
	c_Server := flag.Bool("server", false, "install in server mode")
	c_Client := flag.Bool("client", false, "install in client mode")

	// Configure output verbosity
	c_Verbose := flag.Bool("v", false, "display verbose log output")
	c_VeryVerbose := flag.Bool("vv", false, "display very verbose log output (implies -v)")

	// Parse command line arguments
	flag.Parse()
	config := Config{ConfigFile: *c_ConfigFile, File: "modpack.json"}
	config.ReadFile()
	config.TargetDirectory = *c_TargetDirectory
	config.Concurrency = *c_Concurrency
	config.Server = *c_Server
	config.Client = *c_Client
	config.Verbose = *c_Verbose
	config.VeryVerbose = *c_VeryVerbose
	if c_File.set {
		config.File = c_File.value
	}
	if c_Remote.set {
		config.Remote = c_Remote.value
	}
	if config.VeryVerbose {
		config.Verbose = true
	}
	return &config
}

func (this *Config) ReadFile() error {
	data, err := ioutil.ReadFile(this.ConfigFile)
	if err != nil {
		return err
	}
	if err = json.Unmarshal(data, &this); err != nil {
		return err
	}
	return nil
}

func GetSpec(config Config) (*spec.Spec, error) {
	if config.Remote != "" {
		return spec.FromRemote(config.Remote, true, config.VeryVerbose)
	} else {
		return spec.FromFile(config.File, true, config.VeryVerbose)
	}
}

func main() {
	// Pause at program completion
	if terminal.IsTerminal(int(os.Stdout.Fd())) {
		defer bufio.NewReader(os.Stdin).ReadBytes([]byte("\n")[0])
		defer fmt.Print("Press [enter] to exit...\n")
	}

	// Parse configuration options
	config := ParseConfig()

	s, err := GetSpec(*config)
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
	os.MkdirAll(config.TargetDirectory, 0755)
	os.Chdir(config.TargetDirectory)

	// Start mod downloads
	respch, count, err := s.Mods.Fetch(config.Concurrency, config.Verbose)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	// Track mod download progress
	modTracker := output.DownloadTracker{"mods", respch, nil, pb_timer, count, len(s.Mods.Items)}
	modTracker.Log()

	// Start config downloads
	respch, count, err = s.Config.Fetch(config.Concurrency, config.Verbose)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	// Track config download progress
	configTracker := output.DownloadTracker{"configs", respch, nil, pb_timer, count, len(s.Config.Items)}
	configTracker.Log()

	if config.Client || config.Server {
		// Download the Forge intaller
		_, err := s.Forge.Fetch(config.Verbose)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}
		//forgeTracker := output.DownloadTracker{"forge installer", respch, nil, pb_timer, count, 1}
		//forgeTracker.Log()

		if config.Server {
			// Install Forge server files
			fmt.Print("Installing Forge server files... ")
			if config.Verbose {
				fmt.Print("\n")
			}
			if err := s.Forge.Install(config.Verbose); err != nil {
				fmt.Fprintf(os.Stderr, "\n%v\n", err)
				os.Exit(1)
			} else {
				fmt.Print("Done.\n")
			}
		}
	}
	fmt.Print("Installation complete!\n")
}
