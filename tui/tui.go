package tui

import (
	"fmt"
	"time"

	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/atotto/clipboard"

	"github.com/tinkthemaker/locket/vault"
)

const clipboardTTL = 30 * time.Second

type screen int

const (
	screenWelcome screen = iota
	screenUnlock
	screenBrowse
	screenAdd
	screenConfirm
)

type listItem struct {
	name    string
	value   string
	created string
}

func (i listItem) Title() string       { return i.name }
func (i listItem) Description() string { return "added " + i.created }
func (i listItem) FilterValue() string { return i.name }

type (
	clearClipboardMsg struct{ value string }
	clearStatusMsg    int
)

type Model struct {
	screen screen
	width  int
	height int

	vault      *vault.Vault
	passphrase string

	passInput    textinput.Model
	confirmInput textinput.Model
	nameInput    textinput.Model
	valueInput   textinput.Model

	list   list.Model
	reveal bool

	addFocus   int
	deleteName string

	errMsg    string
	statusMsg string
	statusGen int
}

func Run() error {
	m, err := initialModel()
	if err != nil {
		return err
	}
	_, err = tea.NewProgram(m).Run()
	return err
}

func initialModel() (Model, error) {
	exists, err := vault.Exists()
	if err != nil {
		return Model{}, err
	}

	pass := textinput.New()
	pass.EchoMode = textinput.EchoPassword
	pass.Prompt = "→ "
	pass.Placeholder = "master password"
	pass.SetWidth(40)
	pass.Focus()

	confirm := textinput.New()
	confirm.EchoMode = textinput.EchoPassword
	confirm.Prompt = "→ "
	confirm.Placeholder = "confirm password"
	confirm.SetWidth(40)

	name := textinput.New()
	name.Prompt = "→ "
	name.Placeholder = "e.g. openai"
	name.SetWidth(40)

	value := textinput.New()
	value.EchoMode = textinput.EchoPassword
	value.Prompt = "→ "
	value.Placeholder = "your secret value"
	value.SetWidth(40)

	l := list.New(nil, list.NewDefaultDelegate(), 0, 0)
	l.SetShowTitle(false)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)
	l.SetShowHelp(false)
	l.DisableQuitKeybindings()

	m := Model{
		list:         l,
		passInput:    pass,
		confirmInput: confirm,
		nameInput:    name,
		valueInput:   value,
	}
	if exists {
		m.screen = screenUnlock
	} else {
		m.screen = screenWelcome
	}
	return m, nil
}

func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		listW := msg.Width - 6
		listH := msg.Height - 10
		if listH < 5 {
			listH = 5
		}
		if listW < 20 {
			listW = 20
		}
		m.list.SetSize(listW, listH)
		return m, nil

	case clearClipboardMsg:
		cur, _ := clipboard.ReadAll()
		if cur == msg.value {
			_ = clipboard.WriteAll("")
			m, cmd := m.setStatus("Clipboard cleared.")
			return m, cmd
		}
		return m, nil

	case clearStatusMsg:
		if int(msg) == m.statusGen {
			m.statusMsg = ""
		}
		return m, nil
	}

	switch m.screen {
	case screenWelcome:
		return m.updateWelcome(msg)
	case screenUnlock:
		return m.updateUnlock(msg)
	case screenBrowse:
		return m.updateBrowse(msg)
	case screenAdd:
		return m.updateAdd(msg)
	case screenConfirm:
		return m.updateConfirm(msg)
	}
	return m, nil
}

func (m Model) View() tea.View {
	var body string
	switch m.screen {
	case screenWelcome:
		body = m.viewWelcome()
	case screenUnlock:
		body = m.viewUnlock()
	case screenBrowse:
		body = m.viewBrowse()
	case screenAdd:
		body = m.viewAdd()
	case screenConfirm:
		body = m.viewConfirm()
	}
	v := tea.NewView(body)
	v.AltScreen = true
	return v
}

func (m Model) setStatus(s string) (Model, tea.Cmd) {
	m.statusMsg = s
	m.statusGen++
	gen := m.statusGen
	return m, tea.Tick(3*time.Second, func(time.Time) tea.Msg {
		return clearStatusMsg(gen)
	})
}

// ----- welcome -----

