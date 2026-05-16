package vault

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"
)

type Entry struct {
	Name    string `json:"name"`
	Value   string `json:"value"`
	Created string `json:"created"`
}

type Vault struct {
	Keys []Entry `json:"keys"`
}

var ErrNotFound = errors.New("key not found")

func Path() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "locket", "vault.age"), nil
}

func Exists() (bool, error) {
	p, err := Path()
	if err != nil {
		return false, err
	}
	_, err = os.Stat(p)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func Load(passphrase string) (*Vault, error) {
	p, err := Path()
	if err != nil {
		return nil, err
	}
	ct, err := os.ReadFile(p)
	if err != nil {
		return nil, err
	}
	pt, err := decrypt(ct, passphrase)
	if err != nil {
		return nil, err
	}
	var v Vault
	if err := json.Unmarshal(pt, &v); err != nil {
		return nil, fmt.Errorf("parse vault: %w", err)
	}
	return &v, nil
}

func (v *Vault) Save(passphrase string) error {
	p, err := Path()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o700); err != nil {
		return err
	}
	pt, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	ct, err := encrypt(pt, passphrase)
	if err != nil {
		return err
	}
	tmp := p + ".tmp"
	if err := os.WriteFile(tmp, ct, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, p)
}

func (v *Vault) Add(name, value string) error {
	for _, e := range v.Keys {
		if e.Name == name {
			return fmt.Errorf("key %q already exists", name)
		}
	}
	v.Keys = append(v.Keys, Entry{
		Name:    name,
		Value:   value,
		Created: time.Now().Format("2006-01-02"),
	})
	return nil
}

func (v *Vault) Get(name string) (string, error) {
	for _, e := range v.Keys {
		if e.Name == name {
			return e.Value, nil
		}
	}
	return "", ErrNotFound
}

func (v *Vault) Names() []string {
	out := make([]string, 0, len(v.Keys))
	for _, e := range v.Keys {
		out = append(out, e.Name)
	}
	sort.Strings(out)
	return out
}

func (v *Vault) Remove(name string) error {
	for i, e := range v.Keys {
		if e.Name == name {
			v.Keys = append(v.Keys[:i], v.Keys[i+1:]...)
			return nil
		}
	}
	return ErrNotFound
}
