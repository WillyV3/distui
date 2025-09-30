
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

1. **Setup Tasks**:
   - T-SETUP-1: Add doublestar dependency to go.mod
   - T-SETUP-2: Create internal/workflow package structure
   - T-SETUP-3: Update internal/models/types.go with new structs

2. **Bug Fix Tasks** (High Priority):
   - T-BUG-1: Fix dot file handling in internal/gitcleanup/categorize.go
   - T-BUG-2: Add test for dot file categorization

3. **Smart Commit Preferences Tasks**:
   - T-SCP-1: [P] Create CategoryRules model in internal/models/types.go
   - T-SCP-2: [P] Add smart_commit section parsing to internal/config/loader.go
   - T-SCP-3: [P] Create pattern matching logic with doublestar
   - T-SCP-4: Update internal/gitcleanup/categorize.go to use custom rules
   - T-SCP-5: Create handlers/smart_commit_prefs_handler.go
   - T-SCP-6: Create views/smart_commit_prefs_view.go
   - T-SCP-7: Integrate into Configure View Advanced tab
   - T-SCP-8: Add default rules reset functionality
   - T-SCP-9: [P] Write unit tests for pattern matching
   - T-SCP-10: [P] Write integration tests for preferences UI

4. **Workflow Generation Tasks**:
   - T-WF-1: [P] Create internal/workflow/template.go with embedded YAML
   - T-WF-2: [P] Create internal/workflow/generator.go
   - T-WF-3: Add workflow_generation section parsing to config loader
   - T-WF-4: Create handlers/workflow_gen_handler.go
   - T-WF-5: Create views/workflow_gen_view.go
   - T-WF-6: Integrate into Configure View Advanced tab
   - T-WF-7: Add preview modal for workflow YAML
   - T-WF-8: Add file generation with user consent
   - T-WF-9: [P] Write tests for template generation
   - T-WF-10: [P] Write tests for workflow validation

5. **Integration Tasks**:
   - T-INT-1: Update Configure View to handle Advanced tab expansion
   - T-INT-2: Add navigation between smart commit prefs and workflow gen
   - T-INT-3: Wire up save/load for both feature configs
   - T-INT-4: Test full integration with existing features
   - T-INT-5: Update quickstart.md scenarios

6. **Polish Tasks**:
   - T-POL-1: Add error handling for invalid patterns
   - T-POL-2: Add loading states for async operations
   - T-POL-3: Add keyboard shortcuts documentation
   - T-POL-4: Performance testing for pattern matching

**Ordering Strategy**:
- Setup first (dependencies, models)
- Bug fix next (blocks smart commit improvements)
- Features in parallel where independent
- Integration after both features complete
- Polish last

**Dependency Rules**:
- T-SETUP-* must complete before feature tasks
- T-BUG-* should complete before T-SCP-4
- T-SCP-1,2,3 can run parallel
- T-WF-1,2 can run parallel
- T-SCP-5,6,7 depend on T-SCP-1,2,3,4
- T-WF-4,5,6 depend on T-WF-1,2,3
- T-INT-* depend on all feature tasks
- T-POL-* can run parallel after integration

**Estimated Output**: ~35 tasks in tasks.md

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
- [x] Phase 0: Research complete (/plan command) - research.md created
- [x] Phase 1: Design complete (/plan command) - data-model.md updated, contracts updated, quickstart.md created, CLAUDE.md updated
- [x] Phase 2: Task planning complete (/plan command - describe approach only) - 35 tasks planned
- [ ] Phase 3: Tasks generated (/tasks command) - READY TO EXECUTE
- [ ] Phase 4: Implementation complete
- [ ] Phase 5: Validation passed

**Gate Status**:
- [x] Initial Constitution Check: PASS - All 10 principles met
- [x] Post-Design Constitution Check: PASS - No new violations
- [x] All NEEDS CLARIFICATION resolved - No unknowns remain
- [x] Complexity deviations documented - NONE (no violations)

**Artifacts Created**:
- [x] `/specs/001-build-a-terminal/plan.md` - This file
- [x] `/specs/001-build-a-terminal/research.md` - Technical decisions documented
- [x] `/specs/001-build-a-terminal/quickstart.md` - Integration test scenarios
- [x] `/specs/001-build-a-terminal/data-model.md` - Updated with new entities (already existed)
- [x] `/specs/001-build-a-terminal/contracts/project.yaml` - Updated with new config sections (already done)
- [x] `/CLAUDE.md` - Updated with new dependencies

**Ready for Next Phase**: Execute `/tasks` command to generate tasks.md

---
*Based on Constitution v1.3.0 (TUI Layout Integrity) - See `.specify/memory/constitution.md`*
