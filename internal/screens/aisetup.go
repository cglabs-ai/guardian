package screens

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/guardian-sh/guardian/internal/ai"
	"github.com/guardian-sh/guardian/internal/ui"
)

type AISetupStep int

const (
	AIStepKey AISetupStep = iota
	AIStepValidating
	AIStepMenu
	AIStepScanning
	AIStepResults
)

type AISetupModel struct {
	step        AISetupStep
	keyInput    textinput.Model
	cursor      int
	scanResults *ai.ScanResults
	err         error
	validKey    bool
}

func NewAISetup() AISetupModel {
	keyInput := textinput.New()
	keyInput.Placeholder = "AIza..."
	keyInput.Focus()
	keyInput.CharLimit = 64
	keyInput.Width = 50
	keyInput.EchoMode = textinput.EchoPassword
	keyInput.EchoCharacter = '*'

	// Check if key already exists
	existingKey := loadExistingKey()
	if existingKey != "" {
		keyInput.SetValue(existingKey)
	}

	return AISetupModel{
		step:     AIStepKey,
		keyInput: keyInput,
	}
}

func loadExistingKey() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	keyPath := filepath.Join(homeDir, ".guardian", "credentials")
	data, err := os.ReadFile(keyPath)
	if err != nil {
		return ""
	}

	return strings.TrimSpace(string(data))
}

func saveKey(key string) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	guardianDir := filepath.Join(homeDir, ".guardian")
	if err := os.MkdirAll(guardianDir, 0700); err != nil {
		return err
	}

	keyPath := filepath.Join(guardianDir, "credentials")
	return os.WriteFile(keyPath, []byte(key), 0600)
}

func (m AISetupModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m AISetupModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.step {
		case AIStepKey:
			return m.updateKey(msg)
		case AIStepMenu:
			return m.updateMenu(msg)
		case AIStepResults:
			return m.updateResults(msg)
		}

	case keyValidMsg:
		m.validKey = msg.valid
		m.err = msg.err
		if msg.valid {
			saveKey(m.keyInput.Value())
			m.step = AIStepMenu
		} else {
			m.step = AIStepKey
		}
		return m, nil

	case scanCompleteMsg:
		m.scanResults = msg.results
		m.err = msg.err
		m.step = AIStepResults
		return m, nil
	}

	return m, nil
}

func (m AISetupModel) updateKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Enter):
		if m.keyInput.Value() != "" {
			m.step = AIStepValidating
			return m, validateKey(m.keyInput.Value())
		}
	case key.Matches(msg, keys.Back), key.Matches(msg, keys.Quit):
		return m, goBack()
	}

	var cmd tea.Cmd
	m.keyInput, cmd = m.keyInput.Update(msg)
	return m, cmd
}

func (m AISetupModel) updateMenu(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Up):
		if m.cursor > 0 {
			m.cursor--
		}
	case key.Matches(msg, keys.Down):
		if m.cursor < 1 {
			m.cursor++
		}
	case key.Matches(msg, keys.Enter):
		if m.cursor == 0 {
			m.step = AIStepScanning
			return m, doScan(m.keyInput.Value())
		} else {
			return m, goBack()
		}
	case key.Matches(msg, keys.Back), key.Matches(msg, keys.Quit):
		return m, goBack()
	}
	return m, nil
}

func (m AISetupModel) updateResults(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y", "enter":
		// Apply configuration and go to interactive mode
		return m, switchScreen(ScreenInteractive, nil)
	case "n", "N", "esc":
		m.step = AIStepMenu
		return m, nil
	case "q":
		return m, tea.Quit
	}
	return m, nil
}

func (m AISetupModel) View() string {
	var s strings.Builder

	s.WriteString(ui.SmallLogo())
	s.WriteString("\n\n")

	switch m.step {
	case AIStepKey:
		s.WriteString(m.viewKey())
	case AIStepValidating:
		s.WriteString(m.viewValidating())
	case AIStepMenu:
		s.WriteString(m.viewMenu())
	case AIStepScanning:
		s.WriteString(m.viewScanning())
	case AIStepResults:
		s.WriteString(m.viewResults())
	}

	return s.String()
}

