package screens

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/guardian-sh/guardian/internal/scaffolding"
	"github.com/guardian-sh/guardian/internal/ui"
)

type QuickStartStep int

const (
	StepSelectStack QuickStartStep = iota
	StepSourceDir
	StepExcludeDirs
	StepConfirm
	StepInstalling
	StepDone
)

type Stack struct {
	Label    string
	Value    string
	Language string
}

var stacks = []Stack{
	{Label: "Python", Value: "python", Language: "python"},
	{Label: "Python + FastAPI", Value: "python-fastapi", Language: "python"},
	{Label: "Python + Django", Value: "python-django", Language: "python"},
	{Label: "TypeScript", Value: "typescript", Language: "typescript"},
	{Label: "TypeScript + React", Value: "typescript-react", Language: "typescript"},
	{Label: "TypeScript + Node", Value: "typescript-node", Language: "typescript"},
	{Label: "Go", Value: "go", Language: "go"},
	{Label: "PHP + Laravel", Value: "php-laravel", Language: "php"},
}

type QuickStartModel struct {
	step          QuickStartStep
	cursor        int
	selectedStack Stack
	sourceDir     textinput.Model
	excludeDirs   textinput.Model
	detectedSrc   string
	detectedExcl  string
	filesToCreate []string
	installedIdx  int
	err           error
}

func NewQuickStart() QuickStartModel {
	srcInput := textinput.New()
	srcInput.Placeholder = "src/"
	srcInput.Focus()
	srcInput.CharLimit = 256
	srcInput.Width = 50

	exclInput := textinput.New()
	exclInput.Placeholder = "tests/, __pycache__/, node_modules/"
	exclInput.CharLimit = 512
	exclInput.Width = 50

	// Try to detect source directory
	detectedSrc := detectSourceDir()
	if detectedSrc != "" {
		srcInput.SetValue(detectedSrc)
	}

	// Detect common exclusions
	detectedExcl := detectExclusions()
	if detectedExcl != "" {
		exclInput.SetValue(detectedExcl)
	}

	return QuickStartModel{
		step:         StepSelectStack,
		cursor:       0,
		sourceDir:    srcInput,
		excludeDirs:  exclInput,
		detectedSrc:  detectedSrc,
		detectedExcl: detectedExcl,
	}
}

func detectSourceDir() string {
	candidates := []string{"src", "app", "lib", "pkg"}
	for _, dir := range candidates {
		if info, err := os.Stat(dir); err == nil && info.IsDir() {
			return dir + "/"
		}
	}
	return ""
}

func detectExclusions() string {
	var found []string
	candidates := []string{"tests", "test", "__pycache__", "node_modules", "migrations", ".venv", "venv", "dist", "build"}
	for _, dir := range candidates {
		if info, err := os.Stat(dir); err == nil && info.IsDir() {
			found = append(found, dir+"/")
		}
	}
	if len(found) > 0 {
		return strings.Join(found, ", ")
	}
	return ""
}

func (m QuickStartModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m QuickStartModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.step {
		case StepSelectStack:
			return m.updateStackSelection(msg)
		case StepSourceDir:
			return m.updateSourceDir(msg)
		case StepExcludeDirs:
			return m.updateExcludeDirs(msg)
		case StepConfirm:
			return m.updateConfirm(msg)
		case StepDone:
			return m.updateDone(msg)
		}

	case installTickMsg:
		if m.installedIdx < len(m.filesToCreate) {
			m.installedIdx++
			if m.installedIdx < len(m.filesToCreate) {
				return m, installTick()
			}
		}
		m.step = StepDone
		return m, nil

	case installCompleteMsg:
		m.step = StepDone
		return m, nil

	case installErrorMsg:
		m.err = msg.err
		m.step = StepDone
		return m, nil
	}

	return m, nil
}

