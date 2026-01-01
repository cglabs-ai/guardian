package ui

import (
	"github.com/charmbracelet/lipgloss"
)

// Forest green color palette
var (
	// Primary colors
	Green       = lipgloss.Color("#2E8B57") // Sea green - primary
	DarkGreen   = lipgloss.Color("#1E5631") // Dark forest
	LightGreen  = lipgloss.Color("#3CB371") // Medium sea green
	BrightGreen = lipgloss.Color("#00FF7F") // Spring green - accents

	// Neutral colors
	White   = lipgloss.Color("#FFFFFF")
	Gray    = lipgloss.Color("#808080")
	DimGray = lipgloss.Color("#545454")
	Black   = lipgloss.Color("#000000")

	// Status colors
	Red    = lipgloss.Color("#FF6B6B")
	Yellow = lipgloss.Color("#FFE66D")
	Cyan   = lipgloss.Color("#4ECDC4")
)

// Styles
var (
	// Box styles
	BoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Green).
			Padding(1, 2)

	HeaderBox = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Green).
			Padding(0, 2).
			MarginBottom(1)

	// Text styles
	TitleStyle = lipgloss.NewStyle().
			Foreground(BrightGreen).
			Bold(true)

	SubtitleStyle = lipgloss.NewStyle().
			Foreground(LightGreen)

	NormalStyle = lipgloss.NewStyle().
			Foreground(White)

	DimStyle = lipgloss.NewStyle().
			Foreground(DimGray)

	HighlightStyle = lipgloss.NewStyle().
			Foreground(BrightGreen).
			Bold(true)

	// Menu styles
	SelectedStyle = lipgloss.NewStyle().
			Foreground(BrightGreen).
			Bold(true)

	UnselectedStyle = lipgloss.NewStyle().
			Foreground(Gray)

	CursorStyle = lipgloss.NewStyle().
			Foreground(BrightGreen).
			Bold(true)

	// Status styles
	SuccessStyle = lipgloss.NewStyle().
			Foreground(LightGreen)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(Red)

	WarningStyle = lipgloss.NewStyle().
			Foreground(Yellow)

	InfoStyle = lipgloss.NewStyle().
			Foreground(Cyan)

	// Input styles
	InputStyle = lipgloss.NewStyle().
			Foreground(White)

	PlaceholderStyle = lipgloss.NewStyle().
			Foreground(DimGray)

	// Code block style
	CodeStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#1a1a1a")).
			Foreground(LightGreen).
			Padding(0, 1)

	// Prompt box style (for Claude prompts)
	PromptBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(LightGreen).
			Padding(1, 2).
			MarginTop(1).
			MarginBottom(1)

	PromptHeaderStyle = lipgloss.NewStyle().
			Background(Green).
			Foreground(White).
			Bold(true).
			Padding(0, 2)

	// Issue styles
	CriticalStyle = lipgloss.NewStyle().
			Foreground(Red).
			Bold(true)

	WarningIssueStyle = lipgloss.NewStyle().
			Foreground(Yellow)

	InfoIssueStyle = lipgloss.NewStyle().
			Foreground(Cyan)

	// File path style
	FilePathStyle = lipgloss.NewStyle().
			Foreground(LightGreen)

	LineNumStyle = lipgloss.NewStyle().
			Foreground(DimGray)

	RuleNameStyle = lipgloss.NewStyle().
			Foreground(Cyan)
)

// Helper functions
func Success(s string) string {
	return SuccessStyle.Render("✓ " + s)
}

func Error(s string) string {
	return ErrorStyle.Render("✗ " + s)
}

func Warning(s string) string {
	return WarningStyle.Render("⚠ " + s)
}

func Info(s string) string {
	return InfoStyle.Render("ℹ " + s)
}

func Bullet(s string) string {
	return DimStyle.Render("  ├─ ") + NormalStyle.Render(s)
}

func LastBullet(s string) string {
	return DimStyle.Render("  └─ ") + NormalStyle.Render(s)
}

func Indent(s string) string {
	return "    " + s
}

func Divider() string {
	return DimStyle.Render("──────────────────────────────────────────────────────────────")
}
