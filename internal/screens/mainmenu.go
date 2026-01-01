package screens

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/guardian-sh/guardian/internal/ui"
)

const Version = "0.1.0"

// Screen represents which screen is currently active
type Screen int

const (
	ScreenMainMenu Screen = iota
	ScreenQuickStart
	ScreenAISetup
	ScreenAbout
	ScreenInteractive
)

// MainMenuModel is the main menu
type MainMenuModel struct {
	cursor int
	items  []ui.MenuItem
}

func NewMainMenu() MainMenuModel {
	return MainMenuModel{
		cursor: 0,
		items: []ui.MenuItem{
			{Label: "Quick Start", Description: "Add guards to this project", Value: "quickstart"},
			{Label: "AI Setup", Description: "Smart config with your API key", Value: "ai"},
			{Label: "About", Description: "What Guardian catches", Value: "about"},
		},
	}
}

func (m MainMenuModel) Init() tea.Cmd {
	return nil
}

func (m MainMenuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Up):
			if m.cursor > 0 {
				m.cursor--
			}
		case key.Matches(msg, keys.Down):
			if m.cursor < len(m.items)-1 {
				m.cursor++
			}
		case key.Matches(msg, keys.Enter):
			return m, selectMenuItem(m.items[m.cursor].Value)
		case key.Matches(msg, keys.Quit):
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m MainMenuModel) View() string {
	var s string

	// Logo box
	s += ui.LogoBox()
	s += "\n\n"

	// Version line
	s += "  " + ui.VersionLine(Version)
	s += "\n\n"

	// Question
	s += ui.NormalStyle.Render("  ? ")
	s += ui.TitleStyle.Render("What would you like to do?")
	s += "\n\n"

	// Menu
	s += ui.RenderMenu(m.items, m.cursor)
	s += "\n"

	// Help
	s += ui.DimStyle.Render("  ↑/↓ navigate · enter select · q quit")

	return s
}

// Key bindings
type keyMap struct {
	Up    key.Binding
	Down  key.Binding
	Enter key.Binding
	Quit  key.Binding
	Back  key.Binding
	Tab   key.Binding
	Help  key.Binding
}

var keys = keyMap{
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
	),
	Back: key.NewBinding(
		key.WithKeys("esc", "backspace"),
	),
	Tab: key.NewBinding(
		key.WithKeys("tab"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
	),
}

// Messages
type MenuSelectMsg struct {
	Value string
}

func selectMenuItem(value string) tea.Cmd {
	return func() tea.Msg {
		return MenuSelectMsg{Value: value}
	}
}

type GoBackMsg struct{}

func goBack() tea.Cmd {
	return func() tea.Msg {
		return GoBackMsg{}
	}
}

type SwitchScreenMsg struct {
	Screen Screen
	Data   interface{}
}

func switchScreen(screen Screen, data interface{}) tea.Cmd {
	return func() tea.Msg {
		return SwitchScreenMsg{Screen: screen, Data: data}
	}
}
