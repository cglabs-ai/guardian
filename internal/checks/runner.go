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
	output, _ := cmd.CombinedOutput()

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

		// Skip excluded directories
		if info.IsDir() {
			name := info.Name()
			if name == ".git" || name == "node_modules" || name == "__pycache__" ||
				name == ".venv" || name == "venv" || name == ".guardian" {
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
	relPath := path

	// File size check
	if len(lines) > 500 {
		issues = append(issues, Issue{
			File:     relPath,
			Line:     1,
			Rule:     "file-size",
			Message:  "File has " + strconv.Itoa(len(lines)) + " lines (max 500)",
			Severity: "warning",
		})
	}

	// Line-by-line checks
	for i, line := range lines {
		lineNum := i + 1

		// Mock data patterns
		mockPatterns := []string{
			`test@example\.com`,
			`example@`,
			`@test\.com`,
			`fake_`,
			`_fake`,
			`mock_`,
			`_mock`,
			`dummy_`,
			`placeholder`,
			`test_user`,
			`test_password`,
			`changeme`,
			`your_.*_here`,
		}

		for _, pattern := range mockPatterns {
			if matched, _ := regexp.MatchString(pattern, strings.ToLower(line)); matched {
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

		// Print statements (Python)
		if strings.Contains(line, "print(") && !strings.HasPrefix(strings.TrimSpace(line), "#") {
			issues = append(issues, Issue{
				File:     relPath,
				Line:     lineNum,
				Rule:     "ban-print",
				Message:  "Remove print() - use logging instead",
				Severity: "info",
			})
		}

		// Console.log (JS/TS)
		if strings.Contains(line, "console.log(") && !strings.HasPrefix(strings.TrimSpace(line), "//") {
			issues = append(issues, Issue{
				File:     relPath,
				Line:     lineNum,
				Rule:     "ban-console",
				Message:  "Remove console.log() - use proper logging",
				Severity: "info",
			})
		}

		// Bare except (Python)
		if matched, _ := regexp.MatchString(`except\s*:`, line); matched {
			issues = append(issues, Issue{
				File:     relPath,
				Line:     lineNum,
				Rule:     "ban-except",
				Message:  "Avoid bare except: - catch specific exceptions",
				Severity: "warning",
			})
		}

		// eval/exec
		if strings.Contains(line, "eval(") || strings.Contains(line, "exec(") {
			issues = append(issues, Issue{
				File:     relPath,
				Line:     lineNum,
				Rule:     "ban-eval",
				Message:  "Avoid eval()/exec() - security risk",
				Severity: "critical",
			})
		}

		// Star imports
		if matched, _ := regexp.MatchString(`from\s+\S+\s+import\s+\*`, line); matched {
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

		// Dangerous commands
		dangerousPatterns := []string{
			`rm\s+-rf`,
			`DROP\s+TABLE`,
			`DROP\s+DATABASE`,
			`DELETE\s+FROM\s+\w+\s*;`,
			`TRUNCATE\s+TABLE`,
		}

		for _, pattern := range dangerousPatterns {
			if matched, _ := regexp.MatchString("(?i)"+pattern, line); matched {
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

		// Secret patterns
		secretPatterns := []string{
			`api_key\s*=\s*["'][^"']+["']`,
			`password\s*=\s*["'][^"']+["']`,
			`secret\s*=\s*["'][^"']+["']`,
			`AWS_SECRET`,
			`PRIVATE_KEY`,
		}

		for _, pattern := range secretPatterns {
			if matched, _ := regexp.MatchString("(?i)"+pattern, line); matched {
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

		// SQL injection (f-strings in queries)
		if strings.Contains(line, "f\"SELECT") || strings.Contains(line, "f'SELECT") ||
			strings.Contains(line, "f\"INSERT") || strings.Contains(line, "f'INSERT") ||
			strings.Contains(line, "f\"UPDATE") || strings.Contains(line, "f'UPDATE") ||
			strings.Contains(line, "f\"DELETE") || strings.Contains(line, "f'DELETE") {
			issues = append(issues, Issue{
				File:     relPath,
				Line:     lineNum,
				Rule:     "sql-injection",
				Message:  "f-string in SQL query - use parameterized queries",
				Severity: "critical",
			})
		}

		// subprocess with shell=True
		if strings.Contains(line, "shell=True") {
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
	// Try format: "file.py:45 [rule] message"
	re := regexp.MustCompile(`^(.+):(\d+)\s+\[([^\]]+)\]\s+(.+)$`)
	matches := re.FindStringSubmatch(line)

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

	// Try format: "FAIL file.py:45 - rule: message"
	re2 := regexp.MustCompile(`^FAIL\s+(.+):(\d+)\s+-\s+([^:]+):\s+(.+)$`)
	matches2 := re2.FindStringSubmatch(line)

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

	// Common exclusions
	exclusions := []string{".git", "node_modules", "__pycache__", ".venv", "venv", "tests", "test"}

	filepath.Walk(dir, func(path string, fileInfo os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if fileInfo.IsDir() {
			name := fileInfo.Name()
			for _, excl := range exclusions {
				if name == excl {
					info.Excluded = append(info.Excluded, name+"/")
					return filepath.SkipDir
				}
			}
			return nil
		}

		ext := filepath.Ext(path)
		if ext != ".py" && ext != ".js" && ext != ".ts" && ext != ".tsx" && ext != ".go" {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		lines := len(strings.Split(string(content), "\n"))

		relPath, _ := filepath.Rel(dir, path)
		info.Files = append(info.Files, FileInfo{
			Path:  relPath,
			Lines: lines,
		})

		info.FileCount++
		info.TotalLines += lines

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
