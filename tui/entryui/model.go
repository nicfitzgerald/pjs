package entryui

import (
	"log"

	"github.com/bashbunni/project-management/entry"
	"github.com/bashbunni/project-management/tui/constants"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/paginator"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
)

var cmd tea.Cmd

// BackMsg change state back to project view
type BackMsg bool
type PrintedMsg struct{}

// Model implements tea.Model
type Model struct {
	viewport        viewport.Model
	er              *entry.GormRepository
	activeProjectID uint
	p               *tea.Program
	alerts          string
	error           string
	windowSize      tea.WindowSizeMsg
	paginator       paginator.Model
	entries         []entry.Entry
}

// Init run any intial IO on program start
func (m Model) Init() tea.Cmd {
	return nil
}

// New initialize the entryui model for your program
func New(er *entry.GormRepository, activeProjectID uint, p *tea.Program, windowSize tea.WindowSizeMsg) *Model {
	m := Model{er: er, activeProjectID: activeProjectID, windowSize: windowSize}
	m.p = p
	m.viewport = viewport.New(windowSize.Width, calculateHeight(windowSize.Height))
	m.viewport.Style = lipgloss.NewStyle().
		Align(lipgloss.Bottom)

	// init paginator
	m.paginator = paginator.New()
	m.paginator.Type = paginator.Dots
	m.paginator.ActiveDot = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "235", Dark: "252"}).Render("•")
	m.paginator.InactiveDot = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "250", Dark: "238"}).Render("•")

	// get entries
	var err error
	if m.entries, err = er.GetEntriesByProjectID(m.activeProjectID); err != nil {
		log.Fatalf("failed to get entries: %v", err)
	}
	m.paginator.SetTotalPages(len(m.entries))

	// set content
	m.setViewportContent()
	return &m
}

func (m *Model) setViewportContent() {
	var content string
	if len(m.entries) == 0 {
		content = "There are no entries for this project :)"
	} else {
		content = entry.FormatEntry(m.entries[m.paginator.Page])
	}
	str, err := glamour.Render(content, "dark")
	if err != nil {
		m.error = "could not render content with glamour"
	}
	m.viewport.SetContent(str)
}

// Update handle IO and commands
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.viewport.Width = msg.Width
		m.viewport.Height = calculateHeight(msg.Height)
	case errMsg:
		m.error = msg.Error()
	case editorFinishedMsg:
		if msg.err != nil {
			return m, tea.Quit
		}
		cmd = m.createEntryCmd(msg.file)
	case updateEntryListMsg:
		return m, m.updateEntriesCmd
	case PrintedMsg:
		m.alerts = "printed!"
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, constants.Keymap.Create):
			return m, openEditorCmd()
		case key.Matches(msg, constants.Keymap.Back):
			return m, func() tea.Msg {
				return BackMsg(true)
			}
		case msg.String() == "p":
			return m, func() tea.Msg {
				entries, err := m.er.GetEntriesByProjectID(m.activeProjectID)
				if err != nil {
					return errMsg{err}
				}
				err = entry.OutputEntriesToPDF(entries)
				if err != nil {
					return errMsg{err}
				}
				return PrintedMsg{}
			}
		case msg.String() == "ctrl+c":
			return m, tea.Quit
		case msg.String() == "q":
			return m, tea.Quit
		default:
			// m.viewport, cmd = m.viewport.Update(msg)
			// cmds = append(cmds, cmd)
		}
	}
	m.paginator, cmd = m.paginator.Update(msg)
	return m, cmd
}

func (m Model) helpView() string {
	return constants.HelpStyle("\n ↑/↓: navigate  • esc: back • c: create entry • d: delete entry • p: print • q: quit\n")
}

func (m Model) errorView() string {
	return constants.ErrStyle(m.error)
}

func (m Model) alertView() string {
	return constants.AlertStyle(m.alerts)
}

// View return the text UI to be output to the terminal
func (m Model) View() string {
	m.setViewportContent()
	formatted := lipgloss.JoinVertical(lipgloss.Left, "\n", m.viewport.View(), m.helpView(), m.errorView(), m.paginator.View(), m.alertView())
	return constants.DocStyle.Render(formatted)
}

/* helpers */

func calculateHeight(height int) int {
	return height - height/7
}

// TODO: don't pipe from Stdin, break up functionality more
/*
Questions:
I might be able to use the ExecCommand interface
then SetStdOut to an io.Writer that I've created?

Scratchpad
- p.Send? - but I don't need to send it back to the program.
I just need to suspend bubble tea so I can pipe data to pandoc.
*/
