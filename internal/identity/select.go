package identity

import (
	"os"
	"os/exec"
	"strings"

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

func Select() (*Profile, error) {
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
	keys := make([]string, 0, len(config.Profiles))
	for k := range config.Profiles {
		keys = append(keys, k)
	}
	r := strings.NewReader(strings.Join(keys, "\n"))
	cmd := exec.Command("fzf")
	cmd.Stdin = r
	cmd.Stderr = os.Stderr
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	key := strings.TrimSpace(string(out))
	return config.Profiles[key], nil
}
