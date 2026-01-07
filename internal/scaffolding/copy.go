package scaffolding

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

//go:embed files/*
var scaffoldingFiles embed.FS

// InstallConfig holds configuration for installation
type InstallConfig struct {
	Language    string   // python, typescript, go, php
	Stack       string   // python-fastapi, typescript-react, etc.
	SourceDir   string   // src/
	ExcludeDirs []string // tests/, __pycache__/, etc.
}

// Install copies scaffolding files to the target directory
func Install(config InstallConfig) error {
	guardianDir := ".guardian"

	// Track if we created the directory (for cleanup on error)
	createdDir := false
	if _, err := os.Stat(guardianDir); os.IsNotExist(err) {
		createdDir = true
	}

	// Create .guardian directory
	if err := os.MkdirAll(guardianDir, 0755); err != nil {
		return fmt.Errorf("failed to create .guardian directory: %w", err)
	}

	// Cleanup on error - remove partially created files
	cleanup := func() {
		if createdDir {
			os.RemoveAll(guardianDir)
		}
		os.Remove("guardian_config.toml")
	}

	// Copy language-specific files
	srcDir := filepath.Join("files", config.Language)
	files, err := scaffoldingFiles.ReadDir(srcDir)
	if err != nil {
		// Fall back to generating files in-memory
		if err := generateFiles(config); err != nil {
			cleanup()
			return err
		}
		return nil
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		srcPath := filepath.Join(srcDir, file.Name())
		content, err := scaffoldingFiles.ReadFile(srcPath)
		if err != nil {
			continue
		}

		// Determine destination
		destPath := file.Name()
		if strings.HasPrefix(destPath, "check_") || destPath == "guardian.py" || destPath == "guardian.js" {
			destPath = filepath.Join(guardianDir, destPath)
		}

		if err := os.WriteFile(destPath, content, 0644); err != nil {
			cleanup()
			return fmt.Errorf("failed to write %s: %w", destPath, err)
		}
	}

	// Generate config file
	if err := generateConfig(config); err != nil {
		cleanup()
		return err
	}

	// Generate/update pre-commit config
	if err := generatePreCommitConfig(config); err != nil {
		cleanup()
		return err
	}

	return nil
}

// generateFiles generates scaffolding files in-memory (when embeds aren't available)
func generateFiles(config InstallConfig) error {
	var err error
	switch config.Language {
	case "python":
		err = generatePythonFiles(config)
	case "typescript":
		err = generateTypeScriptFiles(config)
	case "go":
		err = generateGoFiles(config)
	default:
		err = generatePythonFiles(config) // Default to Python
	}

	if err != nil {
		return err
	}

	// Generate config file
	if err := generateConfig(config); err != nil {
		return err
	}

	// Generate/update pre-commit config
	if err := generatePreCommitConfig(config); err != nil {
		return err
	}

	return nil
}

func generatePythonFiles(config InstallConfig) error {
	files := map[string]string{
		".guardian/check_file_size.py":        pythonCheckFileSize,
		".guardian/check_function_size.py":    pythonCheckFunctionSize,
		".guardian/check_dangerous.py":        pythonCheckDangerous,
		".guardian/check_mock_data.py":        pythonCheckMockData,
		".guardian/check_security.py":         pythonCheckSecurity,
		".guardian/check_star_imports.py":     pythonCheckStarImports,
		".guardian/check_mutable_defaults.py": pythonCheckMutableDefaults,
		".guardian/check_todo_markers.py":     pythonCheckTodoMarkers,
		".guardian/check_subprocess_shell.py": pythonCheckSubprocessShell,
		".guardian/check_bare_except.py":      pythonCheckBareExcept,
		".guardian/guardian.py":               pythonGuardian,
	}

	for path, content := range files {
		dir := filepath.Dir(path)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
		if err := os.WriteFile(path, []byte(content), 0755); err != nil {
			return err
		}
	}

	return nil
}

