package screens

import (
	"fmt"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/guardian-sh/guardian/internal/checks"
	"github.com/guardian-sh/guardian/internal/prompts"
	"github.com/guardian-sh/guardian/internal/ui"
)

type InteractiveMode int

const (
	ModeCommand InteractiveMode = iota
	ModeResults
	ModeHelp
	ModePrompt
	ModePromptResult
	ModeExplain
	ModeDryRun
)

type InteractiveModel struct {
	mode         InteractiveMode
	input        textinput.Model
	cursor       int
	issues       []checks.Issue
	helpCursor   int
	promptCursor int
	promptText   string
	explainIdx   int
	dryRunInfo   *checks.DryRunInfo
	data         interface{}
}

func NewInteractive(data interface{}) InteractiveModel {
	input := textinput.New()
	input.Placeholder = "type a command..."
	input.Focus()
	input.CharLimit = 256
	input.Width = 60

	return InteractiveModel{
		mode:  ModeCommand,
		input: input,
		data:  data,
	}
}

func (m InteractiveModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m InteractiveModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.mode {
		case ModeCommand:
			return m.updateCommand(msg)
		case ModeResults:
			return m.updateResults(msg)
		case ModeHelp:
			return m.updateHelp(msg)
		case ModePrompt:
			return m.updatePrompt(msg)
		case ModePromptResult:
			return m.updatePromptResult(msg)
		case ModeExplain:
			return m.updateExplain(msg)
		case ModeDryRun:
			return m.updateDryRun(msg)
		}

	case checksCompleteMsg:
		m.issues = msg.issues
		m.mode = ModeResults
		return m, nil

	case dryRunCompleteMsg:
		m.dryRunInfo = msg.info
		m.mode = ModeDryRun
		return m, nil

	case promptGeneratedMsg:
		m.promptText = msg.prompt
		m.mode = ModePromptResult
		clipboard.WriteAll(msg.prompt)
		return m, nil
	}

	return m, nil
}

func (m InteractiveModel) updateCommand(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Enter):
		cmd := strings.TrimSpace(m.input.Value())
		m.input.SetValue("")
		return m.handleCommand(cmd)
	case key.Matches(msg, keys.Quit):
		return m, tea.Quit
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m InteractiveModel) handleCommand(cmd string) (tea.Model, tea.Cmd) {
	switch strings.ToLower(cmd) {
	case "/run", "run":
		return m, runChecks()
	case "/dry-run", "dry-run", "/dryrun", "dryrun":
		return m, runDryRun()
	case "/help", "help", "?":
		m.mode = ModeHelp
		m.helpCursor = 0
		return m, nil
	case "/prompt", "prompt":
		m.mode = ModePrompt
		m.promptCursor = 0
		return m, nil
	case "/config", "config":
		return m, openConfig()
	case "/exit", "exit", "quit", "q":
		return m, tea.Quit
	default:
		// Check if it's /explain N
		if strings.HasPrefix(cmd, "/explain ") || strings.HasPrefix(cmd, "explain ") {
			parts := strings.Fields(cmd)
			if len(parts) >= 2 {
				var idx int
				fmt.Sscanf(parts[1], "%d", &idx)
				if idx > 0 && idx <= len(m.issues) {
					m.explainIdx = idx - 1
					m.mode = ModeExplain
					return m, nil
				}
			}
		}
		// Check if it's /prompt followed by something
		if strings.HasPrefix(cmd, "/prompt ") {
			m.mode = ModePrompt
			m.promptCursor = 0
			return m, nil
		}
	}
	return m, nil
}

func (m InteractiveModel) updateResults(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Back):
		m.mode = ModeCommand
		return m, nil
	case key.Matches(msg, keys.Quit):
		return m, tea.Quit
	}

	cmd := strings.ToLower(msg.String())
	switch cmd {
	case "p":
		m.mode = ModePrompt
		m.promptCursor = 0
	case "e":
		if len(m.issues) > 0 {
			m.explainIdx = 0
			m.mode = ModeExplain
		}
	}

	return m, nil
}

func (m InteractiveModel) updateHelp(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	helpItems := []string{
		"What is pre-commit?",
		"What does Guardian check for?",
		"How do I fix an issue?",
		"How do I turn off a rule?",
	}

	switch {
	case key.Matches(msg, keys.Up):
		if m.helpCursor > 0 {
			m.helpCursor--
		}
	case key.Matches(msg, keys.Down):
		if m.helpCursor < len(helpItems)-1 {
			m.helpCursor++
		}
	case key.Matches(msg, keys.Enter):
		// Generate prompt for selected help topic
		return m, generateHelpPrompt(helpItems[m.helpCursor])
	case key.Matches(msg, keys.Back), key.Matches(msg, keys.Quit):
		m.mode = ModeCommand
		return m, nil
	}

	return m, nil
}

