package tui

import (
	"fmt"
	"strings"

	"github.com/FScoward/rakutree/internal/git"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type viewState int

const (
	menuView viewState = iota
	listView
	addView
	pathSelectView
	customPathView
	removeView
)

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 1)

	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7D56F4")).
			Bold(true)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000")).
			Bold(true)

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#04B575")).
			Bold(true)
)

type item struct {
	title string
	desc  string
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.desc }
func (i item) FilterValue() string { return i.title }

type Model struct {
	state            viewState
	list             list.Model
	pathInput        textinput.Model
	worktrees        []git.Worktree
	branches         []string
	selectedBranch   string
	pathSuggestions  []git.PathSuggestion
	err              error
	message          string
	quitting         bool
	width            int
	height           int
}

func NewModel() Model {
	ti := textinput.New()
	ti.Placeholder = "Enter worktree path (e.g., ../feature-branch)"
	ti.Focus()
	ti.CharLimit = 256
	ti.Width = 50

	items := []list.Item{
		item{title: "List Worktrees", desc: "View all existing worktrees"},
		item{title: "Add Worktree", desc: "Create a new worktree"},
		item{title: "Remove Worktree", desc: "Delete an existing worktree"},
		item{title: "Quit", desc: "Exit the application"},
	}

	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Git Worktree Manager"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)

	return Model{
		state:     menuView,
		list:      l,
		pathInput: ti,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.list.SetSize(msg.Width, msg.Height-4)
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			if m.state == menuView {
				m.quitting = true
				return m, tea.Quit
			}
			// Go back to menu from other views
			m.state = menuView
			m.err = nil
			m.message = ""
			return m, nil

		case "esc":
			if m.state != menuView {
				m.state = menuView
				m.err = nil
				m.message = ""
				return m, nil
			}

		case "enter":
			return m.handleEnter()
		}
	}

	switch m.state {
	case menuView, listView, addView, removeView, pathSelectView:
		var cmd tea.Cmd
		m.list, cmd = m.list.Update(msg)
		return m, cmd
	case customPathView:
		var cmd tea.Cmd
		m.pathInput, cmd = m.pathInput.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m Model) handleEnter() (tea.Model, tea.Cmd) {
	switch m.state {
	case menuView:
		selected := m.list.SelectedItem()
		if selected == nil {
			return m, nil
		}

		switch selected.(item).title {
		case "List Worktrees":
			m.state = listView
			worktrees, err := git.ListWorktrees()
			if err != nil {
				m.err = err
				m.state = menuView
				return m, nil
			}
			m.worktrees = worktrees

			items := make([]list.Item, len(worktrees))
			for i, wt := range worktrees {
				branch := wt.Branch
				if branch == "" {
					branch = "detached"
				}
				items[i] = item{
					title: wt.Path,
					desc:  fmt.Sprintf("Branch: %s | Commit: %.7s", branch, wt.Commit),
				}
			}
			m.list.SetItems(items)
			m.list.Title = "Worktrees (press ESC to go back)"

		case "Add Worktree":
			branches, err := git.ListBranches()
			if err != nil {
				m.err = err
				return m, nil
			}
			m.branches = branches

			items := make([]list.Item, len(branches))
			for i, branch := range branches {
				items[i] = item{title: branch, desc: ""}
			}
			m.list.SetItems(items)
			m.list.Title = "Select a branch (press ESC to cancel)"
			m.state = addView

		case "Remove Worktree":
			worktrees, err := git.ListWorktrees()
			if err != nil {
				m.err = err
				return m, nil
			}
			// Filter out the main worktree (first one)
			if len(worktrees) > 1 {
				m.worktrees = worktrees[1:]
			} else {
				m.message = "No additional worktrees to remove"
				return m, nil
			}

			items := make([]list.Item, len(m.worktrees))
			for i, wt := range m.worktrees {
				branch := wt.Branch
				if branch == "" {
					branch = "detached"
				}
				items[i] = item{
					title: wt.Path,
					desc:  fmt.Sprintf("Branch: %s", branch),
				}
			}
			m.list.SetItems(items)
			m.list.Title = "Select worktree to remove (press ESC to cancel)"
			m.state = removeView

		case "Quit":
			m.quitting = true
			return m, tea.Quit
		}

	case addView:
		selected := m.list.SelectedItem()
		if selected == nil {
			return m, nil
		}

		branch := selected.(item).title
		m.selectedBranch = branch

		// Get path suggestions based on branch
		suggestions, err := git.SuggestPaths(branch)
		if err != nil {
			m.err = err
			m.state = menuView
			return m, nil
		}
		m.pathSuggestions = suggestions

		// Show path selection screen
		items := make([]list.Item, len(suggestions))
		for i, sug := range suggestions {
			title := sug.Path
			if sug.IsCustom {
				title = "✏️  Custom path..."
			}
			items[i] = item{
				title: title,
				desc:  sug.Description,
			}
		}
		m.list.SetItems(items)
		m.list.Title = fmt.Sprintf("Select path for '%s' (ESC to cancel)", branch)
		m.state = pathSelectView

	case pathSelectView:
		selected := m.list.SelectedItem()
		if selected == nil {
			return m, nil
		}

		// Find the selected suggestion
		selectedIndex := -1
		for i, item := range m.list.Items() {
			if item == selected {
				selectedIndex = i
				break
			}
		}

		if selectedIndex < 0 || selectedIndex >= len(m.pathSuggestions) {
			return m, nil
		}

		suggestion := m.pathSuggestions[selectedIndex]

		// If custom path selected, show input
		if suggestion.IsCustom {
			m.pathInput.SetValue("")
			m.pathInput.Placeholder = "Enter custom path (e.g., ../my-worktree)"
			m.state = customPathView
			return m, nil
		}

		// Otherwise, use the suggested path
		err := git.AddWorktree(suggestion.Path, m.selectedBranch)
		if err != nil {
			m.err = err
		} else {
			m.message = fmt.Sprintf("Successfully added worktree at %s", suggestion.Path)
		}
		m.state = menuView
		m.resetMenuItems()

	case customPathView:
		path := m.pathInput.Value()
		if path == "" {
			m.err = fmt.Errorf("path cannot be empty")
			m.state = menuView
			m.resetMenuItems()
			return m, nil
		}

		err := git.AddWorktree(path, m.selectedBranch)
		if err != nil {
			m.err = err
		} else {
			m.message = fmt.Sprintf("Successfully added worktree at %s", path)
		}
		m.pathInput.SetValue("")
		m.state = menuView
		m.resetMenuItems()

	case removeView:
		selected := m.list.SelectedItem()
		if selected == nil {
			return m, nil
		}

		path := selected.(item).title
		err := git.RemoveWorktree(path)
		if err != nil {
			m.err = err
		} else {
			m.message = fmt.Sprintf("Successfully removed worktree at %s", path)
		}
		m.state = menuView
		m.resetMenuItems()
	}

	return m, nil
}

