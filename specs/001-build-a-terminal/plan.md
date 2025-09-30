
# Implementation Plan: Smart Commit Preferences & GitHub Workflow Generation

**Branch**: `001-build-a-terminal` | **Date**: 2025-09-30 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/Users/williamvansickleiii/charmtuitemplate/distui/distui-app/specs/001-build-a-terminal/spec.md`

## Execution Flow (/plan command scope)
```
1. Load feature spec from Input path
   → If not found: ERROR "No feature spec at {path}"
2. Fill Technical Context (scan for NEEDS CLARIFICATION)
   → Detect Project Type from file system structure or context (web=frontend+backend, mobile=app+api)
   → Set Structure Decision based on project type
3. Fill the Constitution Check section based on the content of the constitution document.
4. Evaluate Constitution Check section below
   → If violations exist: Document in Complexity Tracking
   → If no justification possible: ERROR "Simplify approach first"
   → Update Progress Tracking: Initial Constitution Check
5. Execute Phase 0 → research.md
   → If NEEDS CLARIFICATION remain: ERROR "Resolve unknowns"
6. Execute Phase 1 → contracts, data-model.md, quickstart.md, agent-specific template file (e.g., `CLAUDE.md` for Claude Code, `.github/copilot-instructions.md` for GitHub Copilot, `GEMINI.md` for Gemini CLI, `QWEN.md` for Qwen Code or `AGENTS.md` for opencode).
7. Re-evaluate Constitution Check section
   → If new violations: Refactor design, return to Phase 1
   → Update Progress Tracking: Post-Design Constitution Check
