package screens

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/guardian-sh/guardian/internal/ui"
)

type AboutModel struct{}

func NewAbout() AboutModel {
	return AboutModel{}
}

func (m AboutModel) Init() tea.Cmd {
	return nil
}

func (m AboutModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Enter), key.Matches(msg, keys.Back), key.Matches(msg, keys.Quit):
			return m, goBack()
		}
	}
	return m, nil
}

func (m AboutModel) View() string {
	var s strings.Builder

	s.WriteString(ui.SmallLogo())
	s.WriteString("\n\n")

	s.WriteString(ui.TitleStyle.Render("  ● About"))
	s.WriteString("\n\n")

	s.WriteString(ui.NormalStyle.Render("  Guardian catches AI slop before it hits your codebase."))
	s.WriteString("\n\n")

	// Free checks
	s.WriteString(ui.TitleStyle.Render("  Free checks "))
	s.WriteString(ui.DimStyle.Render("(no AI, <200ms):"))
	s.WriteString("\n\n")

	freeChecks := []struct {
		name string
		desc string
	}{
		{"file-size", "Files over 500 lines"},
		{"func-size", "Functions over 50 lines"},
		{"mock-data", "test_, fake_, example@, placeholder"},
		{"ban-print", "print() statements"},
		{"ban-except", "Bare except: blocks"},
		{"ban-eval", "eval(), exec()"},
		{"ban-star", "from x import *"},
		{"mutable-default", "def foo(items=[])"},
		{"todo-markers", "TODO, FIXME, HACK"},
		{"dangerous-cmds", "rm -rf, DROP TABLE, DELETE FROM"},
		{"secret-patterns", "api_key=, password=, hardcoded tokens"},
		{"subprocess-shell", "shell=True"},
		{"sql-injection", "f-strings in SQL"},
	}

	for i, check := range freeChecks {
		prefix := "├─"
		if i == len(freeChecks)-1 {
			prefix = "└─"
		}
		s.WriteString(ui.DimStyle.Render("    " + prefix + " "))
		s.WriteString(ui.HighlightStyle.Render(padRight(check.name, 18)))
		s.WriteString(ui.NormalStyle.Render(check.desc))
		s.WriteString("\n")
	}

	s.WriteString("\n")

	// AI features
	s.WriteString(ui.TitleStyle.Render("  AI features "))
	s.WriteString(ui.DimStyle.Render("(BYOK):"))
	s.WriteString("\n\n")

	aiFeatures := []struct {
		name string
		desc string
	}{
		{"smart-scan", "Auto-detect stack, patterns, conflicts"},
		{"prompt-gen", "Generate Claude prompts for fixes"},
	}

	for i, feature := range aiFeatures {
		prefix := "├─"
		if i == len(aiFeatures)-1 {
			prefix = "└─"
		}
		s.WriteString(ui.DimStyle.Render("    " + prefix + " "))
		s.WriteString(ui.HighlightStyle.Render(padRight(feature.name, 18)))
		s.WriteString(ui.NormalStyle.Render(feature.desc))
		s.WriteString("\n")
	}

	s.WriteString("\n")
	s.WriteString(ui.NormalStyle.Render("  All checks configurable in "))
	s.WriteString(ui.FilePathStyle.Render("guardian_config.toml"))
	s.WriteString("\n\n")

	s.WriteString(ui.Divider())
	s.WriteString("\n\n")

	s.WriteString(ui.SubtitleStyle.Render("  guardian.sh"))
	s.WriteString(ui.DimStyle.Render(" · "))
	s.WriteString(ui.DimStyle.Render("github.com/guardian-sh/guardian"))
	s.WriteString(ui.DimStyle.Render(" · "))
	s.WriteString(ui.DimStyle.Render("MIT License"))
	s.WriteString("\n\n")

	s.WriteString(ui.DimStyle.Render("  Press any key to go back..."))

	return s.String()
}

func padRight(s string, length int) string {
	if len(s) >= length {
		return s
	}
	return s + strings.Repeat(" ", length-len(s))
}
