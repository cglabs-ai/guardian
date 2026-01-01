package main

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/guardian-sh/guardian/internal/checks"
	"github.com/guardian-sh/guardian/internal/scaffolding"
	"github.com/guardian-sh/guardian/internal/screens"
	"github.com/guardian-sh/guardian/internal/ui"
)

const version = "0.1.0"

func main() {
	if len(os.Args) < 2 {
		// No arguments - launch interactive mode
		runInteractive()
		return
	}

	cmd := strings.ToLower(os.Args[1])

	switch cmd {
	case "check", "run":
		runCheck()
	case "add":
		runAdd()
	case "config":
		runConfig()
	case "version", "--version", "-v":
		fmt.Printf("guardian %s\n", version)
	case "help", "--help", "-h":
		printHelp()
	default:
		fmt.Printf("Unknown command: %s\n\n", cmd)
		printHelp()
		os.Exit(1)
	}
}

func runInteractive() {
	p := tea.NewProgram(
		screens.NewApp(),
		tea.WithAltScreen(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

func runCheck() {
	fix := false
	for _, arg := range os.Args[2:] {
		if arg == "--fix" || arg == "-f" {
			fix = true
		}
	}

	fmt.Println(ui.SmallLogo())
	fmt.Println()

	issues := checks.RunAll(".")

	if len(issues) == 0 {
		fmt.Println(ui.Success("No issues found"))
		return
	}

	// Group by file
	fileIssues := make(map[string][]checks.Issue)
	for _, issue := range issues {
		fileIssues[issue.File] = append(fileIssues[issue.File], issue)
	}

	// Print issues
	critical, warnings, info := 0, 0, 0
	for file, issues := range fileIssues {
		fmt.Printf("\n%s\n", ui.FilePathStyle.Render(file))

		for _, issue := range issues {
			severity := ""
			switch issue.Severity {
			case "critical":
				severity = ui.CriticalStyle.Render(fmt.Sprintf("[%s]", issue.Rule))
				critical++
			case "warning":
				severity = ui.WarningIssueStyle.Render(fmt.Sprintf("[%s]", issue.Rule))
				warnings++
			default:
				severity = ui.InfoIssueStyle.Render(fmt.Sprintf("[%s]", issue.Rule))
				info++
			}

			fmt.Printf("  %s  %s  %s\n",
				ui.LineNumStyle.Render(fmt.Sprintf(":%d", issue.Line)),
				severity,
				issue.Message,
			)
		}
	}

	fmt.Println()
	fmt.Println(ui.Divider())

	// Summary
	parts := []string{}
	if critical > 0 {
		parts = append(parts, ui.CriticalStyle.Render(fmt.Sprintf("%d critical", critical)))
	}
	if warnings > 0 {
		parts = append(parts, ui.WarningStyle.Render(fmt.Sprintf("%d warnings", warnings)))
	}
	if info > 0 {
		parts = append(parts, ui.InfoStyle.Render(fmt.Sprintf("%d info", info)))
	}

	fmt.Printf("\n%s\n", strings.Join(parts, ui.DimStyle.Render(" · ")))

	if fix {
		fmt.Println()
		fmt.Println(ui.Warning("Auto-fix not yet implemented. Use /prompt to generate Claude fixes."))
	}

	fmt.Println()
	fmt.Println(ui.DimStyle.Render("Run 'guardian' for interactive mode with /prompt to generate fixes."))

	if critical > 0 {
		os.Exit(1)
	}
}

func runAdd() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: guardian add <language>")
		fmt.Println()
		fmt.Println("Languages:")
		fmt.Println("  python          Python project")
		fmt.Println("  python-fastapi  Python + FastAPI")
		fmt.Println("  python-django   Python + Django")
		fmt.Println("  typescript      TypeScript project")
		fmt.Println("  typescript-react TypeScript + React")
		fmt.Println("  go              Go project")
		os.Exit(1)
	}

	lang := strings.ToLower(os.Args[2])

	// Map stack to language
	language := lang
	stack := lang
	if strings.HasPrefix(lang, "python") {
		language = "python"
	} else if strings.HasPrefix(lang, "typescript") {
		language = "typescript"
	}

	fmt.Println(ui.SmallLogo())
	fmt.Println()

	fmt.Printf("Adding Guardian for %s...\n\n", lang)

	// Install scaffolding files
	config := scaffolding.InstallConfig{
		Language:    language,
		Stack:       stack,
		SourceDir:   "src",
		ExcludeDirs: []string{"tests", "__pycache__", "node_modules"},
	}

	if err := scaffolding.Install(config); err != nil {
		fmt.Println(ui.Error(fmt.Sprintf("Failed to install: %v", err)))
		os.Exit(1)
	}

	fmt.Println(ui.Success("Created .guardian/ checks"))
	fmt.Println(ui.Success("Created guardian_config.toml"))
	fmt.Println(ui.Success("Created .pre-commit-config.yaml"))

	fmt.Println()
	fmt.Println("Run 'guardian' to enter interactive mode.")
}

func runConfig() {
	// Open config in editor
	configPath := "guardian_config.toml"

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		fmt.Println(ui.Error("No guardian_config.toml found"))
		fmt.Println()
		fmt.Println("Run 'guardian add <language>' to create one.")
		os.Exit(1)
	}

	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vim"
	}

	fmt.Printf("Opening %s in %s...\n", configPath, editor)

	// In a real implementation, we'd exec the editor
	// For now, just print the path
	fmt.Println()
	fmt.Println(ui.DimStyle.Render("Edit: " + configPath))
}

func printHelp() {
	fmt.Println(ui.Logo())
	fmt.Println()
	fmt.Println("Usage: guardian [command]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  (none)         Launch interactive mode")
	fmt.Println("  check          Run all checks")
	fmt.Println("  check --fix    Run checks and auto-fix (coming soon)")
	fmt.Println("  add <lang>     Add Guardian to project")
	fmt.Println("  config         Open configuration")
	fmt.Println("  version        Print version")
	fmt.Println("  help           Print this help")
	fmt.Println()
	fmt.Println("Interactive commands:")
	fmt.Println("  /run           Check your code now")
	fmt.Println("  /dry-run       Preview what would be checked")
	fmt.Println("  /help          Explain something")
	fmt.Println("  /prompt        Generate a prompt for Claude")
	fmt.Println("  /config        Open configuration")
	fmt.Println("  /exit          Leave Guardian")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  guardian                    # Interactive mode")
	fmt.Println("  guardian add python         # Add to Python project")
	fmt.Println("  guardian add typescript     # Add to TypeScript project")
	fmt.Println("  guardian check              # Run checks in CI")
	fmt.Println()
	fmt.Println("Learn more: https://guardian.sh")
}
