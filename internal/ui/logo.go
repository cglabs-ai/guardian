package ui

import "github.com/charmbracelet/lipgloss"

const logoArt = `
   ╔═╗ ╦ ╦ ╔═╗ ╦═╗ ╔╦╗ ╦ ╔═╗ ╔╗╔
   ║ ╦ ║ ║ ╠═╣ ╠╦╝  ║║ ║ ╠═╣ ║║║
   ╚═╝ ╚═╝ ╩ ╩ ╩╚═ ═╩╝ ╩ ╩ ╩ ╝╚╝
`

const tagline = "Stop AI slop before it hits your codebase."

func Logo() string {
	logoStyle := lipgloss.NewStyle().
		Foreground(BrightGreen).
		Bold(true)

	taglineStyle := lipgloss.NewStyle().
		Foreground(LightGreen).
		Italic(true)

	logo := logoStyle.Render(logoArt)
	tag := taglineStyle.Render("   " + tagline)

	return logo + "\n" + tag
}

func LogoBox() string {
	content := Logo()

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(Green).
		Padding(0, 2).
		Align(lipgloss.Center)

	return box.Render(content)
}

func SmallLogo() string {
	return TitleStyle.Render("GUARDIAN") + DimStyle.Render(" · ") + SubtitleStyle.Render("guardian.sh")
}

func VersionLine(version string) string {
	return DimStyle.Render("v"+version) +
		DimStyle.Render("                                              ") +
		SubtitleStyle.Render("guardian.sh")
}