func (m *Model) resetMenuItems() {
	items := []list.Item{
		item{title: "List Worktrees", desc: "View all existing worktrees"},
		item{title: "Add Worktree", desc: "Create a new worktree"},
		item{title: "Remove Worktree", desc: "Delete an existing worktree"},
		item{title: "Quit", desc: "Exit the application"},
	}
	m.list.SetItems(items)
	m.list.Title = "Git Worktree Manager"
}

func (m Model) View() string {
	if m.quitting {
		return ""
	}

	var s strings.Builder

	// Show error or success message
	if m.err != nil {
		s.WriteString(errorStyle.Render(fmt.Sprintf("Error: %v\n\n", m.err)))
	} else if m.message != "" {
		s.WriteString(successStyle.Render(m.message + "\n\n"))
	}

	switch m.state {
	case menuView, listView, removeView:
		s.WriteString(m.list.View())
		if m.state == menuView {
			s.WriteString("\n\n")
			s.WriteString("Use ↑/↓ to navigate, Enter to select, q to quit")
		}
	case addView:
		s.WriteString(m.list.View())
		s.WriteString("\n\n")
		s.WriteString("Press Enter to select branch, ESC to cancel")
	case pathSelectView:
		s.WriteString(m.list.View())
		s.WriteString("\n\n")
		s.WriteString("💡 Suggestions are learned from your existing worktrees\n")
		s.WriteString("Press Enter to select, ESC to cancel")
	case customPathView:
		s.WriteString(titleStyle.Render(fmt.Sprintf("Custom path for '%s'", m.selectedBranch)))
		s.WriteString("\n\n")
		s.WriteString("Enter custom path:\n")
		s.WriteString(m.pathInput.View())
		s.WriteString("\n\n")
		s.WriteString("Press Enter to confirm, ESC to cancel")
	}

	return s.String()
}