func (m QuickStartModel) updateStackSelection(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Up):
		if m.cursor > 0 {
			m.cursor--
		}
	case key.Matches(msg, keys.Down):
		if m.cursor < len(stacks)-1 {
			m.cursor++
		}
	case key.Matches(msg, keys.Enter):
		m.selectedStack = stacks[m.cursor]
		m.step = StepSourceDir
		return m, nil
	case key.Matches(msg, keys.Back), key.Matches(msg, keys.Quit):
		return m, goBack()
	}
	return m, nil
}

func (m QuickStartModel) updateSourceDir(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Enter):
		m.step = StepExcludeDirs
		m.sourceDir.Blur()
		m.excludeDirs.Focus()
		return m, nil
	case key.Matches(msg, keys.Back):
		m.step = StepSelectStack
		return m, nil
	}

	var cmd tea.Cmd
	m.sourceDir, cmd = m.sourceDir.Update(msg)
	return m, cmd
}

func (m QuickStartModel) updateExcludeDirs(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Enter):
		m.step = StepConfirm
		m.excludeDirs.Blur()
		m.filesToCreate = m.getFilesToCreate()
		return m, nil
	case key.Matches(msg, keys.Back):
		m.step = StepSourceDir
		m.excludeDirs.Blur()
		m.sourceDir.Focus()
		return m, nil
	}

	var cmd tea.Cmd
	m.excludeDirs, cmd = m.excludeDirs.Update(msg)
	return m, cmd
}

func (m QuickStartModel) updateConfirm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y", "enter":
		m.step = StepInstalling
		m.installedIdx = 0
		return m, m.doInstall()
	case "n", "N":
		return m, goBack()
	case "esc", "backspace":
		m.step = StepExcludeDirs
		m.excludeDirs.Focus()
		return m, nil
	}
	return m, nil
}

func (m QuickStartModel) updateDone(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Enter):
		// Go to interactive mode
		return m, switchScreen(ScreenInteractive, InteractiveData{
			Stack:       m.selectedStack,
			SourceDir:   m.sourceDir.Value(),
			ExcludeDirs: m.excludeDirs.Value(),
		})
	case key.Matches(msg, keys.Quit), key.Matches(msg, keys.Back):
		return m, tea.Quit
	}
	return m, nil
}

func (m QuickStartModel) getFilesToCreate() []string {
	files := []string{
		".guardian/check_file_size.py",
		".guardian/check_function_size.py",
		".guardian/check_dangerous.py",
		".guardian/check_mock_data.py",
		".guardian/check_security.py",
		".guardian/guardian.py",
		"guardian_config.toml",
		".pre-commit-config.yaml",
	}

	// Adjust based on language
	switch m.selectedStack.Language {
	case "typescript":
		files = []string{
			".guardian/check_file_size.js",
			".guardian/check_function_size.js",
			".guardian/check_dangerous.js",
			".guardian/check_mock_data.js",
			".guardian/check_security.js",
			".guardian/guardian.js",
			"guardian.config.js",
			".pre-commit-config.yaml",
		}
	case "go":
		files = []string{
			".guardian/check_file_size.go",
			".guardian/check_function_size.go",
			".guardian/check_dangerous.go",
			".guardian/check_mock_data.go",
			".guardian/guardian.go",
			"guardian_config.toml",
			".pre-commit-config.yaml",
		}
	}

	return files
}

func (m QuickStartModel) doInstall() tea.Cmd {
	return func() tea.Msg {
		err := scaffolding.Install(scaffolding.InstallConfig{
			Language:    m.selectedStack.Language,
			Stack:       m.selectedStack.Value,
			SourceDir:   m.sourceDir.Value(),
			ExcludeDirs: strings.Split(m.excludeDirs.Value(), ","),
		})
		if err != nil {
			return installErrorMsg{err: err}
		}
		return installCompleteMsg{}
	}
}

type installTickMsg struct{}
type installCompleteMsg struct{}
type installErrorMsg struct{ err error }

func installTick() tea.Cmd {
	return func() tea.Msg {
		return installTickMsg{}
	}
}

