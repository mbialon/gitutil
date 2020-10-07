package identity

import (
	"github.com/mitchellh/go-homedir"
	"github.com/pelletier/go-toml"
)

type Config struct {
	Profiles map[string]*Profile
}

type Profile struct {
	Name        string `toml:"name"`
	Email       string `toml:"email"`
	SignOff     bool   `toml:"signoff"`
	GPGSign     bool   `toml:"gpgsign"`
	IdentityKey string `toml:"identity_key"`
}

func ReadFile() (*Config, error) {
	fpath, err := homedir.Expand("~/.gitprofiles")
	if err != nil {
		return nil, err
	}
	tree, err := toml.LoadFile(fpath)
	if err != nil {
		return nil, err
	}
	config := new(Config)
	if err := tree.Unmarshal(config); err != nil {
		return nil, err
	}
	return config, nil
}
