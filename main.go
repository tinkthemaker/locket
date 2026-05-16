package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"

	"github.com/tinkthemaker/locket/tui"
	"github.com/tinkthemaker/locket/vault"
)

func main() {
	var err error
	if len(os.Args) < 2 {
		err = runTUI()
	} else {
		switch os.Args[1] {
		case "add":
			err = runAdd()
		case "get":
			err = runGet(os.Args[2:])
		case "ls", "list":
			err = runLs()
		case "rm", "remove":
			err = runRm(os.Args[2:])
		case "-h", "--help", "help":
			printUsage()
		default:
			fmt.Fprintf(os.Stderr, "unknown command: %s\n\n", os.Args[1])
			printUsage()
			os.Exit(2)
		}
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`locket - tiny encrypted key vault

Usage:
  locket              Open the TUI to browse keys
  locket add          Add a new key (prompts for name and value)
  locket get NAME     Print the value for NAME to stdout
  locket ls           List all key names
  locket rm NAME      Remove a key

Vault location: ~/.config/locket/vault.age`)
}

func readPassword(prompt string) (string, error) {
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

func readLine(prompt string) (string, error) {
	fmt.Fprint(os.Stderr, prompt)
	s, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimRight(s, "\r\n"), nil
}

func openVault() (*vault.Vault, string, error) {
	exists, err := vault.Exists()
	if err != nil {
		return nil, "", err
	}
	if !exists {
		fmt.Fprintln(os.Stderr, "No vault found. Creating a new one.")
		pw, err := readPassword("Choose master password: ")
		if err != nil {
			return nil, "", err
		}
		if pw == "" {
			return nil, "", errors.New("password cannot be empty")
		}
		pw2, err := readPassword("Confirm: ")
		if err != nil {
			return nil, "", err
		}
		if pw != pw2 {
			return nil, "", errors.New("passwords do not match")
		}
		v := &vault.Vault{}
		if err := v.Save(pw); err != nil {
			return nil, "", err
		}
		path, _ := vault.Path()
		fmt.Fprintf(os.Stderr, "Vault created at %s\n", path)
		return v, pw, nil
	}
	pw, err := readPassword("Master password: ")
	if err != nil {
		return nil, "", err
	}
	v, err := vault.Load(pw)
	if err != nil {
		return nil, "", err
	}
	return v, pw, nil
}

func runAdd() error {
	v, pw, err := openVault()
	if err != nil {
		return err
	}
	name, err := readLine("Name: ")
	if err != nil {
		return err
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return errors.New("name cannot be empty")
	}
	value, err := readPassword("Value: ")
	if err != nil {
		return err
	}
	if value == "" {
		return errors.New("value cannot be empty")
	}
	if err := v.Add(name, value); err != nil {
		return err
	}
	if err := v.Save(pw); err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "Added %q\n", name)
	return nil
}

func runGet(args []string) error {
	if len(args) != 1 {
		return errors.New("usage: locket get NAME")
	}
	v, _, err := openVault()
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

func runLs() error {
	v, _, err := openVault()
	if err != nil {
		return err
	}
	for _, n := range v.Names() {
		fmt.Println(n)
	}
	return nil
}

func runRm(args []string) error {
	if len(args) != 1 {
		return errors.New("usage: locket rm NAME")
	}
	v, pw, err := openVault()
	if err != nil {
		return err
	}
	if _, err := v.Get(args[0]); err != nil {
		return err
	}
	confirm, err := readLine(fmt.Sprintf("Remove %q? [y/N]: ", args[0]))
	if err != nil {
		return err
	}
	if strings.ToLower(strings.TrimSpace(confirm)) != "y" {
		fmt.Fprintln(os.Stderr, "Cancelled")
		return nil
	}
	if err := v.Remove(args[0]); err != nil {
		return err
	}
	if err := v.Save(pw); err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "Removed %q\n", args[0])
	return nil
}

func runTUI() error {
	v, _, err := openVault()
	if err != nil {
		return err
	}
	return tui.Run(v)
}
