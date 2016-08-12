package git

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"path"
	"strings"
)

// Repository values represent a public repository on GitHub.
type Repository struct {
	Owner string
	Name  string
}

// NewRepository takes a string in the form "<owner>/<repo>" and
// returns a new Repository value with the corresponding properties.
func NewRepository(repository string) (*Repository, error) {
	split := strings.Split(repository, "/")
	if len(split) != 2 {
		return nil, fmt.Errorf("error: invalid repository name")
	}
	return &Repository{split[0], split[1]}, nil
}

// Url returns the content URL for the Repository.
func (this *Repository) ContentPath() string {
	return fmt.Sprintf("api.github.com/repos/%s/%s/contents",
		this.Owner, this.Name)
}

// Explore returns a list of Content values for each item in the path.
// An empty path maps to the root of the repository path structure.
func (this *Repository) Explore(p string) (ContentList, error) {
	content := ContentList{}
	resp, err := http.Get("https://" + path.Join(this.ContentPath(), p))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	results, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(results, &content)
	return content, err
}

// Aggregate returns a flat list of Content values from the path and all
// subdirectories under it, recursively. Only file elements are returned.
func (this *Repository) Aggregate(p string) (ContentList, error) {
	full := ContentList{}
	content, err := this.Explore(p)
	if err != nil {
		return nil, err
	}
	for _, el := range content {
		if el.Type == "dir" {
			// Recurse into subdirectory and add all files to the list
			c, err := this.Aggregate(el.Path)
			if err != nil {
				return nil, err
			}
			full = append(full, c...)
		} else {
			// Add file to the list
			full = append(full, el)
		}
	}
	return full, nil
}
