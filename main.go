package main

import (
	"errors"
	"fmt"
	"os"

	"golang.org/x/term"

	"github.com/tinkthemaker/locket/tui"
	"github.com/tinkthemaker/locket/vault"
)

func readPasswordTTY(prompt string) (string, error) {
	fd := int(os.Stdin.Fd())
	if !term.IsTerminal(fd) {
		return "", errors.New("stdin is not a terminal; cannot read password securely")
	}
	fmt.Fprint(os.Stderr, prompt)
	b, err := term.ReadPassword(fd)
	fmt.Fprintln(os.Stderr)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func main() {
	if len(os.Args) >= 2 {
		switch os.Args[1] {
		case "-h", "--help", "help":
			printUsage()
			return
		case "get":
			if err := runGet(os.Args[2:]); err != nil {
				fmt.Fprintln(os.Stderr, "error:", err)
				os.Exit(1)
			}
			return
		case "where":
			if err := runWhere(); err != nil {
				fmt.Fprintln(os.Stderr, "error:", err)
				os.Exit(1)
			}
			return
		default:
			fmt.Fprintf(os.Stderr, "unknown command: %s\n\n", os.Args[1])
			printUsage()
			os.Exit(2)
		}
	}

	if err := tui.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`locket — tiny encrypted key vault

Usage:
  locket              Open the TUI (add, view, copy, remove keys)
  locket get NAME     Print the value of NAME to stdout (for scripts)
  locket where        Print the vault file path

On first run, locket walks you through creating a master password.`)
}

func runGet(args []string) error {
	if len(args) != 1 {
		return errors.New("usage: locket get NAME")
	}
	pw, err := readPasswordTTY("Master password: ")
	if err != nil {
		return err
	}
	v, err := vault.Load(pw)
	if err != nil {
		return err
	}
	val, err := v.Get(args[0])
	if err != nil {
		return err
	}
	fmt.Println(val)
	return nil
}

func runWhere() error {
	p, err := vault.Path()
	if err != nil {
		return err
	}
	fmt.Println(p)
	return nil
}
