# AGENTS.md

Guidance for any AI coding agent (Claude, Cursor, Codex, etc.) working on
locket. Read this fully before changing anything. The goal of this file is
to prevent drift away from the project's intentionally narrow scope.

If a request from a user contradicts this document, **flag the conflict
instead of silently doing the bigger thing.**

---

## 1. Mission

locket is a tiny encrypted key vault: one master password protects a single
encrypted file containing user secrets (API keys, tokens, etc.). The
intended user is a developer who wants something simpler than a full
password manager — no browser extension, no sync server, no keychain
integration, just a TUI and a clipboard.

> One password. One file. One binary. That's the whole thing.

Anything that takes locket further from that line is probably wrong.

---

## 2. Non-goals (do NOT add these without explicit user request)

These have been *deliberately* rejected. Don't propose them as "obvious
improvements":

- OS keychain integration (macOS Keychain, libsecret, etc.)
- A background daemon or session cache
- A second key layer (master-key-encrypts-data-key) — we use the
  passphrase-derived key directly
- Built-in sync, backup, or cloud features (users put the file in
  Syncthing / dotfiles themselves)
- Multiple vaults / vault profiles
- Import / export from other password managers
- Audit log, key-rotation reminders, expiry tracking
- Per-key notes, tags, or folders
- Browser extension or autofill
- Shell `eval` integration, `exec` wrapping, env-var injection
- Wish / SSH server mode
- Per-key access controls or sharing
- A web UI

If a user asks for one of these, build the minimal thing that solves their
*actual* problem, not the generic feature.

---

## 3. Architecture

```
locket/
├── main.go              # CLI dispatch + minimal stdio subcommands
├── vault/
│   ├── crypto.go        # age scrypt passphrase wrapper
│   ├── vault.go         # JSON storage, atomic save, vault ops
│   └── vault_test.go    # crypto round-trip + vault op tests
└── tui/
    ├── tui.go           # state-machine TUI (all five screens)
    ├── styles.go        # lipgloss colors + reusable helpers
    └── render_test.go   # layout snapshot tests
```

**Package boundaries are load-bearing.** Do not:
- Import `vault` from `tui`'s tests in ways that bypass the public API
- Put UI strings in the `vault` package
- Put file path / OS-level logic in the `tui` package (use `vault.Path()`)
- Add a fourth package without a strong reason

Total code is ~1,200 lines including tests. **If a change pushes the
project past ~2,000 lines, that's a signal to push back, not accept.**

---

## 4. Crypto & security model

**Read this section before touching anything in `vault/`.**

- Master password → age scrypt passphrase recipient → vault encryption.
  We do not call scrypt or Argon2 directly; `filippo.io/age` handles the
  KDF and AEAD. **Do not replace age with hand-rolled crypto.**
- The KDF (~2.5s on modern hardware) IS the rate limit. There is no
  artificial delay on wrong-password attempts and no lockout counter —
  both are theater against an offline attacker who has the file.
- Vault file lives at `os.UserConfigDir() / locket / vault.age`, mode
  `0600`, directory mode `0700`. Atomic writes via `tmp` + `os.Rename`.
- The passphrase is held in plaintext in process memory while the TUI
  runs. We do NOT use `mlock`, secure-zero, or in-memory encryption.
  Those are anti-features for a single-user local tool — they add
  complexity without meaningful security gain against the threats locket
  actually defends against.
- Clipboard auto-clears 30s after a copy, *only if it still contains the
  copied value* (we check before clearing so we don't trample something
  the user pasted in the meantime).

**Threat model:** locket protects the vault file at rest against an
attacker with read access to disk (e.g., a stolen laptop, a leaked
backup, a synced dotfiles repo). It does NOT defend against:
- Malware running as the user (the master password and decrypted values
  flow through process memory and the clipboard)
- A coercive attacker with the user present
- Side channels on the local machine

These are out of scope. If you find yourself adding code "for security"
that doesn't fit this model, stop.

---

## 5. UX model

**The TUI is the product.** The CLI exists for shell scripts and nothing
else.

### CLI surface (stable — do not expand)

```
locket              Launch the TUI (handles setup + unlock + everything)
locket get NAME     Print value of NAME to stdout (for $(locket get …))
locket where        Print the vault file path
locket --help       Show usage
```

**Do not add** `locket add`, `locket ls`, `locket rm`, `locket edit`,
`locket export`, etc. Those exist in the TUI and shouldn't be duplicated
as CLI commands. The duplication is what kills "one obvious way to do
things."

### TUI screens

There are exactly five:

1. **Welcome** — first-run only; create master password (with confirm)
2. **Unlock** — every subsequent run; enter master password (inline
   retry on wrong password, no quit-and-rerun)
