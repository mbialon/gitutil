package identity

import (
	"github.com/mbialon/gitutil/internal/git"
)

func Get() (*Profile, error) {
	var p Profile
	var c git.Config
	p.Name = c.Name()
	p.Email = c.Email()
	p.SignOff = c.SignOff()
	p.GPGSign = c.GPGSign()
	return &p, c.Err()
}
