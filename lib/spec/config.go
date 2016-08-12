package spec

import (
    "github.com/cavaliercoder/grab"
    "github.com/faceless-saint/m3/lib/net"
    "github.com/faceless-saint/m3/lib/git"
    "strings"
)

// Config values represent Forge mod configurations hosted on GitHub.
type Config struct {
	Repository string
	Path       string
	Items      net.Downloadables
}

// Fetch downloads all config files from the repository.
func (this *Config) Fetch(num int, verbose bool) (<-chan *grab.Response, int, error) {
	repo, err := git.NewRepository(this.Repository)
	if err != nil {
		return nil, 0, err
	}
	configs, err := repo.Aggregate(this.Path)
	if err != nil {
		return nil, 0, err
	}
	for _, el := range configs.JustFiles() {
		conf := el
		conf.Path = strings.Replace(conf.Path, this.Path+"/", "", 1)
		this.Items = append(this.Items, &conf)
	}
	return this.Items.GetFiles("config", num)
}
