package ui

import (
	"fmt"
	"strings"
)

// MenuItem represents a menu option
type MenuItem struct {
	Label       string
	Description string
	Value       string
}

// RenderMenu renders a selectable menu
func RenderMenu(items []MenuItem, cursor int) string {
	var b strings.Builder

	for i, item := range items {
		if i == cursor {
			b.WriteString(CursorStyle.Render("  ❯ ● "))
			b.WriteString(SelectedStyle.Render(item.Label))
			if item.Description != "" {
				b.WriteString(DimStyle.Render("  " + item.Description))
			}
		} else {
			b.WriteString(DimStyle.Render("    ○ "))
			b.WriteString(UnselectedStyle.Render(item.Label))
		}
		b.WriteString("\n")
	}

	return b.String()
}

// RenderPrompt renders an input prompt
func RenderPrompt(label string, value string, placeholder string, focused bool) string {
	var b strings.Builder

	b.WriteString(NormalStyle.Render("  ? "))
	b.WriteString(TitleStyle.Render(label))
	b.WriteString("\n")

	if focused {
		b.WriteString(CursorStyle.Render("  › "))
		if value == "" {
			b.WriteString(PlaceholderStyle.Render(placeholder))
		} else {
			b.WriteString(InputStyle.Render(value))
		}
		b.WriteString(CursorStyle.Render("█"))
	} else {
		b.WriteString(DimStyle.Render("  › "))
		b.WriteString(DimStyle.Render(value))
	}

	return b.String()
}

// RenderConfirmation renders a yes/no confirmation
func RenderConfirmation(label string, defaultYes bool) string {
	var b strings.Builder

	b.WriteString(NormalStyle.Render("  ? "))
	b.WriteString(TitleStyle.Render(label))

	if defaultYes {
		b.WriteString(DimStyle.Render(" (Y/n) "))
	} else {
		b.WriteString(DimStyle.Render(" (y/N) "))
	}

	return b.String()
}

// RenderProgress renders a progress indicator
func RenderProgress(steps []string, current int) string {
	var b strings.Builder

	for i, step := range steps {
		if i < current {
			b.WriteString(SuccessStyle.Render("  ✓ " + step))
		} else if i == current {
			b.WriteString(HighlightStyle.Render("  ├─ " + step))
		} else {
			b.WriteString(DimStyle.Render("  ├─ " + step))
		}
		b.WriteString("\n")
	}

	return b.String()
}

// RenderFileList renders a list of files being created
func RenderFileList(files []string, completed int) string {
	var b strings.Builder

	for i, file := range files {
		if i < completed {
			b.WriteString(Success(file))
		} else {
			b.WriteString(DimStyle.Render("  · " + file))
		}
		b.WriteString("\n")
	}

	return b.String()
}

// RenderIssue renders a single issue
func RenderIssue(filepath string, line int, rule string, message string, severity string) string {
	var b strings.Builder

	// File:line
	b.WriteString(FilePathStyle.Render(filepath))
	b.WriteString(LineNumStyle.Render(fmt.Sprintf(":%d", line)))
	b.WriteString("   ")

	// Rule name with color based on severity
	switch severity {
	case "critical":
		b.WriteString(CriticalStyle.Render(rule))
	case "warning":
		b.WriteString(WarningIssueStyle.Render(rule))
	default:
		b.WriteString(InfoIssueStyle.Render(rule))
	}

	// Message
	b.WriteString("   ")
	b.WriteString(NormalStyle.Render(message))

	return b.String()
}

// RenderIssueGroup renders issues grouped by file
func RenderIssueGroup(filepath string, issues []struct {
	Line     int
	Rule     string
	Message  string
	Severity string
}) string {
	var b strings.Builder

	b.WriteString("\n")
	b.WriteString(FilePathStyle.Render("  " + filepath))
	b.WriteString("\n")

	for _, issue := range issues {
		b.WriteString(LineNumStyle.Render(fmt.Sprintf("    :%d", issue.Line)))
		b.WriteString("   ")

		switch issue.Severity {
		case "critical":
			b.WriteString(CriticalStyle.Render(fmt.Sprintf("[%s]", issue.Rule)))
		case "warning":
			b.WriteString(WarningIssueStyle.Render(fmt.Sprintf("[%s]", issue.Rule)))
		default:
			b.WriteString(InfoIssueStyle.Render(fmt.Sprintf("[%s]", issue.Rule)))
		}

		b.WriteString("  ")
		b.WriteString(NormalStyle.Render(issue.Message))
		b.WriteString("\n")
	}

	return b.String()
}

// RenderHelp renders help text at the bottom
func RenderHelp(keys map[string]string) string {
	var parts []string

	for key, desc := range keys {
		parts = append(parts, DimStyle.Render(key)+" "+SubtitleStyle.Render(desc))
	}

	return strings.Join(parts, DimStyle.Render(" · "))
}

// RenderCommandPrompt renders the interactive command prompt
func RenderCommandPrompt(input string) string {
	var b strings.Builder

	b.WriteString(CursorStyle.Render("  › "))
	b.WriteString(InputStyle.Render(input))
	b.WriteString(CursorStyle.Render("█"))

	return b.String()
}

// RenderDetected renders auto-detected values
func RenderDetected(label string, value string) string {
	return DimStyle.Render("    detected: ") + SubtitleStyle.Render(value)
}
