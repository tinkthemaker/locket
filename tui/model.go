package tui

import (
	"fmt"
	"time"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/atotto/clipboard"

	"github.com/tinkthemaker/locket/vault"
)

const clipboardTTL = 30 * time.Second

type item struct {
	name    string
	value   string
	created string
}

func (i item) Title() string       { return i.name }
func (i item) Description() string { return i.created }
func (i item) FilterValue() string { return i.name }

type clearClipboardMsg struct{ value string }

type keyMap struct {
	Copy   key.Binding
	Reveal key.Binding
	Quit   key.Binding
}

var keys = keyMap{
	Copy:   key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "copy")),
	Reveal: key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "reveal")),
	Quit:   key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit")),
}

type model struct {
	list   list.Model
	reveal bool
}

func Run(v *vault.Vault) error {
	items := make([]list.Item, 0, len(v.Keys))
	for _, e := range v.Keys {
		items = append(items, item{name: e.Name, value: e.Value, created: e.Created})
	}
	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = "locket"
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(true)
	l.AdditionalShortHelpKeys = func() []key.Binding {
		return []key.Binding{keys.Copy, keys.Reveal}
	}
	l.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{keys.Copy, keys.Reveal}
	}

	_, err := tea.NewProgram(model{list: l}).Run()
	return err
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		h := msg.Height
		if m.reveal {
			h -= 2
		}
		m.list.SetSize(msg.Width, h)

	case clearClipboardMsg:
		current, err := clipboard.ReadAll()
		if err == nil && current == msg.value {
			_ = clipboard.WriteAll("")
			return m, m.list.NewStatusMessage("Clipboard cleared")
		}
		return m, nil

	case tea.KeyPressMsg:
		if m.list.FilterState() == list.Filtering {
			break
		}
		switch {
		case key.Matches(msg, keys.Quit):
			return m, tea.Quit
		case key.Matches(msg, keys.Reveal):
			m.reveal = !m.reveal
			return m, nil
		case key.Matches(msg, keys.Copy):
			sel, ok := m.list.SelectedItem().(item)
			if !ok {
				return m, nil
			}
			if err := clipboard.WriteAll(sel.value); err != nil {
				return m, m.list.NewStatusMessage("Clipboard error: " + err.Error())
			}
			value := sel.value
			return m, tea.Batch(
				m.list.NewStatusMessage(fmt.Sprintf("Copied %q — clears in %ds", sel.name, int(clipboardTTL.Seconds()))),
				tea.Tick(clipboardTTL, func(time.Time) tea.Msg {
					return clearClipboardMsg{value: value}
				}),
			)
		}
	}
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m model) View() tea.View {
	parts := []string{m.list.View()}
	if m.reveal {
		if sel, ok := m.list.SelectedItem().(item); ok {
			parts = append(parts,
				revealLabelStyle.Render("value:")+revealStyle.Render(sel.value))
		}
	}
	v := tea.NewView(lipgloss.JoinVertical(lipgloss.Left, parts...))
	v.AltScreen = true
	return v
}
