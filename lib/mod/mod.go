/* Mod is a library containing type definitions and utility functions
 * pertaining to Minecraft mod files.
 */
package mod

import (
	"fmt"
	"github.com/faceless-saint/m3/lib/net"
	"hash"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
)

// Raw values act as universal import containers for mod types. All mod
// types must implement m3net.Downloadable.
type Raw struct {
	Name     string
	Version  string
	Checksum string
	Url      string
	Curse    string
}

// New initializes a new mod type from the imported Raw value. The
// specific mod type is chosen dynamically based on the defined data
// fields, using the most feature-rich implementation supported.
func New(mod *Raw) (net.Downloadable, error) {
	if mod.Name == "" {
		// Missing required 'name' property.
		return nil, &InitError{*mod, "mod init error: missing required property 'name'"}
	}
	// Determine checksum and hashing algorithm
	hashsum := strings.SplitN(mod.Checksum, ":", 2)
	if len(hashsum) < 2 {
		// Default to SHA256 if no algorithm is given
		hashsum = []string{"sha256", hashsum[0]}
	}
	h, err := net.NewHash(hashsum[0])
	if err != nil {
		return nil, err
	}
	base := RemoteMod{mod.Name, mod.Version, hashsum[1], h, mod.Url}

	/* * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * *
	 * When defining new mod types, add their initialization checks  *
	 * here to allow them to be loaded dynamically. They should be   *
     * specified in decending order of preference.                   *
	 * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * */
	switch {
	// Use a CurseMod value.
	case mod.Curse != "":
		curseMod := CurseMod{base, mod.Curse}
		err := curseMod.Init()
		return &curseMod, err

	// Use a basic RemoteMod value. (last resort)
	case mod.Url != "":
		return &base, nil

	// Error: no mod implementations were satisfied.
	default:
		return nil, &InitError{*mod, "mod init error: not enough data to implement Mod - need one of 'curse', 'url'"}
	}
}

// RemoteMod values represent mod files hosted at a generic URL.
type RemoteMod struct {
    Name     string
	Version  string
	checksum string
	hash     hash.Hash
	url      string
}

func (this *RemoteMod) Url() string      { return this.url }
func (this *RemoteMod) Checksum() string { return this.checksum }
func (this *RemoteMod) Hash() hash.Hash  { return this.hash }
func (this *RemoteMod) Filename() string {
	if this.Version != "" {
		return fmt.Sprintf("%s-%s-%s.jar",
			this.Name, this.Version, net.Digest(this.Url(), 6))
	} else {
		return fmt.Sprintf("%s-%s.jar",
			this.Name, net.Digest(this.Url(), 6))
	}
}

// CurseMod values represent mod files hosted on Curseforge. This adds
// extra functionality inluding update checking. The "Curse" property
// can either be the full download URL or the Curse file ID.
type CurseMod struct {
	RemoteMod
    // Curse is the file download ID on Curseforge.
	Curse string
}

// Curseforge download URL is constructed as follows:
//    url_head + {mod_name} + url_mid + {file_id} + url_tail
const curse_url_head = "https://minecraft.curseforge.com/projects/"
const curse_url_mid = "/files/"
const curse_url_tail = "/download"

// Init validates the Curse property for the CurseMod as either a Curse
// download URL or a file ID. It then sets the download URL accordingly.
func (this *CurseMod) Init() error {
	// Match Curse field to either a URL or a file ID
	re_url := regexp.MustCompile(curse_url_head + this.Name +
		curse_url_mid + "[0-9]+" + curse_url_tail)
	re_id := regexp.MustCompile("^[0-9]+$")
	if re_url.Match([]byte(this.Curse)) {
		// Curse field contains a Curseforge URL - convert to file ID
		this.Curse = string(strings.Replace(strings.Replace(this.Curse,
			curse_url_head+this.Name+curse_url_mid, "", 1),
			curse_url_tail, "", 1))
	} else if !re_id.Match([]byte(this.Curse)) {
		// Curse field is not a valid file ID
		return &InitError{Raw{Curse: this.Curse},
			"mod init error: 'curse' property is invalid"}
	}

	// Set download URL
	this.url = curse_url_head + this.Name + curse_url_mid +
		this.Curse + curse_url_tail
	return nil
}

// GetLatest gets the latest file ID for the Mod from Curseforge.
func (this *CurseMod) GetLatest() (string, error) {
	re := regexp.MustCompile("href=\"/projects/" +
		this.Name + "/files/[0-9]+/download")
	resp, err := http.Get("https://minecraft.curseforge.com/projects/" +
		this.Name + "/files?sort=releasetype")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	page, err := ioutil.ReadAll(resp.Body)
	return string(strings.Replace(strings.Replace(string(re.Find(page)),
		"href=\"/projects/"+this.Name+"/files/", "", 1),
		"/download", "", 1)), nil
}

// SetTarget sets the CurseMod value to track the given remote file from
// Curseforge. If a complete Curseforge URL is given, the file id will
// be extracted from it.
func (this *CurseMod) SetTarget(curse string) error {
	this.Curse = curse
	return this.Init()
}