func (m InteractiveModel) updatePrompt(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	promptItems := []string{
		"I have issues and don't know how to fix them",
		"I need to set up pre-commit but don't know how",
		"I don't understand what Guardian is telling me",
		"I want to change the rules but don't know how",
		"Something else",
	}

	switch {
	case key.Matches(msg, keys.Up):
		if m.promptCursor > 0 {
			m.promptCursor--
		}
	case key.Matches(msg, keys.Down):
		if m.promptCursor < len(promptItems)-1 {
			m.promptCursor++
		}
	case key.Matches(msg, keys.Enter):
		return m, generatePrompt(promptItems[m.promptCursor], m.issues)
	case key.Matches(msg, keys.Back), key.Matches(msg, keys.Quit):
		m.mode = ModeCommand
		return m, nil
	}

	return m, nil
}

func (m InteractiveModel) updatePromptResult(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Back), key.Matches(msg, keys.Quit), key.Matches(msg, keys.Enter):
		m.mode = ModeCommand
		return m, nil
	}
	return m, nil
}

func (m InteractiveModel) updateExplain(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Back), key.Matches(msg, keys.Quit), key.Matches(msg, keys.Enter):
		m.mode = ModeResults
		return m, nil
	case msg.String() == "p":
		// Generate prompt for this specific issue
		if m.explainIdx < len(m.issues) {
			return m, generatePromptForIssue(m.issues[m.explainIdx])
		}
	}
	return m, nil
}

func (m InteractiveModel) updateDryRun(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Back), key.Matches(msg, keys.Quit), key.Matches(msg, keys.Enter):
		m.mode = ModeCommand
		return m, nil
	}
	return m, nil
}

func (m InteractiveModel) View() string {
	var s strings.Builder

	switch m.mode {
	case ModeCommand:
		s.WriteString(m.viewCommand())
	case ModeResults:
		s.WriteString(m.viewResults())
	case ModeHelp:
		s.WriteString(m.viewHelp())
	case ModePrompt:
		s.WriteString(m.viewPrompt())
	case ModePromptResult:
		s.WriteString(m.viewPromptResult())
	case ModeExplain:
		s.WriteString(m.viewExplain())
	case ModeDryRun:
		s.WriteString(m.viewDryRun())
	}

	return s.String()
}

func (m InteractiveModel) viewCommand() string {
	var s strings.Builder

	s.WriteString(ui.SmallLogo())
	s.WriteString("\n\n")

	s.WriteString(ui.TitleStyle.Render("  What now?"))
	s.WriteString(ui.DimStyle.Render(" (type a command or question)"))
	s.WriteString("\n\n")

	commands := []struct {
		cmd  string
		desc string
	}{
		{"/run", "Check your code now"},
		{"/dry-run", "Preview what would be checked"},
		{"/help", "Explain something"},
		{"/prompt", "Generate a prompt for Claude"},
		{"/exit", "Leave Guardian"},
	}

	for _, c := range commands {
		s.WriteString(ui.HighlightStyle.Render("  " + padRight(c.cmd, 14)))
		s.WriteString(ui.DimStyle.Render(c.desc))
		s.WriteString("\n")
	}

	s.WriteString("\n")
	s.WriteString(ui.CursorStyle.Render("  › "))
	s.WriteString(m.input.View())
	s.WriteString("\n")

	return s.String()
}

