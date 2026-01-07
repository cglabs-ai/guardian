package checks

import (
	"bufio"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// Pre-compiled regexes for performance (compiled once at package init)
var (
	// Mock data patterns
	mockPatternRegexes = []*regexp.Regexp{
		regexp.MustCompile(`test@example\.com`),
		regexp.MustCompile(`example@`),
		regexp.MustCompile(`@test\.com`),
		regexp.MustCompile(`fake_`),
		regexp.MustCompile(`_fake`),
		regexp.MustCompile(`mock_`),
		regexp.MustCompile(`_mock`),
		regexp.MustCompile(`dummy_`),
		regexp.MustCompile(`placeholder`),
		regexp.MustCompile(`test_user`),
		regexp.MustCompile(`test_password`),
		regexp.MustCompile(`changeme`),
		regexp.MustCompile(`your_.*_here`),
	}

	// Code pattern regexes
	printRe     = regexp.MustCompile(`\bprint\s*\(`)
	bareExceptRe = regexp.MustCompile(`except\s*:`)
	evalRe      = regexp.MustCompile(`(?:^|[=(:,\s])eval\s*\(`)
	execRe      = regexp.MustCompile(`(?:^|[=(:,\s])exec\s*\(`)
	starImportRe = regexp.MustCompile(`from\s+\S+\s+import\s+\*`)
	sqlInjectionRe = regexp.MustCompile(`(?i)f["'](?:SELECT|INSERT|UPDATE|DELETE)`)

	// Dangerous command patterns
	dangerousPatternRegexes = []*regexp.Regexp{
		regexp.MustCompile(`(?i)rm\s+-rf`),
		regexp.MustCompile(`(?i)DROP\s+TABLE`),
		regexp.MustCompile(`(?i)DROP\s+DATABASE`),
		regexp.MustCompile(`(?i)DELETE\s+FROM\s+\w+\s*;`),
		regexp.MustCompile(`(?i)TRUNCATE\s+TABLE`),
	}

	// Secret patterns
	secretPatternRegexes = []*regexp.Regexp{
		regexp.MustCompile(`(?i)api_key\s*=\s*["'][^"']+["']`),
		regexp.MustCompile(`(?i)password\s*=\s*["'][^"']+["']`),
		regexp.MustCompile(`(?i)secret\s*=\s*["'][^"']+["']`),
		regexp.MustCompile(`(?i)AWS_SECRET`),
		regexp.MustCompile(`(?i)PRIVATE_KEY`),
	}

	// Output parsing regexes
	issueFormatRe = regexp.MustCompile(`^(.+):(\d+)\s+\[([^\]]+)\]\s+(.+)$`)
	failFormatRe  = regexp.MustCompile(`^FAIL\s+(.+):(\d+)\s+-\s+([^:]+):\s+(.+)$`)

	// Shared exclusion list for directory skipping (used by both RunAll and DryRun)
	excludedDirs = map[string]bool{
		".git":        true,
		"node_modules": true,
		"__pycache__":  true,
		".venv":        true,
		"venv":         true,
		".guardian":    true,
	}
)

// Issue represents a single code issue
type Issue struct {
	File     string
	Line     int
	Rule     string
	Message  string
	Severity string // "critical", "warning", "info"
}

// DryRunInfo contains info about what would be checked
type DryRunInfo struct {
	Files      []FileInfo
	Excluded   []string
	FileCount  int
	TotalLines int
}

// FileInfo contains info about a single file
type FileInfo struct {
	Path  string
	Lines int
}

// RunAll runs all checks in the given directory
func RunAll(dir string) []Issue {
	var issues []Issue

	// Check if guardian.py exists
	guardianPath := filepath.Join(dir, ".guardian", "guardian.py")
	if _, err := os.Stat(guardianPath); os.IsNotExist(err) {
		// Try running individual checks
		issues = append(issues, runBuiltinChecks(dir)...)
		return issues
	}

	// Run the guardian.py script
	cmd := exec.Command("python3", guardianPath)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()

	if err != nil {
		// Python script failed - fall back to builtin checks
		// This handles: python3 not installed, script errors, etc.
		issues = append(issues, runBuiltinChecks(dir)...)
		return issues
	}

	// Parse output
	issues = append(issues, parseGuardianOutput(string(output))...)

	return issues
}

// runBuiltinChecks runs checks without external scripts
func runBuiltinChecks(dir string) []Issue {
	var issues []Issue

	// Walk directory
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		// Skip excluded directories (using shared exclusion list)
		if info.IsDir() {
			if excludedDirs[info.Name()] {
				return filepath.SkipDir
			}
			return nil
		}

		// Only check Python and JS/TS files
		ext := filepath.Ext(path)
		if ext != ".py" && ext != ".js" && ext != ".ts" && ext != ".tsx" {
			return nil
		}

		// Run checks on file
		fileIssues := checkFile(path)
		issues = append(issues, fileIssues...)

		return nil
	})

	return issues
}

