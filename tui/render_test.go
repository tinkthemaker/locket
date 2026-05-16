package tui

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/tinkthemaker/locket/vault"
)

// Verifies the password input is focused immediately after initialModel
// (no Tab required) so typing actually reaches it. Without this, the
// unlock screen would silently swallow keystrokes because textinput.Update
// short-circuits on !focus.
func TestPasswordInputFocusedAtStart(t *testing.T) {
	m, err := initialModel()
	if err != nil {
		t.Fatalf("initialModel: %v", err)
	}
	if !m.passInput.Focused() {
		t.Fatal("passInput should be focused at startup but isn't")
	}

	updated, _ := m.Update(tea.KeyPressMsg{Code: 'h', Text: "h"})
	if got := updated.(Model).passInput.Value(); got != "h" {
		t.Fatalf("typing 'h' should land in passInput; got %q", got)
	}
}

// Renders each screen at a fixed terminal size so we can eyeball the layout.
// Run with: go test ./tui/ -run Render -v
func TestRenderScreens(t *testing.T) {
	const w, h = 80, 24

	v := &vault.Vault{Keys: []vault.Entry{
		{Name: "openai", Value: "sk-test-123", Created: "2026-05-15"},
		{Name: "stripe_prod", Value: "sk_live_xyz", Created: "2026-05-10"},
		{Name: "github_token", Value: "ghp_abc", Created: "2026-05-01"},
	}}

	cases := []struct {
		name  string
		setup func() Model
	}{
		{"welcome", func() Model {
			m, _ := initialModel()
			m.screen = screenWelcome
			m.passInput.Focus()
			return m
		}},
		{"welcome_error", func() Model {
			m, _ := initialModel()
			m.screen = screenWelcome
			m.errMsg = "Passwords don't match. Try again."
			m.passInput.SetValue("hunter2hunter2")
			m.confirmInput.Focus()
			return m
		}},
		{"unlock", func() Model {
			m, _ := initialModel()
			m.screen = screenUnlock
			m.passInput.Focus()
			return m
		}},
		{"unlock_error", func() Model {
			m, _ := initialModel()
			m.screen = screenUnlock
			m.errMsg = "Wrong password — try again."
			m.passInput.Focus()
			return m
		}},
		{"browse", func() Model {
			m, _ := initialModel()
			m.vault = v
			mm, _ := m.enterBrowse("")
			return mm
		}},
		{"browse_with_status", func() Model {
			m, _ := initialModel()
			m.vault = v
			mm, _ := m.enterBrowse("")
			mm.statusMsg = `Copied "openai" — clears in 30s.`
			return mm
		}},
		{"browse_reveal", func() Model {
			m, _ := initialModel()
			m.vault = v
			mm, _ := m.enterBrowse("")
			mm.reveal = true
			return mm
		}},
		{"browse_empty", func() Model {
			m, _ := initialModel()
			m.vault = &vault.Vault{}
			mm, _ := m.enterBrowse("")
			return mm
		}},
		{"add", func() Model {
			m, _ := initialModel()
			m.vault = v
			m.screen = screenAdd
			m.nameInput.SetValue("anthropic")
			m.nameInput.Focus()
			return m
		}},
		{"confirm", func() Model {
			m, _ := initialModel()
			m.vault = v
			mm, _ := m.enterBrowse("")
			mm.deleteName = "stripe_prod"
			mm.screen = screenConfirm
			return mm
		}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := tc.setup()
			model, _ := m.Update(tea.WindowSizeMsg{Width: w, Height: h})
			view := model.(Model).View()
			if strings.TrimSpace(view.Content) == "" {
				t.Fatalf("%s rendered empty content", tc.name)
			}
			t.Logf("\n=== %s ===\n%s", tc.name, view.Content)
		})
	}
}
