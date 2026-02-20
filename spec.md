# gh-packs

### Context is a dependency.

**Product & Engineering Specification**
Version 1.0 · February 2026

*A GitHub CLI extension that makes AI context composable, versioned, and routable across repositories.*

---

## 1. Executive Summary

**gh-packs** is a `gh` CLI extension that treats AI context as a versioned dependency. It enables AI coding assistants to autonomously discover, load, and reason over cross-repository context — producing correct, informed output when working across service boundaries.

We're investing heavily in AI-consumable context: `.context` directories, architecture documentation, convention guides. This context dramatically improves AI output within a single repository. But our software spans hundreds of repositories, and the moment an AI assistant needs to work across service boundaries, it's blind.

The context exists. There's no way to route it.

**gh-packs** solves this with three mechanisms:

1. **Versioned context packages** with a lockfile, pulled from git repos using `gh` — zero infrastructure required
2. **An inter-repo dependency graph** that maps how services, libraries, and platforms connect
3. **Sub-agent resolution** via `claude --print` that spawns an independent AI session to reason over the full graph and return a synthesized, task-specific briefing — keeping the main session's context window lean

The result: AI assistants that understand system topology and produce cross-service code that works against real contracts, not hallucinated interfaces.

---

## 2. The Problem

### 2.1 Context Is Siloed

Every repository is an island of AI context. Teams invest real engineering effort producing `.context` directories, system prompts, and convention guides. Within a single repo, this context transforms AI output from plausible-but-wrong to precise-and-correct.

But software doesn't live in one repo. A checkout service calls a payments API. A frontend consumes a backend's REST endpoints. Shared types define contracts between systems. Protocol buffers specify service interfaces.

The moment a developer asks an AI assistant to work across these boundaries, the assistant loses all context for everything outside the current repository.

### 2.2 The Manual Workaround Is Failing

Developers compensate through manual context injection: copying API documentation into chat, pasting proto definitions, verbally describing how another service works. This is slow, error-prone, incomplete, and defeats the purpose of having AI context at all.

The knowledge exists in structured, machine-readable form. It's locked in another repo.

### 2.3 Context Is Write-Only

The gap is not context production — teams are already doing that. The gap is **context orchestration**: getting the right context to the right AI session at the right moment, without human intervention.

Teams produce context that nobody outside their repo consumes. The context layer is write-only. There is no delivery mechanism.

---

## 3. The Solution

**gh-packs** makes AI context composable, versioned, and autonomously routable across repositories. It runs as a `gh` CLI extension — no separate binary, no new authentication, no infrastructure. It consists of three layers.

### 3.1 The Package Layer

Every repository can be both a producer and consumer of context packs. A pack is the contents of a repo's context directory (`.context/`, `.linkedin/context/`, or any configured path) at a tagged git version.

Consuming repos declare dependencies in a `packs.yaml` manifest with semver ranges. Running `gh packs install` resolves versions against git tags, pulls the context directory via `gh`, and writes a `packs.lock` file with exact commit pins and content hashes.

Deterministic. Reproducible. Auditable.

### 3.2 The Graph Layer

Each `packs.yaml` defines edges in a dependency graph. Across repositories, these manifests form a complete, version-controlled map of how services connect. The graph encodes not just what depends on what, but *why*: which files trigger which dependencies, what relationship each dependency has, and what protocols are used.

This graph is traversable by both humans (`gh packs graph`) and AI agents.

### 3.3 The Resolution Layer

The core innovation. When an AI assistant encounters cross-service work, it runs `gh packs resolve "<task description>"`. This spawns a sub-agent via `claude --print` — an independent AI session with its own context window.

The sub-agent loads the full dependency graph and all relevant pack contents (potentially 50KB+), reasons over them, and returns a tight, task-specific briefing (typically 1–3KB) to stdout. The main AI session reads only this synthesized output. The sub-agent then terminates.

Zero context window pollution in the main session.

---

## 4. Architecture

### 4.1 System Components

Four components: the `gh` extension binary (Go), the `gh` CLI (authentication and git operations), the local file system (manifests, lockfiles, installed packs), and the sub-agent (`claude --print`).

