package prompts

import (
	"fmt"
	"strings"

	"github.com/guardian-sh/guardian/internal/checks"
)

// Explanation holds explanation text for a rule
type Explanation struct {
	Problem string
	Why     string
	Fix     string
}

// Generate creates a Claude prompt based on user selection
func Generate(selection string, issues []checks.Issue) string {
	switch selection {
	case "I have issues and don't know how to fix them":
		return generateFixPrompt(issues)
	case "I need to set up pre-commit but don't know how":
		return generateSetupPrompt()
	case "I don't understand what Guardian is telling me":
		return generateExplainPrompt(issues)
	case "I want to change the rules but don't know how":
		return generateConfigPrompt()
	default:
		return generateGenericPrompt()
	}
}

// generateFixPrompt creates a prompt to fix issues
func generateFixPrompt(issues []checks.Issue) string {
	var sb strings.Builder

	sb.WriteString(`I ran Guardian (a code quality tool) and it found problems
I don't know how to fix. Here's what it said:

---
`)

	for _, issue := range issues {
		sb.WriteString(fmt.Sprintf("%s:%d - %s\n", issue.File, issue.Line, issue.Rule))
		sb.WriteString(fmt.Sprintf("  %s\n", issue.Message))
	}

	sb.WriteString(`---

Please:
1. Explain each problem in simple terms
2. Show me the fix for each one
3. After you fix them, run ` + "`guardian check`" + ` to make sure they pass
4. If any still fail, keep fixing until they all pass

IMPORTANT: Don't just delete code. Fix it properly. If you're unsure
about something, ask me before making changes.`)

	return sb.String()
}

// generateSetupPrompt creates a prompt to set up pre-commit
func generateSetupPrompt() string {
	return `I just installed Guardian (guardian.sh) in my project. It
created these files:

- .guardian/ (folder with check scripts)
- guardian_config.toml
- .pre-commit-config.yaml

I need help setting up pre-commit so Guardian runs
automatically when I commit code.

Please:
1. Install pre-commit if needed
2. Set up the git hooks
3. Test that it works with a small change
4. Explain what you did in simple terms

My system: [tell Claude your OS - macOS/Linux/Windows]

If anything goes wrong, explain what happened and how to fix it.`
}

// generateExplainPrompt creates a prompt to explain issues
func generateExplainPrompt(issues []checks.Issue) string {
	var sb strings.Builder

	sb.WriteString(`Guardian found these issues in my code but I don't understand
what they mean or why they matter:

---
`)

	for _, issue := range issues {
		sb.WriteString(fmt.Sprintf("%s:%d - [%s] %s\n", issue.File, issue.Line, issue.Rule, issue.Message))
	}

	sb.WriteString(`---

For each issue, please explain:
1. What the problem actually is (in simple terms)
2. Why it matters (what could go wrong)
3. How to fix it (show me the actual code change)

I'm not an expert, so please avoid jargon and explain like
I'm a beginner.`)

	return sb.String()
}

// generateConfigPrompt creates a prompt to configure Guardian
func generateConfigPrompt() string {
	return `I'm using Guardian (guardian.sh) and I want to customize the rules.

The config file is guardian_config.toml in my project root.

I need help with:
- Turning off rules that don't apply to my project
- Adding custom patterns to detect
- Excluding certain files or directories

Please:
1. First, show me my current config (read guardian_config.toml)
2. Ask me what I want to change
3. Update the config for me
4. Explain what each change does

If the config file doesn't exist, help me create one with
sensible defaults for my project.`
}

// generateGenericPrompt creates a generic help prompt
func generateGenericPrompt() string {
	return `I'm using Guardian (guardian.sh), a code quality tool that catches
common mistakes before I commit code.

I need help with something but I'm not sure how to explain it.

Guardian checks for things like:
- Files that are too long (over 500 lines)
- Functions that are too complex (over 50 lines)
- Test/mock data left in production code
- Security issues like hardcoded passwords
- Dangerous commands like rm -rf

Can you help me understand how to use it better? Ask me questions
to figure out what I need.`
}

// GenerateHelp creates a prompt for help topics
func GenerateHelp(topic string) string {
	switch topic {
	case "What is pre-commit?":
		return `Explain to me in simple terms:

1. What is pre-commit?
2. What does it do when I try to commit code?
3. Why would I want to use it?
4. How do I install it?

I'm not a developer, so please avoid technical jargon.
Use analogies if they help explain things.`

	case "What does Guardian check for?":
		return `I installed Guardian (guardian.sh) but I don't fully understand
what it's checking for.

Can you explain each of these checks in simple terms:
- file-size: Files over 500 lines
- func-size: Functions over 50 lines
- mock-data: Test data in production code
- ban-print: print() statements
- ban-eval: eval() and exec()
- dangerous-cmds: Commands like rm -rf
- secret-patterns: Hardcoded passwords
- sql-injection: SQL security issues

For each one, explain:
1. What it catches
2. Why it's a problem
3. A simple example

Keep it beginner-friendly.`

	case "How do I fix an issue?":
		return `Guardian found an issue in my code and I don't know how to fix it.

I need you to:
1. Look at the file Guardian mentioned
2. Find the specific line with the issue
3. Explain what's wrong
4. Show me how to fix it
5. After fixing, run 'guardian check' to verify

Please explain your changes so I can learn for next time.`

	case "How do I turn off a rule?":
		return `I want to disable a Guardian rule because [explain why].

The config file should be guardian_config.toml in my project.

Please:
1. Read my current config
2. Show me which rule controls this check
3. Update the config to disable it
4. Explain what the change does

If the config file doesn't exist, help me create one.`

	default:
		return generateGenericPrompt()
	}
}

