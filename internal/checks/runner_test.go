package checks

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Helper to create temp file with content and run checks
func checkCode(t *testing.T, filename, content string) []Issue {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, filename)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}
	return checkFile(path)
}

// Helper to assert issue count
func assertIssueCount(t *testing.T, issues []Issue, expected int, context string) {
	t.Helper()
	if len(issues) != expected {
		t.Errorf("%s: expected %d issues, got %d", context, expected, len(issues))
		for _, issue := range issues {
			t.Logf("  - %s:%d [%s] %s", issue.File, issue.Line, issue.Rule, issue.Message)
		}
	}
}

// Helper to assert specific rule triggered
func assertHasRule(t *testing.T, issues []Issue, rule string, context string) {
	t.Helper()
	for _, issue := range issues {
		if issue.Rule == rule {
			return
		}
	}
	t.Errorf("%s: expected rule '%s' to trigger, but it didn't", context, rule)
}

// Helper to assert specific rule did NOT trigger
func assertNoRule(t *testing.T, issues []Issue, rule string, context string) {
	t.Helper()
	for _, issue := range issues {
		if issue.Rule == rule {
			t.Errorf("%s: rule '%s' should NOT trigger, but it did at line %d", context, rule, issue.Line)
		}
	}
}

// ============================================================================
// EVAL/EXEC DETECTION - The False Positive Bug We Fixed
// ============================================================================

func TestEval_TruePositives(t *testing.T) {
	tests := []struct {
		name string
		code string
	}{
		{"simple eval", `result = eval("1+1")`},
		{"eval with variable", `x = eval(user_input)`},
		{"exec call", `exec("print('hello')")`},
		{"eval in function", `def f(): return eval(x)`},
		{"eval after assignment", `x = 1; y = eval(z)`},
		{"eval with complex expr", `result = eval(compile(source, '<string>', 'eval'))`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			issues := checkCode(t, "test.py", tt.code)
			assertHasRule(t, issues, "ban-eval", tt.name)
		})
	}
}

func TestEval_FalsePositives(t *testing.T) {
	tests := []struct {
		name string
		code string
	}{
		{"eval in string", `msg = "Never use eval() in production"`},
		{"eval in comment", `# This checks for eval() usage`},
		{"eval in docstring", `"""Check if code contains eval()"""`},
		{"eval in f-string", `error = f"Found eval() at line {n}"`},
		{"eval in contains check", `if "eval(" in code:`},
		{"eval in error message", `raise Error("Don't use eval()")`},
		{"eval in print", `print("Warning: eval() is dangerous")`},
		{"eval in log", `logger.warning("Detected eval() usage")`},
		{"eval in dict", `rules = {"eval": "forbidden"}`},
		{"eval in list", `banned = ["eval(", "exec("]`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			issues := checkCode(t, "test.py", tt.code)
			assertNoRule(t, issues, "ban-eval", tt.name)
		})
	}
}

func TestEval_EdgeCases(t *testing.T) {
	// Mixed - has both real eval AND mentions in strings
	code := `
"""This module checks for eval() usage."""
# Comment mentioning eval()
HELP = "Never use eval() in production"

def check_for_eval(code):
    if "eval(" in code:
        return True
    return False

# This IS actually using eval
result = eval("1 + 1")
`
	issues := checkCode(t, "test.py", code)

	// Should have exactly 1 ban-eval issue (the real one on last line)
	evalIssues := []Issue{}
	for _, issue := range issues {
		if issue.Rule == "ban-eval" {
			evalIssues = append(evalIssues, issue)
		}
	}

	if len(evalIssues) != 1 {
		t.Errorf("expected exactly 1 eval issue for real usage, got %d", len(evalIssues))
		for _, issue := range evalIssues {
			t.Logf("  line %d: %s", issue.Line, issue.Message)
		}
	}
}

// ============================================================================
// SQL INJECTION DETECTION
// ============================================================================

