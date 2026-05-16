package vault

import (
	"testing"
)

func TestCryptoRoundTrip(t *testing.T) {
	plaintext := []byte(`{"keys":[{"name":"openai","value":"sk-test","created":"2026-05-15"}]}`)
	ct, err := encrypt(plaintext, "correct horse battery staple")
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}
	pt, err := decrypt(ct, "correct horse battery staple")
	if err != nil {
		t.Fatalf("decrypt: %v", err)
	}
	if string(pt) != string(plaintext) {
		t.Fatalf("round-trip mismatch: got %q want %q", pt, plaintext)
	}
}

func TestDecryptWrongPassword(t *testing.T) {
	ct, err := encrypt([]byte("secret"), "right")
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}
	if _, err := decrypt(ct, "wrong"); err != ErrIncorrectPassword {
		t.Fatalf("expected ErrIncorrectPassword, got %v", err)
	}
}

func TestVaultOps(t *testing.T) {
	v := &Vault{}
	if err := v.Add("openai", "sk-1"); err != nil {
		t.Fatalf("add: %v", err)
	}
	if err := v.Add("stripe", "sk_live"); err != nil {
		t.Fatalf("add: %v", err)
	}
	if err := v.Add("openai", "dup"); err == nil {
		t.Fatal("expected duplicate-name error")
	}

	got, err := v.Get("openai")
	if err != nil || got != "sk-1" {
		t.Fatalf("get openai: %q %v", got, err)
	}
	if _, err := v.Get("nope"); err != ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}

	names := v.Names()
	if len(names) != 2 || names[0] != "openai" || names[1] != "stripe" {
		t.Fatalf("names: %v", names)
	}

	if err := v.Remove("openai"); err != nil {
		t.Fatalf("remove: %v", err)
	}
	if _, err := v.Get("openai"); err != ErrNotFound {
		t.Fatal("openai should be gone")
	}
	if err := v.Remove("openai"); err != ErrNotFound {
		t.Fatalf("remove missing: %v", err)
	}
}
