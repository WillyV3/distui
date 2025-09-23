# Bubble Tea Tmux Session Monitoring TUI - 2025 Best Practices Guide

## Overview

This guide provides comprehensive best practices for building a tmux session monitoring TUI using Bubble Tea in 2025. Based on current Bubble Tea patterns and your slaygent manager requirements, this focuses on table-based data display with minimal complexity.

## Architecture Principles for 2025

### 1. The Single-File Approach (Phase 1)
Start with everything in `main.go` until you prove it works:

```go
package main

import (
    "fmt"
    "os"
    "os/exec"
    "strings"
    "time"

    "github.com/charmbracelet/bubbles/table"
    "github.com/charmbracelet/bubbles/timer"
    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/lipgloss"
)

// Single model - keep it simple
type model struct {
    table       table.Model
    refreshTimer timer.Model
    err         error
    width       int
    height      int
}
```

### 2. Current Bubble Tea Table Component (2025)

The table component has evolved significantly. Here's the modern pattern:

```go
func initTable() table.Model {
    columns := []table.Column{
        {Title: "Pane ID", Width: 8},
        {Title: "Directory", Width: 30},
        {Title: "Type", Width: 12},
        {Title: "Name", Width: 20},
        {Title: "Registered", Width: 10},
    }

    t := table.New(
        table.WithColumns(columns),
        table.WithFocused(true),
        table.WithHeight(10),
    )

    // 2025 styling patterns
    s := table.DefaultStyles()
    s.Header = s.Header.
        BorderStyle(lipgloss.NormalBorder()).
        BorderForeground(lipgloss.Color("240")).
        BorderBottom(true).
        Bold(false)
    s.Selected = s.Selected.
        Foreground(lipgloss.Color("229")).
        Background(lipgloss.Color("57")).
        Bold(false)
    t.SetStyles(s)

    return t
}
```

## Model-Update-View Pattern for Table Applications

### The Model (Keep It Minimal)

```go
type model struct {
    table       table.Model
    refreshTimer timer.Model
    sessions    []TmuxSession
    err         error
    width       int
    height      int
    quitting    bool
}

type TmuxSession struct {
    PaneID      string
    Directory   string
    CommandType string
    AgentName   string
    Registered  bool
}
```

### The Update Function (2025 Patterns)

```go
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    var cmds []tea.Cmd

    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "q", "ctrl+c":
            m.quitting = true
            return m, tea.Quit
        case "r":
            // Manual refresh
            return m, m.refreshData()
        }

    case tea.WindowSizeMsg:
        m.width = msg.Width
        m.height = msg.Height
        // Adjust table height dynamically
        m.table.SetHeight(m.height - 6) // Leave room for header/footer

    case timer.TickMsg:
        var cmd tea.Cmd
        m.refreshTimer, cmd = m.refreshTimer.Update(msg)
        cmds = append(cmds, cmd)

    case timer.TimeoutMsg:
        // Auto-refresh triggered
        return m, tea.Batch(m.refreshData(), m.refreshTimer.Init())

    case sessionsMsg:
        // Data refresh completed
        m.sessions = msg.sessions
        m.err = msg.err
        m.table.SetRows(m.buildTableRows())

    case tea.QuitMsg:
        return m, tea.Quit
    }

    // Update table component
    var cmd tea.Cmd
    m.table, cmd = m.table.Update(msg)
    cmds = append(cmds, cmd)

    return m, tea.Batch(cmds...)
}
```

### The View Function (Clean and Responsive)

```go
func (m model) View() string {
    if m.quitting {
        return ""
    }

    var content strings.Builder

    // Header
    header := lipgloss.NewStyle().
        Bold(true).
        Foreground(lipgloss.Color("99")).
        Render("Slaygent Manager - Tmux Sessions")
    content.WriteString(header + "\n\n")

    // Error handling
    if m.err != nil {
        errorStyle := lipgloss.NewStyle().
            Foreground(lipgloss.Color("196")).
            Bold(true)
        content.WriteString(errorStyle.Render(fmt.Sprintf("Error: %v", m.err)) + "\n\n")
    }

    // Table
    content.WriteString(m.table.View() + "\n")

    // Footer with controls
    footer := lipgloss.NewStyle().
        Faint(true).
        Render("↑/↓: navigate • r: refresh • q: quit")
    content.WriteString("\n" + footer)

    // Handle window sizing
    if m.width > 0 {
        return lipgloss.NewStyle().
            Width(m.width).
            Height(m.height).
            Render(content.String())
    }

    return content.String()
}
```

## Keyboard Navigation Best Practices

### Standard Navigation Pattern