func TestSQLInjection_TruePositives(t *testing.T) {
	tests := []struct {
		name string
		code string
	}{
		{"f-string SELECT", `query = f"SELECT * FROM users WHERE id = {user_id}"`},
		{"f-string INSERT", `query = f"INSERT INTO users VALUES ({name})"`},
		{"f-string UPDATE", `query = f"UPDATE users SET name = {name}"`},
		{"f-string DELETE", `query = f"DELETE FROM users WHERE id = {id}"`},
		{"single quote f-string", `query = f'SELECT * FROM users WHERE id = {user_id}'`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			issues := checkCode(t, "test.py", tt.code)
			assertHasRule(t, issues, "sql-injection", tt.name)
		})
	}
}

func TestSQLInjection_FalsePositives(t *testing.T) {
	tests := []struct {
		name string
		code string
	}{
		{"parameterized query", `cursor.execute("SELECT * FROM users WHERE id = ?", (user_id,))`},
		{"named params", `cursor.execute("SELECT * FROM users WHERE id = :id", {"id": user_id})`},
		{"regular string", `query = "SELECT * FROM users WHERE id = 1"`},
		{"comment with SQL", `# Never use f"SELECT..." - use parameterized queries`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			issues := checkCode(t, "test.py", tt.code)
			assertNoRule(t, issues, "sql-injection", tt.name)
		})
	}
}

// ============================================================================
// SECRET PATTERN DETECTION
// ============================================================================

func TestSecrets_TruePositives(t *testing.T) {
	tests := []struct {
		name string
		code string
	}{
		{"api_key assignment", `api_key = "sk-12345abcdef"`},
		{"password assignment", `password = "supersecret123"`},
		{"secret assignment", `secret = "my_secret_value"`},
		{"AWS_SECRET in code", `AWS_SECRET = "AKIAIOSFODNN7EXAMPLE"`},
		{"PRIVATE_KEY reference", `PRIVATE_KEY = "-----BEGIN RSA PRIVATE KEY-----"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			issues := checkCode(t, "test.py", tt.code)
			assertHasRule(t, issues, "secret-pattern", tt.name)
		})
	}
}

func TestSecrets_FalsePositives(t *testing.T) {
	tests := []struct {
		name string
		code string
	}{
		{"env var lookup", `api_key = os.environ.get("API_KEY")`},
		{"empty password", `password = ""`},
		{"placeholder", `secret = os.getenv("SECRET")`},
		{"comment about secrets", `# api_key = "never hardcode this"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			issues := checkCode(t, "test.py", tt.code)
			assertNoRule(t, issues, "secret-pattern", tt.name)
		})
	}
}

// ============================================================================
// DANGEROUS COMMANDS
// ============================================================================

func TestDangerousCommands_TruePositives(t *testing.T) {
	tests := []struct {
		name string
		code string
	}{
		{"rm -rf", `os.system("rm -rf /")`},
		{"DROP TABLE", `cursor.execute("DROP TABLE users")`},
		{"DROP DATABASE", `cursor.execute("DROP DATABASE production")`},
		{"DELETE FROM without WHERE", `cursor.execute("DELETE FROM users;")`},
		{"TRUNCATE TABLE", `cursor.execute("TRUNCATE TABLE logs")`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			issues := checkCode(t, "test.py", tt.code)
			assertHasRule(t, issues, "dangerous-cmd", tt.name)
		})
	}
}

func TestDangerousCommands_FalsePositives(t *testing.T) {
	tests := []struct {
		name string
		code string
	}{
		{"DELETE with WHERE", `cursor.execute("DELETE FROM users WHERE expired = true")`},
		{"comment about rm", `# Never run rm -rf without checking first`},
		{"safe removal", `os.remove("temp.txt")`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			issues := checkCode(t, "test.py", tt.code)
			assertNoRule(t, issues, "dangerous-cmd", tt.name)
		})
	}
}

// ============================================================================
// PRINT/CONSOLE STATEMENTS
// ============================================================================