func (m QuickStartModel) View() string {
	var s strings.Builder

	s.WriteString(ui.SmallLogo())
	s.WriteString("\n\n")

	switch m.step {
	case StepSelectStack:
		s.WriteString(m.viewStackSelection())
	case StepSourceDir:
		s.WriteString(m.viewSourceDir())
	case StepExcludeDirs:
		s.WriteString(m.viewExcludeDirs())
	case StepConfirm:
		s.WriteString(m.viewConfirm())
	case StepInstalling:
		s.WriteString(m.viewInstalling())
	case StepDone:
		s.WriteString(m.viewDone())
	}

	return s.String()
}

func (m QuickStartModel) viewStackSelection() string {
	var s strings.Builder

	s.WriteString(ui.TitleStyle.Render("  ● Quick Start"))
	s.WriteString("\n\n")

	s.WriteString(ui.NormalStyle.Render("  ? "))
	s.WriteString(ui.TitleStyle.Render("Select your stack:"))
	s.WriteString("\n\n")

	for i, stack := range stacks {
		if i == m.cursor {
			s.WriteString(ui.CursorStyle.Render("  ❯ ● "))
			s.WriteString(ui.SelectedStyle.Render(stack.Label))
		} else {
			s.WriteString(ui.DimStyle.Render("    ○ "))
			s.WriteString(ui.UnselectedStyle.Render(stack.Label))
		}
		s.WriteString("\n")
	}

	s.WriteString("\n")
	s.WriteString(ui.DimStyle.Render("  ↑/↓ navigate · enter select · esc back"))

	return s.String()
}

func (m QuickStartModel) viewSourceDir() string {
	var s strings.Builder

	s.WriteString(ui.TitleStyle.Render("  ● Quick Start"))
	s.WriteString(ui.DimStyle.Render(" > "))
	s.WriteString(ui.SubtitleStyle.Render(m.selectedStack.Label))
	s.WriteString("\n\n")

	s.WriteString(ui.NormalStyle.Render("  ? "))
	s.WriteString(ui.TitleStyle.Render("Source directory?"))
	s.WriteString("\n")

	if m.detectedSrc != "" {
		s.WriteString(ui.RenderDetected("", m.detectedSrc))
		s.WriteString("\n")
	}

	s.WriteString("\n")
	s.WriteString("  ")
	s.WriteString(ui.CursorStyle.Render("› "))
	s.WriteString(m.sourceDir.View())
	s.WriteString("\n\n")

	s.WriteString(ui.DimStyle.Render("  ↑/↓ navigate · enter confirm · type to override"))

	return s.String()
}

func (m QuickStartModel) viewExcludeDirs() string {
	var s strings.Builder

	s.WriteString(ui.TitleStyle.Render("  ● Quick Start"))
	s.WriteString(ui.DimStyle.Render(" > "))
	s.WriteString(ui.SubtitleStyle.Render(m.selectedStack.Label))
	s.WriteString("\n\n")

	s.WriteString(ui.NormalStyle.Render("  ? "))
	s.WriteString(ui.TitleStyle.Render("Exclude directories?"))
	s.WriteString(ui.DimStyle.Render(" (comma-separated)"))
	s.WriteString("\n")

	if m.detectedExcl != "" {
		s.WriteString(ui.RenderDetected("", m.detectedExcl))
		s.WriteString("\n")
	}

	s.WriteString("\n")
	s.WriteString("  ")
	s.WriteString(ui.CursorStyle.Render("› "))
	s.WriteString(m.excludeDirs.View())
	s.WriteString("\n\n")

	s.WriteString(ui.DimStyle.Render("  enter confirm · esc back"))

	return s.String()
}