func generateTypeScriptFiles(config InstallConfig) error {
	// Note: function size check removed - brace counting doesn't work for arrow functions,
	// class methods, or functions with object literals. Needs proper AST parsing.
	files := map[string]string{
		".guardian/check_file_size.js":   tsCheckFileSize,
		".guardian/check_dangerous.js":   tsCheckDangerous,
		".guardian/check_mock_data.js":   tsCheckMockData,
		".guardian/check_console_log.js": tsCheckConsoleLog,
		".guardian/guardian.js":          tsGuardian,
	}

	for path, content := range files {
		dir := filepath.Dir(path)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
		if err := os.WriteFile(path, []byte(content), 0755); err != nil {
			return err
		}
	}

	return nil
}

func generateGoFiles(config InstallConfig) error {
	// Go projects typically use go vet, staticcheck, etc.
	// We'll generate a simple wrapper
	files := map[string]string{
		".guardian/guardian.sh": goGuardianScript,
	}

	for path, content := range files {
		dir := filepath.Dir(path)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
		if err := os.WriteFile(path, []byte(content), 0755); err != nil {
			return err
		}
	}

	return nil
}

func generateConfig(config InstallConfig) error {
	// Clean up exclude dirs
	excludes := []string{}
	for _, dir := range config.ExcludeDirs {
		dir = strings.TrimSpace(dir)
		dir = strings.Trim(dir, "/")
		if dir != "" {
			excludes = append(excludes, dir)
		}
	}

	content := fmt.Sprintf(`# Guardian Configuration
# Stop AI slop before it hits your codebase.

[project]
src_root = "%s"
exclude_dirs = [%s]

[limits]
max_file_lines = 500
max_function_lines = 50

[limits.custom_file_limits]
# "some/big/file.py" = 700

[quality]
ban_print = true
ban_bare_except = true
ban_mutable_defaults = true
ban_star_imports = true
ban_todo_markers = true

# Mock/fake data detection
ban_mock_data = true
mock_patterns = [
    "mock_", "_mock", "fake_", "_fake", "dummy_", "_dummy",
    "test_user", "test_email", "test_password",
    "example@", "@example.com", "@test.com",
    "placeholder", "sample_", "hardcoded",
    "changeme", "replace_me", "your_", "xxx",
    "lorem ipsum", "foo_bar", "asdf",
]

[security]
ban_eval_exec = true
ban_subprocess_shell = true
ban_dangerous_commands = true
dangerous_patterns = [
    "rm -rf",
    "DROP TABLE",
    "DROP DATABASE",
    "DELETE FROM",
    "TRUNCATE TABLE",
]
secret_patterns = [
    "api_key", "apikey", "api-key",
    "secret", "password", "passwd",
    "private_key", "privatekey",
    "access_token", "auth_token",
]
`, strings.TrimSuffix(config.SourceDir, "/"), formatExcludes(excludes))

	return os.WriteFile("guardian_config.toml", []byte(content), 0644)
}

func formatExcludes(excludes []string) string {
	if len(excludes) == 0 {
		return ""
	}
	quoted := make([]string, len(excludes))
	for i, e := range excludes {
		quoted[i] = fmt.Sprintf(`"%s"`, e)
	}
	return strings.Join(quoted, ", ")
}

