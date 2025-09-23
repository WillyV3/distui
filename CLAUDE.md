# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**Willy's Back Shed** is a production-ready SSH-served TUI application built with Bubble Tea, featuring a multi-page personal dashboard with 47+ example applications. The main application serves as WillyV3's digital tool shed, providing navigation to blog posts, homebrew packages, and GitHub repositories through a beautiful, animated terminal interface.

### 0. No Fallbacks Rule - Fail Fast and Obvious
- **NEVER write fallback behavior** - if something is wrong, fail immediately
- **NO silent defaults** - missing data should panic or show clear error
- **NO recovery attempts** - let errors bubble up visibly

### Core Architecture

The application follows a clean separation of concerns with a proven 3-layer architecture:

- **Main Application** (`main.go`): Direct SSH server on port 2234 with Ed25519 key generation
- **App Router** (`app.go`): Central routing with enum-based page state management
- **Handlers** (`handlers/`): Page-specific update logic separated by functionality
- **Views** (`views/`): Page-specific rendering logic with consistent layout patterns
- **Components** (`components/`): Reusable UI elements and animations
- **Examples** (`examples/`): 47+ comprehensive TUI pattern demonstrations

#### Technical Stack
- **SSH Server**: Direct connection using Charm's Wish library
- **TUI Framework**: Bubble Tea with Model-Update-View pattern
- **UI Components**: Bubbles library for proven terminal components
- **Styling**: Lipgloss for consistent theming and responsive layouts
- **Security**: Ed25519 SSH keys with public key authentication

## Development Commands

### Main Application Operations
```bash
# Run the SSH TUI server (connects on localhost:2234)
go run main.go

# Build the production application
go build -o willysbackshed main.go

# Connect to the running SSH server
ssh localhost -p 2234

# Build with optimization
go build -ldflags="-s -w" -o willysbackshed main.go
```

### Example Applications (47+ Available)
```bash
# Navigate to examples directory
cd examples

# Run individual examples for testing patterns
go run simple/main.go          # Basic Bubble Tea structure
go run chat/main.go            # Real-time chat interface
go run table/main.go           # Data table with sorting
go run spinner/main.go         # Loading animations
go run list-fancy/main.go      # Advanced list components

# Build specific example
go build -o example-app simple/main.go

# Build all examples for distribution
go build ./...
```

### Development Testing
```bash
# Test main application
go test ./...

# Test specific components
go test ./handlers/
go test ./views/
go test ./components/

# Run with verbose output
go test -v ./...

# Test examples (individual testing)
cd examples && go test ./simple/
```

## Example Categories and Architecture

### Core UI Patterns
- **simple**: Basic Bubble Tea application structure
- **composable-views**: Multiple view composition patterns
- **focus-blur**: Component focus management
- **fullscreen**: Full-screen application layouts
- **altscreen-toggle**: Alternate screen buffer usage

### Input Handling
- **textinput**: Single text input components
- **textinputs**: Multiple text input management
- **textarea**: Multi-line text editing
- **autocomplete**: Auto-completion functionality
- **mouse**: Mouse interaction patterns

### Data Display
- **table**: Tabular data presentation with sorting/filtering
- **list-simple/list-fancy/list-default**: Various list implementations
- **pager**: Document viewing and pagination
- **paginator**: Page-based navigation

### Interactive Components
- **chat**: Real-time chat interface
- **credit-card-form**: Form validation patterns
- **file-picker**: File system navigation
- **progress-animated/progress-static**: Progress indicators
- **spinner/spinners**: Loading indicators

### Advanced Patterns
- **tui-daemon-combo**: Combining TUI with background services
- **send-msg**: Inter-component messaging
- **exec**: External command execution
- **http**: HTTP client integration
- **realtime**: Real-time data updates

## Configuration System

### Main Application Configuration
- **config.yaml**: YAML-based endpoint configuration
- **config**: SSH config format support
- Supports both local TUI apps and remote SSH proxying

### Example Configuration Structure
```yaml
listen: 127.0.0.1
port: 2223
endpoints:
  - name: "example-app"
    description: "Demo application"
  - name: "remote-host"
    address: "host:22"
    user: "username"
```

## Development Patterns

### Bubble Tea Application Structure
1. **Model**: Application state management
2. **Init()**: Initial commands and setup
3. **Update()**: Message handling and state transitions
4. **View()**: UI rendering and layout

### Component Integration
- Use Bubbles components for common UI elements
- Implement custom components following Bubble Tea patterns
- Handle keyboard/mouse events through the Update() method
- Manage component focus and navigation

### Styling Best Practices
- Use Lipgloss for consistent styling across components
- Define style constants for reusable design elements
- Handle terminal size changes gracefully
- Support both light and dark terminal themes

## Key Files and Structure

- `main.go`: Wishlist-style SSH server with multiple endpoints
- `examples/`: Comprehensive collection of Bubble Tea patterns
- `config.yaml`/`config`: Configuration files for server setup
- `docs/`: Documentation and guides
- Individual example directories contain focused demonstrations

## Integration Notes

### SSH Server Integration
The main application demonstrates serving Bubble Tea apps over SSH:
- SSH key generation and management
- Multiple endpoint configuration
- Middleware composition for request handling
- Session management for concurrent users

### Terminal Compatibility
- Uses activeterm middleware for better terminal support
- Handles various terminal emulators and SSH clients
- Supports mouse interaction where available
- Graceful fallbacks for limited terminal capabilities

## Development Workflow

### User Handles Building
- User will run `go build` and `go run` commands
- Claude should not execute build/run commands

### Adding New Pages (3-Step Pattern)
1. **Add page enum**: `newPage pageState = iota` in app.go
2. **Create handler**: `handlers/newpagehandler.go` with `UpdateNewPage()` function
3. **Create view**: `views/newpageview.go` with `RenderNewPageContent()` function
4. **Add routing**: Case in app.go switch statement calling handler
- Layout automatically applied via `m.renderPage(title, content)`
- No complex registration needed - clean separation of concerns
- **Keep handlers in correct files** - handlers go in `handlers/` package, views in `views/` package

### Creating New Examples
1. Create new directory in `examples/`
2. Follow the standard Bubble Tea pattern (Init/Update/View)
3. Include appropriate keyboard shortcuts (q/esc/ctrl+c to quit)
4. Add README.md with description and screenshots if applicable

### Testing Applications
- Test in various terminal emulators
- Verify keyboard shortcuts work correctly
- Check responsiveness to terminal resize events
- Test with different terminal color capabilities

### Debugging
- Use the logging middleware for request debugging
- Add debug output to terminal for state inspection
- Test both local execution and SSH connectivity