func (m AISetupModel) viewKey() string {
	var s strings.Builder

	s.WriteString(ui.TitleStyle.Render("  ● AI Setup"))
	s.WriteString("\n\n")

	s.WriteString(ui.NormalStyle.Render("  Guardian's AI features use Gemini Flash for smart configuration."))
	s.WriteString("\n")
	s.WriteString(ui.NormalStyle.Render("  You provide your own key. Typical cost: <$0.01 per project."))
	s.WriteString("\n\n")

	s.WriteString(ui.NormalStyle.Render("  ? "))
	s.WriteString(ui.TitleStyle.Render("Gemini API Key:"))
	s.WriteString("\n\n")

	s.WriteString("  ")
	s.WriteString(ui.CursorStyle.Render("› "))
	s.WriteString(m.keyInput.View())
	s.WriteString("\n\n")

	if m.err != nil {
		s.WriteString(ui.Error(m.err.Error()))
		s.WriteString("\n\n")
	}

	s.WriteString(ui.DimStyle.Render("    Get a key: "))
	s.WriteString(ui.SubtitleStyle.Render("ai.google.dev/gemini-api"))
	s.WriteString("\n\n")

	s.WriteString(ui.DimStyle.Render("    ⚠ Key stored in ~/.guardian/credentials (plaintext)"))
	s.WriteString("\n\n")

	s.WriteString(ui.DimStyle.Render("  enter continue · esc back"))

	return s.String()
}

func (m AISetupModel) viewValidating() string {
	var s strings.Builder

	s.WriteString(ui.TitleStyle.Render("  ● AI Setup"))
	s.WriteString("\n\n")

	s.WriteString(ui.HighlightStyle.Render("  Validating key..."))

	return s.String()
}

func (m AISetupModel) viewMenu() string {
	var s strings.Builder

	s.WriteString(ui.TitleStyle.Render("  ● AI Setup"))
	s.WriteString("\n\n")

	s.WriteString(ui.Success("Key valid. Saved to ~/.guardian/credentials"))
	s.WriteString("\n")
	s.WriteString(ui.DimStyle.Render("  ⚠ Stored in plaintext. Use a restricted API key."))
	s.WriteString("\n\n")

	s.WriteString(ui.NormalStyle.Render("  AI features enabled:"))
	s.WriteString("\n\n")

	items := []string{"Smart Scan      Analyze this project, generate custom config", "Back            Return to main menu"}
	for i, item := range items {
		if i == m.cursor {
			s.WriteString(ui.CursorStyle.Render("  ❯ ● "))
			parts := strings.SplitN(item, "  ", 2)
			s.WriteString(ui.SelectedStyle.Render(parts[0]))
			if len(parts) > 1 {
				s.WriteString(ui.DimStyle.Render("  " + strings.TrimSpace(parts[1])))
			}
		} else {
			s.WriteString(ui.DimStyle.Render("    ○ "))
			s.WriteString(ui.UnselectedStyle.Render(item))
		}
		s.WriteString("\n")
	}

	s.WriteString("\n")
	s.WriteString(ui.DimStyle.Render("  ↑/↓ navigate · enter select · esc back"))

	return s.String()
}

func (m AISetupModel) viewScanning() string {
	var s strings.Builder

	s.WriteString(ui.TitleStyle.Render("  Scanning project..."))
	s.WriteString("\n\n")

	steps := []string{
		"Reading file structure",
		"Detecting frameworks",
		"Finding code patterns",
		"Checking for conflicts",
	}

	for _, step := range steps {
		s.WriteString(ui.HighlightStyle.Render("  ├─ " + step))
		s.WriteString("\n")
	}

	return s.String()
}

