package main

import (
	"bytes"
	"fmt"
	"os"
	"sort"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mbialon/gitutil/internal/identity"
	"github.com/muesli/termenv"
)

func main() {
	config, err := identity.ReadFile()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		os.Exit(1)
	}
	profile, err := identity.Get()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		os.Exit(1)
	}
	p := tea.NewProgram(initialize(config, profile), update, view)
	if err := p.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		os.Exit(1)
	}
}

type Model struct {
	Profiles []Profile
	Current  *identity.Profile

	Cursor int
	Chosen *Profile
	Err    error
}

type Profile struct {
	Label string
	*identity.Profile
}

func initialize(config *identity.Config, current *identity.Profile) func() (tea.Model, tea.Cmd) {
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
		return Model{
			Profiles: profiles,
			Current:  current,
		}, nil
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
	buf.WriteString(termenv.String("Current profile\n\n").Underline().String())

	buf.WriteString(fmt.Sprintf("  %s\n", m.Current.Name))
	buf.WriteString(fmt.Sprintf("  %s\n", m.Current.Email))
	if m.Current.SignOff {
		fmt.Fprintf(&buf, "  +signoff\n")
	}
	if m.Current.GPGSign {
		fmt.Fprintf(&buf, "  +gpgsign\n")
	}
	buf.WriteString("\n")

	buf.WriteString(termenv.String("Choose profile\n\n").Underline().String())

	for i, profile := range m.Profiles {
		gutter := " "
		if m.Cursor == i {
			gutter = "│"
		}
		fmt.Fprintf(&buf, "%s [%s]\n", gutter, profile.Label)
		fmt.Fprintf(&buf, "%s %s\n", gutter, profile.Name)
		fmt.Fprintf(&buf, "%s %s\n", gutter, profile.Email)
		if profile.SignOff {
			fmt.Fprintf(&buf, "%s +signoff\n", gutter)
		}
		if profile.GPGSign {
			fmt.Fprintf(&buf, "%s +gpgsign\n", gutter)
		}
		fmt.Fprintln(&buf)
	}

	buf.WriteString("\nPress q to quit.\n")
	return buf.String()
}

func errorView(m Model) string {
	s := fmt.Sprintf("ERROR: %s\n", m.Err.Error())
	return termenv.String(s).Foreground(termenv.ANSIBrightRed).String()
}

func successView() string {
	return termenv.String("OK").Foreground(termenv.ANSIGreen).String()
}