func generatePreCommitConfig(config InstallConfig) error {
	// Check if .pre-commit-config.yaml exists
	existingContent := ""
	if data, err := os.ReadFile(".pre-commit-config.yaml"); err == nil {
		existingContent = string(data)
	}

	// If Guardian hooks already present, skip
	if strings.Contains(existingContent, "guardian") {
		return nil
	}

	guardianHook := ""
	switch config.Language {
	case "python":
		guardianHook = `
  - repo: local
    hooks:
      - id: guardian-file-size
        name: Check file size
        entry: python .guardian/check_file_size.py
        language: python
        types: [python]
      - id: guardian-function-size
        name: Check function size
        entry: python .guardian/check_function_size.py
        language: python
        types: [python]
      - id: guardian-dangerous
        name: Check dangerous patterns
        entry: python .guardian/check_dangerous.py
        language: python
        types: [python]
      - id: guardian-mock-data
        name: Check mock data
        entry: python .guardian/check_mock_data.py
        language: python
        types: [python]
      - id: guardian-security
        name: Check security patterns
        entry: python .guardian/check_security.py
        language: python
        types: [python]
      - id: guardian-star-imports
        name: Check star imports
        entry: python .guardian/check_star_imports.py
        language: python
        types: [python]
      - id: guardian-mutable-defaults
        name: Check mutable defaults
        entry: python .guardian/check_mutable_defaults.py
        language: python
        types: [python]
      - id: guardian-todo-markers
        name: Check TODO markers
        entry: python .guardian/check_todo_markers.py
        language: python
        types: [python]
      - id: guardian-subprocess-shell
        name: Check subprocess shell
        entry: python .guardian/check_subprocess_shell.py
        language: python
        types: [python]
      - id: guardian-bare-except
        name: Check bare except
        entry: python .guardian/check_bare_except.py
        language: python
        types: [python]
`
	case "typescript":
		guardianHook = `
  - repo: local
    hooks:
      - id: guardian
        name: Guardian checks
        entry: node .guardian/guardian.js
        language: node
        types: [javascript, jsx, typescript, tsx]
`
	default:
		guardianHook = `
  - repo: local
    hooks:
      - id: guardian
        name: Guardian checks
        entry: .guardian/guardian.sh
        language: script
`
	}

	if existingContent == "" {
		// Create new file
		content := `repos:` + guardianHook
		return os.WriteFile(".pre-commit-config.yaml", []byte(content), 0644)
	}

	// Append to existing - ensure newline before our hooks
	newContent := strings.TrimRight(existingContent, "\n") + "\n" + guardianHook
	return os.WriteFile(".pre-commit-config.yaml", []byte(newContent), 0644)
}

// Python check scripts
const pythonCheckFileSize = `#!/usr/bin/env python3
"""Check that Python files don't exceed line limits."""

import sys
from pathlib import Path

MAX_LINES = 500

def main() -> int:
    if len(sys.argv) < 2:
        return 0

    failed = False
    for filepath in sys.argv[1:]:
        path = Path(filepath)
        if not path.exists() or path.suffix != ".py":
            continue

        lines = path.read_text().splitlines()
        if len(lines) > MAX_LINES:
            print(f"{filepath}:1 [file-size] File has {len(lines)} lines (max {MAX_LINES})")
            failed = True

    return 1 if failed else 0

if __name__ == "__main__":
    sys.exit(main())
`

const pythonCheckFunctionSize = `#!/usr/bin/env python3
"""Check that functions don't exceed line limits."""

import ast
import sys
from pathlib import Path

MAX_LINES = 50

def main() -> int:
    if len(sys.argv) < 2:
        return 0

    failed = False
    for filepath in sys.argv[1:]:
        path = Path(filepath)
        if not path.exists() or path.suffix != ".py":
            continue

        try:
            tree = ast.parse(path.read_text())
        except SyntaxError:
            continue

        for node in ast.walk(tree):
            if isinstance(node, (ast.FunctionDef, ast.AsyncFunctionDef)):
                lines = (node.end_lineno or node.lineno) - node.lineno + 1
                if lines > MAX_LINES:
                    print(f"{filepath}:{node.lineno} [func-size] {node.name}() has {lines} lines (max {MAX_LINES})")
                    failed = True

    return 1 if failed else 0

if __name__ == "__main__":
    sys.exit(main())
`