No server. No registry service. No persistent background process.

**Data flow:**

```
Developer gives cross-service task
  → Claude Code reads CLAUDE.md, sees packs instruction
  → Claude Code shells out: gh packs resolve "<task>"
  → CLI loads graph, assembles pack content
  → CLI spawns: claude --print (sub-agent with 30KB+ context)
  → Sub-agent reasons over full graph, synthesizes briefing
  → 1–3KB briefing returns to stdout
  → Claude Code reads briefing, writes implementation using real contracts
```

### 4.2 Why gh

Building on `gh` eliminates infrastructure entirely. GitHub becomes the registry.

- **Authentication** — SSO, PAT, OAuth — already configured for every developer
- **Org-level access control** — respects existing GitHub permissions model
- **Sparse content fetching** — pull only the context directory, not the whole repo
- **Tag listing** — version resolution via GitHub API
- **Distribution** — `gh extension install tunajam/gh-packs` and you're done

### 4.3 Why Sub-Agent over MCP

An MCP server is always-on, always in context. It adds tool definitions and schemas to every message — even when the developer is renaming a variable.

The sub-agent pattern:

- **Spawns only when needed** — zero overhead for single-repo work
- **Gets its own context window** — can load 50KB+ of pack content
- **Returns compressed output** — 1–3KB synthesized briefing
- **Dies immediately** — no persistent connection, no lingering state

50KB goes in, 2KB of task-specific intelligence comes out. The main session never sees the noise.

---

## 5. File Specifications

### 5.1 packs.yaml — The Manifest

Declares the repo's identity and its context dependencies. Equivalent to `package.json`.

```yaml
name: "@mycompany/checkout-service"
version: "4.1.0"
description: "Checkout orchestration service"

pack:
  type: "service"                     # service | library | platform | types
  proto: "proto/checkout/v1/*.proto"  # source of truth for API contract

dependencies:
  "@mycompany/payments-api":
    version: "^2.3.0"
    source: "github.com/mycompany/payments-api"
    path: ".context"                  # configurable per dependency
    reason: "We call their refund and charge endpoints"
    relevance:
      - "src/integrations/payments/**"

  "@mycompany/merchant-service":
    version: "^1.7.0"
    source: "github.com/mycompany/merchant-service"
    path: ".linkedin/context"         # supports any directory convention
    reason: "Merchant lookup for webhook configuration"

  "@mycompany/shared-types":
    version: "^1.0.0"
    source: "github.com/mycompany/shared-types"
    path: ".context"
    reason: "Money, Currency, Address types"

resolve:
  max_depth: 2                        # transitive resolution depth
  system_prompt_append: |
    This service uses gRPC for all sync calls.
    Always reference proto definitions as the source of truth.
```

| Field | Required | Description |
|-------|----------|-------------|
| `name` | Yes | Scoped package name (`@org/name`) |
| `version` | Yes | Semver version of this repo's context |
| `description` | No | Human-readable description |
| `pack.type` | No | `service`, `library`, `platform`, `types` |
| `pack.proto` | No | Glob path to proto definitions |
| `dependencies.*.version` | Yes | Semver range (`^`, `~`, exact) |
| `dependencies.*.source` | Yes | GitHub repo (`owner/repo` format) |
| `dependencies.*.path` | No | Path within repo (default: `.context`) |
| `dependencies.*.reason` | No | Why this dependency exists (used by sub-agent for relevance) |
| `dependencies.*.relevance` | No | Glob patterns for files that trigger this dependency |
| `resolve.max_depth` | No | Transitive depth (default: `2`) |
| `resolve.system_prompt_append` | No | Additional sub-agent instructions |

### 5.2 packs.lock — The Lockfile

Pins exact versions and integrity hashes. Auto-generated. Committed to version control.

```yaml
lockfile_version: 1
resolved_at: "2026-02-19T14:32:00Z"

packages:
  "@mycompany/payments-api@2.3.1":
    source: "github.com/mycompany/payments-api"
    ref: "v2.3.1"
    commit: "a1b2c3d4e5f6"
    content_hash: "sha256:e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0"

  "@mycompany/shared-types@1.2.0":
    source: "github.com/mycompany/shared-types"
    ref: "v1.2.0"
    commit: "h7i8j9k0l1m2"
    content_hash: "sha256:c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4"
```

