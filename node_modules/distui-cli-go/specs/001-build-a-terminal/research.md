# Technical Research: distui Implementation

## Executive Summary
distui requires a Go-based TUI that manages release distributions without polluting user repositories. The solution leverages Bubble Tea for the TUI framework, direct command execution via os/exec, and file-based YAML storage in ~/.distui. All operations execute directly from the TUI with real-time feedback.

## Technology Selection

### TUI Framework: Bubble Tea
**Rationale**: Industry-standard Go TUI framework with excellent documentation and community support.
- Elm-inspired architecture fits perfectly with state management needs
- Built-in support for concurrent operations (release processes)
- Excellent keyboard handling for navigation requirements
- Active maintenance by Charm team

### Styling: Lipgloss
**Rationale**: Companion library to Bubble Tea for terminal styling.
- Declarative styling approach aligns with code clarity principles
- Supports all required visual elements (progress bars, tables, forms)
- Minimal overhead, pure Go implementation

### Configuration Storage: YAML Files
**Rationale**: Human-readable, version-control friendly, no database dependencies.
- Simple to implement and debug
- Users can manually edit if needed (no vendor lock-in)
- Atomic file operations prevent corruption
- Supports complex nested structures for project configs

### Command Execution: os/exec
**Rationale**: Direct command execution without intermediate scripts.
- Real-time output streaming to TUI
- Full control over process lifecycle
- No shell injection vulnerabilities
- Platform-native execution

## Implementation Patterns

### Configuration Management
```
~/.distui/
├── config.yaml         # Global user settings
├── projects/
│   ├── github-com-user-project1.yaml
│   └── github-com-user-project2.yaml
└── cache/
    └── detections.json # Cached detection results
```

### View Navigation Pattern
- TAB key cycles through views sequentially
- Direct shortcuts (P, G, S) jump to specific views
- View state preserved during navigation
- Breadcrumb trail shows current location

### Release Execution Pattern
1. Pre-flight checks (tests, git status)
2. Version determination
3. Tag creation
4. GoReleaser execution
5. Distribution channel updates (parallel where possible)
6. Post-release verification

## Dependencies

### Core Dependencies
```go
require (
    github.com/charmbracelet/bubbletea v0.27.0
    github.com/charmbracelet/lipgloss v0.13.0
    gopkg.in/yaml.v3 v3.0.1
)
```

### Development Dependencies
```go
require (
    github.com/stretchr/testify v1.9.0
)
```

### External Tool Dependencies
- `gh` CLI (GitHub operations)
- `git` (version control)
- `goreleaser` (release builds)
- `brew` (Homebrew operations)
- `npm` (NPM publishing)

## Risk Analysis

### Technical Risks
1. **Terminal Compatibility**
   - Mitigation: Fallback to basic ANSI, detect terminal capabilities

2. **Command Execution Failures**
   - Mitigation: Comprehensive error handling, rollback capability

3. **File System Permissions**
   - Mitigation: Check ~/.distui writability on startup

4. **Concurrent Access**
   - Mitigation: File locking for project configs

### Operational Risks
1. **Missing External Tools**
   - Mitigation: Graceful degradation, clear error messages

2. **Network Dependencies**
   - Mitigation: Timeout controls, retry logic

3. **Large Output Streams**
   - Mitigation: Buffered streaming, output truncation

## Performance Considerations

### UI Responsiveness
- All I/O operations in goroutines
- Command output streamed, not buffered entirely
- Debounced keyboard input for navigation
- Lazy loading of project lists

### Memory Management
- Stream command output instead of buffering
- Limited history retention (last 10 releases)
- Periodic cache cleanup
- Config files loaded on-demand

### Startup Optimization
- Parallel detection operations
- Cached detection results (5-minute TTL)
- Lazy view initialization
- Minimal initial file I/O

## Security Considerations

### No Shell Injection
- Direct command execution without shell interpretation
- All arguments properly escaped
- No user input in command construction

### Token Management
- Rely on gh CLI's token management
- Never store tokens in distui configs
- NPM tokens via environment variables only

### File Permissions
- Config files created with 0600 (user-only read/write)
- Atomic writes prevent partial updates
- No world-readable sensitive data

## Integration Points

### gh CLI Integration
```go
cmd := exec.Command("gh", "repo", "view", "--json", "name,owner")
// Parse JSON output for repository detection
```

### GoReleaser Integration
```go
cmd := exec.Command("goreleaser", "release", "--clean")
cmd.Env = append(os.Environ(), "GITHUB_TOKEN=...")
// Stream output to TUI
```

### Homebrew Tap Updates
```go
// Download tarball, calculate SHA256
// Update formula file with sed-like replacement
// Git commit and push to tap repository
```

## Testing Strategy

### Unit Testing
- Mock command execution for predictable tests
- Table-driven tests for configuration logic
- Property-based testing for state machines

### Integration Testing
- Real command execution in isolated environment
- Temporary directory for test configurations
- Docker containers for external tool simulation

### Manual Testing Checklist
- [ ] Navigation with TAB and shortcuts
- [ ] Release execution under 30 seconds
- [ ] Error handling for failed commands
- [ ] Configuration persistence
- [ ] Multi-project management

## Implementation Timeline

### Week 1: Foundation
- Configuration management
- Basic TUI structure
- Navigation implementation

### Week 2: Core Features
- Project detection
- Release execution
- Distribution channels

### Week 3: Polish
- Error handling
- Progress indicators
- Performance optimization

### Week 4: Testing & Documentation
- Comprehensive test coverage
- User documentation
- Example configurations