func (m AISetupModel) viewResults() string {
	var s strings.Builder

	headerBox := ui.HeaderBox.Render("Smart Scan Results")
	s.WriteString(headerBox)
	s.WriteString("\n\n")

	if m.err != nil {
		s.WriteString(ui.Error("Scan failed: " + m.err.Error()))
		s.WriteString("\n\n")
		s.WriteString(ui.DimStyle.Render("  enter to try again · esc back"))
		return s.String()
	}

	if m.scanResults == nil {
		s.WriteString(ui.DimStyle.Render("  No results available"))
		return s.String()
	}

	// Detected section
	s.WriteString(ui.TitleStyle.Render("  Detected:"))
	s.WriteString("\n")
	s.WriteString(ui.DimStyle.Render("    Language:     "))
	s.WriteString(ui.NormalStyle.Render(m.scanResults.Language))
	s.WriteString("\n")
	s.WriteString(ui.DimStyle.Render("    Framework:    "))
	s.WriteString(ui.NormalStyle.Render(m.scanResults.Framework))
	s.WriteString("\n")
	s.WriteString(ui.DimStyle.Render("    Source:       "))
	s.WriteString(ui.NormalStyle.Render(m.scanResults.SourceDir))
	s.WriteString("\n")
	s.WriteString(ui.DimStyle.Render("    Tests:        "))
	s.WriteString(ui.NormalStyle.Render(m.scanResults.TestDir))
	s.WriteString("\n\n")

	// Found patterns
	if len(m.scanResults.MockPatterns) > 0 {
		s.WriteString(ui.TitleStyle.Render("  Found patterns in your code:"))
		s.WriteString("\n")
		s.WriteString(ui.DimStyle.Render("    Mock data:    "))
		s.WriteString(ui.NormalStyle.Render(strings.Join(m.scanResults.MockPatterns, ", ")))
		s.WriteString("\n")
	}

	// Secrets found
	if len(m.scanResults.SecretsFound) > 0 {
		s.WriteString(ui.WarningStyle.Render("    Secrets:      Found " + strconv.Itoa(len(m.scanResults.SecretsFound)) + " possible exposed keys"))
		s.WriteString("\n")
	}
	s.WriteString("\n")

	// Recommendations
	if len(m.scanResults.Recommendations) > 0 {
		s.WriteString(ui.TitleStyle.Render("  Recommendations:"))
		s.WriteString("\n")
		for _, rec := range m.scanResults.Recommendations {
			s.WriteString(ui.Success(rec))
			s.WriteString("\n")
		}
		s.WriteString("\n")
	}

	// Conflicts
	s.WriteString(ui.TitleStyle.Render("  Conflicts:"))
	s.WriteString("\n")
	if len(m.scanResults.Conflicts) == 0 {
		s.WriteString(ui.Success("None - safe to install"))
	} else {
		for _, c := range m.scanResults.Conflicts {
			s.WriteString(ui.Warning(c))
			s.WriteString("\n")
		}
	}
	s.WriteString("\n\n")

	s.WriteString(ui.RenderConfirmation("Apply this configuration?", true))
	s.WriteString("\n\n")
	s.WriteString(ui.DimStyle.Render("  y/enter apply · n/esc back"))

	return s.String()
}

// Messages
type keyValidMsg struct {
	valid bool
	err   error
}

func validateKey(apiKey string) tea.Cmd {
	return func() tea.Msg {
		// Simulate validation delay
		time.Sleep(500 * time.Millisecond)

		valid, err := ai.ValidateKey(apiKey)
		return keyValidMsg{valid: valid, err: err}
	}
}

type scanCompleteMsg struct {
	results *ai.ScanResults
	err     error
}

func doScan(apiKey string) tea.Cmd {
	return func() tea.Msg {
		results, err := ai.ScanProject(apiKey, ".")
		return scanCompleteMsg{results: results, err: err}
	}
}