### 5.3 Pack Structure

A pack is a repo's context directory at a specific version. Structure is flexible, but the recommended convention:

```
.context/
  CONTEXT.md          # Architecture, design decisions, system purpose
  API.md              # Endpoints, schemas, method signatures
  CONVENTIONS.md      # Naming, error handling, auth patterns
  INTEGRATION.md      # How to integrate from a consumer's perspective
  proto/              # Proto definitions (source of truth for contracts)
    service/v1/
      service.proto
      types.proto
```

Content is **authored for AI consumption** — structured, opinionated, focused on what an LLM needs to generate correct code. Not general-purpose human documentation.

### 5.4 .packs/resolve-prompt.md — Sub-Agent System Prompt

```markdown
You are a context resolution agent for a software engineering team.

You will receive:
1. A dependency manifest showing how services relate
2. Full context packs for each installed dependency
3. A developer task description

Produce a tight, task-specific briefing that another AI assistant will use
to write code. Include ONLY what is needed for this specific task:

- Exact API endpoints with method signatures and request/response shapes
- Auth patterns for any cross-service calls
- Event schemas if async communication is involved
- Relevant shared types with exact field definitions
- Critical gotchas, rate limits, or constraints

Rules:
- Be precise. Use exact names, paths, and types from the packs.
- Be brief. The consumer has a limited context window.
- Be opinionated. If there is a right way to do it, say so.
- Skip anything not directly relevant to the task.
- Do not explain what you are doing. Just deliver the briefing.
```

Fully customizable per repo. Extendable via `resolve.system_prompt_append` in `packs.yaml`.

---

## 6. CLI Reference

All commands run as `gh packs <command>`.

### Installation

```bash
gh extension install tunajam/gh-packs
```

### Commands

| Command | Description |
|---------|-------------|
| `gh packs init` | Scaffold `packs.yaml`, `.context/`, and `.packs/resolve-prompt.md` |
| `gh packs add <source>` | Add a dependency, resolve version, update manifest and lockfile |
| `gh packs install` | Install all dependencies from lockfile with integrity verification |
| `gh packs resolve "<task>"` | Spawn sub-agent to synthesize task-specific briefing |
| `gh packs show <name>` | Output full contents of a single installed pack |
| `gh packs graph` | Print the dependency tree |
| `gh packs remove <name>` | Remove a dependency |
| `gh packs update` | Update dependencies within semver ranges |
| `gh packs validate` | Lint manifest and verify lockfile integrity |

### gh packs init

```
$ gh packs init

Detected: github.com/mycompany/checkout-service

Created:
  packs.yaml               ← manifest (edit to add dependencies)
  .context/CONTEXT.md      ← your repo's context (fill this in)
  .packs/resolve-prompt.md ← sub-agent system prompt (customizable)
```

### gh packs add

```
$ gh packs add github.com/mycompany/payments-api

Scanning github.com/mycompany/payments-api via gh...
Found .context/ directory
Tags: v2.3.1 (latest), v2.3.0, v2.2.0, v2.1.0
Resolving ^2.3.0 → v2.3.1

  + @mycompany/payments-api@2.3.1       (8.4 KB)
  + @mycompany/shared-types@1.2.0       (3.1 KB)  [transitive]

Updated packs.yaml
Updated packs.lock
Installed to .context/packs/
```

### gh packs install

```
$ gh packs install

Installing from packs.lock...
  @mycompany/payments-api@2.3.1      ✓ verified
  @mycompany/merchant-service@1.7.0  ✓ verified
  @mycompany/shared-types@1.2.0      ✓ verified

3 packs installed in 1.8s
```

### gh packs resolve