```go
case tea.KeyMsg:
    switch msg.String() {
    case "q", "ctrl+c", "esc":
        return m, tea.Quit
    case "up", "k":
        // Table handles this automatically
    case "down", "j":
        // Table handles this automatically
    case "r", "f5":
        return m, m.refreshData()
    case "enter":
        // Handle selection if needed
        selected := m.table.SelectedRow()
        if len(selected) > 0 {
            return m, m.handleSelection(selected)
        }
    case "?", "h":
        // Show help modal
        return m, m.showHelp()
    }
```

### Mouse Support (2025 Enhancement)

```go
// In main function
p := tea.NewProgram(
    initialModel(),
    tea.WithAltScreen(),
    tea.WithMouseCellMotion(), // Enable mouse support
)
```

## Performance Patterns for Data Refreshing

### Async Command Pattern

```go
type sessionsMsg struct {
    sessions []TmuxSession
    err      error
}

func (m model) refreshData() tea.Cmd {
    return func() tea.Msg {
        sessions, err := fetchTmuxSessions()
        return sessionsMsg{
            sessions: sessions,
            err:      err,
        }
    }
}

func fetchTmuxSessions() ([]TmuxSession, error) {
    // Non-blocking command execution
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    cmd := exec.CommandContext(ctx, "tmux", "list-panes", "-a", "-F",
        "#{pane_id}\t#{pane_current_path}\t#{pane_current_command}\t#{session_name}")

    output, err := cmd.Output()
    if err != nil {
        return nil, fmt.Errorf("tmux command failed: %w", err)
    }

    return parseTmuxOutput(string(output))
}
```

### Timer-Based Auto-Refresh

```go
func (m model) Init() tea.Cmd {
    return tea.Batch(
        m.refreshData(),
        m.refreshTimer.Init(),
    )
}

// Configure refresh timer
func initialModel() model {
    return model{
        table:        initTable(),
        refreshTimer: timer.NewWithInterval(5*time.Second, time.Second),
        sessions:     []TmuxSession{},
    }
}
```

## Tmux Integration Patterns

### Modern Tmux Command Formatting (2025)

```go
func fetchTmuxSessions() ([]TmuxSession, error) {
    // Use comprehensive format string for all data at once
    formatStr := "#{pane_id}\t#{pane_current_path}\t#{pane_current_command}\t#{session_name}\t#{window_name}\t#{pane_title}"

    cmd := exec.Command("tmux", "list-panes", "-a", "-F", formatStr)
    output, err := cmd.Output()
    if err != nil {
        // Handle tmux not running
        if strings.Contains(err.Error(), "no server running") {
            return []TmuxSession{}, nil
        }
        return nil, err
    }

    return parseTmuxOutput(string(output))
}

func parseTmuxOutput(output string) ([]TmuxSession, error) {
    var sessions []TmuxSession

    lines := strings.Split(strings.TrimSpace(output), "\n")
    for _, line := range lines {
        if line == "" {
            continue
        }

        fields := strings.Split(line, "\t")
        if len(fields) < 4 {
            continue
        }

        session := TmuxSession{
            PaneID:      fields[0],
            Directory:   fields[1],
            CommandType: detectCommandType(fields[2]),
            AgentName:   extractAgentName(fields[3]),
            Registered:  checkRegistration(fields[3]),
        }

        sessions = append(sessions, session)
    }

    return sessions, nil
}
```

### Command Type Detection

```go
func detectCommandType(command string) string {
    command = strings.ToLower(command)

    switch {
    case strings.Contains(command, "claude"):
        return "claude"
    case strings.Contains(command, "opencode"):
        return "opencode"
    case strings.Contains(command, "cursor"):
        return "cursor"
    case strings.Contains(command, "python"):
        return "python"
    case strings.Contains(command, "node"):
        return "node"
    case command == "zsh" || command == "bash" || command == "fish":
        return "shell"
    default:
        return "unknown"
    }
}
```

## State Management for Minimal Applications

### Keep State Flat and Simple

```go
// ❌ DON'T: Complex nested state
type model struct {
    ui struct {
        table struct {
            data    []Row
            styles  TableStyles
            config  TableConfig
        }
        layout struct {
            width  int
            height int
        }
    }
    data struct {
        sessions map[string]Session
        agents   map[string]Agent
    }
}

// ✅ DO: Flat, simple state
type model struct {
    table       table.Model
    sessions    []TmuxSession
    width       int
    height      int
    err         error
    loading     bool
}
```

### State Updates Pattern