const pythonCheckDangerous = `#!/usr/bin/env python3
"""Check for dangerous commands and patterns."""

import re
import sys
from pathlib import Path

DANGEROUS_PATTERNS = [
    (r"rm\s+-rf", "rm -rf command"),
    (r"DROP\s+TABLE", "DROP TABLE statement"),
    (r"DROP\s+DATABASE", "DROP DATABASE statement"),
    (r"DELETE\s+FROM\s+\w+\s*;", "DELETE without WHERE"),
    (r"TRUNCATE\s+TABLE", "TRUNCATE TABLE statement"),
    (r"shutil\.rmtree", "shutil.rmtree call"),
    (r"os\.remove", "os.remove call"),
]

def main() -> int:
    if len(sys.argv) < 2:
        return 0

    failed = False
    for filepath in sys.argv[1:]:
        path = Path(filepath)
        if not path.exists():
            continue

        content = path.read_text()
        for i, line in enumerate(content.splitlines(), 1):
            for pattern, desc in DANGEROUS_PATTERNS:
                if re.search(pattern, line, re.IGNORECASE):
                    print(f"{filepath}:{i} [dangerous-cmd] {desc} detected")
                    failed = True

    return 1 if failed else 0

if __name__ == "__main__":
    sys.exit(main())
`

const pythonCheckMockData = `#!/usr/bin/env python3
"""Check for mock/test data in production code."""

import re
import sys
from pathlib import Path

MOCK_PATTERNS = [
    r"test@example\.com",
    r"example@",
    r"@test\.com",
    r"fake_\w+",
    r"\w+_fake",
    r"mock_\w+",
    r"\w+_mock",
    r"dummy_\w+",
    r"placeholder",
    r"test_user",
    r"test_password",
    r"changeme",
    r"your_\w+_here",
    r"lorem\s+ipsum",
]

def main() -> int:
    if len(sys.argv) < 2:
        return 0

    failed = False
    for filepath in sys.argv[1:]:
        path = Path(filepath)
        if not path.exists() or path.suffix != ".py":
            continue

        # Skip test files
        if "test" in filepath.lower():
            continue

        content = path.read_text()
        for i, line in enumerate(content.splitlines(), 1):
            # Skip comments
            if line.strip().startswith("#"):
                continue

            for pattern in MOCK_PATTERNS:
                if re.search(pattern, line, re.IGNORECASE):
                    print(f"{filepath}:{i} [mock-data] Possible test/mock data detected")
                    failed = True
                    break

    return 1 if failed else 0

if __name__ == "__main__":
    sys.exit(main())
`

