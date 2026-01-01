package screens

import (
	tea "github.com/charmbracelet/bubbletea"
)

// AppModel is the root model that manages screen transitions
type AppModel struct {
	currentScreen Screen
	mainMenu      MainMenuModel
	quickStart    QuickStartModel
	aiSetup       AISetupModel
	about         AboutModel
	interactive   InteractiveModel
}

// NewApp creates a new application
func NewApp() AppModel {
	return AppModel{
		currentScreen: ScreenMainMenu,
		mainMenu:      NewMainMenu(),
	}
}

func (m AppModel) Init() tea.Cmd {
	return m.mainMenu.Init()
}

func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Global quit
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}

	case MenuSelectMsg:
		switch msg.Value {
		case "quickstart":
			m.currentScreen = ScreenQuickStart
			m.quickStart = NewQuickStart()
			return m, m.quickStart.Init()
		case "ai":
			m.currentScreen = ScreenAISetup
			m.aiSetup = NewAISetup()
			return m, m.aiSetup.Init()
		case "about":
			m.currentScreen = ScreenAbout
			m.about = NewAbout()
			return m, m.about.Init()
		}

	case GoBackMsg:
		m.currentScreen = ScreenMainMenu
		m.mainMenu = NewMainMenu()
		return m, m.mainMenu.Init()

	case SwitchScreenMsg:
		switch msg.Screen {
		case ScreenInteractive:
			m.currentScreen = ScreenInteractive
			m.interactive = NewInteractive(msg.Data)
			return m, m.interactive.Init()
		case ScreenMainMenu:
			m.currentScreen = ScreenMainMenu
			m.mainMenu = NewMainMenu()
			return m, m.mainMenu.Init()
		}
	}

	// Route to current screen
	var cmd tea.Cmd
	switch m.currentScreen {
	case ScreenMainMenu:
		var newModel tea.Model
		newModel, cmd = m.mainMenu.Update(msg)
		m.mainMenu = newModel.(MainMenuModel)
	case ScreenQuickStart:
		var newModel tea.Model
		newModel, cmd = m.quickStart.Update(msg)
		m.quickStart = newModel.(QuickStartModel)
	case ScreenAISetup:
		var newModel tea.Model
		newModel, cmd = m.aiSetup.Update(msg)
		m.aiSetup = newModel.(AISetupModel)
	case ScreenAbout:
		var newModel tea.Model
		newModel, cmd = m.about.Update(msg)
		m.about = newModel.(AboutModel)
	case ScreenInteractive:
		var newModel tea.Model
		newModel, cmd = m.interactive.Update(msg)
		m.interactive = newModel.(InteractiveModel)
	}

	return m, cmd
}

func (m AppModel) View() string {
	switch m.currentScreen {
	case ScreenMainMenu:
		return m.mainMenu.View()
	case ScreenQuickStart:
		return m.quickStart.View()
	case ScreenAISetup:
		return m.aiSetup.View()
	case ScreenAbout:
		return m.about.View()
	case ScreenInteractive:
		return m.interactive.View()
	default:
		return m.mainMenu.View()
	}
}