func TestPrint_TruePositives(t *testing.T) {
	tests := []struct {
		name string
		code string
	}{
		{"simple print", `print("hello")`},
		{"print with variable", `print(x)`},
		{"print in function", `def f(): print("debug")`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			issues := checkCode(t, "test.py", tt.code)
			assertHasRule(t, issues, "ban-print", tt.name)
		})
	}
}

func TestPrint_FalsePositives(t *testing.T) {
	tests := []struct {
		name string
		code string
	}{
		{"commented print", `# print("debug")`},
		{"blueprint (Flask)", `app.register_blueprint(api)`},
		{"fingerprint variable", `fingerprint = hash(data)`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			issues := checkCode(t, "test.py", tt.code)
			assertNoRule(t, issues, "ban-print", tt.name)
		})
	}
}

func TestConsoleLog_TruePositives(t *testing.T) {
	tests := []struct {
		name string
		code string
	}{
		{"simple console.log", `console.log("hello")`},
		{"console.log with var", `console.log(data)`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			issues := checkCode(t, "test.ts", tt.code)
			assertHasRule(t, issues, "ban-console", tt.name)
		})
	}
}

func TestConsoleLog_FalsePositives(t *testing.T) {
	tests := []struct {
		name string
		code string
	}{
		{"commented console.log", `// console.log("debug")`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			issues := checkCode(t, "test.ts", tt.code)
			assertNoRule(t, issues, "ban-console", tt.name)
		})
	}
}

// ============================================================================
// BARE EXCEPT
// ============================================================================

func TestBareExcept_TruePositives(t *testing.T) {
	tests := []struct {
		name string
		code string
	}{
		{"bare except", "try:\n    pass\nexcept:\n    pass"},
		{"except with space", "try:\n    pass\nexcept :\n    pass"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			issues := checkCode(t, "test.py", tt.code)
			assertHasRule(t, issues, "ban-except", tt.name)
		})
	}
}

func TestBareExcept_FalsePositives(t *testing.T) {
	tests := []struct {
		name string
		code string
	}{
		{"specific exception", "try:\n    pass\nexcept ValueError:\n    pass"},
		{"multiple exceptions", "try:\n    pass\nexcept (ValueError, TypeError):\n    pass"},
		{"exception as e", "try:\n    pass\nexcept Exception as e:\n    pass"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			issues := checkCode(t, "test.py", tt.code)
			assertNoRule(t, issues, "ban-except", tt.name)
		})
	}
}

// ============================================================================
// STAR IMPORTS
// ============================================================================

func TestStarImport_TruePositives(t *testing.T) {
	tests := []struct {
		name string
		code string
	}{
		{"from module import *", `from os import *`},
		{"from package import *", `from package.module import *`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			issues := checkCode(t, "test.py", tt.code)
			assertHasRule(t, issues, "ban-star", tt.name)
		})
	}
}

func TestStarImport_FalsePositives(t *testing.T) {
	tests := []struct {
		name string
		code string
	}{
		{"import *", `import os`},
		{"from import specific", `from os import path`},
		{"from import multiple", `from os import path, getcwd`},
		{"comment", `# from os import *`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			issues := checkCode(t, "test.py", tt.code)
			assertNoRule(t, issues, "ban-star", tt.name)
		})
	}
}

// ============================================================================
// SUBPROCESS SHELL=TRUE
// ============================================================================

func TestSubprocessShell_TruePositives(t *testing.T) {
	tests := []struct {
		name string
		code string
	}{
		{"shell=True", `subprocess.run(cmd, shell=True)`},
		{"Popen shell=True", `subprocess.Popen(cmd, shell=True)`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			issues := checkCode(t, "test.py", tt.code)
			assertHasRule(t, issues, "subprocess-shell", tt.name)
		})
	}
}