```
$ gh packs resolve "add refund webhooks for merchants"

## Task Context: Refund Webhooks for Merchants

**Refund endpoint to extend:**
POST /api/refunds — returns { id, status, amount }
Add: webhook_dispatch field to refund response

**Merchant webhook config:**
GET /api/merchants/:id → includes webhook_config.url and webhook_config.secret
Auth: service JWT, scope merchant:read
Sign payloads with HMAC-SHA256 using webhook_config.secret

**Event to publish:**
Topic: payment.events
Schema: { type: "refund.completed", merchant_id, refund_id, amount: Money }
Money type: { amount_cents: int, currency: string (ISO 4217) }

**Auth pattern for service-to-service:**
Request JWT from auth-service with scopes: [merchant:read, events:publish]
Token goes in Authorization: Bearer header
```

**Flags:**

| Flag | Purpose |
|------|---------|
| `--files <paths>` | Hint at files being modified for better relevance |
| `--depth shallow\|default\|deep` | Control sub-agent reasoning effort |
| `--dry-run` | Show which packs would load without spawning sub-agent |
| `--raw` | Dump all pack content to stdout instead of spawning sub-agent |

### gh packs graph

```
$ gh packs graph

@mycompany/checkout-service (you are here)
  ├── @mycompany/payments-api@2.3.1
  │   ├── @mycompany/shared-types@1.2.0
  │   └── @mycompany/events-platform@3.1.0
  ├── @mycompany/merchant-service@1.7.0
  │   └── @mycompany/shared-types@1.2.0 (deduped)
  └── @mycompany/shared-types@1.2.0 (deduped)
```

Supports `--format mermaid` for documentation and `--format json` for programmatic use.

---

## 7. Sub-Agent

### 7.1 Invocation

```bash
claude --print \
  --system-prompt "$(cat .packs/resolve-prompt.md)" \
  "## Dependency Manifest
$(cat packs.yaml)

## Installed Packs
${ASSEMBLED_PACK_CONTENT}

## Task
${USER_PROMPT}"
```

The Go CLI assembles inputs, invokes this command, and streams the output. All intelligence lives in a disposable Claude session.

### 7.2 Context Budget

| Input | Size |
|-------|------|
| Typical single pack | 3–10 KB |
| 5 dependencies at full depth | 25–50 KB |
| Sub-agent briefing returned | 1–3 KB |

50KB goes in, 2KB of task-specific intelligence comes out.

### 7.3 What the Sub-Agent Can Do That a Context Dump Cannot

Because it has its own reasoning budget, the sub-agent can:

- **Resolve conflicts** — if two packs describe the same schema differently, determine which is current
- **Synthesize integration patterns** — describe the actual glue between services, not just hand over two API docs
- **Filter by task relevance** — a payments pack might have 40 endpoints; the sub-agent surfaces only the 3 that matter
- **Warn about gotchas** — rate limits, deprecations, known issues
- **Cross-reference versions** — note which fields may not be available on older pack versions

---

## 8. AI Integration

### 8.1 CLAUDE.md Convention

Add this block to any repo's `CLAUDE.md`:

```markdown
## Cross-Service Context

This repo uses gh-packs for cross-repo context resolution.

When your task involves other services, APIs, or shared systems:
  Run: gh packs resolve "<description of what you are doing>"

When you need a specific service's full context:
  Run: gh packs show <pack-name>

When you need to understand system topology:
  Run: gh packs graph

Do not guess at other services' APIs, event schemas, or auth patterns.
Always resolve first.
```

### 8.2 How Claude Code Uses It

Claude Code already runs bash commands. No new protocol or integration surface:

1. Developer gives Claude Code a cross-service task
2. Claude Code reads the `CLAUDE.md` instruction
3. Claude Code runs: `gh packs resolve "<task>"`
4. Claude Code reads the synthesized briefing from stdout
5. Claude Code writes the implementation using real contracts

packs is just another CLI tool — one that happens to spawn its own Claude instance for heavy reasoning.

### 8.3 Tool-Agnostic

`gh packs resolve` outputs to stdout. Any AI tool that can execute shell commands can consume it — Claude Code, Cursor, Copilot, Windsurf, Codex. The sub-agent could be swapped for any LLM CLI in the future.

The `--raw` flag bypasses the sub-agent entirely and dumps assembled context to stdout, making packs useful even without Claude CLI installed.

