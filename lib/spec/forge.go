package spec

import (
	"github.com/cavaliercoder/grab"
	"github.com/faceless-saint/m3/lib/net"
	"hash"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Installer values represent downloadable Forge installer files.
type Installer struct {
	Version        string
	checksum       string
	hash           hash.Hash
	ServerChecksum string
}

// RawInstaller values act as JSON import containers for Installer values.
type RawInstaller struct {
	Version        string
	Checksum       string
	ServerChecksum string
}

// NewInstaller returns a new Install from the given RawInstaller.
func NewInstaller(raw *RawInstaller) (*Installer, error) {
	// Determine checksum and hashing algorithm
	hashsum := strings.SplitN(raw.Checksum, ":", 2)
	if len(hashsum) < 2 {
		// Default to SHA256 if no algorithm is given
		hashsum = []string{"sha256", hashsum[0]}
	}
	h, err := net.NewHash(hashsum[0])
	if err != nil {
		return nil, err
	}
	return &Installer{raw.Version, hashsum[1], h, raw.ServerChecksum}, nil
}

func (this *Installer) Filename() string {
	return "forge-" + this.Version + "-installer.jar"
}
func (this *Installer) Url() string {
	return "http://files.minecraftforge.net/maven/net/minecraftforge/" +
		"forge/" + this.Version + "/forge-" + this.Version + "-installer.jar"
}
func (this *Installer) Checksum() string { return this.checksum }
func (this *Installer) Hash() hash.Hash  { return this.hash }

// Fetch downloads the forge installer for the Spec
func (this *Installer) Fetch(verbose bool) (*grab.Response, error) {
	// Prepare the working directory
	err := this.clean()
	if err != nil {
		return nil, err
	}
	// Download the Forge installer
	return net.GetFile(this, "")
}

// Install runs the Forge installer and installs server files.
func (this *Installer) Install(verbose bool) error {
	cmd := exec.Command("java", "-jar",
		this.Filename(), "--installServer")
	if verbose {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	return cmd.Run()
}

func (this *Installer) clean() error {
	// Remove all invalid Forge installers
	forge_installers, _ := filepath.Glob("forge-*-installer.jar")
	for _, el := range forge_installers {
		if el != this.Filename() {
			err := os.RemoveAll(el)
			if err != nil {
				return err
			}
		} else {
			err := net.VerifyFile(el, this.Checksum(), this.Hash())
			if err != nil {
				return err
			}
		}
	}
	// Remove all invalid Forge server files
	forge_servers, _ := filepath.Glob("forge-*-universal.jar")
	for _, el := range forge_servers {
		if el != "forge-"+this.Version+"-universal.jar" {
			err := os.RemoveAll(el)
			if err != nil {
				return err
			}
			err = os.RemoveAll("libraries")
			if err != nil {
				return err
			}
		} else {
			err := net.VerifyFile(el, this.ServerChecksum, this.Hash())
			if err != nil {
				return err
			}
		}
	}
	return nil
}
