package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
)

var Version = "dev"

var (
	appStyle = lipgloss.NewStyle().Padding(1, 2)

	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFD700")).
			Bold(true).
			Padding(0, 1)

	infoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00BFFF"))

	errStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF4500")).
			Bold(true)

	statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#C0C0C0")).
			Italic(true)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262")).
			MarginTop(1)
)

type noteItem struct {
	title string
	path  string
}

func (i noteItem) Title() string       { return i.title }
func (i noteItem) Description() string { return i.path }
func (i noteItem) FilterValue() string { return i.title }

type state int

const (
	stateList state = iota
	stateReading
	stateCreating
	stateEditing
	stateConfirmDelete
)

type model struct {
	list          list.Model
	viewport      viewport.Model
	textInput     textinput.Model
	textarea      textarea.Model
	notesDir      string
	currentState  state
	selectedNote  noteItem
	ready         bool
	width, height int
}

func initialModel(notesDir string) model {
	items := getNotes(notesDir)
	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Atlas Notes"
	l.Styles.Title = titleStyle
	l.AdditionalFullHelpKeys = func() []key.Binding { return []key.Binding{} }

	ti := textinput.New()
	ti.Placeholder = "Note Name (e.g. My Idea)"

	ta := textarea.New()
	ta.Placeholder = "Start writing your note here..."
	ta.Focus()

	return model{
		list:         l,
		notesDir:     notesDir,
		currentState: stateList,
		textInput:    ti,
		textarea:     ta,
	}
}

func getNotes(dir string) []list.Item {
	files, _ := os.ReadDir(dir)
	var items []list.Item
	for _, f := range files {
		if !f.IsDir() && strings.HasSuffix(f.Name(), ".md") {
			name := strings.TrimSuffix(f.Name(), ".md")
			items = append(items, noteItem{title: name, path: filepath.Join(dir, f.Name())})
		}
	}
	if items == nil {
		return []list.Item{}
	}
	return items
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.list.SetSize(msg.Width-4, msg.Height-6)
		m.textarea.SetWidth(msg.Width - 6)
		m.textarea.SetHeight(msg.Height - 10)
		if !m.ready {
			m.viewport = viewport.New(msg.Width-4, msg.Height-8)
			m.ready = true
		} else {
			m.viewport.Width = msg.Width - 4
			m.viewport.Height = msg.Height - 8
		}

	case tea.KeyMsg:
		switch m.currentState {
		case stateList:
			switch msg.String() {
			case "q", "ctrl+c":
				return m, tea.Quit
			case "enter":
				if i, ok := m.list.SelectedItem().(noteItem); ok {
					m.selectedNote = i
					m.currentState = stateReading
					m.renderNote()
					return m, nil
				}
			case "n":
				m.currentState = stateCreating
				m.textInput.SetValue("")
				m.textInput.Focus()
				return m, textinput.Blink
			case "e":
				if i, ok := m.list.SelectedItem().(noteItem); ok {
					m.selectedNote = i
					m.currentState = stateEditing
					content, _ := os.ReadFile(i.path)
					m.textarea.SetValue(string(content))
					m.textarea.Focus()
					return m, textarea.Blink
				}
			case "d":
				if i, ok := m.list.SelectedItem().(noteItem); ok {
					m.selectedNote = i
					m.currentState = stateConfirmDelete
					return m, nil
				}
			}

		case stateReading:
			switch msg.String() {
			case "esc", "q":
				m.currentState = stateList
				return m, nil
			case "e":
				m.currentState = stateEditing
				content, _ := os.ReadFile(m.selectedNote.path)
				m.textarea.SetValue(string(content))
				m.textarea.Focus()
				return m, textarea.Blink
			}

		case stateCreating:
			switch msg.String() {
			case "enter":
				name := m.textInput.Value()
				if strings.TrimSpace(name) != "" {
					path := createNoteFile(m.notesDir, name)
					m.selectedNote = noteItem{title: name, path: path}
					m.list.SetItems(getNotes(m.notesDir))
					// Immediately start editing the new note
					m.textarea.SetValue("# " + name + "\n\n")
					m.currentState = stateEditing
					m.textarea.Focus()
					return m, textarea.Blink
				}
			case "esc":
				m.currentState = stateList
				return m, nil
			}

		case stateEditing:
			switch msg.String() {
			case "ctrl+s":
				os.WriteFile(m.selectedNote.path, []byte(m.textarea.Value()), 0644)
				m.currentState = stateReading
				m.renderNote()
				return m, nil
			case "esc":
				m.currentState = stateList
				return m, nil
			}

		case stateConfirmDelete:
			switch msg.String() {
			case "enter":
				os.Remove(m.selectedNote.path)
				m.list.SetItems(getNotes(m.notesDir))
				m.currentState = stateList
				return m, nil
			case "esc", "n":
				m.currentState = stateList
				return m, nil
			}
		}
	}

	switch m.currentState {
	case stateList:
		m.list, cmd = m.list.Update(msg)
	case stateReading:
		m.viewport, cmd = m.viewport.Update(msg)
	case stateCreating:
		m.textInput, cmd = m.textInput.Update(msg)
	case stateEditing:
		m.textarea, cmd = m.textarea.Update(msg)
	}

	return m, cmd
}

