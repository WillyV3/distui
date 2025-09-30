<!--
Sync Impact Report
Version change: 1.1.0 → 1.2.0 (Pragmatic repository files amendment)
Modified principles:
  - "Zero Repository Pollution" → "Pragmatic Repository Files" - Allow essential distribution config files with user consent
  - "Structural Discipline" (Code Quality Standards) - Maintained pragmatic file size guidance
Added sections: None
Removed sections: None
Templates requiring updates:
  - ⚠ plan-template.md (update for allowed repo files: .goreleaser.yaml, package.json, CI workflows)
  - ⚠ tasks-template.md (add tasks for config file generation)
  - ✅ constitution command (this file)
Follow-up TODOs:
  - Implement .goreleaser.yaml generation from distui config
  - Implement package.json generation when NPM enabled
  - Add user consent flow before generating repo files
  - Consider refactoring configure_handler.go (989 lines) into smaller composable modules
-->

# distui Constitution

## Core Principles

### I. Pragmatic Repository Files
All distui state MUST be stored in ~/.distui, never in user repositories.
However, distribution tools (GoReleaser, NPM) require configuration files
in the repository to function. distui MAY generate these essential files
with explicit user consent:

**ALLOWED (with user consent):**
- .goreleaser.yaml (required by GoReleaser for releases)
- package.json (required by NPM for publishing)
- .github/workflows/*.yml (optional, if CI/CD generation enabled)

**NEVER ALLOWED:**
- .distui.yaml or similar distui state files
- Generated shell scripts or temporary files
- Lock files, caches, or build artifacts

Generated configuration files MUST be:
- Idempotent (safe to regenerate without side effects)
- Human-readable and editable
- Safe to commit to version control
- Generated only after user confirmation

Projects remain under user control. Once generated, config files belong
to the user and can be modified without distui's involvement.

### II. 30-Second Release Execution
Release execution MUST complete within 30 seconds for typical projects. The TUI
executes all commands directly - no intermediate script generation, no waiting for
CI/CD pipelines. Direct execution means immediate feedback and rapid iteration.

### III. User Agency and Navigation Freedom
Users MUST always control their navigation path. The TUI detects project context
but never forces users into a specific mode. TAB cycles between views, keyboard
shortcuts provide direct jumps, and users choose their workflow - not the tool
choosing for them.

### IV. Stateful Global Intelligence
distui MUST remember all project configurations globally in ~/.distui/projects/.
Working across multiple projects requires no re-configuration. Each project's
settings persist, evolve, and travel with the developer, not the repository.

### V. Clean Go Code Excellence
All code MUST use Bubble Tea for TUI framework and Lipgloss for styling. Go
best practices are non-negotiable: clear interfaces, proper error handling,
idiomatic patterns. The codebase serves as an exemplar of Go TUI development.

### VI. Direct Command Execution
The TUI MUST execute commands directly within its process: go test, gh release
create, brew tap updates, npm publish. No script generation, no intermediate
files. Commands run with real-time output visible in the TUI, errors handled
interactively.

### VII. Developer Choice Architecture
Support BOTH rapid local releases AND optional CI/CD workflow generation. Some
developers want 30-second local releases, others want GitHub Actions automation,
many want both. The tool adapts to developer preference, not vice versa.

### VIII. Smart Detection with Override
Use gh CLI for intelligent detection of repositories, taps, authentication - but
ALWAYS allow user override. Detection provides convenience, not constraints.
Every detected value can be modified, every assumption challenged.

### IX. No Vendor Lock-in
Configuration MUST use readable YAML that works without distui. Once configured,
projects can be released manually using the same commands distui would execute.
The tool adds convenience, not dependency.

### X. Clean Configuration Separation
Global settings in ~/.distui/config.yaml, per-project in ~/.distui/projects/.
Clear boundaries between what affects all projects versus specific ones. No
configuration mixing, no inheritance confusion.

## Code Quality Standards

### Self-Documenting Code
- Every variable, function, and type name MUST be immediately understandable
- Code reads like prose, not puzzles - clarity over cleverness
- NO comments explaining what code does (except API documentation)
- If code needs comments, rewrite the code to be clearer

### Structural Discipline
Files SHOULD be kept concise and focused. While a 100-line guideline is ideal
for single-terminal-screen visibility, pragmatism is required:

- **Essential files only**: Files may exceed 100 lines when they contain only
  essential, non-redundant logic that serves a cohesive purpose
- **No arbitrary splits**: Do not split files artificially just to meet a line
  count if it creates confusion or breaks natural cohesion
- **Natural boundaries**: When files grow large, look for natural module
  boundaries (separate concerns, extract reusable components)
- **Refactoring targets**: Files exceeding 300 lines are strong candidates for
  refactoring into composable modules

Nesting and control flow:
- Nesting MUST be minimized - use early returns and guard clauses
- NO nested conditionals beyond absolutely necessary cases
- Accept repetition if it clarifies control flow
- Helper functions require rigorous justification

### Error Philosophy
- NO try/catch blocks unless absolutely critical
- Failures should be catastrophic and visible
- Business errors handled through types, not catches
- Errors bubble up, not swept under rugs

## Development Workflow

### Test-Driven Development
- Write tests first for new functionality
- Tests must be clear, focused, and fast
- Integration tests for command execution
- Unit tests for business logic
- No test, no merge

### Code Review Standards
- Every PR reviewed against constitution principles
- Alignment score (0-10) for code clarity
- Violations identified with specific fixes
- Refactoring required before merge if standards not met

### Performance Requirements
- TUI responsiveness < 100ms for user actions
- Release execution < 30 seconds typical case
- Memory usage < 50MB for normal operations
- Startup time < 1 second

## Governance

The constitution supersedes all other practices and conventions. It defines the
non-negotiable principles that guide all development decisions.

### Amendment Process
- Proposed changes require written justification
- Breaking changes to principles require major version bump
- New principles require minor version bump
- Clarifications require patch version bump

### Compliance Verification
- All code reviews MUST verify constitutional compliance
- Automated checks where possible (nesting depth, error handling patterns)
- Manual review for clarity and philosophy alignment
- Non-compliant code blocked from merge

### Living Document
This constitution evolves with the project but changes are deliberate and
documented. Each amendment requires clear rationale and migration plan if
breaking existing patterns.

**Version**: 1.2.0 | **Ratified**: 2025-09-28 | **Last Amended**: 2025-09-29