func (m Model) updateWelcome(msg tea.Msg) (tea.Model, tea.Cmd) {
	if k, ok := msg.(tea.KeyPressMsg); ok {
		switch k.String() {
		case "esc", "ctrl+c":
			return m, tea.Quit
		case "tab", "shift+tab", "down", "up":
			if m.passInput.Focused() {
				m.passInput.Blur()
				return m, m.confirmInput.Focus()
			}
			m.confirmInput.Blur()
			return m, m.passInput.Focus()
		case "enter":
			if m.passInput.Focused() {
				if m.passInput.Value() == "" {
					m.errMsg = "Password can't be empty."
					return m, nil
				}
				m.errMsg = ""
				m.passInput.Blur()
				return m, m.confirmInput.Focus()
			}
			return m.submitWelcome()
		}
	}
	var cmds []tea.Cmd
	var cmd tea.Cmd
	m.passInput, cmd = m.passInput.Update(msg)
	cmds = append(cmds, cmd)
	m.confirmInput, cmd = m.confirmInput.Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m Model) submitWelcome() (tea.Model, tea.Cmd) {
	pw := m.passInput.Value()
	confirm := m.confirmInput.Value()
	if pw == "" {
		m.errMsg = "Password can't be empty."
		return m, nil
	}
	if pw != confirm {
		m.errMsg = "Passwords don't match. Try again."
		m.confirmInput.SetValue("")
		m.confirmInput.Focus()
		return m, nil
	}
	v := &vault.Vault{}
	if err := v.Save(pw); err != nil {
		m.errMsg = "Could not save vault: " + err.Error()
		return m, nil
	}
	m.vault = v
	m.passphrase = pw
	m.errMsg = ""
	return m.enterBrowse("Vault created. You're all set.")
}

func (m Model) viewWelcome() string {
	card := lipgloss.JoinVertical(lipgloss.Left,
		headingStyle.Render("Welcome."),
		"",
		helpStyle.Render("Choose a master password — you only need to remember this one."),
		helpStyle.Render("It protects every key you'll add to your vault."),
		"",
		promptStyle.Render("Master password"),
		m.passInput.View(),
		"",
		promptStyle.Render("Confirm"),
		m.confirmInput.View(),
	)

	if m.errMsg != "" {
		card = lipgloss.JoinVertical(lipgloss.Left, card, "", errorStyle.Render(m.errMsg))
	}

	body := lipgloss.JoinVertical(lipgloss.Center,
		banner(),
		"",
		cardStyle.Render(card),
		"",
		helpBar(
			[2]string{"tab", "switch field"},
			[2]string{"enter", "create vault"},
			[2]string{"esc", "quit"},
		),
	)
	return center(m.width, m.height, body)
}

// ----- unlock -----

func (m Model) updateUnlock(msg tea.Msg) (tea.Model, tea.Cmd) {
	if k, ok := msg.(tea.KeyPressMsg); ok {
		switch k.String() {
		case "esc", "ctrl+c":
			return m, tea.Quit
		case "enter":
			return m.submitUnlock()
		}
	}
	var cmd tea.Cmd
	m.passInput, cmd = m.passInput.Update(msg)
	return m, cmd
}

func (m Model) submitUnlock() (tea.Model, tea.Cmd) {
	pw := m.passInput.Value()
	if pw == "" {
		m.errMsg = "Enter your master password."
		return m, nil
	}
	v, err := vault.Load(pw)
	if err != nil {
		m.errMsg = "Wrong password — try again."
		m.passInput.SetValue("")
		return m, nil
	}
	m.vault = v
	m.passphrase = pw
	m.errMsg = ""
	return m.enterBrowse("")
}

func (m Model) viewUnlock() string {
	card := lipgloss.JoinVertical(lipgloss.Left,
		headingStyle.Render("Unlock your vault."),
		"",
		promptStyle.Render("Master password"),
		m.passInput.View(),
	)

	if m.errMsg != "" {
		card = lipgloss.JoinVertical(lipgloss.Left, card, "", errorStyle.Render(m.errMsg))
	}

	body := lipgloss.JoinVertical(lipgloss.Center,
		banner(),
		"",
		cardStyle.Render(card),
		"",
		helpBar(
			[2]string{"enter", "unlock"},
			[2]string{"esc", "quit"},
		),
	)
	return center(m.width, m.height, body)
}

// ----- browse -----

