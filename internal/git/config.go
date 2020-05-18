package git

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
)

type Config struct {
	err error
}

func (c *Config) SetName(s string) {
	c.set("user.name", s)
}

func (c *Config) SetEmail(s string) {
	c.set("user.email", s)
}

func (c *Config) SetSignOff(b bool) {
	c.set("format.signoff", strconv.FormatBool(b))
}

func (c *Config) SetGPGSign(b bool) {
	c.set("commit.gpgsign", strconv.FormatBool(b))
}

func (c *Config) SetSSHCommand(s string) {
	c.set("core.sshCommand", s)
}

func (c *Config) Err() error {
	return c.err
}

func (c *Config) set(key, val string) {
	if c.err != nil {
		return
	}
	cmd := exec.Command("git", "config", key, val)
	cmd.Dir = ""
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		c.err = fmt.Errorf("cannot set %s: %w", key, err)
	}
}
