# gh-packs

**The package manager for AI context.** A GitHub CLI extension that makes AI context composable, versioned, and routable across repositories.

## Install

```bash
gh extension install tunajam/gh-packs
```

## Quick Start

```bash
# Initialize in your repo
gh packs init

# Add a dependency
gh packs add github.com/mycompany/payments-api

# Install all deps from lockfile
gh packs install

# Get a task-specific briefing via sub-agent
gh packs resolve "add refund webhooks for merchants"
```

## What It Does

When AI coding assistants work across service boundaries, they're blind to other repos' context. **packs** fixes this:

1. **Versioned context packages** — pull `.context/` directories from other repos via `gh`, pinned with semver and a lockfile
2. **Dependency graph** — `packs.yaml` maps how your services connect
3. **Sub-agent resolution** — `packs resolve` spawns a Claude session that reads the full graph and returns a tight, task-specific briefing

## Commands

| Command | Description |
|---------|-------------|
| `gh packs init` | Scaffold packs.yaml, .context/, .packs/ |
| `gh packs add <source>` | Add dependency from GitHub repo |
| `gh packs install` | Install all deps from lockfile |
| `gh packs resolve "<task>"` | Sub-agent briefing via claude --print |
| `gh packs show <name>` | Dump single pack to stdout |
| `gh packs graph` | Print dependency tree |
| `gh packs remove <name>` | Remove a dependency |
| `gh packs validate` | Lint manifest and verify integrity |

## Prerequisites

- [gh CLI](https://cli.github.com/) installed and authenticated
- [claude CLI](https://docs.anthropic.com/en/docs/claude-cli) installed (for `resolve`)
- Go 1.22+ (for building from source)

## How It Works

```
Developer: "add refund webhooks"
    → Claude reads CLAUDE.md, sees packs instruction
    → Runs: gh packs resolve "add refund webhooks"
    → packs assembles graph + pack contents
    → Spawns claude --print sub-agent with full context
    → Sub-agent returns 1-3KB task-specific briefing
    → Main session writes code using real contracts
```

50KB of context goes in, 2KB of task-specific intelligence comes out. Zero context window pollution.

## License

MIT
