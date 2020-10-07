package git

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strconv"
)

type Config struct {
	err error
}

func (c *Config) Name() string {
	s := c.get("user.name")
	return s
}

func (c *Config) SetName(s string) {
	c.set("user.name", s)
}

func (c *Config) Email() string {
	s := c.get("user.email")
	return s
}

func (c *Config) SetEmail(s string) {
	c.set("user.email", s)
}

func (c *Config) SignOff() bool {
	return c.parseBool(c.get("format.signoff"))
}

func (c *Config) parseBool(s string) bool {
	if c.err != nil {
		return false
	}
	v, err := strconv.ParseBool(s)
	if err != nil {
		c.err = fmt.Errorf("parse %s: %w", s, err)
	}
	return v
}

func (c *Config) SetSignOff(b bool) {
	c.set("format.signoff", strconv.FormatBool(b))
}

func (c *Config) GPGSign() bool {
	return c.parseBool(c.get("commit.gpgsign"))
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

func (c *Config) get(key string) string {
	if c.err != nil {
		return ""
	}
	cmd := exec.Command("git", "config", key)
	cmd.Dir = ""
	b, err := cmd.Output()
	if err != nil {
		c.err = fmt.Errorf("get %s: %w", key, err)
		return ""
	}
	return string(bytes.TrimSpace(b))
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
		c.err = fmt.Errorf("set %s: %w", key, err)
	}
}
