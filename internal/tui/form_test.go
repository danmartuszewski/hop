package tui

import (
	"testing"

	"github.com/danmartuszewski/hop/internal/config"
)

func TestFormIdentityFilePrefilled(t *testing.T) {
	conn := &config.Connection{
		ID:           "prod",
		Host:         "prod.example.com",
		User:         "admin",
		Port:         2222,
		IdentityFile: "~/.ssh/work_key",
	}

	m := NewFormModel("Edit Connection", conn)

	if got := m.inputs[fieldIdentity].Value(); got != "~/.ssh/work_key" {
		t.Errorf("expected identity field pre-filled with %q, got %q", "~/.ssh/work_key", got)
	}
}

func TestFormGetConnectionIncludesIdentityFile(t *testing.T) {
	m := NewFormModel("Add Connection", nil)
	m.inputs[fieldID].SetValue("box")
	m.inputs[fieldHost].SetValue("box.example.com")
	m.inputs[fieldIdentity].SetValue("~/.ssh/id_ed25519")

	conn, err := m.GetConnection()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if conn.IdentityFile != "~/.ssh/id_ed25519" {
		t.Errorf("expected IdentityFile %q, got %q", "~/.ssh/id_ed25519", conn.IdentityFile)
	}
}

func TestFormGetConnectionTrimsIdentityFile(t *testing.T) {
	m := NewFormModel("Add Connection", nil)
	m.inputs[fieldID].SetValue("box")
	m.inputs[fieldHost].SetValue("box.example.com")
	m.inputs[fieldIdentity].SetValue("  ~/.ssh/id_ed25519  ")

	conn, err := m.GetConnection()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if conn.IdentityFile != "~/.ssh/id_ed25519" {
		t.Errorf("expected trimmed IdentityFile, got %q", conn.IdentityFile)
	}
}

// Editing a connection through the form must not drop fields that the form
// does not expose (proxy jump, agent forwarding, mosh, extra options).
func TestFormPreservesUnexposedFieldsOnEdit(t *testing.T) {
	mosh := true
	conn := &config.Connection{
		ID:           "gateway",
		Host:         "gw.example.com",
		User:         "ops",
		Port:         22,
		IdentityFile: "~/.ssh/gw_key",
		ProxyJump:    "bastion",
		ForwardAgent: true,
		UseMosh:      &mosh,
		Options:      map[string]string{"ServerAliveInterval": "60"},
	}

	m := NewFormModel("Edit Connection", conn)

	// Change only the user; everything else stays as pre-filled.
	m.inputs[fieldUser].SetValue("admin")

	updated, err := m.GetConnection()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if updated.User != "admin" {
		t.Errorf("expected user to be updated to admin, got %q", updated.User)
	}
	if updated.ProxyJump != "bastion" {
		t.Errorf("expected ProxyJump preserved, got %q", updated.ProxyJump)
	}
	if !updated.ForwardAgent {
		t.Error("expected ForwardAgent preserved")
	}
	if updated.UseMosh == nil || !*updated.UseMosh {
		t.Error("expected UseMosh preserved")
	}
	if updated.Options["ServerAliveInterval"] != "60" {
		t.Errorf("expected Options preserved, got %v", updated.Options)
	}
	if updated.IdentityFile != "~/.ssh/gw_key" {
		t.Errorf("expected IdentityFile preserved, got %q", updated.IdentityFile)
	}
}
