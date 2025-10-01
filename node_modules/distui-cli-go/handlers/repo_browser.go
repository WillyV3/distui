package handlers

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

type RepoBrowserModel struct {
	CurrentDirectory string
	Entries          []FileEntry
	Selected         int
	ShowHidden       bool
	Width            int
	Height           int
	KeyMap           RepoBrowserKeyMap
	Error            error
}

type FileEntry struct {
	Name        string
	IsDir       bool
	Size        int64
	Mode        os.FileMode
	ModTime     string
}

type RepoBrowserKeyMap struct {
	GoToTop  key.Binding
	GoToLast key.Binding
	Down     key.Binding
	Up       key.Binding
	PageUp   key.Binding
	PageDown key.Binding
	Back     key.Binding
	Open     key.Binding
	Quit     key.Binding
}

func DefaultRepoBrowserKeyMap() RepoBrowserKeyMap {
	return RepoBrowserKeyMap{
		GoToTop:  key.NewBinding(key.WithKeys("g"), key.WithHelp("g", "first")),
		GoToLast: key.NewBinding(key.WithKeys("G"), key.WithHelp("G", "last")),
		Down:     key.NewBinding(key.WithKeys("j", "down"), key.WithHelp("j", "down")),
		Up:       key.NewBinding(key.WithKeys("k", "up"), key.WithHelp("k", "up")),
		PageUp:   key.NewBinding(key.WithKeys("K", "pgup"), key.WithHelp("pgup", "page up")),
		PageDown: key.NewBinding(key.WithKeys("J", "pgdown"), key.WithHelp("pgdown", "page down")),
		Back:     key.NewBinding(key.WithKeys("h", "backspace", "left"), key.WithHelp("h", "back")),
		Open:     key.NewBinding(key.WithKeys("l", "right", "enter"), key.WithHelp("l", "open")),
		Quit:     key.NewBinding(key.WithKeys("q", "esc"), key.WithHelp("q", "quit")),
	}
}

func NewRepoBrowserModel(width, height int) *RepoBrowserModel {
	m := &RepoBrowserModel{
		CurrentDirectory: ".",
		Selected:         0,
		ShowHidden:       false,
		Width:            width,
		Height:           height,
		KeyMap:           DefaultRepoBrowserKeyMap(),
	}
	m.LoadDirectory()
	return m
}

func (m *RepoBrowserModel) LoadDirectory() {
	entries, err := os.ReadDir(m.CurrentDirectory)
	if err != nil {
		m.Error = err
		return
	}

	m.Entries = []FileEntry{}
	for _, entry := range entries {
		// Skip hidden files unless ShowHidden is true
		if !m.ShowHidden && strings.HasPrefix(entry.Name(), ".") {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		m.Entries = append(m.Entries, FileEntry{
			Name:    entry.Name(),
			IsDir:   entry.IsDir(),
			Size:    info.Size(),
			Mode:    info.Mode(),
			ModTime: info.ModTime().Format("Jan _2 15:04"),
		})
	}

	// Sort: directories first, then alphabetically
	sort.Slice(m.Entries, func(i, j int) bool {
		if m.Entries[i].IsDir != m.Entries[j].IsDir {
			return m.Entries[i].IsDir
		}
		return strings.ToLower(m.Entries[i].Name) < strings.ToLower(m.Entries[j].Name)
	})

	// Reset selection if out of bounds
	if m.Selected >= len(m.Entries) {
		m.Selected = len(m.Entries) - 1
	}
	if m.Selected < 0 {
		m.Selected = 0
	}
}

func (m *RepoBrowserModel) Update(msg tea.Msg) (*RepoBrowserModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.KeyMap.Down):
			if m.Selected < len(m.Entries)-1 {
				m.Selected++
			}
		case key.Matches(msg, m.KeyMap.Up):
			if m.Selected > 0 {
				m.Selected--
			}
		case key.Matches(msg, m.KeyMap.GoToTop):
			m.Selected = 0
		case key.Matches(msg, m.KeyMap.GoToLast):
			m.Selected = len(m.Entries) - 1
		case key.Matches(msg, m.KeyMap.PageDown):
			pageSize := m.Height - 8
			if pageSize < 1 {
				pageSize = 1
			}
			m.Selected += pageSize
			if m.Selected >= len(m.Entries) {
				m.Selected = len(m.Entries) - 1
			}
		case key.Matches(msg, m.KeyMap.PageUp):
			pageSize := m.Height - 8
			if pageSize < 1 {
				pageSize = 1
			}
			m.Selected -= pageSize
			if m.Selected < 0 {
				m.Selected = 0
			}
		case key.Matches(msg, m.KeyMap.Back):
			// Go up one directory
			if m.CurrentDirectory != "." && m.CurrentDirectory != "/" {
				m.CurrentDirectory = filepath.Dir(m.CurrentDirectory)
				m.Selected = 0
				m.LoadDirectory()
			}
		case key.Matches(msg, m.KeyMap.Open):
			// Open directory or do nothing for files
			if m.Selected < len(m.Entries) && m.Entries[m.Selected].IsDir {
				m.CurrentDirectory = filepath.Join(m.CurrentDirectory, m.Entries[m.Selected].Name)
				m.Selected = 0
				m.LoadDirectory()
			}
		}
	}

	return m, nil
}

func (e FileEntry) String() string {
	// Type indicator - single character
	typeChar := "-"
	if e.IsDir {
		typeChar = "/"
	} else if strings.HasSuffix(e.Name, ".go") {
		typeChar = "g"
	} else if strings.HasSuffix(e.Name, ".md") {
		typeChar = "m"
	} else if strings.HasSuffix(e.Name, ".json") {
		typeChar = "j"
	} else if strings.HasSuffix(e.Name, ".yaml") || strings.HasSuffix(e.Name, ".yml") {
		typeChar = "y"
	} else if strings.HasSuffix(e.Name, ".txt") {
		typeChar = "t"
	} else if e.Mode&0111 != 0 && !strings.Contains(e.Name, ".") {
		// Executable file without extension (likely binary)
		typeChar = "b"
	}

	// Fixed width name column (40 chars)
	name := e.Name
	if e.IsDir {
		name = name + "/"
	}
	if len(name) > 40 {
		name = name[:37] + "..."
	}
	nameFormatted := fmt.Sprintf("%-40s", name)

	// Date column
	modTime := e.ModTime

	return fmt.Sprintf("%s  %s  %s", typeChar, nameFormatted, modTime)
}

func (m *RepoBrowserModel) SetSize(width, height int) {
	m.Width = width
	m.Height = height
}