const pythonCheckSecurity = `#!/usr/bin/env python3
"""Check for security issues using AST parsing."""

import ast
import re
import sys
from pathlib import Path

def check_eval_exec(tree, filepath):
    """Check for eval/exec calls using AST - no false positives from strings."""
    issues = []
    for node in ast.walk(tree):
        if isinstance(node, ast.Call):
            # Check for eval() or exec() calls
            if isinstance(node.func, ast.Name) and node.func.id in ("eval", "exec"):
                issues.append(f"{filepath}:{node.lineno} [ban-eval] Avoid {node.func.id}() - security risk")
    return issues

def check_subprocess_shell(tree, filepath):
    """Check for subprocess with shell=True using AST."""
    issues = []
    for node in ast.walk(tree):
        if isinstance(node, ast.Call):
            # Check for subprocess.* calls
            if isinstance(node.func, ast.Attribute):
                if isinstance(node.func.value, ast.Name) and node.func.value.id == "subprocess":
                    for kw in node.keywords:
                        if kw.arg == "shell":
                            if isinstance(kw.value, ast.Constant) and kw.value.value is True:
                                issues.append(f"{filepath}:{node.lineno} [subprocess-shell] Avoid shell=True - security risk")
    return issues

def main() -> int:
    if len(sys.argv) < 2:
        return 0

    failed = False
    for filepath in sys.argv[1:]:
        path = Path(filepath)
        if not path.exists() or path.suffix != ".py":
            continue

        content = path.read_text()

        # Parse AST for accurate detection
        try:
            tree = ast.parse(content)
        except SyntaxError:
            continue

        # AST-based checks (no false positives)
        for issue in check_eval_exec(tree, filepath):
            print(issue)
            failed = True

        for issue in check_subprocess_shell(tree, filepath):
            print(issue)
            failed = True

        # Line-based checks for patterns that can't use AST
        for i, line in enumerate(content.splitlines(), 1):
            # Skip comments
            stripped = line.strip()
            if stripped.startswith("#"):
                continue

            # Skip if line is inside a string (basic heuristic)
            quote_count = line.count('"') + line.count("'")

            # SQL injection - catch f-strings with SQL keywords
            if re.search(r'f["\'](?:SELECT|INSERT|UPDATE|DELETE)', line, re.IGNORECASE):
                print(f"{filepath}:{i} [sql-injection] f-string in SQL query - use parameterized queries")
                failed = True

            # Hardcoded secrets - only if it looks like an assignment, not in a string
            if "=" in line and quote_count <= 2:
                secret_patterns = [
                    r'^[^#]*\bapi_key\s*=\s*["\'][^"\']+["\']',
                    r'^[^#]*\bpassword\s*=\s*["\'][^"\']+["\']',
                    r'^[^#]*\bsecret\s*=\s*["\'][^"\']+["\']',
                ]
                for pattern in secret_patterns:
                    if re.search(pattern, line, re.IGNORECASE):
                        print(f"{filepath}:{i} [secret-pattern] Possible hardcoded secret - use environment variables")
                        failed = True
                        break

    return 1 if failed else 0

if __name__ == "__main__":
    sys.exit(main())
`

const pythonGuardian = `#!/usr/bin/env python3
"""Guardian - Main entry point for all checks."""

import subprocess
import sys
from pathlib import Path

CHECKS = [
    "check_file_size.py",
    "check_function_size.py",
    "check_dangerous.py",
    "check_mock_data.py",
    "check_security.py",
    "check_star_imports.py",
    "check_mutable_defaults.py",
    "check_todo_markers.py",
    "check_subprocess_shell.py",
    "check_bare_except.py",
]

def main() -> int:
    guardian_dir = Path(__file__).parent
    files = sys.argv[1:] if len(sys.argv) > 1 else []

    if not files:
        # Find all Python files
        files = [str(p) for p in Path(".").rglob("*.py") if ".guardian" not in str(p)]

    failed = False
    for check in CHECKS:
        check_path = guardian_dir / check
        if not check_path.exists():
            continue

        result = subprocess.run(
            ["python3", str(check_path)] + files,
            capture_output=False
        )
        if result.returncode != 0:
            failed = True

    return 1 if failed else 0

if __name__ == "__main__":
    sys.exit(main())
`

// TypeScript check scripts
const tsCheckFileSize = `#!/usr/bin/env node
/**
 * Check that files don't exceed line limits.
 */

const fs = require('fs');
const path = require('path');

const MAX_LINES = 500;

function main() {
    const files = process.argv.slice(2);
    if (files.length === 0) process.exit(0);

    let failed = false;

    for (const filepath of files) {
        if (!fs.existsSync(filepath)) continue;

        const content = fs.readFileSync(filepath, 'utf8');
        const lines = content.split('\n').length;

        if (lines > MAX_LINES) {
            console.log(` + "`${filepath}:1 [file-size] File has ${lines} lines (max ${MAX_LINES})`" + `);
            failed = true;
        }
    }

    process.exit(failed ? 1 : 0);
}

main();
`

