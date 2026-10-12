package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestHostKeyPersists(t *testing.T) {
	path := filepath.Join(t.TempDir(), "host_key.pem")
	first, err := loadHostKey(path)
	if err != nil {
		t.Fatal(err)
	}
	second, err := loadHostKey(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(first.PublicKey().Marshal(), second.PublicKey().Marshal()) {
		t.Fatal("host identity changed")
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0600 {
		t.Fatal("private key permissions must be 0600")
	}
}

func TestInvalidKeyIsNotOverwritten(t *testing.T) {
	path := filepath.Join(t.TempDir(), "host_key.pem")
	if err := os.WriteFile(path, []byte("invalid"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := loadHostKey(path); err == nil {
		t.Fatal("accepted invalid key")
	}
	data, _ := os.ReadFile(path)
	if string(data) != "invalid" {
		t.Fatal("overwrote existing key")
	}
}