func (m InteractiveModel) viewResults() string {
	var s strings.Builder

	if len(m.issues) == 0 {
		headerBox := ui.HeaderBox.Render(ui.TitleStyle.Render("GUARDIAN") + ui.DimStyle.Render(" · ") + ui.SuccessStyle.Render("No issues found"))
		s.WriteString(headerBox)
		s.WriteString("\n\n")
		s.WriteString(ui.SuccessStyle.Render("  ✓ Your code looks clean!"))
		s.WriteString("\n\n")
		s.WriteString(ui.DimStyle.Render("  Press any key to continue..."))
		return s.String()
	}

	// Count by severity
	critical, warnings, info := 0, 0, 0
	for _, issue := range m.issues {
		switch issue.Severity {
		case "critical":
			critical++
		case "warning":
			warnings++
		default:
			info++
		}
	}

	// Group issues by file
	fileIssues := make(map[string][]checks.Issue)
	for _, issue := range m.issues {
		fileIssues[issue.File] = append(fileIssues[issue.File], issue)
	}

	header := fmt.Sprintf("%s · %d issues in %d files", ui.TitleStyle.Render("GUARDIAN"), len(m.issues), len(fileIssues))
	headerBox := ui.HeaderBox.Render(header)
	s.WriteString(headerBox)
	s.WriteString("\n")

	for file, issues := range fileIssues {
		s.WriteString("\n")
		s.WriteString(ui.FilePathStyle.Render("  " + file))
		s.WriteString("\n")

		for _, issue := range issues {
			s.WriteString(ui.LineNumStyle.Render(fmt.Sprintf("    :%d", issue.Line)))
			s.WriteString("   ")

			switch issue.Severity {
			case "critical":
				s.WriteString(ui.CriticalStyle.Render(fmt.Sprintf("[%s]", issue.Rule)))
			case "warning":
				s.WriteString(ui.WarningIssueStyle.Render(fmt.Sprintf("[%s]", issue.Rule)))
			default:
				s.WriteString(ui.InfoIssueStyle.Render(fmt.Sprintf("[%s]", issue.Rule)))
			}

			s.WriteString("  ")
			s.WriteString(ui.NormalStyle.Render(issue.Message))
			s.WriteString("\n")
		}
	}

	s.WriteString("\n")
	s.WriteString(ui.Divider())
	s.WriteString("\n")

	// Summary line
	summaryParts := []string{}
	if critical > 0 {
		summaryParts = append(summaryParts, ui.CriticalStyle.Render(fmt.Sprintf("%d critical", critical)))
	}
	if warnings > 0 {
		summaryParts = append(summaryParts, ui.WarningIssueStyle.Render(fmt.Sprintf("%d warnings", warnings)))
	}
	if info > 0 {
		summaryParts = append(summaryParts, ui.InfoIssueStyle.Render(fmt.Sprintf("%d info", info)))
	}
	s.WriteString("  ")
	s.WriteString(strings.Join(summaryParts, ui.DimStyle.Render(" · ")))
	s.WriteString("\n\n")

	s.WriteString(ui.HighlightStyle.Render("  /prompt"))
	s.WriteString(ui.DimStyle.Render("     Get a Claude prompt to fix these"))
	s.WriteString("\n")
	s.WriteString(ui.HighlightStyle.Render("  /explain N"))
	s.WriteString(ui.DimStyle.Render("  Explain issue N in detail"))
	s.WriteString("\n\n")

	s.WriteString(ui.DimStyle.Render("  p prompt · e explain · esc back"))

	return s.String()
}

