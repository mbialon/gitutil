package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"sort"
	"text/template"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/mbialon/gitutil/internal/identity"
	"github.com/muesli/termenv"
)

var (
	version = "dirty"
	commit  = "dirty"
	date    = "dirty"
	builtBy = "dirty"
)

var versionTemplate = template.Must(template.New("").Parse(`Version:  {{.Version}}
Commit:   {{.Commit}}
Date:     {{.Date}}
Built by: {{.BuiltBy}}
`))

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	fs := flag.NewFlagSet("git-identity", flag.ExitOnError)
	var (
		versionFlag = fs.Bool("version", false, "Print version information")
	)
	if err := fs.Parse(args); err != nil {
		return err
	}

	if *versionFlag {
		return versionTemplate.Execute(os.Stdout, struct {
			Version, Commit, Date, BuiltBy string
		}{
			Version: version,
			Commit:  commit,
			Date:    date,
			BuiltBy: builtBy,
		})
	}

	config, err := identity.ReadFile()
	if err != nil {
		return err
	}
	profile, err := identity.Get()
	if err != nil {
		return err
	}
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
	model := Model{
		keys: keys,
		help: help.NewModel(),

		Profiles: profiles,
		Current:  profile,
	}
	p := tea.NewProgram(model)
	return p.Start()
}

type keyMap struct {
	Up      key.Binding
	Down    key.Binding
	SignOff key.Binding
	GPGSign key.Binding
	Enter   key.Binding
	Help    key.Binding
	Quit    key.Binding
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Help, k.Quit}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.SignOff, k.GPGSign, k.Enter},
		{k.Help, k.Quit},
	}
}

var keys = keyMap{
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("^/k", "move up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("v/j", "move down"),
	),
	SignOff: key.NewBinding(
		key.WithKeys("s"),
		key.WithHelp("s", "toggle sign off"),
	),
	GPGSign: key.NewBinding(
		key.WithKeys("shift+s"),
		key.WithHelp("S", "toggle gpg signing"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "select identity"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "toggle help"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "esc", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
}

type Model struct {
	keys keyMap
	help help.Model

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

type setMsg struct {
	err error
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.help.Width = msg.Width
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Up):
			if m.Cursor > 0 {
				m.Cursor--
			}
		case key.Matches(msg, m.keys.Down):
			if m.Cursor < len(m.Profiles)-1 {
				m.Cursor++
			}
		case key.Matches(msg, m.keys.SignOff):
			p := &m.Profiles[m.Cursor]
			p.SignOff = !p.SignOff
		case key.Matches(msg, m.keys.GPGSign):
			p := &m.Profiles[m.Cursor]
			p.GPGSign = !p.GPGSign
		case key.Matches(msg, m.keys.Enter):
			m.Chosen = &m.Profiles[m.Cursor]
			return m, func() tea.Msg {
				err := identity.Set(m.Chosen.Profile)
				return setMsg{err}
			}
		case key.Matches(msg, m.keys.Help):
			m.help.ShowAll = !m.help.ShowAll
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit
		}
	case setMsg:
		if msg.err != nil {
			m.Err = msg.err
		}
		return m, tea.Quit
	}
	return m, nil
}

func (m Model) View() string {
	var view string
	switch {
	case m.Chosen == nil:
		view = profilesView(m)
	case m.Err != nil:
		view = errorView(m)
	default:
		view = successView()
	}
	helpView := m.help.View(m.keys)
	return view + helpView
}

func profilesView(m Model) string {
	var buf bytes.Buffer
	buf.WriteString(termenv.String("Current profile\n\n").Underline().String())

	if m.Current.Name != "" {
		buf.WriteString(fmt.Sprintf("  %s\n", m.Current.Name))
	}
	if m.Current.Email != "" {
		buf.WriteString(fmt.Sprintf("  %s\n", m.Current.Email))
	}
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

	return buf.String()
}

func errorView(m Model) string {
	s := fmt.Sprintf("ERROR: %s\n", m.Err.Error())
	return termenv.String(s).Foreground(termenv.ANSIBrightRed).String()
}

func successView() string {
	return termenv.String("OK").Foreground(termenv.ANSIGreen).String()
}
