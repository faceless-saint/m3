package git

import (
	"hash"
	"github.com/faceless-saint/m3/lib/net"
)


// ContentList values represent lists of Content values.
type ContentList []Content

// Content values represent item returned from the GitHub Content API.
type Content struct {
	Name         string
	Type         string
	Size         int
	Download_url string
	Path         string
	Sha          string
}

func (this *Content) Url() string      { return this.Download_url }
func (this *Content) Filename() string { return this.Path }
func (this *Content) Checksum() string { return this.Sha }
func (this *Content) Hash() hash.Hash {
	return &net.GitHash{}
}
 
// JustFiles returns a new ContentList containing only the file elements
// of the given list (no directory elements).
func (this *ContentList) JustFiles() ContentList {
	newList := ContentList{}
	for _, el := range *this {
		if el.Type == "file" {
			newList = append(newList, el)
		}
	}
	return newList
}
