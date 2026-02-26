# ralphkit

Orchestrate Ralph-style autonomous AI coding loops with Claude Code.

## What is a Ralph Loop?

A Ralph loop runs Claude Code in an autonomous loop until a PRD/spec is fully complete. Each iteration: the agent reads the spec, works on tasks, evaluates completion, and loops until everything is done.

## Installation

```bash
go install github.com/kfroemming/ralphkit@latest
```

Then verify dependencies:

```bash
ralphkit install
```

This checks for Node.js, npm, the Claude CLI (`npm install -g @anthropic-ai/claude-code`), and git. It also creates the `~/.ralphkit/` config directory.

## Quick Start

### 1. Create a PRD

```bash
ralphkit new my-feature
```

This launches an interactive wizard that walks you through describing your project. Claude then expands your notes into a structured PRD with goals, features, acceptance criteria, and more.

### 2. Run the loop

```bash
ralphkit run my-feature.prd.md
```

Claude works through the spec autonomously, iterating until all items are complete or the max iteration limit is reached.

## Commands

### `ralphkit new [name]`

Interactive PRD crafting wizard. Asks about your project, tech stack, features, constraints, and success criteria, then generates a structured PRD with Claude.

### `ralphkit run [prd-file]`

Start a Ralph loop from a PRD/spec file.

| Flag | Description |
|------|-------------|
| `-m, --model` | Claude model (default: `claude-opus-4-6`). Shortcuts: `opus`, `sonnet`, `haiku` |
| `--skip-tests` | Skip tests between iterations |
| `--with-tests` | Run tests between iterations (default) |
| `-n, --max-iterations` | Max iterations before stopping (default: 10) |
| `-w, --worktree` | Git worktree path to run in |
| `-s, --session-name` | Name this session |
| `-d, --dir` | Working directory |
| `--notify` | macOS notification on completion |
| `--dangerously-skip-permissions` | Pass through to Claude CLI |
| `-q, --quiet` | Suppress UI chrome |

### `ralphkit install`

Check and install all dependencies.

### `ralphkit session list`

List all sessions with status (running/stopped/complete), iteration count, start time, and working directory.

### `ralphkit session stop [name]`

Stop a running session by sending SIGINT.

### `ralphkit session clean`

Remove completed and stopped session files.

### `ralphkit worktree add [branch] [path]`

Add a git worktree, creating the branch if it doesn't exist.

### `ralphkit worktree list`

List git worktrees.

### `ralphkit worktree remove [path]`

Remove a git worktree.

### `ralphkit tail [session-name]`

Tail the live output of a running session.

### `ralphkit config show`

Show current configuration.

### `ralphkit config set [key] [value]`

Set a configuration value. Available keys:

- `default_model` — Default Claude model
- `max_iterations` — Default max iterations

## Tips for Good PRDs

- Be specific about acceptance criteria — Claude needs clear "done" conditions
- List features as bullet points with testable outcomes
- Include what's out of scope to prevent scope creep
- Mention tech stack and constraints upfront
- Break large projects into multiple PRDs and run them sequentially

## Configuration

Config lives at `~/.ralphkit/config.yaml`:

```yaml
default_model: claude-opus-4-6
max_iterations: 10
```

Session state is stored in `~/.ralphkit/sessions/`.
