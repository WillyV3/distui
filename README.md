# TUI Template App

A production-ready SSH-served TUI application template built with Bubble Tea. This template provides a clean 3-layer architecture for building terminal user interfaces.

## Features

- **SSH Server**: Direct SSH connection with Ed25519 key generation
- **Multi-page Navigation**: Clean page routing system
- **Responsive Layouts**: Automatically adapts to terminal size
- **Animated Elements**: Spinner animations and smooth transitions
- **ASCII Art Support**: Customizable title graphics
- **Production Ready**: Follows Bubble Tea best practices

## Architecture

The application follows a proven 3-layer architecture:

- **Main Application** (`main.go`): SSH server setup and initialization
- **App Router** (`app.go`): Central routing with enum-based page state management
- **Handlers** (`handlers/`): Page-specific update logic separated by functionality
- **Views** (`views/`): Page-specific rendering logic with consistent layout patterns

## Quick Start

1. **Clone and customize:**
   ```bash
   git clone <this-repo> my-tui-app
   cd my-tui-app
   ```

2. **Update module name:**
   ```bash
   # Edit go.mod and change 'tuitemplate' to your module name
   go mod edit -module my-tui-app
   ```

3. **Customize content:**
   - Edit `ascii-art-txt` for your custom title
   - Update page content in `views/page*_view.go`
   - Modify navigation items in `app.go`

4. **Build and run:**
   ```bash
   go build -o my-tui-app .
   ./my-tui-app
   ```

5. **Connect via SSH:**
   ```bash
   ssh localhost -p 2234
   ```

## Adding New Pages

Follow the 3-step pattern:

1. **Add page enum** in `app.go`:
   ```go
   const (
       homePage pageState = iota
       page1
       page2
       page3
       newPage  // Add here
   )
   ```

2. **Create handler** in `handlers/newpagehandler.go`:
   ```go
   func UpdateNewPage(currentPage, homePage int, msg tea.Msg) (int, bool, tea.Cmd) {
       // Handle keyboard input and navigation
   }
   ```

3. **Create view** in `views/newpage_view.go`:
   ```go
   func RenderNewPageContent() string {
       // Return the page content
   }
   ```

4. **Add routing** in `app.go` switch statement

## Customization

### Styling
- Colors and styles are defined using Lipgloss
- Main theme color: `lipgloss.Color("117")` (cyan blue)
- All pages use the same layout wrapper via `renderPage()`

### Navigation
- `↑/↓` or `j/k`: Navigate menu
- `enter`: Select page
- `esc`: Back to home
- `q`: Quit application

### SSH Configuration
- Default port: 2234
- Keys stored in `.wishlist/server`
- Public key authentication enabled (allows all keys)

## File Structure

```
├── main.go              # SSH server and entry point
├── app.go               # Main application logic and routing
├── ascii-art-txt        # ASCII art title
├── handlers/            # Page update handlers
│   ├── page1handler.go
│   ├── page2handler.go
│   └── page3handler.go
├── views/               # Page view renderers
│   ├── page1_view.go
│   ├── page2_view.go
│   └── page3_view.go
└── examples/            # Example applications (optional)
```

## Dependencies

- **Bubble Tea**: TUI framework
- **Bubbles**: UI components
- **Lipgloss**: Styling and layout
- **Wish**: SSH server framework
- **Keygen**: SSH key generation

## Development

```bash
# Run in development
go run main.go

# Build optimized binary
go build -ldflags="-s -w" -o tuitemplate .

# Test connection
ssh localhost -p 2234
```

## License

This template is provided as-is for building your own TUI applications.