// checkFile runs builtin checks on a single file
func checkFile(path string) []Issue {
	var issues []Issue

	content, err := os.ReadFile(path)
	if err != nil {
		return issues
	}

	lines := strings.Split(string(content), "\n")
	// Fix off-by-one: if file ends with newline, Split adds empty element
	// A 500-line file with trailing newline has 501 elements but is still 500 lines
	lineCount := len(lines)
	if lineCount > 0 && lines[lineCount-1] == "" {
		lineCount--
	}
	relPath := path

	// File size check
	if lineCount > 500 {
		issues = append(issues, Issue{
			File:     relPath,
			Line:     1,
			Rule:     "file-size",
			Message:  "File has " + strconv.Itoa(lineCount) + " lines (max 500)",
			Severity: "warning",
		})
	}

	// Track docstring state for multi-line strings
	inDocstring := false
	docstringDelim := ""

	// Line-by-line checks
	for i, line := range lines {
		lineNum := i + 1
		trimmed := strings.TrimSpace(line)

		// Skip empty lines
		if trimmed == "" {
			continue
		}

		// Track multi-line docstrings (Python)
		if !inDocstring {
			if strings.HasPrefix(trimmed, `"""`) || strings.HasPrefix(trimmed, `'''`) {
				docstringDelim = trimmed[:3]
				// Check if docstring ends on same line
				rest := trimmed[3:]
				if !strings.Contains(rest, docstringDelim) {
					inDocstring = true
				}
				continue // Skip docstring start line
			}
		} else {
			// Inside docstring - check for end
			if strings.Contains(trimmed, docstringDelim) {
				inDocstring = false
			}
			continue // Skip all docstring content
		}

		// Skip comment lines (Python #, JS/TS //)
		isComment := strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, "//")

		// Mock data patterns (using pre-compiled regexes)
		lowerLine := strings.ToLower(line)
		for _, re := range mockPatternRegexes {
			if re.MatchString(lowerLine) {
				issues = append(issues, Issue{
					File:     relPath,
					Line:     lineNum,
					Rule:     "mock-data",
					Message:  "Possible test/mock data detected",
					Severity: "warning",
				})
				break
			}
		}

		// Print statements (Python) - use word boundary to avoid "blueprint", "fingerprint"
		if !isComment && printRe.MatchString(line) {
			issues = append(issues, Issue{
				File:     relPath,
				Line:     lineNum,
				Rule:     "ban-print",
				Message:  "Remove print() - use logging instead",
				Severity: "info",
			})
		}

		// Console.log (JS/TS)
		if !isComment && strings.Contains(line, "console.log(") {
			issues = append(issues, Issue{
				File:     relPath,
				Line:     lineNum,
				Rule:     "ban-console",
				Message:  "Remove console.log() - use proper logging",
				Severity: "info",
			})
		}

		// Bare except (Python)
		if !isComment && bareExceptRe.MatchString(line) {
			issues = append(issues, Issue{
				File:     relPath,
				Line:     lineNum,
				Rule:     "ban-except",
				Message:  "Avoid bare except: - catch specific exceptions",
				Severity: "warning",
			})
		}

		// eval/exec - only flag actual function calls, not strings/comments
		if !isComment {
			// Only match if eval/exec is preceded by = ( , : or start of line
			// This avoids matching "eval(" inside strings like "don't use eval()"
			if evalRe.MatchString(trimmed) {
				// Check if inside a string by counting unescaped quotes before eval
				beforeEval := strings.Split(line, "eval")[0]
				// Count only unescaped quotes by removing escaped ones first
				cleaned := strings.ReplaceAll(beforeEval, `\"`, "")
				cleaned = strings.ReplaceAll(cleaned, `\'`, "")
				doubleQuotes := strings.Count(cleaned, `"`)
				singleQuotes := strings.Count(cleaned, `'`)
				// If either quote type is odd, we're inside a string
				if doubleQuotes%2 == 0 && singleQuotes%2 == 0 {
					issues = append(issues, Issue{
						File:     relPath,
						Line:     lineNum,
						Rule:     "ban-eval",
						Message:  "Avoid eval() - security risk",
						Severity: "critical",
					})
				}
			}
			if execRe.MatchString(trimmed) {
				beforeExec := strings.Split(line, "exec")[0]
				cleaned := strings.ReplaceAll(beforeExec, `\"`, "")
				cleaned = strings.ReplaceAll(cleaned, `\'`, "")
				doubleQuotes := strings.Count(cleaned, `"`)
				singleQuotes := strings.Count(cleaned, `'`)
				if doubleQuotes%2 == 0 && singleQuotes%2 == 0 {
					issues = append(issues, Issue{
						File:     relPath,
						Line:     lineNum,
						Rule:     "ban-eval",
						Message:  "Avoid exec() - security risk",
						Severity: "critical",
					})
				}
			}
		}

		// Star imports
		if !isComment && starImportRe.MatchString(line) {
			issues = append(issues, Issue{
				File:     relPath,
				Line:     lineNum,
				Rule:     "ban-star",
				Message:  "Avoid wildcard imports - import specific names",
				Severity: "warning",
			})
		}

		// TODO/FIXME markers
		upperLine := strings.ToUpper(line)
		if strings.Contains(upperLine, "TODO") || strings.Contains(upperLine, "FIXME") || strings.Contains(upperLine, "HACK") {
			issues = append(issues, Issue{
				File:     relPath,
				Line:     lineNum,
				Rule:     "todo-marker",
				Message:  "Resolve TODO/FIXME before committing",
				Severity: "info",
			})
		}

		// Dangerous commands (using pre-compiled regexes)
		if !isComment {
			for _, re := range dangerousPatternRegexes {
				if re.MatchString(line) {
					issues = append(issues, Issue{
						File:     relPath,
						Line:     lineNum,
						Rule:     "dangerous-cmd",
						Message:  "Dangerous command detected - review carefully",
						Severity: "critical",
					})
					break
				}
			}
		}

		// Secret patterns (using pre-compiled regexes)
		if !isComment {
			for _, re := range secretPatternRegexes {
				if re.MatchString(line) {
					issues = append(issues, Issue{
						File:     relPath,
						Line:     lineNum,
						Rule:     "secret-pattern",
						Message:  "Possible hardcoded secret - use environment variables",
						Severity: "critical",
					})
					break
				}
			}
		}

		// SQL injection (f-strings in queries) - case insensitive
		if !isComment && sqlInjectionRe.MatchString(line) {
			issues = append(issues, Issue{
				File:     relPath,
				Line:     lineNum,
				Rule:     "sql-injection",
				Message:  "f-string in SQL query - use parameterized queries",
				Severity: "critical",
			})
		}

		// subprocess with shell=True
		if !isComment && strings.Contains(line, "shell=True") {
			issues = append(issues, Issue{
				File:     relPath,
				Line:     lineNum,
				Rule:     "subprocess-shell",
				Message:  "Avoid shell=True in subprocess - security risk",
				Severity: "warning",
			})
		}
	}

	return issues
}