func (m Model) enterBrowse(status string) (Model, tea.Cmd) {
	items := make([]list.Item, 0, len(m.vault.Keys))
	for _, e := range m.vault.Keys {
		items = append(items, listItem{name: e.Name, value: e.Value, created: e.Created})
	}
	setItems := m.list.SetItems(items)

	m.screen = screenBrowse
	m.reveal = false
	m.passInput.Blur()
	m.confirmInput.Blur()
	m.nameInput.Blur()
	m.valueInput.Blur()

	cmds := []tea.Cmd{setItems}
	if status != "" {
		mm, cmd := m.setStatus(status)
		return mm, tea.Batch(append(cmds, cmd)...)
	}
	return m, tea.Batch(cmds...)
}

func (m Model) updateBrowse(msg tea.Msg) (tea.Model, tea.Cmd) {
	if k, ok := msg.(tea.KeyPressMsg); ok {
		filtering := m.list.FilterState() == list.Filtering
		if !filtering {
			switch k.String() {
			case "q", "ctrl+c":
				return m, tea.Quit
			case "esc":
				if m.list.FilterState() == list.FilterApplied {
					break
				}
				return m, tea.Quit
			case "r":
				m.reveal = !m.reveal
				return m, nil
			case "n":
				return m.enterAdd()
			case "d", "delete":
				if sel, ok := m.list.SelectedItem().(listItem); ok {
					m.deleteName = sel.name
					m.screen = screenConfirm
					return m, nil
				}
				return m, nil
			case "enter":
				return m.copySelected()
			}
		}
	}
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m Model) copySelected() (tea.Model, tea.Cmd) {
	sel, ok := m.list.SelectedItem().(listItem)
	if !ok {
		return m, nil
	}
	if err := clipboard.WriteAll(sel.value); err != nil {
		mm, cmd := m.setStatus("Clipboard error: " + err.Error())
		return mm, cmd
	}
	mm, statusCmd := m.setStatus(fmt.Sprintf("Copied %q — clears in %ds.", sel.name, int(clipboardTTL.Seconds())))
	value := sel.value
	tick := tea.Tick(clipboardTTL, func(time.Time) tea.Msg {
		return clearClipboardMsg{value: value}
	})
	return mm, tea.Batch(statusCmd, tick)
}

func (m Model) viewBrowse() string {
	count := len(m.vault.Keys)
	var subtitle string
	switch count {
	case 0:
		subtitle = "empty"
	case 1:
		subtitle = "1 key"
	default:
		subtitle = fmt.Sprintf("%d keys", count)
	}

	header := lipgloss.JoinHorizontal(lipgloss.Center,
		smallBanner(),
		helpStyle.Render("  ·  "),
		helpStyle.Render(subtitle),
	)

	var body string
	if count == 0 {
		body = m.viewEmptyBrowse()
	} else {
		body = m.list.View()
		if m.reveal {
			if sel, ok := m.list.SelectedItem().(listItem); ok {
				revealBar := lipgloss.JoinHorizontal(lipgloss.Left,
					revealLabelStyle.Render("revealed "),
					revealValueStyle.Render(sel.value),
				)
				body = lipgloss.JoinVertical(lipgloss.Left, body, "", revealBar)
			}
		}
	}

	status := ""
	if m.statusMsg != "" {
		status = successStyle.Render(m.statusMsg)
	}

	var help string
	if count == 0 {
		help = helpBar(
			[2]string{"n", "new"},
			[2]string{"q", "quit"},
		)
	} else {
		help = helpBar(
			[2]string{"enter", "copy"},
			[2]string{"r", "reveal"},
			[2]string{"n", "new"},
			[2]string{"d", "delete"},
			[2]string{"/", "filter"},
			[2]string{"q", "quit"},
		)
	}

	parts := []string{header, "", body, ""}
	if status != "" {
		parts = append(parts, status, "")
	}
	parts = append(parts, help)

	return lipgloss.NewStyle().Padding(1, 2).Render(
		lipgloss.JoinVertical(lipgloss.Left, parts...),
	)
}

func (m Model) viewEmptyBrowse() string {
	prompt := helpStyle.Render("Press ") + keyHintStyle.Render("n") + helpStyle.Render(" to add your first key.")
	msg := lipgloss.JoinVertical(lipgloss.Center,
		headingStyle.Render("Your vault is empty."),
		"",
		prompt,
	)
	return cardStyle.Render(msg)
}

// ----- add -----

func (m Model) enterAdd() (Model, tea.Cmd) {
	m.nameInput.SetValue("")
	m.valueInput.SetValue("")
	m.addFocus = 0
	m.errMsg = ""
	m.screen = screenAdd
	m.valueInput.Blur()
	return m, m.nameInput.Focus()
}