const tsCheckFunctionSize = `#!/usr/bin/env node
/**
 * Check that functions don't exceed line limits.
 */

const fs = require('fs');

const MAX_LINES = 50;

function main() {
    const files = process.argv.slice(2);
    if (files.length === 0) process.exit(0);

    let failed = false;

    for (const filepath of files) {
        if (!fs.existsSync(filepath)) continue;

        const content = fs.readFileSync(filepath, 'utf8');
        const lines = content.split('\n');

        // Simple function detection (not perfect but catches most cases)
        const funcPattern = /^(\s*)(async\s+)?function\s+(\w+)|^(\s*)const\s+(\w+)\s*=\s*(async\s*)?\(/;

        let inFunction = false;
        let funcStart = 0;
        let funcName = '';
        let braceCount = 0;

        lines.forEach((line, idx) => {
            const match = line.match(funcPattern);
            if (match && !inFunction) {
                inFunction = true;
                funcStart = idx + 1;
                funcName = match[3] || match[5] || 'anonymous';
                braceCount = 0;
            }

            if (inFunction) {
                braceCount += (line.match(/{/g) || []).length;
                braceCount -= (line.match(/}/g) || []).length;

                if (braceCount <= 0 && line.includes('}')) {
                    const funcLines = idx - funcStart + 2;
                    if (funcLines > MAX_LINES) {
                        console.log(` + "`${filepath}:${funcStart} [func-size] ${funcName}() has ${funcLines} lines (max ${MAX_LINES})`" + `);
                        failed = true;
                    }
                    inFunction = false;
                }
            }
        });
    }

    process.exit(failed ? 1 : 0);
}

main();
`

const tsCheckDangerous = `#!/usr/bin/env node
/**
 * Check for dangerous commands and patterns.
 */

const fs = require('fs');

const DANGEROUS_PATTERNS = [
    [/rm\s+-rf/i, 'rm -rf command'],
    [/DROP\s+TABLE/i, 'DROP TABLE statement'],
    [/DROP\s+DATABASE/i, 'DROP DATABASE statement'],
    [/DELETE\s+FROM\s+\w+\s*;/i, 'DELETE without WHERE'],
    [/TRUNCATE\s+TABLE/i, 'TRUNCATE TABLE statement'],
];

function main() {
    const files = process.argv.slice(2);
    if (files.length === 0) process.exit(0);

    let failed = false;

    for (const filepath of files) {
        if (!fs.existsSync(filepath)) continue;

        const content = fs.readFileSync(filepath, 'utf8');
        const lines = content.split('\n');

        lines.forEach((line, idx) => {
            for (const [pattern, desc] of DANGEROUS_PATTERNS) {
                if (pattern.test(line)) {
                    console.log(` + "`${filepath}:${idx + 1} [dangerous-cmd] ${desc} detected`" + `);
                    failed = true;
                }
            }
        });
    }

    process.exit(failed ? 1 : 0);
}

main();
`

const tsCheckMockData = `#!/usr/bin/env node
/**
 * Check for mock/test data in production code.
 */

const fs = require('fs');
const path = require('path');

const MOCK_PATTERNS = [
    /test@example\.com/i,
    /example@/i,
    /@test\.com/i,
    /fake_\w+/i,
    /\w+_fake/i,
    /mock_\w+/i,
    /\w+_mock/i,
    /dummy_\w+/i,
    /placeholder/i,
    /test_user/i,
    /test_password/i,
    /changeme/i,
    /lorem\s+ipsum/i,
];

function main() {
    const files = process.argv.slice(2);
    if (files.length === 0) process.exit(0);

    let failed = false;

    for (const filepath of files) {
        if (!fs.existsSync(filepath)) continue;

        // Skip test files
        if (filepath.includes('test') || filepath.includes('spec')) continue;

        const content = fs.readFileSync(filepath, 'utf8');
        const lines = content.split('\n');

        lines.forEach((line, idx) => {
            // Skip comments
            if (line.trim().startsWith('//') || line.trim().startsWith('/*')) return;

            for (const pattern of MOCK_PATTERNS) {
                if (pattern.test(line)) {
                    console.log(` + "`${filepath}:${idx + 1} [mock-data] Possible test/mock data detected`" + `);
                    failed = true;
                    break;
                }
            }
        });
    }

    process.exit(failed ? 1 : 0);
}

main();
`

