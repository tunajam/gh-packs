# gh-packs

**The package manager for AI context.** Version, share, and resolve cross-repo context for AI coding assistants.

AI assistants are blind at service boundaries. When your payments service calls the auth service, Claude doesn't know the auth API contract. **packs** fixes this — it's npm for `.context/` directories.

## Install

```bash
gh extension install tunajam/gh-packs
```

## Quick Start

```bash
# 1. Initialize in your repo
gh packs init

# 2. Add a dependency (fetches .context/ from the target repo)
gh packs add github.com/mycompany/payments-api --reason "billing integration"

# 3. Install all deps from lockfile (on a fresh clone)
gh packs install

# 4. Get a task-specific briefing via sub-agent
gh packs resolve "add refund webhooks for merchants"
```

## How It Works

```
packs.yaml          ← declares which repos you depend on (like package.json)
packs.lock          ← pins exact versions + content hashes (like package-lock.json)
.context/           ← your repo's context (API docs, schemas, conventions)
.context/packs/     ← installed dependency context (gitignored)
.packs/             ← local config (custom resolve prompt)
```

When you run `gh packs add`, packs:

1. Resolves the semver range against the repo's git tags
2. Downloads the `.context/` directory via the GitHub API
3. Pins the exact commit + content hash in `packs.lock`
4. Stores the files in `.context/packs/<repo>/`

When you run `gh packs resolve "task"`, packs:

1. Loads all installed pack contents
2. Spawns `claude --print` with a system prompt tuned for context resolution
3. The sub-agent reads the full dependency graph and returns a tight, task-specific briefing
4. 50KB of context goes in → 2KB of actionable intelligence comes out

## Commands

### `gh packs init`

Scaffolds `packs.yaml`, `.context/CONTEXT.md`, and `.packs/resolve-prompt.md`.

```bash
gh packs init
# Created:
#   packs.yaml               ← manifest
#   .context/CONTEXT.md      ← your repo's context template
#   .packs/resolve-prompt.md ← sub-agent system prompt (customizable)
```

### `gh packs add <source>`

Add a dependency from a GitHub repo. Resolves version, downloads context, updates manifest and lockfile.

```bash
gh packs add github.com/mycompany/auth-service
gh packs add mycompany/payments --version "~2.1.0" --reason "refund flow"
gh packs add mycompany/shared-types --path "docs/context"
```

| Flag | Description |
|------|-------------|
| `--version` | Semver range (default: `^latest`) |
| `--path` | Context path in source repo (default: `.context`) |
| `--reason` | Why this dependency is needed |

### `gh packs install`

Install all dependencies from `packs.lock`. Verifies content hashes.

```bash
gh packs install
#   @mycompany/auth-service@1.2.3  ✓ verified
#   @mycompany/payments@2.1.0      ✓ verified
# 2 packs installed in 1.2s
```

### `gh packs resolve "<task>"`

Spawns a Claude sub-agent that reads your full dependency graph and returns a task-specific briefing.

```bash
gh packs resolve "add refund webhooks for merchants"
gh packs resolve --dry-run "migrate to v2 auth tokens"
```

### `gh packs show <name>`

Dump an installed pack's contents to stdout.

```bash
gh packs show payments
```

### `gh packs graph`

Print the dependency tree.

```bash
gh packs graph
# @mycompany/checkout (you are here)
#   ├── @mycompany/auth-service@1.2.3
#   └── @mycompany/payments@2.1.0
```

### `gh packs remove <name>`

Remove a dependency. Cleans up manifest, lockfile, and installed files.

```bash
gh packs remove payments
```

### `gh packs validate`

Lint `packs.yaml` and verify lockfile integrity (content hashes match installed files).

```bash
gh packs validate
# ✓ @mycompany/auth-service@1.2.3
# ✓ @mycompany/payments@2.1.0
# ✓ All checks passed
```

## `packs.yaml` Format

```yaml
name: "@mycompany/checkout"
version: "1.0.0"
description: "Checkout service"

pack:
  type: service

dependencies:
  "@mycompany/auth-service":
    version: "^1.0.0"
    source: "github.com/mycompany/auth-service"
    reason: "token validation for checkout flow"
  "@mycompany/payments":
    version: "~2.1.0"
    source: "github.com/mycompany/payments"
    path: "docs/context"
    reason: "charge creation and refund APIs"

resolve:
  max_depth: 2
  system_prompt_append: "Focus on REST APIs, ignore gRPC."
```

### Version Ranges

| Range | Meaning | Example |
|-------|---------|---------|
| `^1.0.0` | Same major | `>=1.0.0, <2.0.0` |
| `~1.2.0` | Same major.minor | `>=1.2.0, <1.3.0` |
| `1.2.3` | Exact | `=1.2.3` |
| `>=1.0.0` | Floor, no ceiling | `>=1.0.0` |
| `*` / `latest` | Highest available | latest tag |

## The Resolve Pattern

The `resolve` command implements a **sub-agent pattern** for context routing:

1. Your main AI session (Claude, Cursor, etc.) encounters a cross-service task
2. It runs `gh packs resolve "task description"`
3. packs assembles the full context: manifest + all installed pack contents
4. A sub-agent (`claude --print`) reads everything and produces a focused briefing
5. The main session gets exactly the context it needs — no window pollution

The resolve prompt is customizable at `.packs/resolve-prompt.md`.

## Requirements

- [GitHub CLI](https://cli.github.com/) (`gh`) — installed and authenticated
- [Claude CLI](https://docs.anthropic.com/en/docs/claude-cli) — required for `resolve` command only
- Source repos need semver git tags (`v1.0.0`, `v1.2.3`, etc.)
- Source repos need a `.context/` directory (or custom path) with context files

## License

MIT