```go
func (m model) buildTableRows() []table.Row {
    rows := make([]table.Row, len(m.sessions))

    for i, session := range m.sessions {
        registered := "✗"
        if session.Registered {
            registered = "✓"
        }

        rows[i] = table.Row{
            session.PaneID,
            truncateString(session.Directory, 30),
            session.CommandType,
            session.AgentName,
            registered,
        }
    }

    return rows
}

func truncateString(s string, maxLen int) string {
    if len(s) <= maxLen {
        return s
    }
    return s[:maxLen-3] + "..."
}
```

## Error Handling Patterns

### Graceful Error Display

```go
func (m model) View() string {
    var content strings.Builder

    // Always show header
    content.WriteString(m.renderHeader())

    // Show errors prominently but don't crash
    if m.err != nil {
        errorMsg := fmt.Sprintf("⚠ Error: %v", m.err)
        errorStyle := lipgloss.NewStyle().
            Foreground(lipgloss.Color("196")).
            Background(lipgloss.Color("52")).
            Padding(0, 1).
            Margin(1, 0)
        content.WriteString(errorStyle.Render(errorMsg) + "\n")

        // Show recovery instructions
        content.WriteString("Press 'r' to retry or 'q' to quit\n")
        return content.String()
    }

    // Normal view
    content.WriteString(m.table.View())
    content.WriteString(m.renderFooter())

    return content.String()
}
```

### System Command Error Handling

```go
func (m model) refreshData() tea.Cmd {
    return func() tea.Msg {
        sessions, err := fetchTmuxSessions()

        // Handle specific error cases
        if err != nil {
            if strings.Contains(err.Error(), "no server running") {
                // Tmux not running - not an error, just empty state
                return sessionsMsg{
                    sessions: []TmuxSession{},
                    err:      nil,
                }
            }
            // Real error
            return sessionsMsg{
                sessions: nil,
                err:      fmt.Errorf("failed to fetch tmux data: %w", err),
            }
        }

        return sessionsMsg{
            sessions: sessions,
            err:      nil,
        }
    }
}
```

## Registration Status Integration

### Python Script Integration Pattern

```go
func checkRegistration(sessionName string) bool {
    // Read from slaygent registration file/database
    cmd := exec.Command("python3", "-c", `
import json
import os
try:
    with open(os.path.expanduser("~/.slaygent/registry.json")) as f:
        registry = json.load(f)
    print("true" if "`+sessionName+`" in registry else "false")
except:
    print("false")
`)

    output, err := cmd.Output()
    if err != nil {
        return false
    }

    return strings.TrimSpace(string(output)) == "true"
}

func extractAgentName(sessionName string) string {
    // Extract meaningful name from session name
    // Remove common prefixes
    name := strings.TrimPrefix(sessionName, "slaygent-")
    name = strings.TrimPrefix(name, "dev-")

    // If it looks like project-feature format, use that
    parts := strings.Split(name, "-")
    if len(parts) >= 2 {
        return strings.Join(parts[:2], "-")
    }

    return name
}
```

## Complete Minimal Example

Here's a complete working example following 2025 best practices:

```go
package main

import (
    "context"
    "fmt"
    "os"
    "os/exec"
    "strings"
    "time"

    "github.com/charmbracelet/bubbles/table"
    "github.com/charmbracelet/bubbles/timer"
    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/lipgloss"
)

type model struct {
    table       table.Model
    refreshTimer timer.Model
    sessions    []TmuxSession
    err         error
    width       int
    height      int
}

type TmuxSession struct {
    PaneID      string
    Directory   string
    CommandType string
    AgentName   string
    Registered  bool
}

type sessionsMsg struct {
    sessions []TmuxSession
    err      error
}

func main() {
    p := tea.NewProgram(
        initialModel(),
        tea.WithAltScreen(),
        tea.WithMouseCellMotion(),
    )

    if _, err := p.Run(); err != nil {
        fmt.Printf("Error: %v", err)
        os.Exit(1)
    }
}

func initialModel() model {
    columns := []table.Column{
        {Title: "Pane ID", Width: 8},
        {Title: "Directory", Width: 30},
        {Title: "Type", Width: 12},
        {Title: "Name", Width: 20},
        {Title: "Registered", Width: 10},
    }

    t := table.New(
        table.WithColumns(columns),
        table.WithFocused(true),
        table.WithHeight(10),
    )

    return model{
        table:        t,
        refreshTimer: timer.NewWithInterval(5*time.Second, time.Second),
    }
}

func (m model) Init() tea.Cmd {
    return tea.Batch(
        m.refreshData(),
        m.refreshTimer.Init(),
    )
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    var cmds []tea.Cmd

    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "q", "ctrl+c":
            return m, tea.Quit
        case "r":
            return m, m.refreshData()
        }

    case tea.WindowSizeMsg:
        m.width = msg.Width
        m.height = msg.Height
        m.table.SetHeight(m.height - 6)

    case timer.TickMsg:
        var cmd tea.Cmd
        m.refreshTimer, cmd = m.refreshTimer.Update(msg)
        cmds = append(cmds, cmd)

    case timer.TimeoutMsg:
        return m, tea.Batch(m.refreshData(), m.refreshTimer.Init())

    case sessionsMsg:
        m.sessions = msg.sessions
        m.err = msg.err
        if m.err == nil {
            m.table.SetRows(m.buildTableRows())
        }
    }

    var cmd tea.Cmd
    m.table, cmd = m.table.Update(msg)
    cmds = append(cmds, cmd)

    return m, tea.Batch(cmds...)
}

func (m model) View() string {
    if m.err != nil {
        return fmt.Sprintf("Error: %v\nPress 'r' to retry, 'q' to quit", m.err)
    }

    var content strings.Builder
    content.WriteString("Slaygent Manager - Tmux Sessions\n\n")
    content.WriteString(m.table.View())
    content.WriteString("\n\n↑/↓: navigate • r: refresh • q: quit")

    return content.String()
}

func (m model) refreshData() tea.Cmd {
    return func() tea.Msg {
        sessions, err := fetchTmuxSessions()
        return sessionsMsg{sessions: sessions, err: err}
    }
}

func (m model) buildTableRows() []table.Row {
    rows := make([]table.Row, len(m.sessions))
    for i, session := range m.sessions {
        registered := "✗"
        if session.Registered {
            registered = "✓"
        }
        rows[i] = table.Row{
            session.PaneID,
            truncateString(session.Directory, 28),
            session.CommandType,
            session.AgentName,
            registered,
        }
    }
    return rows
}

func fetchTmuxSessions() ([]TmuxSession, error) {
    ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
    defer cancel()

    cmd := exec.CommandContext(ctx, "tmux", "list-panes", "-a", "-F",
        "#{pane_id}\t#{pane_current_path}\t#{pane_current_command}\t#{session_name}")

    output, err := cmd.Output()
    if err != nil {
        if strings.Contains(err.Error(), "no server running") {
            return []TmuxSession{}, nil
        }
        return nil, err
    }

    return parseTmuxOutput(string(output))
}

func parseTmuxOutput(output string) ([]TmuxSession, error) {
    var sessions []TmuxSession
    lines := strings.Split(strings.TrimSpace(output), "\n")

    for _, line := range lines {
        if line == "" {
            continue
        }

        fields := strings.Split(line, "\t")
        if len(fields) < 4 {
            continue
        }

        sessions = append(sessions, TmuxSession{
            PaneID:      fields[0],
            Directory:   fields[1],
            CommandType: detectCommandType(fields[2]),
            AgentName:   extractAgentName(fields[3]),
            Registered:  checkRegistration(fields[3]),
        })
    }

    return sessions, nil
}

func detectCommandType(command string) string {
    command = strings.ToLower(command)
    switch {
    case strings.Contains(command, "claude"):
        return "claude"
    case strings.Contains(command, "cursor"):
        return "cursor"
    case command == "zsh" || command == "bash":
        return "shell"
    default:
        return "other"
    }
}

func extractAgentName(sessionName string) string {
    return strings.TrimPrefix(sessionName, "slaygent-")
}

func checkRegistration(sessionName string) bool {
    // Implement your registration check logic
    return false
}

func truncateString(s string, maxLen int) string {
    if len(s) <= maxLen {
        return s
    }
    return s[:maxLen-3] + "..."
}
```

## Dependencies (go.mod)

```go
module tmux-monitor

go 1.21

require (
    github.com/charmbracelet/bubbles v0.17.1
    github.com/charmbracelet/bubbletea v0.25.0
    github.com/charmbracelet/lipgloss v0.9.1
)
```

## Key Takeaways for 2025

1. **Start Simple**: Single file, minimal state, prove it works
2. **Use Modern Table Component**: Leverage built-in styling and behavior
3. **Async Commands**: Never block the UI thread with system commands
4. **Responsive Design**: Handle window resizing properly
5. **Graceful Errors**: Show errors without crashing
6. **Auto-refresh**: Use timers for periodic updates
7. **Mouse Support**: Enable for better UX
8. **Context Timeouts**: Prevent hanging on slow commands

This architecture will give you a solid, maintainable tmux monitoring TUI that follows current Bubble Tea best practices while staying simple and focused.