func (m *model) renderNote() {
	content, err := os.ReadFile(m.selectedNote.path)
	if err != nil {
		m.viewport.SetContent("Error reading note: " + err.Error())
		return
	}
	r, _ := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(m.width-10),
	)
	out, err := r.Render(string(content))
	if err != nil {
		m.viewport.SetContent("Error rendering markdown: " + err.Error())
		return
	}
	m.viewport.SetContent(out)
}

func (m model) View() string {
	var s string

	switch m.currentState {
	case stateList:
		s = m.list.View()
		s += helpStyle.Render("\n[n] New • [e] Edit • [d] Delete • [enter] Read • [q] Quit")
	case stateReading:
		s = fmt.Sprintf("%s\n\n%s\n\n%s",
			titleStyle.Render("Reading: "+m.selectedNote.title),
			m.viewport.View(),
			statusStyle.Render("[esc/q] Back to List • [e] Edit Note Content"),
		)
	case stateCreating:
		s = fmt.Sprintf("%s\n\n%s\n\n%s",
			titleStyle.Render("Create New Note"),
			m.textInput.View(),
			statusStyle.Render("[enter] Create & Write • [esc] Cancel"),
		)
	case stateEditing:
		s = fmt.Sprintf("%s\n\n%s\n\n%s",
			titleStyle.Render("Editing: "+m.selectedNote.title),
			m.textarea.View(),
			statusStyle.Render("[ctrl+s] Save & View • [esc] Cancel/Back to List"),
		)
	case stateConfirmDelete:
		s = fmt.Sprintf("%s\n\n%s\n\n%s",
			errStyle.Render("Confirm Deletion"),
			fmt.Sprintf("Are you sure you want to delete '%s'?", m.selectedNote.title),
			statusStyle.Render("[enter] Confirm Delete • [esc/n] Cancel"),
		)
	}

	return appStyle.Render(s)
}

func createNoteFile(dir, name string) string {
	if !strings.HasSuffix(name, ".md") {
		name += ".md"
	}
	path := filepath.Join(dir, name)
	os.WriteFile(path, []byte("# "+strings.TrimSuffix(name, ".md")+"\n\n"), 0644)
	return path
}

func main() {
	if len(os.Args) > 1 {
		arg := os.Args[1]
		if arg == "-v" || arg == "--version" {
			fmt.Printf("atlas.notes v%s\n", Version)
			return
		}
		if arg == "-h" || arg == "--help" {
			printHelp()
			return
		}
	}

	notesDir, err := getNotesDir()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	p := tea.NewProgram(initialModel(notesDir), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v", err)
		os.Exit(1)
	}
}

func getNotesDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, ".atlas", "atlas.notes.data")
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			return "", err
		}
	}
	return dir, nil
}

func printHelp() {
	fmt.Println(titleStyle.Render("Atlas Notes Help"))
	fmt.Println("Interactive TUI for managing your markdown notes.")
	fmt.Println("\nGlobal Navigation:")
	fmt.Println("  j/k, up/down   Navigate the notes list")
	fmt.Println("  enter          Read the selected note")
	fmt.Println("  n              Create a new note")
	fmt.Println("  e              Edit note content (Built-in Editor)")
	fmt.Println("  d              Delete the selected note")
	fmt.Println("  q, ctrl+c      Quit the application")
	fmt.Println("\nInside Editor:")
	fmt.Println("  ctrl+s         Save and return to reader")
	fmt.Println("  esc            Discard/Cancel and return to list")
	fmt.Println("\nFlags:")
	fmt.Println("  -v, --version  Show version information")
	fmt.Println("  -h, --help     Show this help documentation")
}
