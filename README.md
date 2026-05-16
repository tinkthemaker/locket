# locket

Tiny encrypted key vault for developers. One password, one file, one binary.

```
                  ╭────────────────────────────╮
                  │                            │
                  │       ◆  l o c k e t       │
                  │    tiny encrypted vault    │
                  │                            │
                  ╰────────────────────────────╯

  ╭────────────────────────────────────────────────────────────────────╮
  │   Welcome.                                                         │
  │                                                                    │
  │   Choose a master password — you only need to remember this one.   │
  │   It protects every key you'll add to your vault.                  │
  │                                                                    │
  │   Master password                                                  │
  │   → ••••••••••••                                                   │
  │                                                                    │
  │   Confirm                                                          │
  │   → ••••••••••••                                                   │
  ╰────────────────────────────────────────────────────────────────────╯

           tab switch field   ·   enter create vault   ·   esc quit
```

## Why locket?

If you're a developer who wants:

- A place to put API keys without leaving them in `.env` files, shell history, or pasted into a chat
- Something simpler than a full password manager — no browser extension, no sync server, no cloud account, no daemon
- One master password that protects everything
- A clean TUI you'd actually want to use

…then locket might fit. If you need shared vaults, browser autofill, or family sharing — you want 1Password or Bitwarden, not this.

## Install

```bash
go install github.com/tinkthemaker/locket@latest
```

Or build from source:

```bash
git clone https://github.com/tinkthemaker/locket
cd locket
go build
```

Requires Go 1.24+.

## First run

The first time you run `locket`, it walks you through creating a master password. The vault is created at `~/.config/locket/vault.age` (or the platform equivalent of `os.UserConfigDir()`).

## Daily use

Run `locket` and unlock. You'll land in the browser:

```
  ◆  locket  ·  3 keys


  │ openai
  │ added 2026-05-15

    stripe_prod
    added 2026-05-10

    github_token
    added 2026-05-01




  enter copy   ·   r reveal   ·   n new   ·   d delete   ·   / filter   ·   q quit
```

- `enter` — copy the selected value to your clipboard (auto-clears in 30s)
- `r` — reveal the value inline
- `n` — add a new key
- `d` — delete the selected key
- `/` — filter the list as you type
- `q` — quit

That's the whole interface. No menus, no settings page, no profile picker.

## CLI for shell scripts

For pipelines and substitution, two commands skip the TUI:

```bash
locket get NAME     # print the value to stdout (prompts for master password)
locket where        # print the vault file path
```

Useful as:

```bash
export OPENAI_API_KEY=$(locket get openai)
```

## How it works

Your master password is the only key. It runs through [age](https://github.com/FiloSottile/age)'s scrypt passphrase mode — a memory-hard KDF that takes ~2.5 seconds on modern hardware — to derive the vault encryption key. That delay *is* the rate limit against brute force; locket adds no artificial throttling or lockout counter on top.

Your secrets live inside a single age-encrypted JSON file:

```
~/.config/locket/vault.age    # mode 0600
~/.config/locket/             # mode 0700
```

Writes are atomic (tmp + rename), so a crash during save can't corrupt the vault. The master password is held in process memory only while locket is running — there's no background daemon, no cached key on disk, no OS keychain integration. Every invocation prompts for the password and exits cleanly.

The clipboard auto-clears 30 seconds after a copy, but only if the clipboard still holds the copied value — so it won't trample something else you've pasted in the meantime.

## Sync and backup

The vault is just a file. Two paths that work well:

- **Dotfiles repo**: commit `~/.config/locket/vault.age` to a private git repo
- **Syncthing / Dropbox / iCloud Drive**: put `~/.config/locket/` in a synced folder

Both work because the file is encrypted at rest. locket doesn't ship its own sync.

## What's intentionally not included

These were considered and deliberately left out, because each one expands the attack surface or breaks the "one binary, one file" simplicity:

- OS keychain integration (macOS Keychain, libsecret, Windows DPAPI)
- A background daemon, session cache, or "remember me"
- A second key layer (master-key-encrypts-data-key)
- Built-in sync, backup, or cloud features
- Multiple vault profiles
- Import/export from other password managers
- Audit log, expiry tracking, rotation reminders
- Per-key notes, tags, or folders
- Browser extension or autofill
- Shell `eval` integration

If one of these turns out to matter, file an issue and explain the use case — but the bias is strongly toward keeping locket small.

## Threat model

locket protects the vault file at rest against an attacker who reads your disk: a stolen laptop, a leaked backup, a synced dotfiles repo, a misconfigured cloud bucket. It does **not** defend against malware running as your user (the decrypted values flow through process memory and the clipboard), or a coercive attacker with you present. For those threats you need hardware tokens or a different architecture.

## Development

Build, test, and vet:

```bash
go build ./...
go test ./...
go vet ./...
```

The TUI screen layouts have a snapshot test that prints each screen at 80×24:

```bash
go test ./tui/ -run TestRenderScreens -v
```

If you're an AI coding agent working on this repo, read [`AGENTS.md`](./AGENTS.md) first — it pins down the project's scope and what's deliberately out of bounds.

## License

MIT (see [`LICENSE`](./LICENSE)).