func TestSubprocessShell_FalsePositives(t *testing.T) {
	tests := []struct {
		name string
		code string
	}{
		{"shell=False", `subprocess.run(cmd, shell=False)`},
		{"no shell", `subprocess.run(["ls", "-la"])`},
		{"comment", `# shell=True is dangerous`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			issues := checkCode(t, "test.py", tt.code)
			assertNoRule(t, issues, "subprocess-shell", tt.name)
		})
	}
}

// ============================================================================
// TODO/FIXME MARKERS
// ============================================================================

func TestTodoMarkers_TruePositives(t *testing.T) {
	tests := []struct {
		name string
		code string
	}{
		{"TODO", `# TODO: fix this later`},
		{"FIXME", `# FIXME: broken code`},
		{"HACK", `# HACK: temporary workaround`},
		{"lowercase todo", `# todo: needs attention`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			issues := checkCode(t, "test.py", tt.code)
			assertHasRule(t, issues, "todo-marker", tt.name)
		})
	}
}

// ============================================================================
// MOCK DATA PATTERNS
// ============================================================================

func TestMockData_TruePositives(t *testing.T) {
	tests := []struct {
		name string
		code string
	}{
		{"test@example.com", `email = "test@example.com"`},
		{"fake_user", `user = fake_user`},
		{"mock_data", `data = mock_data`},
		{"placeholder", `value = "placeholder"`},
		{"changeme", `password = "changeme"`},
		{"test_user", `user = test_user`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			issues := checkCode(t, "test.py", tt.code)
			assertHasRule(t, issues, "mock-data", tt.name)
		})
	}
}

// ============================================================================
// FILE SIZE CHECK
// ============================================================================

func TestFileSize_TruePositive(t *testing.T) {
	// Create a file with 501+ lines
	var lines []string
	for i := 0; i < 510; i++ {
		lines = append(lines, "x = 1")
	}
	code := strings.Join(lines, "\n")

	issues := checkCode(t, "test.py", code)
	assertHasRule(t, issues, "file-size", "large file")
}

func TestFileSize_FalsePositive(t *testing.T) {
	// Create a file with exactly 500 lines
	var lines []string
	for i := 0; i < 500; i++ {
		lines = append(lines, "x = 1")
	}
	code := strings.Join(lines, "\n")

	issues := checkCode(t, "test.py", code)
	assertNoRule(t, issues, "file-size", "file at limit")
}

// ============================================================================
// EDGE CASES
// ============================================================================

func TestEdgeCases_EmptyFile(t *testing.T) {
	issues := checkCode(t, "empty.py", "")
	assertIssueCount(t, issues, 0, "empty file")
}

func TestEdgeCases_OnlyComments(t *testing.T) {
	code := `# This is a comment
# Another comment
# eval() exec() print() - all in comments
`
	issues := checkCode(t, "comments.py", code)
	// Should only have todo-marker issues if any, no security issues
	assertNoRule(t, issues, "ban-eval", "only comments")
	assertNoRule(t, issues, "ban-print", "only comments")
}

func TestEdgeCases_OnlyDocstring(t *testing.T) {
	code := `"""
This module uses eval() and exec() for dynamic code.
Never use eval() in production.
Warning: print() statements should be removed.
"""`
	issues := checkCode(t, "docstring.py", code)
	assertNoRule(t, issues, "ban-eval", "only docstring")
	assertNoRule(t, issues, "ban-print", "only docstring")
}

func TestEdgeCases_MixedQuotes(t *testing.T) {
	// Test various quote scenarios - all should NOT trigger (eval is in strings)
	tests := []struct {
		name string
		code string
	}{
		{"apostrophe in double quotes", `msg = "Don't use eval() here"`},
		{"double in single quotes", `msg = 'He said "use eval()" never'`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			issues := checkCode(t, "test.py", tt.code)
			assertNoRule(t, issues, "ban-eval", tt.name)
		})
	}

	// This one SHOULD trigger - real eval call
	t.Run("real eval", func(t *testing.T) {
		issues := checkCode(t, "test.py", `result = eval("1+1")`)
		assertHasRule(t, issues, "ban-eval", "real eval")
	})
}