const tsCheckConsoleLog = `#!/usr/bin/env node
/**
 * Check for console.log statements in production code.
 */

const fs = require('fs');

function main() {
    const files = process.argv.slice(2);
    if (files.length === 0) process.exit(0);

    let failed = false;

    for (const filepath of files) {
        if (!fs.existsSync(filepath)) continue;

        // Skip test files
        if (filepath.includes('test') || filepath.includes('spec')) continue;

        const content = fs.readFileSync(filepath, 'utf8');
        const lines = content.split('\n');

        lines.forEach((line, idx) => {
            // Skip comments
            const trimmed = line.trim();
            if (trimmed.startsWith('//') || trimmed.startsWith('/*')) return;

            if (line.includes('console.log(')) {
                console.log(` + "`${filepath}:${idx + 1} [ban-console] Remove console.log() - use proper logging`" + `);
                failed = true;
            }
        });
    }

    process.exit(failed ? 1 : 0);
}

main();
`

const tsGuardian = `#!/usr/bin/env node
/**
 * Guardian - Main entry point for all checks.
 */

const { execSync } = require('child_process');
const fs = require('fs');
const path = require('path');

const CHECKS = [
    'check_file_size.js',
    'check_dangerous.js',
    'check_mock_data.js',
    'check_console_log.js',
];

function main() {
    const guardianDir = __dirname;
    let files = process.argv.slice(2);

    if (files.length === 0) {
        // Find all JS/TS files
        files = findFiles('.', ['.js', '.ts', '.jsx', '.tsx']);
    }

    let failed = false;

    for (const check of CHECKS) {
        const checkPath = path.join(guardianDir, check);
        if (!fs.existsSync(checkPath)) continue;

        try {
            execSync(` + "`node ${checkPath} ${files.join(' ')}`" + `, { stdio: 'inherit' });
        } catch (e) {
            failed = true;
        }
    }

    process.exit(failed ? 1 : 0);
}

function findFiles(dir, extensions) {
    const results = [];
    const items = fs.readdirSync(dir);

    for (const item of items) {
        const fullPath = path.join(dir, item);
        const stat = fs.statSync(fullPath);

        if (stat.isDirectory()) {
            if (['node_modules', '.git', '.guardian', 'dist', 'build'].includes(item)) continue;
            results.push(...findFiles(fullPath, extensions));
        } else if (extensions.some(ext => item.endsWith(ext))) {
            results.push(fullPath);
        }
    }

    return results;
}

main();
`

const goGuardianScript = `#!/bin/bash
# Guardian checks for Go projects

set -e

echo "Running Go checks..."

# Run go vet
go vet ./...

# Run staticcheck if available
if command -v staticcheck &> /dev/null; then
    staticcheck ./...
fi

# Check for dangerous patterns (|| true prevents exit on no matches)
if grep -rn "os.Remove\|os.RemoveAll\|exec.Command" --include="*.go" . 2>/dev/null; then
    echo "Warning: Dangerous file operations detected"
fi

echo "Guardian checks complete"
`

// Additional Python checks

const pythonCheckStarImports = `#!/usr/bin/env python3
"""Check for star imports (from x import *)."""

import re
import sys
from pathlib import Path

STAR_IMPORT_PATTERN = re.compile(r'^\s*from\s+\w+(?:\.\w+)*\s+import\s+\*')

def main() -> int:
    if len(sys.argv) < 2:
        return 0

    failed = False
    for filepath in sys.argv[1:]:
        path = Path(filepath)
        if not path.exists() or path.suffix != ".py":
            continue

        content = path.read_text()
        for i, line in enumerate(content.splitlines(), 1):
            if STAR_IMPORT_PATTERN.match(line):
                print(f"{filepath}:{i} [ban-star] Avoid 'from x import *' - pollutes namespace")
                failed = True

    return 1 if failed else 0

if __name__ == "__main__":
    sys.exit(main())
`

