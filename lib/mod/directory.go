package mod

import (
	"github.com/cavaliercoder/grab"
	"github.com/faceless-saint/m3/lib/net"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// Directory values represent definitions for a mod directory.
type Directory struct {
	// Ignore is a list of file names to skip during directory cleaning.
	Ignore []string
	// Items is a list of definitions for downloadable mod files.
	Items net.Downloadables
}

// RawDirectory values act as JSON import containers for Directory values.
type RawDirectory struct {
	Ignore []string
	Items  []Raw
}

// NewDirectory returns a new Directory value from the imported RawDirectory.
func NewDirectory(raw *RawDirectory) (*Directory, error) {
	this := Directory{raw.Ignore, nil}
	for _, el := range raw.Items {
		mod, err := New(&el)
		if err != nil {
			return nil, err
		}
		this.Items = append(this.Items, mod)
	}
	return &this, nil
}

// Fetch downloads all mods defined in the Directory to the local "mods"
// directory. Behavior and usage are otherwise identical to FetchTo.
func (this *Directory) Fetch(num int, verbose bool) (<-chan *grab.Response, int, error) {
	return this.FetchTo("mods", num, verbose)
}

// FetchTo downloads all mods defined in the Directory to the given
// local directory 'dir'. The number 'num' provided determines permitted
// number of simultaneous downloads. Warnings will be printed for files
// that lack reference checksums if 'verbose' is true. Clean is run
// immediately preceeding the downloads.
func (this *Directory) FetchTo(dir string, num int, verbose bool) (<-chan *grab.Response, int, error) {
	if err := this.Clean(dir); err != nil {
		return nil, 0, err
	}
	return this.Items.GetFiles(dir, num)
}

// Clean scans the filesystem path and disables any jar files that do
// not match the Directory spec to prepare for a clean mod installation.
// If any disabled mods match file names that are now called for, they
// will automatically be enabled again if they pass checksum validation.
func (this *Directory) Clean(dir string) error {
	if _, err := os.Stat(dir); err != nil {
		// Directory does not exist - nothing to clean
		return nil
	}
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return err
	}
	ignoreMap := make(map[string]*struct{}, len(this.Ignore))
	for _, el := range this.Ignore {
		ignoreMap[el] = new(struct{})
	}
	fileMap := make(map[string]net.Downloadable, len(this.Items))
	for _, el := range this.Items {
		fileMap[el.Filename()] = el
	}
	for _, el := range files {
		if _, ok := ignoreMap[el.Name()]; ok {
			// Mod is being deliberately ignored
			continue
		} else if mod, ok := fileMap[el.Name()]; ok {
			// Mod should exist - verify checksum
			err = net.VerifyFile(filepath.Join(dir, mod.Filename()),
				mod.Checksum(), mod.Hash())
			if err != nil {
				return err
			}
		} else if mod, ok := fileMap[strings.Replace(el.Name(), ".disabled", "", 1)]; ok {
			// Mod should exist but is disabled - enable and verify
			err := os.Link(filepath.Join(dir, el.Name()),
				filepath.Join(dir, mod.Filename()))
			if err != nil {
				return err
			}
			// Verify the checksum of the disabled mod
			err = net.VerifyFile(filepath.Join(dir, mod.Filename()),
				mod.Checksum(), mod.Hash())
			if err != nil {
				return err
			}
			if _, err := os.Stat(filepath.Join(dir, mod.Filename())); err == nil {
				err = os.Remove(filepath.Join(dir, el.Name()))
				if err != nil {
					return err
				}
			}
		} else if filepath.Ext(el.Name()) == ".jar" {
			// Mod file should not exist - disable with extension change
			err := os.Rename(filepath.Join(dir, el.Name()),
				filepath.Join(dir, el.Name()+".disabled"))
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// PruneDir deletes all disabled jar files in the given directory.
func PruneDir(dir string) error {
	files, err := filepath.Glob(".jar.disabled")
	if err != nil {
		return err
	}
	for _, f := range files {
		err := os.Remove(f)
		if err != nil {
			return err
		}
	}
	return nil
}
