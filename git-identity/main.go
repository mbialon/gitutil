package main

import (
	"fmt"
	"os"

	"github.com/mbialon/gitutil/internal/identity"
)

func main() {
	p, err := identity.Select()
	if err != nil {
		fmt.Fprintf(os.Stderr, "fatal: cannot select profile, err: %v\n", err)
		os.Exit(1)
	}
	if err := identity.Set(p); err != nil {
		fmt.Fprintf(os.Stderr, "fatal: cannot set profile, err: %v\n", err)
		os.Exit(1)
	}
}
