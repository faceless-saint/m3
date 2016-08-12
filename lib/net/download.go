package net

import (
	"encoding/hex"
	"fmt"
	"github.com/cavaliercoder/grab"
	"hash"
	"os"
	"path/filepath"
)

// Downloadable is the interface for files hosted at public URLs.
type Downloadable interface {
	// Url returns the download URL for the file.
    Url() string
    // Filename returns the name to save the file as.
	Filename() string
    // Checksum returns the checksum used to validate the file.
	Checksum() string
    // Hash returns the hash.Hash used to compute the checksum.
	Hash() hash.Hash
}

// Downloadables values are lists of objects implementing Downloadable.
type Downloadables []Downloadable

// GetFiles downloads all files in the Downloadables list and saves them
// to the target directory, using at most 'num' simultaneous downloads.
// Returns a channel emitting Response objects as they become available.
func (this *Downloadables) GetFiles(dir string, num int) (<-chan *grab.Response, int, error) {
	reqs, err := this.GetFilesDeferred(dir)
	if err != nil {
		return nil, 0, err
	}
	// Start downloads and return the response channel
	client := grab.NewClient()
	client.UserAgent = "m3"
	return client.DoBatch(num, reqs...), len(reqs), nil
}

// GetFilesDeferred creates a download request for every file in the
// Downloadables list that isn't found in the target directory.
func (this *Downloadables) GetFilesDeferred(dir string) ([]*grab.Request, error) {
	// Prepare each download request
	reqs := []*grab.Request{}
	for _, dl := range *this {
		destination := filepath.Join(dir, dl.Filename())
		if _, err := os.Stat(destination); err != nil {
            req, err := GetFileDeferred(dl, destination)
            if err != nil {
                return nil, err
            }
            reqs = append(reqs, req)
        }
    }
	return reqs, nil
}

// GetFile downloads and saves the file described by the given
// Downloadable object. If no filename is provided then the default
// filename for the object is used.
func GetFile(dl Downloadable, file string) (*grab.Response, error) {
	req, err := GetFileDeferred(dl, file)
	if err != nil {
		return nil, err
	} else if req == nil {
        return nil, nil
    }
	// Execute the request with the default client.
	client := grab.NewClient()
	client.UserAgent = "m3"
	return client.Do(req)
}

// GetFileDeferred returns a request to download the file described by
// the given Downloadable object. If no filename is provided then the
// default filename for the object is used.
func GetFileDeferred(dl Downloadable, file string) (*grab.Request, error) {
	if file == "" {
		file = dl.Filename()
	}
	if _, err := os.Stat(file); err == nil {
        // File already exists - skip download.
        return nil, nil
        err := VerifyFile(file, dl.Checksum(), dl.Hash())
        if err != nil {
            return nil, err
        }
        if _, err := os.Stat(file); err == nil {
            // File verification passed - skip download.
            fmt.Printf("%s already found\n", file)
            return nil, nil
        }
    }
    // Create the download request
    req, err := grab.NewRequest(dl.Url())
	if err != nil {
		return nil, err
	}
    // Attempt to add the file's checksum for verification.
	if len(dl.Checksum()) != 0 {
		if checksum, err := hex.DecodeString(dl.Checksum()); err == nil {
			req.Checksum = checksum
			req.Hash = dl.Hash()
		}
	}
    // Make sure the target directory exists.
	os.MkdirAll(filepath.Dir(file), 0755)
	req.RemoveOnError = true
	req.Filename = file
    return req, err
}
