# Guardian

**Stop AI slop before it hits your codebase.**

Guardian is a universal CLI tool that prevents AI-generated code disasters using deterministic checks. No AI required to run. BYOK (Bring Your Own Key) for optional smart features.

```
╔═╗ ╦ ╦ ╔═╗ ╦═╗ ╔╦╗ ╦ ╔═╗ ╔╗╔
║ ╦ ║ ║ ╠═╣ ╠╦╝  ║║ ║ ╠═╣ ║║║
╚═╝ ╚═╝ ╩ ╩ ╩╚═ ═╩╝ ╩ ╩ ╩ ╝╚╝
```

## Install

```bash
# macOS
brew install guardian

# Linux / macOS (curl)
curl -fsSL guardian.sh/install | sh

# From source
git clone https://github.com/guardian-sh/guardian
cd guardian
make install
```

## Quick Start

```bash
# Interactive mode (recommended)
guardian

# Add to Python project
guardian add python

# Add to TypeScript project  
guardian add typescript

# Run checks in CI
guardian check
```

## What Guardian Catches

### Free Checks (No AI, <200ms)

| Check | What It Catches |
|-------|----------------|
| `file-size` | Files over 500 lines |
| `func-size` | Functions over 50 lines |
| `mock-data` | test@example.com, fake_, placeholder |
| `ban-print` | print() statements |
| `ban-except` | Bare `except:` blocks |
| `ban-eval` | eval(), exec() |
| `ban-star` | `from x import *` |
| `todo-markers` | TODO, FIXME, HACK |
| `dangerous-cmds` | rm -rf, DROP TABLE, DELETE FROM |
| `secret-patterns` | api_key=, password= |
| `subprocess-shell` | shell=True |
| `sql-injection` | f-strings in SQL |

### BYOK Features (Gemini Flash, ~$0.001/use)

- **Smart Scan**: Auto-detect framework, find codebase-specific patterns
- **Prompt Generation**: Generate Claude prompts to fix issues

## Interactive Mode

After running `guardian`, you get an interactive shell:

```
› /run          Check your code now
› /dry-run      Preview what would be checked
› /help         Explain something
› /prompt       Generate a prompt for Claude
› /config       Open configuration
› /exit         Leave Guardian
```

### The `/prompt` Feature

The killer feature for non-technical Claude Code users:

```
› /prompt

What do you need Claude to help with?

❯ ● I have issues and don't know how to fix them
  ○ I need to set up pre-commit but don't know how
  ○ I don't understand what Guardian is telling me
  ○ I want to change the rules but don't know how
```

Guardian generates a complete, copy-paste prompt:

```
┌─────────────────────────────────────────────────────────────┐
│  COPY THIS INTO CLAUDE                                      │
└─────────────────────────────────────────────────────────────┘

I ran Guardian (a code quality tool) and it found problems
I don't know how to fix. Here's what it said:

---
src/api/users.py:45 - "test@example.com" looks like test data
src/api/users.py:142 - create_user() is 67 lines, should be under 50
---

Please:
1. Explain each problem in simple terms
2. Show me the fix for each one
3. After you fix them, run `guardian check` to make sure they pass

✓ Copied to clipboard
```

Paste into Claude Code. Done.

## Configuration

Guardian creates `guardian_config.toml` in your project:

```toml
[project]
src_root = "src"
exclude_dirs = ["tests", "__pycache__", "node_modules"]

[limits]
max_file_lines = 500
max_function_lines = 50

[quality]
ban_print = true
ban_bare_except = true
ban_mock_data = true
mock_patterns = [
    "mock_", "fake_", "dummy_",
    "test@example.com", "placeholder",
]

[security]
ban_eval_exec = true
ban_dangerous_commands = true
dangerous_patterns = ["rm -rf", "DROP TABLE"]
```

## CI Integration

```yaml
# GitHub Actions
- name: Run Guardian
  run: |
    curl -fsSL guardian.sh/install | sh
    guardian check
```

```yaml
# pre-commit
repos:
  - repo: local
    hooks:
      - id: guardian
        name: Guardian checks
        entry: guardian check
        language: system
```

## How It Works

1. `guardian add python` copies check scripts to `.guardian/` in your project
2. You own the files - edit them however you want
3. Checks run via pre-commit hooks or manually
4. No runtime dependency on Guardian being installed

## Project Structure

```
your-project/
├── .guardian/           # Check scripts (you own these)
│   ├── check_file_size.py
│   ├── check_function_size.py
│   ├── check_dangerous.py
│   ├── check_mock_data.py
│   ├── check_security.py
│   └── guardian.py
├── guardian_config.toml # Configuration
└── .pre-commit-config.yaml
```

## Development

```bash
# Clone
git clone https://github.com/guardian-sh/guardian
cd guardian

# Build
make build

# Run locally
./build/guardian

# Run tests
make test

# Build for all platforms
make build-all
```

## License

MIT

---

**Guardian**: AI writes code fast. Guardian makes sure it doesn't delete your home directory while doing it.