// GenerateForIssue creates a prompt for a specific issue
func GenerateForIssue(issue checks.Issue) string {
	return fmt.Sprintf(`Guardian found this issue and I need help fixing it:

File: %s
Line: %d
Rule: %s
Message: %s

Please:
1. Show me the code at that line
2. Explain what's wrong in simple terms
3. Show me the correct way to write it
4. Fix it for me
5. Run 'guardian check' to verify the fix worked

Don't just delete the code - fix it properly. If you need to ask
me questions about what the code should do, please ask.`, issue.File, issue.Line, issue.Rule, issue.Message)
}

// GenerateStopDestructionPrompt creates a prompt to stop Claude from being destructive
func GenerateStopDestructionPrompt() string {
	return `IMPORTANT RULES FOR THIS SESSION:

1. Do NOT delete any files without asking me first
2. Do NOT use rm -rf under any circumstances
3. Do NOT modify files outside of the src/ directory
4. Do NOT run any database commands that delete data
5. Before running any destructive command, show me what
   you're about to do and wait for my approval

If you need to clean up files or make major changes,
show me the list first and let me confirm.

If you're unsure whether something is safe, ASK FIRST.

---

Now, here's what I actually need help with:
[describe your task here]`
}

// GetExplanation returns a detailed explanation for a rule
func GetExplanation(rule string) Explanation {
	explanations := map[string]Explanation{
		"file-size": {
			Problem: "This file has more than 500 lines of code.",
			Why:     "Large files are hard to understand, test, and maintain. They often contain multiple responsibilities that should be separated.",
			Fix:     "Split the file into smaller modules. Group related functions together and move them to separate files.",
		},
		"func-size": {
			Problem: "This function has more than 50 lines of code.",
			Why:     "Long functions are hard to understand and test. They usually do too many things at once.",
			Fix:     "Break the function into smaller helper functions. Each function should do one thing well.",
		},
		"mock-data": {
			Problem: "This looks like test or placeholder data (test@example.com, fake_, dummy_, etc.)",
			Why:     "Test data in production can expose fake accounts, break functionality, or confuse real users.",
			Fix:     "Replace with real data or use environment variables. If it's intentional test code, move it to a test file.",
		},
		"ban-print": {
			Problem: "You're using print() for output.",
			Why:     "Print statements get lost in production, can't be filtered by log level, and are hard to find later.",
			Fix:     "Use a logging library instead: import logging; logging.info('message')",
		},
		"ban-except": {
			Problem: "You're catching all exceptions with bare 'except:'",
			Why:     "This catches everything including KeyboardInterrupt and SystemExit, hiding real errors and making debugging impossible.",
			Fix:     "Catch specific exceptions: except ValueError: or except (TypeError, ValueError):",
		},
		"ban-eval": {
			Problem: "You're using eval() or exec() to run code.",
			Why:     "These execute arbitrary code, which is a massive security risk. Attackers can run any code they want.",
			Fix:     "Almost always there's a safer alternative. For JSON use json.loads(). For math use ast.literal_eval().",
		},
		"ban-star": {
			Problem: "You're using 'from module import *'",
			Why:     "This pollutes your namespace, makes it unclear where names come from, and can cause name conflicts.",
			Fix:     "Import specific names: from module import func1, func2",
		},
		"todo-marker": {
			Problem: "There's a TODO, FIXME, or HACK comment in the code.",
			Why:     "These markers indicate unfinished work that shouldn't go to production.",
			Fix:     "Either complete the TODO or create a ticket to track it and remove the comment.",
		},
		"dangerous-cmd": {
			Problem: "This code contains a dangerous command like rm -rf, DROP TABLE, or DELETE FROM.",
			Why:     "These commands can permanently destroy data. One mistake can wipe databases or file systems.",
			Fix:     "Add confirmation prompts, use safe defaults, or implement soft-delete instead of hard-delete.",
		},
		"secret-pattern": {
			Problem: "This looks like a hardcoded password, API key, or secret.",
			Why:     "Secrets in code get committed to git, shared with everyone, and are very hard to rotate.",
			Fix:     "Use environment variables: api_key = os.environ['API_KEY']",
		},
		"sql-injection": {
			Problem: "You're building SQL queries with f-strings or string concatenation.",
			Why:     "This allows SQL injection attacks. Users can input malicious SQL that drops tables or steals data.",
			Fix:     "Use parameterized queries: cursor.execute('SELECT * FROM users WHERE id = ?', (user_id,))",
		},
		"subprocess-shell": {
			Problem: "You're using shell=True in subprocess.",
			Why:     "This passes commands through a shell, enabling command injection attacks.",
			Fix:     "Pass commands as a list instead: subprocess.run(['ls', '-la'])",
		},
		"ban-console": {
			Problem: "You're using console.log() for output.",
			Why:     "Console statements clutter production logs and can expose sensitive information.",
			Fix:     "Use a proper logging library or remove before committing.",
		},
	}

	if exp, ok := explanations[rule]; ok {
		return exp
	}

	return Explanation{
		Problem: "Guardian detected an issue with this code.",
		Why:     "This pattern can lead to bugs, security issues, or maintenance problems.",
		Fix:     "Review the code and fix the identified issue.",
	}
}