8. Plan Phase 2 → Describe task generation approach (DO NOT create tasks.md)
9. STOP - Ready for /tasks command
```

**IMPORTANT**: The /plan command STOPS at step 7. Phases 2-4 are executed by other commands:
- Phase 2: /tasks command creates tasks.md
- Phase 3-4: Implementation execution (manual or via tools)

## Summary

This plan adds two major enhancements and one bug fix to the production-ready distui v0.0.31:

1. **Smart Commit Preferences** (FR-040 to FR-047): Allow project-level customization of file categorization rules. Users can define custom file extensions and glob patterns for each category (config, code, docs, build, test, assets, data). Settings stored in project YAML, with implementation kept separate from configure_handler.go per user request.

2. **GitHub Actions Workflow Generation** (FR-048 to FR-056): Optional/opt-in CI/CD workflow file generation. Respects developer autonomy - easy to disable, never forced. Generates .github/workflows/release.yml based on enabled distribution channels with preview before creation.

3. **Bug Fix**: Fix handling of dot files/directories (e.g., .github, .goreleaser) in commit settings interface.

Technical approach follows existing patterns: separate handlers for new features, YAML-based configuration, TUI forms for editing, constitutional compliance (no repository pollution except with consent, user agency preserved).

## Technical Context
**Language/Version**: Go 1.21+
**Primary Dependencies**: Bubble Tea v0.27.0, Lipgloss v0.13.0, yaml.v3, doublestar (glob matching)
**Storage**: YAML files in ~/.distui/projects/<id>/config.yaml
**Testing**: go test, table-driven tests
**Target Platform**: Terminal (Linux, macOS, Windows with proper terminal emulator)
**Project Type**: Single (TUI application)
**Performance Goals**: <100ms UI response, <500ms config save, instant pattern validation
**Constraints**: Files under 300 lines (strong refactoring target), no nested conditionals, separate handlers from configure_handler.go
**Scale/Scope**: 3 new handlers (~200 lines each), 2 new models (~150 lines each), YAML contract updates, 1 bug fix

## Constitution Check
*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### Principle I: Pragmatic Repository Files ✅
- Smart commit preferences stored in ~/.distui/projects/<id>/config.yaml (NOT in repository)
- GitHub workflow generation creates .github/workflows/*.yml WITH EXPLICIT USER CONSENT
- Workflow files are optional, user-controlled, and can be disabled
- **PASS**: Follows constitution - no forced repository pollution

### Principle II: 30-Second Release Execution ✅
- New features don't impact release execution time
- Configuration editing happens outside release workflow
- **PASS**: No impact on release performance

### Principle III: User Agency and Navigation Freedom ✅
- Smart commit preferences are optional (defaults provided)
- Workflow generation is opt-in (easy to disable)
- No forced navigation paths or modes
- **PASS**: User maintains full control

### Principle IV: Stateful Global Intelligence ✅
- Smart commit preferences per-project in ~/.distui/projects/
- Settings persist and travel with developer
- **PASS**: Follows global configuration pattern

### Principle V: Clean Go Code Excellence ✅
- New handlers separate from configure_handler.go per user request
- Bubble Tea + Lipgloss for all UI
- Self-documenting code, early returns, minimal nesting
- **PASS**: Maintains code quality standards

### Principle VI: Direct Command Execution ✅
- No script generation (workflow files are YAML config, not scripts)
- Pattern validation happens in-process
- **PASS**: No intermediate execution layer

### Principle VII: Developer Choice Architecture ✅
- Supports BOTH local releases AND optional CI/CD generation
- Workflow generation is opt-in, not required
- **PASS**: Respects developer preference

### Principle VIII: Smart Detection with Override ✅
- Default categorization rules provided
- Users can override any/all rules
- **PASS**: Detection + override pattern followed

### Principle IX: No Vendor Lock-in ✅
- YAML configuration readable without distui
- Workflow files are standard GitHub Actions YAML
- **PASS**: No proprietary formats

### Principle X: Clean Configuration Separation ✅
- Smart commit preferences in project config (not global)
- Workflow settings in project config
- **PASS**: Clear boundaries maintained

### Code Quality Standards ✅
- Separate handlers for new features (not in configure_handler.go)
- Files kept under 300 lines
- Self-documenting names, no comments except API docs
- Early returns, minimal nesting
- **PASS**: Follows all quality standards

**GATE STATUS: PASS** - All constitutional requirements met, no violations to document

## Project Structure

### Documentation (this feature)
```
specs/[###-feature]/
├── plan.md              # This file (/plan command output)
├── research.md          # Phase 0 output (/plan command)
├── data-model.md        # Phase 1 output (/plan command)
├── quickstart.md        # Phase 1 output (/plan command)
├── contracts/           # Phase 1 output (/plan command)
└── tasks.md             # Phase 2 output (/tasks command - NOT created by /plan)
```

### Source Code (repository root)
```
distui-app/
├── app.go                                    # Main TUI (no changes)
├── go.mod                                    # Add doublestar dependency
├── handlers/
│   ├── configure_handler.go                 # Existing (minimal changes)
│   ├── smart_commit_prefs_handler.go        # NEW: Smart commit preferences
│   ├── workflow_gen_handler.go              # NEW: GitHub workflow generation
│   └── cleanup_handler.go                   # FIX: Dot file handling bug
├── views/
│   ├── smart_commit_prefs_view.go           # NEW: Preferences UI
│   └── workflow_gen_view.go                 # NEW: Workflow generation UI
├── internal/
│   ├── models/
│   │   └── types.go                         # UPDATE: Add SmartCommitPrefs, WorkflowConfig
│   ├── config/
│   │   └── loader.go                        # UPDATE: Load/save new config sections
│   ├── gitcleanup/
│   │   ├── categorize.go                    # UPDATE: Use custom rules if enabled
│   │   └── dotfiles.go                      # NEW: Fix dot file handling
│   └── workflow/
│       ├── generator.go                     # NEW: Generate GitHub Actions YAML
│       └── template.go                      # NEW: Workflow template
└── specs/001-build-a-terminal/
    ├── plan.md                              # This file
    ├── research.md                          # Phase 0 output
    ├── data-model.md                        # UPDATE: New entities
    ├── contracts/                           # UPDATE: project.yaml already done
    └── quickstart.md                        # Phase 1 output
```

**Structure Decision**: Single project TUI application. New features implemented as separate handlers and views following existing patterns. Bug fix isolated to gitcleanup package.

## Phase 0: Outline & Research
1. **Extract unknowns from Technical Context** above:
   - For each NEEDS CLARIFICATION → research task
   - For each dependency → best practices task
   - For each integration → patterns task

2. **Generate and dispatch research agents**:
   ```
   For each unknown in Technical Context:
     Task: "Research {unknown} for {feature context}"
   For each technology choice:
     Task: "Find best practices for {tech} in {domain}"
   ```

3. **Consolidate findings** in `research.md` using format:
   - Decision: [what was chosen]
   - Rationale: [why chosen]
   - Alternatives considered: [what else evaluated]

**Output**: research.md with all NEEDS CLARIFICATION resolved

## Phase 1: Design & Contracts
*Prerequisites: research.md complete*

1. **Extract entities from feature spec** → `data-model.md`:
   - Entity name, fields, relationships
   - Validation rules from requirements
   - State transitions if applicable

2. **Generate API contracts** from functional requirements:
   - For each user action → endpoint
   - Use standard REST/GraphQL patterns
   - Output OpenAPI/GraphQL schema to `/contracts/`

3. **Generate contract tests** from contracts:
   - One test file per endpoint
   - Assert request/response schemas
   - Tests must fail (no implementation yet)

4. **Extract test scenarios** from user stories:
   - Each story → integration test scenario
   - Quickstart test = story validation steps

5. **Update agent file incrementally** (O(1) operation):
   - Run `.specify/scripts/bash/update-agent-context.sh claude`
     **IMPORTANT**: Execute it exactly as specified above. Do not add or remove any arguments.
   - If exists: Add only NEW tech from current plan
   - Preserve manual additions between markers
   - Update recent changes (keep last 3)
   - Keep under 150 lines for token efficiency
   - Output to repository root

**Output**: data-model.md, /contracts/*, failing tests, quickstart.md, agent-specific file

## Phase 2: Task Planning Approach
*This section describes what the /tasks command will do - DO NOT execute during /plan*

**Task Generation Strategy**:
- Load `.specify/templates/tasks-template.md` as base
- Generate tasks from Phase 1 design docs (contracts, data model, quickstart)
- Each contract → contract test task [P]
- Each entity → model creation task [P] 
- Each user story → integration test task
- Implementation tasks to make tests pass

**Ordering Strategy**:
- TDD order: Tests before implementation 
- Dependency order: Models before services before UI
- Mark [P] for parallel execution (independent files)

**Estimated Output**: 25-30 numbered, ordered tasks in tasks.md

**IMPORTANT**: This phase is executed by the /tasks command, NOT by /plan

## Phase 3+: Future Implementation
*These phases are beyond the scope of the /plan command*

**Phase 3**: Task execution (/tasks command creates tasks.md)  
**Phase 4**: Implementation (execute tasks.md following constitutional principles)  
**Phase 5**: Validation (run tests, execute quickstart.md, performance validation)

## Complexity Tracking
*Fill ONLY if Constitution Check has violations that must be justified*

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| [e.g., 4th project] | [current need] | [why 3 projects insufficient] |
| [e.g., Repository pattern] | [specific problem] | [why direct DB access insufficient] |


## Progress Tracking
*This checklist is updated during execution flow*

**Phase Status**:
- [ ] Phase 0: Research complete (/plan command)
- [ ] Phase 1: Design complete (/plan command)
- [ ] Phase 2: Task planning complete (/plan command - describe approach only)
- [ ] Phase 3: Tasks generated (/tasks command)
- [ ] Phase 4: Implementation complete
- [ ] Phase 5: Validation passed

**Gate Status**:
- [ ] Initial Constitution Check: PASS
- [ ] Post-Design Constitution Check: PASS
- [ ] All NEEDS CLARIFICATION resolved
- [ ] Complexity deviations documented

---
*Based on Constitution v2.1.1 - See `/memory/constitution.md`*
