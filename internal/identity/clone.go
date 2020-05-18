package identity

import (
	"fmt"
	"os"
	"os/exec"
)

func Clone(p *Profile, repo string) error {
	cmd := exec.Command("git", "clone", repo)
	cmd.Env = append(os.Environ(), fmt.Sprintf("GIT_SSH_COMMAND=ssh -i %s -F /dev/null", p.IdentityKey))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