---

## 9. Version Resolution

### Semver Ranges

| Range | Matches | Behavior |
|-------|---------|----------|
| `^2.3.0` | >=2.3.0, <3.0.0 | Caret: compatible with major |
| `~2.3.0` | >=2.3.0, <2.4.0 | Tilde: compatible with minor |
| `2.3.1` | Exactly 2.3.1 | Pinned |
| `*` | Any | Latest |

Tags must follow `v{major}.{minor}.{patch}` convention. The resolver lists all tags via `gh api`, filters to those matching the semver range, and selects the highest match.

### Transitive Resolution

When a pulled pack contains its own `packs.yaml`, dependencies resolve transitively up to `max_depth` (default: 2). Duplicates are deduplicated — highest version satisfying all constraints wins. Conflicting ranges fail with a clear error.

---

## 10. Competitive Landscape

| Tool | What It Does | gh-packs Differentiator |
|------|-------------|------------------------|
| **Repomix** | Serializes a single repo for AI | Resolves *across* repos with a dependency graph |
| **knowhub** | Syncs config files across repos | Package manager semantics: semver, lockfile, resolution |
| **CorePack AI** | Public library context marketplace | Internal org context with dependency graphs + sub-agent |
| **CTX** | Aggregates code from sources into markdown | Versioned, lockable, with AI-powered synthesis |
| **Qodo / Moderne** | Enterprise multi-repo AI platforms | Lightweight CLI, zero infrastructure, developer-first |

**The gap:** nobody combines versioned context packages, an inter-repo dependency graph, sub-agent resolution, and zero-infrastructure deployment on top of `gh`.

---

## 11. Adoption Path

### Getting Started

1. Pick two repos that frequently integrate
2. Run `gh packs init` in the consuming repo
3. Run `gh packs add` pointing to the producing repo
4. Add the `CLAUDE.md` block
5. Try `gh packs resolve` on a real cross-service task
6. Compare the output quality to working without packs

### The Flywheel

1. A team publishes their context directory (already happening)
2. A consuming team runs `gh packs add` and gets versioned context
3. Claude produces correct cross-service code via `gh packs resolve`
4. The consuming team trusts AI output more and tells other teams
5. Other teams request packs for services that don't have them yet
6. More teams publish and improve their context

### Second-Order Effects

When context has consumers, quality improves organically:

- **Context becomes a team artifact** with real consumers and observable quality
- **New services ship with context** — it becomes part of service creation, not an afterthought
- **Context review joins code review** — API changes get `.context/` updates in the same PR

---

## 12. Future Directions

- **Registry** — hosted resolution at packs.sh for cross-org and public packs
- **Breaking change detection** — diff proto definitions between versions, notify affected consumers
- **Org-wide graph** — impact analysis, coupling detection, onboarding acceleration
- **Public ecosystem** — community packs for Stripe, AWS, Supabase, etc.
- **Multi-agent routing** — planning agents query the graph, implementation agents resolve specific packs

---

## 13. MVP Scope

### What Ships

| Priority | Feature |
|----------|---------|
| P0 | `gh packs init` — scaffold manifest and context directory |
| P0 | `gh packs add` — add dependency via `gh`, resolve version |
| P0 | `gh packs install` — install from lockfile with integrity checks |
| P0 | `gh packs resolve` — sub-agent via `claude --print` |
| P0 | `packs.yaml` manifest + `packs.lock` lockfile |
| P1 | `gh packs show`, `graph`, `remove`, `update`, `validate` |

### Prerequisites

- Go 1.22+ (CLI binary)
- `gh` CLI installed and authenticated
- `claude` CLI installed (for `packs resolve`; `--raw` works without it)
- Repos with context directories tagged with semver versions

### What MVP Proves

1. Versioned context packages work — lockfile pins exact context per dependency
2. Sub-agent resolution works — Claude gets cross-service intelligence without context window bloat
3. Git-as-registry works — zero infrastructure, `gh` handles everything
4. AI output quality improves measurably for cross-service tasks

---

*Context is a dependency. Version it. Install it. Resolve it.*
