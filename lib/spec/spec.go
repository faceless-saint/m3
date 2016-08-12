/* Spec is a library containing type definitions and utility functions
 * pertaining to Minecraft modpack specifications. This includes the
 * management of Forge installers and mod configurations. The management
 * of mod files themselves are delegated to the 'm3/lib/mod' package.
 */
package spec

import (
	"encoding/json"
	"fmt"
	"github.com/faceless-saint/m3/lib/mod"
	"io/ioutil"
	"net/http"
)

// Spec values represent complete modpack specifications.
type Spec struct {
	Forge Installer
	/* "forge": {
	 *      "version": "",
	 *      "checksum": "",
	 *      "serverChecksum": ""
	 * }
	 */
	Config Config
	/* "config": {
	 *      "repository": "",
	 *      "path": ""
	 * }
	 */
	Mods mod.Directory
	/* "mods": {
	     *      "items": [
	     *          {
		 *              "name": "",
		 *              "version": "",
		 *              "checksum":"",
		 *              "url": "",
		 *              "curse": ""
	     *          }...
	     *      ],
	     *      "ignore": [""...]
	     * }
	*/
}

// Raw values act as JSON import containers for Spec values
type Raw struct {
	Forge  RawInstaller
	Config Config
	Mods   mod.RawDirectory
}

// FromFile returns a new Spec parsed from the given JSON file.
func FromFile(file string) (*Spec, error) {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}
	fmt.Printf("Local spec: %s\n", file)
	return FromJSON(data)
}

// FromRemote returns a new Spec parsed from remote JSON data.
func FromRemote(url string) (*Spec, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	page, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	fmt.Printf("Remote spec: %s\n", url)
	return FromJSON(page)
}

// FromGitHub returns a new Spec parsed from the given GitHub content.
func FromGitHub(repository, path string) (*Spec, error) {
	url := fmt.Sprintf("https://raw.githubusercontent.com/%s/master/%s", repository, path)
	return FromRemote(url)
}

// FromJSON returns a new Spec parsed raw JSON data.
func FromJSON(data []byte) (*Spec, error) {
	var raw Raw
	err := json.Unmarshal(data, &raw)
	if err != nil {
		return nil, err
	}
	mods, err := mod.NewDirectory(&raw.Mods)
	if err != nil {
		return nil, err
	}
	forge, err := NewInstaller(&raw.Forge)
	if err != nil {
		return nil, err
	}
	spec := Spec{Forge: *forge, Config: raw.Config, Mods: *mods}
	fmt.Printf("Forge version: %s\nConfig source: %v\n",
		spec.Forge.Version, spec.Config.Repository)
	return &spec, nil
}