func (m QuickStartModel) viewConfirm() string {
	var s strings.Builder

	s.WriteString(ui.TitleStyle.Render("  Ready to install:"))
	s.WriteString("\n\n")

	s.WriteString(ui.DimStyle.Render("    Stack:        "))
	s.WriteString(ui.NormalStyle.Render(m.selectedStack.Label))
	s.WriteString("\n")

	s.WriteString(ui.DimStyle.Render("    Source:       "))
	s.WriteString(ui.NormalStyle.Render(m.sourceDir.Value()))
	s.WriteString("\n")

	s.WriteString(ui.DimStyle.Render("    Exclude:      "))
	s.WriteString(ui.NormalStyle.Render(m.excludeDirs.Value()))
	s.WriteString("\n\n")

	s.WriteString(ui.TitleStyle.Render("  Will create:"))
	s.WriteString("\n\n")

	for _, file := range m.filesToCreate {
		s.WriteString(ui.DimStyle.Render("    "))
		s.WriteString(ui.FilePathStyle.Render(file))
		s.WriteString("\n")
	}

	s.WriteString("\n")
	s.WriteString(ui.RenderConfirmation("Continue?", true))
	s.WriteString("\n\n")

	s.WriteString(ui.DimStyle.Render("  y/enter continue · n/esc back"))

	return s.String()
}

func (m QuickStartModel) viewInstalling() string {
	var s strings.Builder

	s.WriteString(ui.TitleStyle.Render("  Installing..."))
	s.WriteString("\n\n")

	for i, file := range m.filesToCreate {
		if i < m.installedIdx {
			s.WriteString(ui.Success("Created " + file))
		} else if i == m.installedIdx {
			s.WriteString(ui.HighlightStyle.Render("  ├─ " + file))
		} else {
			s.WriteString(ui.DimStyle.Render("  · " + file))
		}
		s.WriteString("\n")
	}

	return s.String()
}

func (m QuickStartModel) viewDone() string {
	var s strings.Builder

	if m.err != nil {
		s.WriteString(ui.Error("Installation failed"))
		s.WriteString("\n\n")
		s.WriteString(ui.ErrorStyle.Render("  " + m.err.Error()))
		s.WriteString("\n\n")
		s.WriteString(ui.DimStyle.Render("  enter to try again · q to quit"))
		return s.String()
	}

	for _, file := range m.filesToCreate {
		s.WriteString(ui.Success("Created " + file))
		s.WriteString("\n")
	}

	s.WriteString("\n")
	s.WriteString(ui.Divider())
	s.WriteString("\n\n")

	s.WriteString(ui.SuccessStyle.Render("  Done. Guardian is active."))
	s.WriteString("\n\n")

	s.WriteString(ui.TitleStyle.Render("  What now?"))
	s.WriteString(ui.DimStyle.Render(" (type a command or question)"))
	s.WriteString("\n\n")

	s.WriteString(ui.HighlightStyle.Render("  /run"))
	s.WriteString(ui.DimStyle.Render("          Check your code now"))
	s.WriteString("\n")

	s.WriteString(ui.HighlightStyle.Render("  /dry-run"))
	s.WriteString(ui.DimStyle.Render("      Preview what would be checked"))
	s.WriteString("\n")

	s.WriteString(ui.HighlightStyle.Render("  /help"))
	s.WriteString(ui.DimStyle.Render("         Explain something"))
	s.WriteString("\n")

	s.WriteString(ui.HighlightStyle.Render("  /prompt"))
	s.WriteString(ui.DimStyle.Render("       Generate a prompt for Claude"))
	s.WriteString("\n")

	s.WriteString(ui.HighlightStyle.Render("  /exit"))
	s.WriteString(ui.DimStyle.Render("         Leave Guardian"))
	s.WriteString("\n\n")

	s.WriteString(ui.DimStyle.Render("  Press enter to continue..."))

	return s.String()
}

// InteractiveData passed to interactive screen
type InteractiveData struct {
	Stack       Stack
	SourceDir   string
	ExcludeDirs string
}

// Utility function for relative paths
func relPath(path string) string {
	cwd, err := os.Getwd()
	if err != nil {
		return path
	}
	rel, err := filepath.Rel(cwd, path)
	if err != nil {
		return path
	}
	return rel
}
