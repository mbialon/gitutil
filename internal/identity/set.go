package identity

import (
	"fmt"

	"github.com/mbialon/gitutil/internal/git"
)

func Set(p *Profile) error {
	var c git.Config
	c.SetName(p.Name)
	c.SetEmail(p.Email)
	c.SetSignOff(p.SignOff)
	c.SetGPGSign(p.GPGSign)
	if p.IdentityKey != "" {
		c.SetSSHCommand(fmt.Sprintf("ssh -i %s -F /dev/null", p.IdentityKey))
	}
	return c.Err()
}