const pythonCheckMutableDefaults = `#!/usr/bin/env python3
"""Check for mutable default arguments."""

import ast
import sys
from pathlib import Path

MUTABLE_TYPES = (ast.List, ast.Dict, ast.Set)

def main() -> int:
    if len(sys.argv) < 2:
        return 0

    failed = False
    for filepath in sys.argv[1:]:
        path = Path(filepath)
        if not path.exists() or path.suffix != ".py":
            continue

        try:
            tree = ast.parse(path.read_text())
        except SyntaxError:
            continue

        for node in ast.walk(tree):
            if isinstance(node, (ast.FunctionDef, ast.AsyncFunctionDef)):
                for default in node.args.defaults + node.args.kw_defaults:
                    if default and isinstance(default, MUTABLE_TYPES):
                        print(f"{filepath}:{node.lineno} [mutable-default] {node.name}() has mutable default argument - use None instead")
                        failed = True
                        break

    return 1 if failed else 0

if __name__ == "__main__":
    sys.exit(main())
`

const pythonCheckTodoMarkers = `#!/usr/bin/env python3
"""Check for TODO, FIXME, HACK markers."""

import re
import sys
from pathlib import Path

TODO_PATTERN = re.compile(r'#\s*(TODO|FIXME|HACK|XXX)\b', re.IGNORECASE)

def main() -> int:
    if len(sys.argv) < 2:
        return 0

    failed = False
    for filepath in sys.argv[1:]:
        path = Path(filepath)
        if not path.exists() or path.suffix != ".py":
            continue

        content = path.read_text()
        for i, line in enumerate(content.splitlines(), 1):
            match = TODO_PATTERN.search(line)
            if match:
                marker = match.group(1).upper()
                print(f"{filepath}:{i} [todo-marker] {marker} found - address before committing")
                failed = True

    return 1 if failed else 0

if __name__ == "__main__":
    sys.exit(main())
`

const pythonCheckSubprocessShell = `#!/usr/bin/env python3
"""Check for subprocess with shell=True."""

import re
import sys
from pathlib import Path

SHELL_TRUE_PATTERN = re.compile(r'subprocess\.\w+\([^)]*shell\s*=\s*True')

def main() -> int:
    if len(sys.argv) < 2:
        return 0

    failed = False
    for filepath in sys.argv[1:]:
        path = Path(filepath)
        if not path.exists() or path.suffix != ".py":
            continue

        content = path.read_text()
        for i, line in enumerate(content.splitlines(), 1):
            # Skip comments
            if line.strip().startswith("#"):
                continue

            if SHELL_TRUE_PATTERN.search(line):
                print(f"{filepath}:{i} [subprocess-shell] Avoid shell=True - command injection risk")
                failed = True

    return 1 if failed else 0

if __name__ == "__main__":
    sys.exit(main())
`

const pythonCheckBareExcept = `#!/usr/bin/env python3
"""Check for bare except clauses."""

import re
import sys
from pathlib import Path

# Matches "except:" without any exception type
BARE_EXCEPT_PATTERN = re.compile(r'^\s*except\s*:\s*(#.*)?$')

def main() -> int:
    if len(sys.argv) < 2:
        return 0

    failed = False
    for filepath in sys.argv[1:]:
        path = Path(filepath)
        if not path.exists() or path.suffix != ".py":
            continue

        content = path.read_text()
        for i, line in enumerate(content.splitlines(), 1):
            if BARE_EXCEPT_PATTERN.match(line):
                print(f"{filepath}:{i} [ban-except] Avoid bare 'except:' - catch specific exceptions")
                failed = True

    return 1 if failed else 0

if __name__ == "__main__":
    sys.exit(main())
`