func TestEdgeCases_NestedQuotes(t *testing.T) {
	code := `msg = "He said 'don't use eval()' yesterday"
warning = 'She said "eval() is dangerous"'`

	issues := checkCode(t, "nested.py", code)
	assertNoRule(t, issues, "ban-eval", "nested quotes")
}

func TestEdgeCases_UnicodeContent(t *testing.T) {
	code := `# Привет мир
message = "こんにちは"
eval("1+1")  # Should still catch this`

	issues := checkCode(t, "unicode.py", code)
	assertHasRule(t, issues, "ban-eval", "unicode content")
}

func TestEdgeCases_VeryLongLine(t *testing.T) {
	// Very long line with proper eval call (preceded by =)
	// Note: "aaaa...eval()" correctly does NOT trigger because eval needs
	// to be preceded by operator/space (avoids false positives like "retrieval()")
	longLine := strings.Repeat("x = 1; ", 700) + `result = eval("x")`
	issues := checkCode(t, "long.py", longLine)
	// Should still catch eval even on very long lines
	assertHasRule(t, issues, "ban-eval", "very long line")
}

// ============================================================================
// FILE TYPE HANDLING
// ============================================================================

func TestFileTypes_Python(t *testing.T) {
	issues := checkCode(t, "test.py", `print("hello")`)
	assertHasRule(t, issues, "ban-print", "python file")
}

func TestFileTypes_JavaScript(t *testing.T) {
	issues := checkCode(t, "test.js", `console.log("hello")`)
	assertHasRule(t, issues, "ban-console", "javascript file")
}

func TestFileTypes_TypeScript(t *testing.T) {
	issues := checkCode(t, "test.ts", `console.log("hello")`)
	assertHasRule(t, issues, "ban-console", "typescript file")
}

func TestFileTypes_TSX(t *testing.T) {
	issues := checkCode(t, "test.tsx", `console.log("hello")`)
	assertHasRule(t, issues, "ban-console", "tsx file")
}

func TestFileTypes_Unsupported(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.rb")
	os.WriteFile(path, []byte(`puts "hello"`), 0644)

	issues := checkFile(path)
	assertIssueCount(t, issues, 0, "unsupported file type")
}

// ============================================================================
// DIRECTORY WALKING (RunAll and DryRun)
// ============================================================================

func TestRunAll_ExcludesGitDir(t *testing.T) {
	dir := t.TempDir()

	// Create .git directory with a Python file
	gitDir := filepath.Join(dir, ".git")
	os.MkdirAll(gitDir, 0755)
	os.WriteFile(filepath.Join(gitDir, "hooks.py"), []byte(`eval("bad")`), 0644)

	// Create a regular Python file
	os.WriteFile(filepath.Join(dir, "main.py"), []byte(`print("hello")`), 0644)

	issues := RunAll(dir)

	// Should only find issue in main.py, not in .git
	for _, issue := range issues {
		if strings.Contains(issue.File, ".git") {
			t.Error("RunAll should exclude .git directory")
		}
	}
}

func TestRunAll_ExcludesNodeModules(t *testing.T) {
	dir := t.TempDir()

	nmDir := filepath.Join(dir, "node_modules")
	os.MkdirAll(nmDir, 0755)
	os.WriteFile(filepath.Join(nmDir, "lib.js"), []byte(`console.log("x")`), 0644)

	issues := RunAll(dir)

	for _, issue := range issues {
		if strings.Contains(issue.File, "node_modules") {
			t.Error("RunAll should exclude node_modules directory")
		}
	}
}