func (m Model) updateAdd(msg tea.Msg) (tea.Model, tea.Cmd) {
	if k, ok := msg.(tea.KeyPressMsg); ok {
		switch k.String() {
		case "esc":
			return m.cancelAdd()
		case "ctrl+c":
			return m, tea.Quit
		case "tab", "shift+tab", "down", "up":
			return m.toggleAddFocus()
		case "enter":
			if m.addFocus == 0 {
				return m.toggleAddFocus()
			}
			return m.submitAdd()
		}
	}
	var cmds []tea.Cmd
	var cmd tea.Cmd
	if m.addFocus == 0 {
		m.nameInput, cmd = m.nameInput.Update(msg)
	} else {
		m.valueInput, cmd = m.valueInput.Update(msg)
	}
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m Model) toggleAddFocus() (tea.Model, tea.Cmd) {
	if m.addFocus == 0 {
		m.nameInput.Blur()
		m.addFocus = 1
		return m, m.valueInput.Focus()
	}
	m.valueInput.Blur()
	m.addFocus = 0
	return m, m.nameInput.Focus()
}

func (m Model) cancelAdd() (tea.Model, tea.Cmd) {
	m.nameInput.Blur()
	m.valueInput.Blur()
	return m.enterBrowse("")
}

func (m Model) submitAdd() (tea.Model, tea.Cmd) {
	name := m.nameInput.Value()
	value := m.valueInput.Value()
	if name == "" {
		m.errMsg = "Name can't be empty."
		return m, nil
	}
	if value == "" {
		m.errMsg = "Value can't be empty."
		return m, nil
	}
	if err := m.vault.Add(name, value); err != nil {
		m.errMsg = err.Error()
		return m, nil
	}
	if err := m.vault.Save(m.passphrase); err != nil {
		m.errMsg = "Could not save vault: " + err.Error()
		return m, nil
	}
	return m.enterBrowse(fmt.Sprintf("Added %q.", name))
}

func (m Model) viewAdd() string {
	card := lipgloss.JoinVertical(lipgloss.Left,
		headingStyle.Render("New key."),
		"",
		helpStyle.Render("Pick a short name you'll remember. The value is hidden as you type."),
		"",
		promptStyle.Render("Name"),
		m.nameInput.View(),
		"",
		promptStyle.Render("Value"),
		m.valueInput.View(),
	)

	if m.errMsg != "" {
		card = lipgloss.JoinVertical(lipgloss.Left, card, "", errorStyle.Render(m.errMsg))
	}

	body := lipgloss.JoinVertical(lipgloss.Center,
		smallBanner(),
		"",
		cardStyle.Render(card),
		"",
		helpBar(
			[2]string{"tab", "switch field"},
			[2]string{"enter", "save"},
			[2]string{"esc", "cancel"},
		),
	)
	return center(m.width, m.height, body)
}

// ----- confirm delete -----

func (m Model) updateConfirm(msg tea.Msg) (tea.Model, tea.Cmd) {
	if k, ok := msg.(tea.KeyPressMsg); ok {
		switch k.String() {
		case "y", "Y", "enter":
			return m.doDelete()
		case "n", "N", "esc":
			m.deleteName = ""
			return m.enterBrowse("")
		case "ctrl+c":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m Model) doDelete() (tea.Model, tea.Cmd) {
	name := m.deleteName
	if err := m.vault.Remove(name); err != nil {
		m.deleteName = ""
		return m.enterBrowse("Could not remove: " + err.Error())
	}
	if err := m.vault.Save(m.passphrase); err != nil {
		return m.enterBrowse("Could not save vault: " + err.Error())
	}
	m.deleteName = ""
	return m.enterBrowse(fmt.Sprintf("Removed %q.", name))
}

func (m Model) viewConfirm() string {
	line := helpStyle.Render("Permanently delete ") +
		titleStyle.Render(m.deleteName) +
		helpStyle.Render(" from your vault?")
	card := lipgloss.JoinVertical(lipgloss.Left,
		dangerLabelStyle.Render("Remove key?"),
		"",
		line,
	)

	body := lipgloss.JoinVertical(lipgloss.Center,
		smallBanner(),
		"",
		cardStyle.Render(card),
		"",
		helpBar(
			[2]string{"y", "yes, remove"},
			[2]string{"n", "no, keep it"},
		),
	)
	return center(m.width, m.height, body)
}
