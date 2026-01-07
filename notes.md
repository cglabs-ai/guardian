# Guardian.sh Product Specification

## Version 1.0 | December 2025

---

# Executive Summary

**Guardian** is a universal CLI tool that prevents AI-generated code disasters using deterministic checks. It catches the 65% of AI coding failures that are preventable with traditional tooling—before they hit production.

**Core Thesis:** AI writes code fast. Guardian makes sure it doesn't delete your home directory while doing it.

**Differentiator:** Unlike CodeRabbit (PR review for engineers), Guardian stops slop before commit and helps non-technical Claude Code users via prompt generation. No AI required to run. BYOK (Bring Your Own Key) for optional smart features.

---

# Table of Contents

1. [Problem Statement](https://claude.ai/chat/5cafc564-0330-44a7-a723-c7976a582d48#1-problem-statement)
2. [Target Users](https://claude.ai/chat/5cafc564-0330-44a7-a723-c7976a582d48#2-target-users)
3. [Product Overview](https://claude.ai/chat/5cafc564-0330-44a7-a723-c7976a582d48#3-product-overview)
4. [CLI Specification](https://claude.ai/chat/5cafc564-0330-44a7-a723-c7976a582d48#4-cli-specification)
5. [Web Platform Specification](https://claude.ai/chat/5cafc564-0330-44a7-a723-c7976a582d48#5-web-platform-specification)
6. [Dashboard Specification](https://claude.ai/chat/5cafc564-0330-44a7-a723-c7976a582d48#6-dashboard-specification)
7. [Technical Architecture](https://claude.ai/chat/5cafc564-0330-44a7-a723-c7976a582d48#7-technical-architecture)
8. [Pricing &amp; Monetization](https://claude.ai/chat/5cafc564-0330-44a7-a723-c7976a582d48#8-pricing--monetization)
9. [Roadmap](https://claude.ai/chat/5cafc564-0330-44a7-a723-c7976a582d48#9-roadmap)
10. [Success Metrics](https://claude.ai/chat/5cafc564-0330-44a7-a723-c7976a582d48#10-success-metrics)
11. [Appendices](https://claude.ai/chat/5cafc564-0330-44a7-a723-c7976a582d48#11-appendices)

---

# 1. Problem Statement

## 1.1 The AI Coding Disaster Landscape

AI coding tools (Claude, GPT, Copilot, Cursor) are shipping broken, dangerous, and embarrassing code at scale:

| Failure Category               | Prevalence | Examples                                                   |
| ------------------------------ | ---------- | ---------------------------------------------------------- |
| Catastrophic Data Deletion     | High       | `rm -rf /`,`DROP TABLE users`, recursive file deletion |
| Security Vulnerabilities       | Very High  | SQL injection, hardcoded secrets, eval() usage             |
| Code Quality Degradation       | Universal  | 500+ line files, 100+ line functions, god objects          |
| Mock Data in Production        | High       | test@example.com, fake_user, placeholder values            |
| Infinite Loops & Runaway Costs | Medium     | Claude Code billing $3,600/day, stuck loops                |
| Hallucinated Dependencies      | Medium     | Importing non-existent libraries                           |
| Context Loss Disasters         | Medium     | AI forgets project structure, rewrites working code        |

## 1.2 Prevention Analysis

Research across 190+ documented AI coding failures reveals:

**65% of disasters are preventable** with traditional, deterministic tooling:

* Catastrophic deletion: **95% preventable** (pattern matching)
* Security vulnerabilities: **85% preventable** (static analysis)
* Code quality issues: **75% preventable** (linting, metrics)
* Build/compilation errors: **90% preventable** (CI/CD)
* Database schema issues: **80% preventable** (migration tools)

**35% require new approaches** (genuinely AI-specific):

* Hallucinated libraries: 5% preventable (build-time only)
* Context misunderstanding: 20% preventable
* Confident wrong code: 15% preventable
* Infinite loops: 40% preventable

## 1.3 The Gap

Existing tools don't address this problem:

* **Linters** (ESLint, Ruff): General code quality, not AI-specific patterns
* **Pre-commit hooks** : Require setup expertise most AI users lack
* **CodeRabbit** : PR review (too late), designed for engineers
* **IDE extensions** : Fragmented, don't work with CLI-based AI tools

**No tool exists that:**

1. Catches AI-specific failure patterns
2. Works before code is committed
3. Helps non-technical users fix issues
4. Requires zero AI to run (deterministic, fast, free)

---

# 2. Target Users

## 2.1 Primary: The "Vibe Coder"

**Profile:**

* Uses Claude Code, Cursor, or Copilot to build things
* Limited traditional programming background
* Doesn't fully understand git, pre-commit, or linting
* Ships code that works but has hidden problems
* Age range: 25-55, often non-technical professionals

**Pain Points:**

* "Claude wrote code that deleted my files"
* "I don't know how to ask Claude to fix the issues properly"
* "I shipped test data to production and didn't know"
* "My functions are huge but I don't know how to split them"

**Jobs to Be Done:**

1. Catch problems before I embarrass myself
2. Explain issues in plain English
3. Tell me exactly what to paste into Claude to fix it
4. Make me look like I know what I'm doing

## 2.2 Secondary: The Solo Developer

**Profile:**

* Experienced developer working alone or in small team
* Uses AI tools to move faster
* Knows the risks but doesn't have time to catch everything
* Wants a safety net, not a lecture

**Pain Points:**

* "I know I should review AI output but I'm moving too fast"
* "I need something that just catches the obvious stuff"
* "I don't want another tool that slows me down"

**Jobs to Be Done:**

1. Catch the dangerous stuff automatically
2. Don't slow me down (<200ms checks)
3. Let me configure what matters for my project
4. Integrate with my existing workflow

## 2.3 Tertiary: The Team Lead

**Profile:**

* Managing developers who use AI tools
* Concerned about code quality and security
* Wants visibility into what's being caught
* Needs to enforce standards without micromanaging

**Pain Points:**

* "I don't know what AI-generated code is slipping through"
* "I need metrics to justify tooling decisions"
* "I want guardrails that don't require constant enforcement"

**Jobs to Be Done:**

1. See what issues are being prevented
2. Set team-wide standards
3. Get reports for stakeholders
4. Reduce review burden

---

# 3. Product Overview

## 3.1 Product Components

```
┌─────────────────────────────────────────────────────────────────┐
│                        GUARDIAN.SH                               │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐          │
│  │   CLI Tool   │  │   Website    │  │  Dashboard   │          │
│  │              │  │              │  │              │          │
│  │ • Checks     │  │ • Landing    │  │ • Stats      │          │
│  │ • Prompts    │  │ • Docs       │  │ • Issues     │          │
│  │ • Config     │  │ • Install    │  │ • Projects   │          │
│  │ • BYOK AI    │  │ • Auth       │  │ • Settings   │          │
│  └──────────────┘  └──────────────┘  └──────────────┘          │
│         │                  │                  │                  │
│         └──────────────────┼──────────────────┘                  │
│                            │                                     │
│                    ┌───────┴───────┐                            │
│                    │   Optional    │                            │
│                    │   Sync API    │                            │
│                    └───────────────┘                            │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

## 3.2 Core Principles

1. **Deterministic First** : All core checks run without AI, cloud, or network
2. **Speed Matters** : <200ms for full check suite
3. **User Owns Files** : Scaffolding copied to project, user can edit
4. **Plain English** : No jargon, explain like user is a beginner
5. **Prompt Generation** : The killer feature—write prompts for users
6. **BYOK Optional** : AI features use user's API key, ~$0.001/use

## 3.3 Competitive Positioning

| Feature                | Guardian | CodeRabbit   | ESLint | Pre-commit |
| ---------------------- | -------- | ------------ | ------ | ---------- |
| Pre-commit prevention  | ✅       | ❌ (PR only) | ✅     | ✅         |
| AI-specific patterns   | ✅       | Partial      | ❌     | ❌         |
| Non-technical friendly | ✅       | ❌           | ❌     | ❌         |
| Prompt generation      | ✅       | ❌           | ❌     | ❌         |
| Zero config start      | ✅       | ✅           | ❌     | ❌         |
| No AI required         | ✅       | ❌           | ✅     | ✅         |
| Speed (<200ms)         | ✅       | ❌           | ✅     | Varies     |
| Free tier              | ✅       | Limited      | ✅     | ✅         |

---

# 4. CLI Specification

## 4.1 Distribution

### 4.1.1 Installation Methods

```bash
# macOS (Homebrew)
brew install guardian

# macOS/Linux (curl)
curl -fsSL guardian.sh/install | sh

# Windows (scoop) - V2
scoop install guardian

# From source
git clone https://github.com/guardian-sh/guardian
cd guardian && make install
```

### 4.1.2 Binary Specifications

| Property     | Value                                                               |
| ------------ | ------------------------------------------------------------------- |
| Language     | Go 1.22+                                                            |
| Binary Size  | <15MB                                                               |
| Dependencies | None (single binary)                                                |
| Platforms    | darwin-amd64, darwin-arm64, linux-amd64, linux-arm64, windows-amd64 |

## 4.2 Commands

### 4.2.1 `guardian` (Interactive Mode)

**Description:** Launches the interactive TUI application.

**Flow:**

```
┌──────────────────────────────────────────────────────────────────┐
│    ╔═╗ ╦ ╦ ╔═╗ ╦═╗ ╔╦╗ ╦ ╔═╗ ╔╗╔                                │
│    ║ ╦ ║ ║ ╠═╣ ╠╦╝  ║║ ║ ╠═╣ ║║║                                │
│    ╚═╝ ╚═╝ ╩ ╩ ╩╚═ ═╩╝ ╩ ╩ ╩ ╝╚╝                                │
│    Stop AI slop before it hits your codebase.                   │
└──────────────────────────────────────────────────────────────────┘

  v0.1.0                                              guardian.sh

  ? What would you like to do?
  ❯ ● Quick Start       Add guards to this project
    ○ AI Setup          Smart config with your API key
    ○ About             What Guardian catches
```

**Keyboard Navigation:**

* `↑/↓` or `j/k`: Navigate menu
* `Enter`: Select option
* `Esc`: Go back
* `q` or `Ctrl+C`: Quit

### 4.2.2 `guardian add <language>`

**Description:** Copies check scripts and config to current project.

**Supported Languages:**

| Language       | Stack Variants                                                              |
| -------------- | --------------------------------------------------------------------------- |
| `python`     | `python`,`python-fastapi`,`python-django`,`python-flask`            |
| `typescript` | `typescript`,`typescript-react`,`typescript-node`,`typescript-next` |
| `go`         | `go`                                                                      |
| `php`        | `php`,`php-laravel`                                                     |

**Files Created:**

```
project/
├── .guardian/
│   ├── check_file_size.py      # File length checker
│   ├── check_function_size.py  # Function length checker
│   ├── check_dangerous.py      # Dangerous command detector
│   ├── check_mock_data.py      # Test data detector
│   ├── check_security.py       # Security pattern checker
│   └── guardian.py             # Main runner script
├── guardian_config.toml        # Configuration file
└── .pre-commit-config.yaml     # Pre-commit hooks (created/updated)
```

**Behavior:**

1. Detects if `.guardian/` exists → prompts to overwrite
2. Detects if `guardian_config.toml` exists → prompts to merge/overwrite
3. Detects if `.pre-commit-config.yaml` exists → appends Guardian hooks
4. Offers to install pre-commit if not present

### 4.2.3 `guardian check`

**Description:** Runs all checks on the current project.

**Flags:**

| Flag                | Description                         |
| ------------------- | ----------------------------------- |
| `--fix`           | Auto-fix issues where possible (V2) |
| `--json`          | Output results as JSON              |
| `--quiet`         | Only output errors                  |
| `--config <path>` | Use specific config file            |

**Exit Codes:**

| Code | Meaning                     |
| ---- | --------------------------- |
| 0    | No issues found             |
| 1    | Issues found (any severity) |
| 2    | Configuration error         |
| 3    | Runtime error               |

**Output Format:**

```
GUARDIAN · 12 issues in 4 files

  src/api/users.py
    :45   [mock-data]      test@example.com detected
    :142  [func-size]      create_user() has 67 lines (max 50)

  src/db/queries.py
    :34   [sql-injection]  f-string in SQL query

──────────────────────────────────────────────────────────────
  3 critical · 6 warnings · 3 info

Run 'guardian' for interactive mode with /prompt to generate fixes.
```

### 4.2.4 `guardian config`

**Description:** Opens `guardian_config.toml` in the user's default editor.

**Behavior:**

1. Checks for `$EDITOR` environment variable
2. Falls back to `vim` (Unix) or `notepad` (Windows)
3. Creates default config if none exists

### 4.2.5 `guardian version`

**Description:** Prints version information.

**Output:**

```
guardian 0.1.0 (darwin-arm64)
```

### 4.2.6 `guardian help`

**Description:** Prints help information.

## 4.3 Interactive Commands

After setup or when running `guardian`, users enter interactive mode:

### 4.3.1 `/run`

**Description:** Executes all checks immediately.

**Output:** Same as `guardian check` but inline in TUI.

### 4.3.2 `/dry-run`

**Description:** Shows what would be checked without running checks.

**Output:**

```
  Would check:

    src/api/users.py (234 lines)
    src/api/auth.py (156 lines)
    src/db/queries.py (89 lines) ⚠ large
    src/utils/helpers.py (45 lines)

  Would skip:
    tests/, __pycache__/, node_modules/, .venv/

  4 files · ~524 lines · <1 second

  Press any key to continue...
```

### 4.3.3 `/help`

**Description:** Launches help topic selector.

**Topics:**

1. "What is pre-commit?"
2. "What does Guardian check for?"
3. "How do I fix an issue?"
4. "How do I turn off a rule?"

**Behavior:** Generates a Claude prompt for selected topic → copies to clipboard.

### 4.3.4 `/prompt`

**Description:** THE KILLER FEATURE. Generates Claude prompts for users who don't know how to ask.

**Menu:**

```
  /prompt

  What do you need Claude to help with?

  ❯ ● I have issues and don't know how to fix them
    ○ I need to set up pre-commit but don't know how
    ○ I don't understand what Guardian is telling me
    ○ I want to change the rules but don't know how
    ○ Something else
```

**Generated Prompts:**

#### "I have issues and don't know how to fix them"

```
┌─────────────────────────────────────────────────────────────┐
│  COPY THIS INTO CLAUDE                                      │
└─────────────────────────────────────────────────────────────┘

I ran Guardian (a code quality tool) and it found problems
I don't know how to fix. Here's what it said:

---
src/api/users.py:45 - mock-data
  "test@example.com" looks like test data

src/api/users.py:142 - func-size
  create_user() is 67 lines (max 50)

src/db/queries.py:34 - sql-injection
  f-string in SQL query - use parameterized queries
---

Please:
1. Explain each problem in simple terms
2. Show me the fix for each one
3. After you fix them, run `guardian check` to make sure they pass
4. If any still fail, keep fixing until they all pass

IMPORTANT: Don't just delete code. Fix it properly.

─────────────────────────────────────────────────────────────

✓ Copied to clipboard

Now paste this into Claude Code.
```

#### "I need to set up pre-commit but don't know how"

```
I just installed Guardian (guardian.sh) in my project. It
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

If anything goes wrong, explain what happened and how to fix it.
```

#### "I don't understand what Guardian is telling me"

```
Guardian found these issues in my code but I don't understand
what they mean or why they matter:

---
[issues listed here]
---

For each issue, please explain:
1. What the problem actually is (in simple terms)
2. Why it matters (what could go wrong)
3. How to fix it (show me the actual code change)

I'm not an expert, so please avoid jargon and explain like
I'm a beginner.
```

#### "I want to change the rules but don't know how"

```
I'm using Guardian (guardian.sh) and I want to customize the rules.

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
sensible defaults for my project.
```

#### "Something else"

```
  Something else

  What's going on? (type naturally, Guardian will write the prompt)

  › claude keeps trying to delete files and I dont know how to stop it
```

**Behavior:** Uses BYOK Gemini to generate a custom prompt based on user's natural language input.

**Example Output for "claude keeps deleting files":**

```
┌─────────────────────────────────────────────────────────────┐
│  COPY THIS INTO CLAUDE                                      │
└─────────────────────────────────────────────────────────────┘

IMPORTANT RULES FOR THIS SESSION:

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
[describe your task here]
```

### 4.3.5 `/explain <n>`

**Description:** Provides detailed explanation of issue number N.

**Output:**

```
  Issue: func-size
  Location: src/api/users.py:142

  ──────────────────────────────────────────────────────────────

  What's wrong:

    This function has more than 50 lines of code.

  Why it's dangerous:

    Long functions are hard to understand and test. They usually
    do too many things at once. AI tools often generate massive
    functions because they don't know when to stop.

  How to fix:

    Break the function into smaller helper functions. Each
    function should do one thing well. A good rule: if you
    can't describe what a function does in one sentence,
    it's doing too much.

  ──────────────────────────────────────────────────────────────

  /prompt fix    Get a Claude prompt to fix this

  p prompt · esc back
```

### 4.3.6 `/config`

**Description:** Opens config file in default editor.

### 4.3.7 `/exit`

**Description:** Exits Guardian interactive mode.

## 4.4 Check Specifications

### 4.4.1 Free Checks (No AI, <200ms)

| Check ID             | Severity | Description                    | Default   |
| -------------------- | -------- | ------------------------------ | --------- |
| `file-size`        | warning  | Files over N lines             | 500 lines |
| `func-size`        | warning  | Functions over N lines         | 50 lines  |
| `mock-data`        | warning  | Test/placeholder data patterns | Enabled   |
| `ban-print`        | info     | print() statements             | Enabled   |
| `ban-console`      | info     | console.log() statements       | Enabled   |
| `ban-except`       | warning  | Bare `except:`blocks         | Enabled   |
| `ban-eval`         | critical | eval(), exec() usage           | Enabled   |
| `ban-star`         | warning  | `from x import *`            | Enabled   |
| `mutable-default`  | warning  | `def foo(items=[])`          | Enabled   |
| `todo-markers`     | info     | TODO, FIXME, HACK comments     | Enabled   |
| `dangerous-cmds`   | critical | rm -rf, DROP TABLE, etc.       | Enabled   |
| `secret-patterns`  | critical | Hardcoded API keys, passwords  | Enabled   |
| `subprocess-shell` | warning  | shell=True in subprocess       | Enabled   |
| `sql-injection`    | critical | f-strings in SQL queries       | Enabled   |

### 4.4.2 Check: `file-size`

**Purpose:** Catches bloated files that are hard to maintain.

**Detection:**

```python
if line_count > config.limits.max_file_lines:
    report_issue()
```

**Configuration:**

```toml
[limits]
max_file_lines = 500

[limits.custom_file_limits]
"src/generated/schema.py" = 2000  # Allow larger generated files
```

**Why It Matters:**

* AI tools generate large files without knowing when to split
* Large files are hard to review, test, and maintain
* Often indicates multiple responsibilities in one file

### 4.4.3 Check: `func-size`

**Purpose:** Catches functions that do too much.

**Detection:**

```python
# Uses AST parsing for Python
# Uses regex/heuristics for other languages
if function_line_count > config.limits.max_function_lines:
    report_issue()
```

**Configuration:**

```toml
[limits]
max_function_lines = 50
```

**Why It Matters:**

* AI generates long functions because it doesn't know when to stop
* Long functions are hard to understand and test
* Often indicates function doing multiple things

### 4.4.4 Check: `mock-data`

**Purpose:** Catches test/placeholder data in production code.

**Detection Patterns:**

```python
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
    r"foo_?bar",
    r"asdf",
    r"xxx+",
]
```

**Configuration:**

```toml
[quality]
ban_mock_data = true
mock_patterns = [
    "mock_", "_mock", "fake_", "_fake",
    "test@example.com", "placeholder",
    # Add custom patterns here
]
```

**Why It Matters:**

* AI uses placeholder data to demonstrate functionality
* Shipping test data to production exposes fake accounts
* Can break functionality when users encounter placeholder values

### 4.4.5 Check: `ban-print`

**Purpose:** Catches debug print statements.

**Detection:**

```python
# Python
if "print(" in line and not is_comment(line):
    report_issue()

# JavaScript/TypeScript
if "console.log(" in line and not is_comment(line):
    report_issue()
```

**Configuration:**

```toml
[quality]
ban_print = true
```

**Why It Matters:**

* Print statements clutter production logs
* Can expose sensitive information
* Should use proper logging with levels

### 4.4.6 Check: `ban-except`

**Purpose:** Catches bare exception handlers.

**Detection:**

```python
if re.match(r"except\s*:", line):
    report_issue()
```

**Configuration:**

```toml
[quality]
ban_bare_except = true
```

**Why It Matters:**

* Catches everything including KeyboardInterrupt, SystemExit
* Hides real errors, makes debugging impossible
* AI often uses bare except for "safety"

### 4.4.7 Check: `ban-eval`

**Purpose:** Catches dangerous code execution.

**Detection:**

```python
if "eval(" in line or "exec(" in line:
    report_issue()
```

**Configuration:**

```toml
[security]
ban_eval_exec = true
```

**Why It Matters:**

* Executes arbitrary code
* Massive security vulnerability
* Almost always a better alternative exists

### 4.4.8 Check: `dangerous-cmds`

**Purpose:** Catches destructive commands.

**Detection Patterns:**

```python
DANGEROUS_PATTERNS = [
    r"rm\s+-rf",
    r"rm\s+-r\s+/",
    r"DROP\s+TABLE",
    r"DROP\s+DATABASE",
    r"DELETE\s+FROM\s+\w+\s*;",  # DELETE without WHERE
    r"TRUNCATE\s+TABLE",
    r"shutil\.rmtree",
    r"os\.remove",
    r"fs\.rmdir",
    r"fs\.unlink",
]
```

**Configuration:**

```toml
[security]
ban_dangerous_commands = true
dangerous_patterns = [
    "rm -rf",
    "DROP TABLE",
    "DELETE FROM",
]
```

**Why It Matters:**

* AI generates destructive commands without understanding consequences
* One mistake can wipe databases or file systems
* Should use soft-delete, backups, or confirmation prompts

### 4.4.9 Check: `secret-patterns`

**Purpose:** Catches hardcoded secrets.

**Detection Patterns:**

```python
SECRET_PATTERNS = [
    r'api_key\s*=\s*["\'][^"\']+["\']',
    r'password\s*=\s*["\'][^"\']+["\']',
    r'secret\s*=\s*["\'][^"\']+["\']',
    r'AWS_SECRET',
    r'PRIVATE_KEY',
    r'auth_token\s*=',
    r'access_token\s*=',
]
```

**Configuration:**

```toml
[security]
secret_patterns = [
    "api_key", "apikey", "api-key",
    "secret", "password", "passwd",
    "private_key", "access_token",
]
```

**Why It Matters:**

* Secrets in code get committed to git
* Shared with everyone who has repo access
* Very hard to rotate once exposed

### 4.4.10 Check: `sql-injection`

**Purpose:** Catches SQL injection vulnerabilities.

**Detection:**

```python
# f-strings in SQL
if re.search(r'f["\']SELECT|f["\']INSERT|f["\']UPDATE|f["\']DELETE', line):
    report_issue()

# String concatenation in SQL
if re.search(r'execute\([^)]*\+', line):
    report_issue()
```

**Configuration:**

```toml
[security]
# Enabled by default, no config option
```

**Why It Matters:**

* AI loves f-strings for convenience
* Allows attackers to inject malicious SQL
* Can steal data, drop tables, or take over database

## 4.5 BYOK AI Features

### 4.5.1 Overview

BYOK (Bring Your Own Key) features use the user's Gemini API key for AI-powered functionality. Typical cost: <$0.01 per project setup.

### 4.5.2 Smart Scan

**Purpose:** Analyzes project to generate optimal configuration.

**Process:**

1. Gather project info locally (files, directories, dependencies)
2. Send to Gemini Flash with analysis prompt
3. Parse response for recommendations
4. Generate customized `guardian_config.toml`

**Detects:**

* Language and framework
* Source directory structure
* Test directory location
* Custom mock data patterns in codebase
* Potential hardcoded secrets
* Existing configuration conflicts

**Output:**

```
  Smart Scan Results

  Detected:
    Language:     Python
    Framework:    FastAPI
    Source:       src/
    Tests:        tests/

  Found patterns in your code:
    Mock data:    dummy_user, test_api_key, placeholder_id

  Secrets:        Found 2 possible exposed keys

  Recommendations:
    ✓ Enable SQL injection checks (SQLAlchemy integration)
    ✓ Enable async checks (FastAPI is async-first)
    ✓ Add custom mock patterns to config

  Conflicts:
    ✓ None - safe to install

  ? Apply this configuration? (Y/n)
```

### 4.5.3 Prompt Generation (Custom)

**Purpose:** Generates Claude prompts from natural language input.

**Process:**

1. User types natural description of their problem
2. Send to Gemini with prompt engineering context
3. Generate structured Claude prompt
4. Copy to clipboard

**Example:**

```
User input: "i keep getting errors about types but i dont understand typescript"

Generated prompt:
I'm new to TypeScript and keep getting type errors I don't understand.
My project uses [detected framework].

Please:
1. Look at the errors in my code
2. Explain what each type error means in simple terms
3. Show me how to fix them
4. Teach me the pattern so I understand for next time

I'm a beginner with types, so please be patient and explain
everything clearly.
```

### 4.5.4 API Key Management

**Storage:** `~/.guardian/credentials`

**Permissions:** 0600 (read/write owner only)

**Format:** Plain text API key

**Security:**

* Never transmitted except to Gemini API
* Never logged or stored in project files
* User can regenerate at any time

## 4.6 Configuration Specification

### 4.6.1 File: `guardian_config.toml`

```toml
# Guardian Configuration
# https://guardian.sh/docs/config

[project]
# Root directory for source files
src_root = "src"

# Directories to exclude from checks
exclude_dirs = [
    "tests",
    "test",
    "__pycache__",
    "node_modules",
    ".venv",
    "venv",
    "dist",
    "build",
    "migrations",
]

# Files to exclude (glob patterns)
exclude_files = [
    "*.min.js",
    "*.generated.*",
]

[limits]
# Maximum lines per file
max_file_lines = 500

# Maximum lines per function
max_function_lines = 50

# Override limits for specific files
[limits.custom_file_limits]
# "path/to/large/file.py" = 1000

[quality]
# Ban print() / console.log() statements
ban_print = true

# Ban bare except: blocks
ban_bare_except = true

# Ban mutable default arguments
ban_mutable_defaults = true

# Ban wildcard imports (from x import *)
ban_star_imports = true

# Flag TODO/FIXME/HACK markers
ban_todo_markers = true

# Detect mock/test data in production code
ban_mock_data = true

# Patterns that indicate mock/test data
mock_patterns = [
    "mock_", "_mock",
    "fake_", "_fake",
    "dummy_", "_dummy",
    "test_user", "test_email", "test_password",
    "example@", "@example.com", "@test.com",
    "placeholder", "sample_", "hardcoded",
    "changeme", "replace_me", "your_", "xxx",
    "lorem ipsum", "foo_bar", "asdf",
]

[security]
# Ban eval() and exec()
ban_eval_exec = true

# Ban subprocess with shell=True
ban_subprocess_shell = true

# Ban dangerous commands
ban_dangerous_commands = true

# Patterns for dangerous commands
dangerous_patterns = [
    "rm -rf",
    "rm -r /",
    "DROP TABLE",
    "DROP DATABASE",
    "DELETE FROM",
    "TRUNCATE TABLE",
]

# Patterns for hardcoded secrets
secret_patterns = [
    "api_key", "apikey", "api-key",
    "secret", "password", "passwd",
    "private_key", "privatekey",
    "access_token", "auth_token",
    "AWS_SECRET", "GITHUB_TOKEN",
]

[output]
# Output format: "text", "json", "github"
format = "text"

# Colorize output
color = true

# Show file paths relative to project root
relative_paths = true
```

### 4.6.2 Stack-Specific Additions

#### Python + FastAPI

```toml
[fastapi]
# Check for async/await issues
check_async = true

# SQLAlchemy injection patterns
sqlalchemy_injection = true

# Route security (auth decorators)
route_security = true
```

#### TypeScript + React

```toml
[react]
# React hooks rules
check_hooks = true

# JSX accessibility
jsx_a11y = true

# useEffect dependencies
effect_deps = true
```

---

# 5. Web Platform Specification

## 5.1 Overview

**Domain:** guardian.sh

**Purpose:** Marketing site, documentation, and authentication gateway.

**Stack:** Next.js 14, Tailwind CSS, Vercel

## 5.2 Design System

### 5.2.1 Colors

| Name       | Hex                   | Usage             |
| ---------- | --------------------- | ----------------- |
| Forest 50  | #f0fdf4               | Backgrounds       |
| Forest 100 | #dcfce7               | Hover states      |
| Forest 400 | #4ade80               | Secondary text    |
| Forest 500 | #22c55e               | Primary green     |
| Forest 600 | #16a34a               | Buttons, icons    |
| Forest 700 | #15803d               | Darker accents    |
| Spring     | #00ff7f               | Highlights, glow  |
| Dark 900   | #0a0a0a               | Page background   |
| Dark 800   | #111111               | Card backgrounds  |
| Dark 700   | #1a1a1a               | Input backgrounds |
| Dark 600   | #222222               | Borders           |
| White      | #ffffff               | Primary text      |
| White/50   | rgba(255,255,255,0.5) | Secondary text    |
| White/10   | rgba(255,255,255,0.1) | Borders           |

### 5.2.2 Typography

**Font Stack:**

```css
font-sans: "SF Pro Display", -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
font-mono: "SF Mono", Menlo, Monaco, "Cascadia Code", Consolas, monospace;
```

**Scale:**

| Name       | Size     | Line Height | Usage                |
| ---------- | -------- | ----------- | -------------------- |
| display-lg | 4.5rem   | 1           | Hero headline        |
| display    | 3.5rem   | 1.1         | Section headlines    |
| display-sm | 2.5rem   | 1.2         | Subsection headlines |
| xl         | 1.25rem  | 1.75        | Large body           |
| base       | 1rem     | 1.5         | Body text            |
| sm         | 0.875rem | 1.25        | Secondary text       |
| xs         | 0.75rem  | 1           | Labels               |

### 5.2.3 Components

**Buttons:**

```css
.btn-primary {
  @apply px-6 py-3 bg-forest-600 text-white font-medium rounded-lg;
  /* Hover: bg-forest-500, glow effect */
}

.btn-secondary {
  @apply px-6 py-3 bg-dark-600 text-white border border-white/10 rounded-lg;
  /* Hover: bg-dark-500, border-white/20 */
}
```

**Code Blocks:**

* macOS-style window chrome (red/yellow/green dots)
* Dark gradient background
* Syntax highlighting
* Copy button

**Cards:**

* Dark background (#1a1a1a)
* Subtle border (white/5)
* Rounded corners (12px)
* Hover: translateY(-2px), border color change

## 5.3 Pages

### 5.3.1 Landing Page (`/`)

**Sections:**

1. **Navigation**
   * Logo (G icon + "Guardian")
   * Links: GitHub, Install, Login (hidden)
   * Fixed, blur background
2. **Hero**
   * ASCII logo (animated fade-in)
   * Headline: "Stop AI slop before it hits your codebase."
   * Subhead: "Deterministic checks that catch disasters..."
   * Install command (click to copy)
   * Secondary install method
3. **What Guardian Catches**
   * Grid of check cards
   * Check name, severity badge, description
   * "+8 more checks" footer
4. **The Killer Feature**
   * "Don't know how to ask Claude to fix it?"
   * Feature description
   * Interactive code block showing /prompt menu
   * Benefit list
5. **Three Commands**
   * Install, Add, Run
   * Step indicators (01, 02, 03)
   * Copyable code block
6. **CTA**
   * "AI writes code fast. Guardian makes sure it doesn't delete your home directory."
   * Install button
   * GitHub link
7. **Footer**
   * Logo
   * Links: GitHub, Twitter
   * MIT License

### 5.3.2 Login Page (`/login`) [Hidden V1]

**Flow:**

1. Email input
2. "Continue with Email" button
3. "Continue with GitHub" button (OAuth)
4. Magic link sent confirmation

**State:**

* Email input
* Loading state
* Success state (check your email)
* Error state

### 5.3.3 Documentation (`/docs`) [V2]

**Structure:**

* Getting Started
  * Installation
  * Quick Start
  * Configuration
* Checks
  * Overview
  * Individual check docs
* Interactive Mode
  * Commands
  * Prompt Generation
* API Reference [V3]

## 5.4 Routes

| Path            | Purpose             | Status    |
| --------------- | ------------------- | --------- |
| `/`           | Landing page        | V1        |
| `/login`      | Authentication      | Hidden V1 |
| `/dashboard`  | Usage stats         | Hidden V1 |
| `/docs`       | Documentation       | V2        |
| `/install`    | Curl install script | V1        |
| `/api/auth/*` | Auth endpoints      | Hidden V1 |

---

# 6. Dashboard Specification

## 6.1 Overview

**Purpose:** Track Guardian usage, issues prevented, and configure settings.

**Access:** Authenticated users only.

**Status:** Hidden in V1 (built but not linked).

## 6.2 Sidebar Navigation

```
┌─────────────────────┐
│ [G] Guardian        │
├─────────────────────┤
│ ○ Overview          │
│ ○ Issues            │
│ ○ Projects          │
│ ○ Settings          │
├─────────────────────┤
│ [Avatar] h@aleq.ai  │
│ Pro Plan            │
└─────────────────────┘
```

## 6.3 Overview Page

### 6.3.1 Stats Cards

| Metric           | Description                                   |
| ---------------- | --------------------------------------------- |
| Issues Prevented | Total issues caught before commit             |
| Checks Run       | Total check executions                        |
| Files Scanned    | Total files analyzed                          |
| Time Saved       | Estimated time saved (issues × avg fix time) |

### 6.3.2 Weekly Chart

* Bar chart showing issues per day
* Hover for exact count
* 7-day rolling window

### 6.3.3 Recent Issues

* Last 5 issues caught
* File, line, rule, message, timestamp
* Click to expand details
* "View all" link

## 6.4 Issues Page

### 6.4.1 Filters

* Search (file name, message)
* Severity (All, Critical, Warning, Info)
* Project (All, or specific)
* Date range

### 6.4.2 Table

| Column  | Description                 |
| ------- | --------------------------- |
| Rule    | Check that caught the issue |
| File    | File path and line number   |
| Message | Issue description           |
| When    | Relative timestamp          |

### 6.4.3 Issue Detail Modal

* Full file path
* Code snippet around issue
* Explanation
* Suggested fix
* "Generate Claude prompt" button

## 6.5 Projects Page

### 6.5.1 Project Cards

* Project name
* Issue count or "Clean" badge
* Last check timestamp
* Click to view project details

### 6.5.2 Add Project

* "Add Project" card
* Instructions for syncing
* API key display

## 6.6 Settings Page

### 6.6.1 Account

* Email (editable)
* Name (editable)
* Save button

### 6.6.2 API Key

* Masked key display
* Copy button
* Regenerate button (with confirmation)

### 6.6.3 Plan

* Current plan name and price
* Status badge (Active)
* "Manage Billing" button

### 6.6.4 Danger Zone

* Delete account button
* Confirmation modal
* Permanent deletion warning

## 6.7 Data Sync

### 6.7.1 Sync Mechanism

Guardian CLI → Dashboard sync is opt-in:

```bash
# Enable sync
guardian config --set sync.enabled=true
guardian config --set sync.api_key=grd_sk_xxxxx

# Disable sync
guardian config --set sync.enabled=false
```

### 6.7.2 Data Transmitted

| Data         | Transmitted | Notes                  |
| ------------ | ----------- | ---------------------- |
| Issue count  | Yes         | Per check run          |
| File names   | Optional    | Can be anonymized      |
| Line numbers | Yes         | No code content        |
| Check types  | Yes         | Which checks triggered |
| Timestamps   | Yes         | For trending           |
| Code content | Never       | Privacy preserved      |

### 6.7.3 Privacy

* No code is ever transmitted
* File paths can be anonymized
* All data encrypted in transit
* User can delete all data at any time

---

# 7. Technical Architecture

## 7.1 CLI Architecture

```
guardian/
├── main.go                    # Entry point
├── internal/
│   ├── ui/
│   │   ├── styles.go         # Color theme, lipgloss styles
│   │   ├── logo.go           # ASCII art
│   │   └── components.go     # Reusable UI components
│   ├── screens/
│   │   ├── app.go            # Screen router
│   │   ├── mainmenu.go       # Main menu screen
│   │   ├── quickstart.go     # Setup flow
│   │   ├── aisetup.go        # BYOK configuration
│   │   ├── interactive.go    # Post-setup commands
│   │   └── about.go          # Information screen
│   ├── checks/
│   │   └── runner.go         # Check execution engine
│   ├── prompts/
│   │   └── generator.go      # Claude prompt generation
│   ├── scaffolding/
│   │   └── copy.go           # File copying logic
│   ├── config/
│   │   └── config.go         # TOML parsing
│   └── ai/
│       └── gemini.go         # Gemini API integration
├── go.mod
├── go.sum
└── Makefile
```

### 7.1.1 Dependencies

| Package   | Version | Purpose          |
| --------- | ------- | ---------------- |
| bubbletea | v0.26.4 | TUI framework    |
| bubbles   | v0.18.0 | TUI components   |
| lipgloss  | v0.11.0 | Terminal styling |
| go-toml   | v2.2.2  | TOML parsing     |
| clipboard | v0.1.4  | Clipboard access |

### 7.1.2 Build Process

```makefile
# Build for current platform
make build

# Build for all platforms
make build-all

# Creates:
# build/guardian-darwin-amd64
# build/guardian-darwin-arm64
# build/guardian-linux-amd64
# build/guardian-linux-arm64
# build/guardian-windows-amd64.exe
```

## 7.2 Web Architecture

```
guardian-web/
├── app/
│   ├── layout.tsx            # Root layout
│   ├── page.tsx              # Landing page
│   ├── globals.css           # Global styles
│   ├── login/
│   │   └── page.tsx          # Login page
│   └── dashboard/
│       └── page.tsx          # Dashboard page
├── public/
│   ├── favicon.svg           # Favicon
│   ├── icon.svg              # Full icon
│   └── install               # Install script
├── tailwind.config.ts        # Tailwind configuration
├── next.config.mjs           # Next.js configuration
├── vercel.json               # Vercel routing
└── package.json
```

### 7.2.1 Dependencies

| Package     | Version | Purpose         |
| ----------- | ------- | --------------- |
| next        | 14.2.0  | React framework |
| react       | 18.x    | UI library      |
| tailwindcss | 3.4.x   | Styling         |
| typescript  | 5.x     | Type safety     |

## 7.3 Infrastructure

### 7.3.1 CLI Distribution

| Channel         | Provider                     | Notes                      |
| --------------- | ---------------------------- | -------------------------- |
| Homebrew        | homebrew-core or tap         | Primary macOS distribution |
| Curl script     | Vercel (guardian.sh/install) | Universal installer        |
| GitHub Releases | GitHub                       | Binary downloads           |
| Scoop           | scoop bucket                 | Windows (V2)               |

### 7.3.2 Web Hosting

| Component | Provider    | Notes               |
| --------- | ----------- | ------------------- |
| Website   | Vercel      | Next.js hosting     |
| Domain    | guardian.sh | Primary domain      |
| CDN       | Vercel Edge | Global distribution |

### 7.3.3 Backend (V2+)

| Component | Provider         | Notes              |
| --------- | ---------------- | ------------------ |
| API       | Vercel Functions | Serverless         |
| Database  | Supabase         | PostgreSQL         |
| Auth      | Supabase Auth    | Magic link + OAuth |
| Analytics | Supabase         | Usage metrics      |

---

# 8. Pricing & Monetization

## 8.1 V1: Free

All V1 features are free:

* CLI tool
* All deterministic checks
* BYOK AI features (user pays Gemini directly)
* Dashboard (when enabled)

## 8.2 V2+: Freemium Model

### 8.2.1 Free Tier

* Unlimited local checks
* All deterministic rules
* BYOK AI features
* 1 project in dashboard
* 30 days data retention

### 8.2.2 Pro Tier ($19/month)

* Unlimited projects
* 1 year data retention
* Team features (shared configs)
* Priority support
* Custom rules (V3)

### 8.2.3 Team Tier ($99/month)

* Everything in Pro
* 10 seats included
* Team dashboard
* Audit logs
* SSO (V3)

## 8.3 Revenue Projections

| Milestone | Users   | Pro @ 5% | Revenue/mo |
| --------- | ------- | -------- | ---------- |
| Launch    | 1,000   | 50       | $950       |
| 6 months  | 10,000  | 500      | $9,500     |
| 1 year    | 50,000  | 2,500    | $47,500    |
| 2 years   | 200,000 | 10,000   | $190,000   |

---

# 9. Roadmap

## 9.1 V1.0 (Launch)

**Target:** January 2026

**Features:**

* [X] Go CLI with interactive TUI
* [X] Python scaffolding (all checks)
* [X] TypeScript scaffolding (all checks)
* [X] BYOK Gemini integration
* [X] Prompt generation feature
* [X] Pre-commit integration
* [X] Website (landing page)
* [X] Login page (hidden)
* [X] Dashboard (hidden)
* [ ] Homebrew formula
* [ ] Documentation

**Metrics:**

* 1,000 installs
* 100 daily active users
* <5% error rate

## 9.2 V1.1 (Polish)

**Target:** February 2026

**Features:**

* [ ] Go scaffolding
* [ ] PHP scaffolding
* [ ] Windows support (scoop)
* [ ] `guardian check --fix` for simple fixes
* [ ] Custom rule definitions
* [ ] VS Code extension (notifications only)

## 9.3 V2.0 (Dashboard)

**Target:** Q2 2026

**Features:**

* [ ] Dashboard public launch
* [ ] CLI ↔ Dashboard sync
* [ ] Team features
* [ ] Billing integration (Stripe)
* [ ] API for CI/CD integration
* [ ] GitHub Action

## 9.4 V3.0 (Enterprise)

**Target:** Q4 2026

**Features:**

* [ ] SSO (SAML, OIDC)
* [ ] Custom rule marketplace
* [ ] AI-suggested fixes
* [ ] Slack/Teams integration
* [ ] Audit logs
* [ ] On-premise deployment

---

# 10. Success Metrics

## 10.1 Acquisition

| Metric             | Target (6mo) | Target (1yr) |
| ------------------ | ------------ | ------------ |
| Total installs     | 10,000       | 50,000       |
| Weekly installs    | 500          | 1,500        |
| GitHub stars       | 1,000        | 5,000        |
| Homebrew downloads | 5,000        | 25,000       |

## 10.2 Activation

| Metric                        | Target |
| ----------------------------- | ------ |
| Install → First check        | 70%    |
| First check → Setup complete | 50%    |
| Setup → Pre-commit enabled   | 40%    |

## 10.3 Engagement

| Metric               | Target          |
| -------------------- | --------------- |
| Daily active users   | 20% of installs |
| Weekly active users  | 40% of installs |
| Avg checks/user/week | 15              |
| /prompt usage rate   | 30%             |

## 10.4 Retention

| Metric            | Target |
| ----------------- | ------ |
| Week 1 retention  | 60%    |
| Week 4 retention  | 40%    |
| Week 12 retention | 25%    |

## 10.5 Revenue (V2+)

| Metric                 | Target (6mo)           | Target (1yr) |
| ---------------------- | ---------------------- | ------------ |
| Free → Pro conversion | 3%                     | 5%           |
| Pro MRR                | $5,000       | $25,000 |              |
| Churn rate             | <5%/mo                 | <3%/mo       |

---

# 11. Appendices

## 11.1 Appendix A: Check Scripts (Python)

### check_file_size.py

```python
#!/usr/bin/env python3
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
```

### check_function_size.py

```python
#!/usr/bin/env python3
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
```

## 11.2 Appendix B: AI Failure Research Sources

Key incidents documented:

1. **Claude Code rm -rf disasters** (2025)
   * User lost entire home directory
   * AI executed recursive deletion without confirmation
2. **$3,600/day billing loop** (2025)
   * Claude Code stuck in infinite task loop
   * Generated thousands of API calls
3. **CVE-2025-55284 DNS exfiltration** (2025)
   * AI-generated code vulnerable to data theft
   * MCP protocol security issue
4. **Production database deletions** (Multiple)
   * DELETE FROM without WHERE
   * DROP TABLE in migration scripts
5. **Hardcoded credentials in commits** (Ongoing)
   * API keys in source code
   * Detected by GitHub secret scanning

## 11.3 Appendix C: Competitive Analysis

### CodeRabbit

* **Model:** AI PR review
* **Pricing:** $15/user/month
* **Pros:** Deep analysis, integrations
* **Cons:** Too late (PR stage), engineer-focused

### ESLint/Ruff

* **Model:** Traditional linting
* **Pricing:** Free
* **Pros:** Fast, comprehensive
* **Cons:** No AI-specific patterns, complex config

### Pre-commit

* **Model:** Hook framework
* **Pricing:** Free
* **Pros:** Flexible, extensible
* **Cons:** Requires setup expertise

### Guardian Positioning

* Pre-commit timing (before PR)
* AI-specific patterns
* Non-technical user support
* Zero-config start
* Prompt generation (unique)

## 11.4 Appendix D: User Research Quotes

> "Claude keeps trying to delete files and I don't know how to stop it"
> — User feedback, December 2025

> "I shipped test@example.com to production three times"
> — Solo developer, Reddit

> "I don't know how to ask Claude to fix it properly"
> — Non-technical founder, Twitter

> "My functions are 200 lines because that's what Claude generates"
> — Junior developer, Discord

---

# Document History

| Version | Date     | Author        | Changes               |
| ------- | -------- | ------------- | --------------------- |
| 1.0     | Dec 2025 | Guardian Team | Initial specification |

---

*End of Product Specification*
