package main

import (
	"bytes"
	"fmt"
	"os"
	"sort"
	"strings"
	"text/tabwriter"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mbialon/gitutil/internal/identity"
)

func main() {
	config, err := identity.ReadFile()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		os.Exit(1)
	}
	p := tea.NewProgram(initialize(config), update, view)
	if err := p.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		os.Exit(1)
	}
}

type Model struct {
	Profiles []Profile
	Cursor   int
	Chosen   *Profile
	Err      error
}

type Profile struct {
	Label string
	*identity.Profile
}

func initialize(config *identity.Config) func() (tea.Model, tea.Cmd) {
	return func() (tea.Model, tea.Cmd) {
		var profiles []Profile
		for k, v := range config.Profiles {
			profiles = append(profiles, Profile{
				Label:   k,
				Profile: v,
			})
		}
		sort.Slice(profiles, func(i, j int) bool {
			return profiles[i].Label < profiles[j].Label
		})
		return Model{Profiles: profiles}, nil
	}
}

func update(msg tea.Msg, model tea.Model) (tea.Model, tea.Cmd) {
	m, _ := model.(Model)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "up", "k":
			if m.Cursor > 0 {
				m.Cursor--
			}
		case "down", "j":
			if m.Cursor < len(m.Profiles)-1 {
				m.Cursor++
			}
		case "enter":
			m.Chosen = &m.Profiles[m.Cursor]
			return m, func() tea.Msg {
				err := identity.Set(m.Chosen.Profile)
				return setMsg{err}
			}
		}
	case setMsg:
		if msg.err != nil {
			m.Err = msg.err
		}
		return m, tea.Quit
	}
	return m, nil
}

type setMsg struct {
	err error
}

func view(model tea.Model) string {
	m, _ := model.(Model)

	if m.Chosen == nil {
		return profilesView(m)
	}
	if m.Err != nil {
		return errorView(m)
	}
	return successView()
}

func profilesView(m Model) string {
	var buf bytes.Buffer
	buf.WriteString("Choose profile\n\n")

	tw := tabwriter.NewWriter(&buf, 2, 2, 1, ' ', 0)
	for i, profile := range m.Profiles {
		cursor := " "
		if m.Cursor == i {
			cursor = ">"
		}

		fmt.Fprintf(tw, "%s\t[%s]\t%s\t<%s>", cursor, profile.Label, profile.Name, profile.Email)
		var options []string
		if profile.SignOff {
			options = append(options, "+signoff")
		}
		if profile.GPGSign {
			options = append(options, "+gpg-sign")
		}
		if len(options) > 0 {
			fmt.Fprintf(tw, "\t%s", strings.Join(options, " "))
		}
		fmt.Fprintln(tw)
	}
	tw.Flush()

	buf.WriteString("\nPress q to quit.\n")
	return buf.String()
}

func errorView(m Model) string {
	var buf bytes.Buffer
	buf.WriteString("ERROR: ")
	buf.WriteString(m.Err.Error())
	buf.WriteString("\n")
	return buf.String()
}

func successView() string {
	return "OK\n"
}