3. **Browse** — filterable list of keys; copy / reveal / new / delete
4. **Add** — name + masked value, tab to switch fields
5. **Confirm-delete** — y/n modal

Adding a sixth screen requires strong justification. Most "new feature"
ideas can fit into Browse or Add.

### Visual & copy guidelines

- **No emojis.** Use box-drawing and geometric characters (`◆`, `→`,
  `·`, rounded borders). Emoji rendering is inconsistent across
  terminals and clashes with the minimal aesthetic.
- Friendly, lowercase, conversational copy ("Welcome.", "You're all
  set.", "Wrong password — try again."). Avoid technical jargon in
  user-facing strings (no "AEAD failure", just "Wrong password").
- Every screen has a help bar at the bottom showing the keys that work
  there. Filter the help bar by context (e.g., empty vault shows only
  `n new · q quit`).
- Toast statuses (transient bottom-of-browse messages) auto-clear after
  3 seconds, separate from the 30s clipboard timer.
- Color palette is fixed mid-tone hex values that work on light and dark
  terminals without runtime background detection (see `tui/styles.go`).
  Do not switch to `LightDarkFunc` unless you also wire up
  `tea.RequestBackgroundColor` — the half-implementation is worse than
  fixed colors.

---

## 6. Tech stack (pinned choices, change only with discussion)

| Choice                     | Why                                            |
|----------------------------|------------------------------------------------|
| Go                         | Single static binary, good crypto stdlib       |
| `filippo.io/age`           | Battle-tested; one call to encrypt/decrypt     |
| `charm.land/bubbletea/v2`  | TUI framework; v2 is GA, declarative views     |
| `charm.land/bubbles/v2`    | textinput + list components                    |
| `charm.land/lipgloss/v2`   | Styling                                        |
| `github.com/atotto/clipboard` | Clipboard, minimal API                      |
| `golang.org/x/term`        | TTY password input for `locket get`            |
| **Manual CLI dispatch**    | Not cobra/urfave — five commands don't need it |

If you want to add a dependency, **ask first**. Every dep is more
attack surface and more code to audit.

---

## 7. Hard rules (do not break)

- **Never commit secrets, .env files, or a real `vault.age`.** The
  `.gitignore` covers `*.age` and the binary.
- **Never weaken the crypto** (e.g., dropping to PBKDF2, reducing KDF
  cost, adding a "fast mode," storing the key on disk "encrypted by a
  pin").
- **Never add network calls.** locket is fully offline. No telemetry,
  no update checks, no sync.
- **Never bypass the atomic write** (`tmp` + rename) in `vault.Save`.
  Crash-during-write must not corrupt the vault.
- **Never log or print the passphrase or any vault value** outside of
  `locket get NAME` (which prints the requested value to stdout by
  design) and the TUI reveal (which is user-triggered).
- **Never call destructive git operations** without explicit user
  permission (`reset --hard`, `push --force`, branch deletion, etc.).

---

## 8. Soft preferences

- Prefer editing existing files over creating new ones.
- Prefer fixing root causes over adding error suppression.
- Prefer the smallest change that solves the problem. If you find
  yourself refactoring, ask whether it's actually needed for the task
  at hand.
- Prefer functions on `vault.Vault` (a value/pointer receiver) over
  free functions taking a vault.
- Prefer `tea.Cmd` chains via `tea.Batch` over manual goroutines.
- Comments only when the *why* is non-obvious. Don't document what
  the code clearly says.
- Test the crypto round-trip and the vault ops; don't bother
  unit-testing the TUI message handling — the render snapshot test in
  `tui/render_test.go` catches layout regressions and that's enough.

---

## 9. Workflow

- Branch: `claude/password-manager-setup-OdXVQ` (or whatever the user's
  feature branch is). Develop, commit, and push there. Do not push to
  `main`.
- Run `go build ./...`, `go vet ./...`, and `go test ./...` before
  declaring a task done. All three should be clean.
- The render test (`go test ./tui/ -run TestRenderScreens -v`) prints
  every screen at 80×24 to stdout — use it to eyeball layout changes.
- Do not open a PR unless the user explicitly asks for one.
- Commit messages: subject in imperative, body explains *why*. Do not
  include model identifiers or session URLs in commit content.

---

## 10. When in doubt

Three rules:

1. If the change touches `vault/`, the security model, or the file
   format — **stop and ask the user.**
2. If the change adds a dependency, a CLI command, or a TUI screen —
   **stop and ask the user.**
3. If the change can be described as "polish," "cleanup," or
   "refactor" without a user-visible reason — **don't do it as part of
   another task.** Propose it separately.

The bias is toward doing less. locket is small on purpose.