// parseGuardianOutput parses output from guardian.py
func parseGuardianOutput(output string) []Issue {
	var issues []Issue

	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()

		// Parse format: "FAIL file.py:45 - rule: message"
		// or: "file.py:45 [rule] message"
		if strings.HasPrefix(line, "FAIL") || strings.Contains(line, "[") {
			issue := parseIssueLine(line)
			if issue.File != "" {
				issues = append(issues, issue)
			}
		}
	}

	return issues
}

func parseIssueLine(line string) Issue {
	// Try format: "file.py:45 [rule] message" (using pre-compiled regex)
	matches := issueFormatRe.FindStringSubmatch(line)

	if len(matches) == 5 {
		lineNum, _ := strconv.Atoi(matches[2])
		return Issue{
			File:     matches[1],
			Line:     lineNum,
			Rule:     matches[3],
			Message:  matches[4],
			Severity: getSeverity(matches[3]),
		}
	}

	// Try format: "FAIL file.py:45 - rule: message" (using pre-compiled regex)
	matches2 := failFormatRe.FindStringSubmatch(line)

	if len(matches2) == 5 {
		lineNum, _ := strconv.Atoi(matches2[2])
		return Issue{
			File:     matches2[1],
			Line:     lineNum,
			Rule:     matches2[3],
			Message:  matches2[4],
			Severity: getSeverity(matches2[3]),
		}
	}

	return Issue{}
}

func getSeverity(rule string) string {
	criticalRules := map[string]bool{
		"ban-eval":       true,
		"dangerous-cmd":  true,
		"secret-pattern": true,
		"sql-injection":  true,
	}

	if criticalRules[rule] {
		return "critical"
	}

	infoRules := map[string]bool{
		"ban-print":   true,
		"ban-console": true,
		"todo-marker": true,
	}

	if infoRules[rule] {
		return "info"
	}

	return "warning"
}

// DryRun returns info about what would be checked
func DryRun(dir string) *DryRunInfo {
	info := &DryRunInfo{
		Excluded: []string{},
	}

	filepath.Walk(dir, func(path string, fileInfo os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		// Use shared exclusion list (same as runBuiltinChecks)
		if fileInfo.IsDir() {
			if excludedDirs[fileInfo.Name()] {
				info.Excluded = append(info.Excluded, fileInfo.Name()+"/")
				return filepath.SkipDir
			}
			return nil
		}

		ext := filepath.Ext(path)
		// Match the same file types as runBuiltinChecks
		if ext != ".py" && ext != ".js" && ext != ".ts" && ext != ".tsx" {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		// Fix off-by-one: consistent with checkFile line counting
		lines := strings.Split(string(content), "\n")
		lineCount := len(lines)
		if lineCount > 0 && lines[lineCount-1] == "" {
			lineCount--
		}

		relPath, _ := filepath.Rel(dir, path)
		info.Files = append(info.Files, FileInfo{
			Path:  relPath,
			Lines: lineCount,
		})

		info.FileCount++
		info.TotalLines += lineCount

		return nil
	})

	// Remove duplicate exclusions
	seen := make(map[string]bool)
	uniqueExcl := []string{}
	for _, e := range info.Excluded {
		if !seen[e] {
			seen[e] = true
			uniqueExcl = append(uniqueExcl, e)
		}
	}
	info.Excluded = uniqueExcl

	return info
}