func (m InteractiveModel) viewHelp() string {
	var s strings.Builder

	s.WriteString(ui.SmallLogo())
	s.WriteString("\n\n")

	s.WriteString(ui.TitleStyle.Render("  /help"))
	s.WriteString("\n\n")

	s.WriteString(ui.NormalStyle.Render("  What do you need help with?"))
	s.WriteString("\n\n")

	helpItems := []string{
		"What is pre-commit?",
		"What does Guardian check for?",
		"How do I fix an issue?",
		"How do I turn off a rule?",
	}

	for i, item := range helpItems {
		if i == m.helpCursor {
			s.WriteString(ui.CursorStyle.Render("  ❯ ● "))
			s.WriteString(ui.SelectedStyle.Render(item))
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

func (m InteractiveModel) viewPrompt() string {
	var s strings.Builder

	s.WriteString(ui.SmallLogo())
	s.WriteString("\n\n")

	s.WriteString(ui.TitleStyle.Render("  /prompt"))
	s.WriteString("\n\n")

	s.WriteString(ui.NormalStyle.Render("  What do you need Claude to help with?"))
	s.WriteString("\n\n")

	promptItems := []string{
		"I have issues and don't know how to fix them",
		"I need to set up pre-commit but don't know how",
		"I don't understand what Guardian is telling me",
		"I want to change the rules but don't know how",
		"Something else",
	}

	for i, item := range promptItems {
		if i == m.promptCursor {
			s.WriteString(ui.CursorStyle.Render("  ❯ ● "))
			s.WriteString(ui.SelectedStyle.Render(item))
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

func (m InteractiveModel) viewPromptResult() string {
	var s strings.Builder

	promptHeader := ui.PromptHeaderStyle.Render("COPY THIS INTO CLAUDE")
	promptBox := ui.PromptBoxStyle.Render(m.promptText)

	s.WriteString(promptHeader)
	s.WriteString("\n")
	s.WriteString(promptBox)
	s.WriteString("\n\n")

	s.WriteString(ui.Success("Copied to clipboard"))
	s.WriteString("\n\n")

	s.WriteString(ui.NormalStyle.Render("  Now paste this into Claude Code."))
	s.WriteString("\n\n")

	s.WriteString(ui.DimStyle.Render("  Press any key to continue..."))

	return s.String()
}

func (m InteractiveModel) viewExplain() string {
	var s strings.Builder

	if m.explainIdx >= len(m.issues) {
		s.WriteString(ui.DimStyle.Render("  No issue to explain"))
		return s.String()
	}

	issue := m.issues[m.explainIdx]

	s.WriteString(ui.TitleStyle.Render(fmt.Sprintf("  Issue: %s", issue.Rule)))
	s.WriteString("\n")
	s.WriteString(ui.FilePathStyle.Render(fmt.Sprintf("  Location: %s:%d", issue.File, issue.Line)))
	s.WriteString("\n\n")

	s.WriteString(ui.Divider())
	s.WriteString("\n\n")

	// Get explanation
	explanation := prompts.GetExplanation(issue.Rule)

	s.WriteString(ui.TitleStyle.Render("  What's wrong:"))
	s.WriteString("\n\n")
	s.WriteString(ui.NormalStyle.Render("    " + explanation.Problem))
	s.WriteString("\n\n")

	s.WriteString(ui.TitleStyle.Render("  Why it's dangerous:"))
	s.WriteString("\n\n")
	s.WriteString(ui.NormalStyle.Render("    " + explanation.Why))
	s.WriteString("\n\n")

	s.WriteString(ui.TitleStyle.Render("  How to fix:"))
	s.WriteString("\n\n")
	s.WriteString(ui.NormalStyle.Render("    " + explanation.Fix))
	s.WriteString("\n\n")

	s.WriteString(ui.Divider())
	s.WriteString("\n\n")

	s.WriteString(ui.HighlightStyle.Render("  /prompt fix"))
	s.WriteString(ui.DimStyle.Render("    Get a Claude prompt to fix this"))
	s.WriteString("\n\n")

	s.WriteString(ui.DimStyle.Render("  p prompt · esc back"))

	return s.String()
}

func (m InteractiveModel) viewDryRun() string {
	var s strings.Builder

	s.WriteString(ui.SmallLogo())
	s.WriteString("\n\n")

	s.WriteString(ui.TitleStyle.Render("  /dry-run"))
	s.WriteString("\n\n")

	if m.dryRunInfo == nil {
		s.WriteString(ui.DimStyle.Render("  No information available"))
		return s.String()
	}

	s.WriteString(ui.NormalStyle.Render("  Would check:"))
	s.WriteString("\n\n")

	for _, file := range m.dryRunInfo.Files {
		line := fmt.Sprintf("    %s (%d lines)", file.Path, file.Lines)
		if file.Lines > 500 {
			s.WriteString(ui.FilePathStyle.Render(line))
			s.WriteString(ui.WarningStyle.Render(" ⚠ large"))
		} else {
			s.WriteString(ui.FilePathStyle.Render(line))
		}
		s.WriteString("\n")
	}

	s.WriteString("\n")
	s.WriteString(ui.NormalStyle.Render("  Would skip:"))
	s.WriteString("\n")
	s.WriteString(ui.DimStyle.Render("    " + strings.Join(m.dryRunInfo.Excluded, ", ")))
	s.WriteString("\n\n")

	s.WriteString(ui.NormalStyle.Render(fmt.Sprintf("  %d files · ~%d lines · <1 second", m.dryRunInfo.FileCount, m.dryRunInfo.TotalLines)))
	s.WriteString("\n\n")

	s.WriteString(ui.DimStyle.Render("  Press any key to continue..."))

	return s.String()
}

// Messages
type checksCompleteMsg struct {
	issues []checks.Issue
}

func runChecks() tea.Cmd {
	return func() tea.Msg {
		issues := checks.RunAll(".")
		return checksCompleteMsg{issues: issues}
	}
}

type dryRunCompleteMsg struct {
	info *checks.DryRunInfo
}

func runDryRun() tea.Cmd {
	return func() tea.Msg {
		info := checks.DryRun(".")
		return dryRunCompleteMsg{info: info}
	}
}

type promptGeneratedMsg struct {
	prompt string
}

func generatePrompt(selection string, issues []checks.Issue) tea.Cmd {
	return func() tea.Msg {
		prompt := prompts.Generate(selection, issues)
		return promptGeneratedMsg{prompt: prompt}
	}
}

func generateHelpPrompt(topic string) tea.Cmd {
	return func() tea.Msg {
		prompt := prompts.GenerateHelp(topic)
		return promptGeneratedMsg{prompt: prompt}
	}
}

func generatePromptForIssue(issue checks.Issue) tea.Cmd {
	return func() tea.Msg {
		prompt := prompts.GenerateForIssue(issue)
		return promptGeneratedMsg{prompt: prompt}
	}
}

func openConfig() tea.Cmd {
	return func() tea.Msg {
		// TODO: Open config in editor
		return nil
	}
}