func TestRunAll_ExcludesVenv(t *testing.T) {
	dir := t.TempDir()

	for _, venv := range []string{"venv", ".venv"} {
		venvDir := filepath.Join(dir, venv)
		os.MkdirAll(venvDir, 0755)
		os.WriteFile(filepath.Join(venvDir, "activate.py"), []byte(`exec("bad")`), 0644)
	}

	issues := RunAll(dir)

	for _, issue := range issues {
		if strings.Contains(issue.File, "venv") {
			t.Error("RunAll should exclude venv directories")
		}
	}
}

func TestDryRun_CountsFiles(t *testing.T) {
	dir := t.TempDir()

	os.WriteFile(filepath.Join(dir, "a.py"), []byte("x=1\ny=2\nz=3"), 0644)
	os.WriteFile(filepath.Join(dir, "b.py"), []byte("x=1"), 0644)
	os.WriteFile(filepath.Join(dir, "c.go"), []byte("x=1"), 0644) // Should be excluded

	info := DryRun(dir)

	if info.FileCount != 2 {
		t.Errorf("expected 2 files, got %d", info.FileCount)
	}
}

func TestDryRun_CountsLines(t *testing.T) {
	dir := t.TempDir()

	os.WriteFile(filepath.Join(dir, "a.py"), []byte("x=1\ny=2\nz=3"), 0644) // 3 lines
	os.WriteFile(filepath.Join(dir, "b.py"), []byte("x=1"), 0644)           // 1 line

	info := DryRun(dir)

	if info.TotalLines != 4 {
		t.Errorf("expected 4 total lines, got %d", info.TotalLines)
	}
}

// ============================================================================
// OUTPUT PARSING
// ============================================================================

func TestParseGuardianOutput_StandardFormat(t *testing.T) {
	output := `main.py:10 [ban-eval] Avoid eval() - security risk
utils.py:25 [sql-injection] f-string in SQL query - use parameterized queries`

	issues := parseGuardianOutput(output)

	if len(issues) != 2 {
		t.Errorf("expected 2 issues, got %d", len(issues))
	}

	if issues[0].File != "main.py" || issues[0].Line != 10 || issues[0].Rule != "ban-eval" {
		t.Errorf("first issue parsed incorrectly: %+v", issues[0])
	}
}

func TestParseGuardianOutput_FailFormat(t *testing.T) {
	output := `FAIL main.py:10 - ban-eval: Avoid eval() - security risk`

	issues := parseGuardianOutput(output)

	if len(issues) != 1 {
		t.Errorf("expected 1 issue, got %d", len(issues))
	}

	if issues[0].File != "main.py" || issues[0].Line != 10 {
		t.Errorf("issue parsed incorrectly: %+v", issues[0])
	}
}

func TestParseGuardianOutput_IgnoresNonIssueLines(t *testing.T) {
	output := `Running checks...
Scanning 10 files...
main.py:10 [ban-eval] Avoid eval()
Done!`

	issues := parseGuardianOutput(output)

	if len(issues) != 1 {
		t.Errorf("expected 1 issue, got %d", len(issues))
	}
}

// ============================================================================
// SEVERITY CLASSIFICATION
// ============================================================================

func TestGetSeverity_Critical(t *testing.T) {
	criticalRules := []string{"ban-eval", "dangerous-cmd", "secret-pattern", "sql-injection"}

	for _, rule := range criticalRules {
		if sev := getSeverity(rule); sev != "critical" {
			t.Errorf("%s should be critical, got %s", rule, sev)
		}
	}
}

func TestGetSeverity_Info(t *testing.T) {
	infoRules := []string{"ban-print", "ban-console", "todo-marker"}

	for _, rule := range infoRules {
		if sev := getSeverity(rule); sev != "info" {
			t.Errorf("%s should be info, got %s", rule, sev)
		}
	}
}

func TestGetSeverity_Warning(t *testing.T) {
	warningRules := []string{"ban-except", "ban-star", "mock-data", "file-size", "subprocess-shell"}

	for _, rule := range warningRules {
		if sev := getSeverity(rule); sev != "warning" {
			t.Errorf("%s should be warning, got %s", rule, sev)
		}
	}
}
