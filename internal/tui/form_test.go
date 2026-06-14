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

// Duplicating must pre-fill the form with the source's values plus the
// suggested ID, and must NOT be in editing mode (it saves as a new connection).
func TestNewDuplicateFormModelPrefills(t *testing.T) {
	conn := &config.Connection{
		ID:           "web-prod",
		Host:         "prod.example.com",
		User:         "admin",
		Port:         2222,
		IdentityFile: "~/.ssh/work_key",
		Project:      "web",
		Env:          "prod",
		Tags:         []string{"prod", "web"},
	}

	m := NewDuplicateFormModel(conn, "web-prod-copy")

	if m.IsEditing() {
		t.Error("expected duplicate form NOT to be in editing mode")
	}
	if m.OriginalID() != "" {
		t.Errorf("expected empty originalID for duplicate, got %q", m.OriginalID())
	}
	if got := m.inputs[fieldID].Value(); got != "web-prod-copy" {
		t.Errorf("expected pre-filled ID web-prod-copy, got %q", got)
	}
	if got := m.inputs[fieldHost].Value(); got != "prod.example.com" {
		t.Errorf("expected host pre-filled, got %q", got)
	}
	if got := m.inputs[fieldTags].Value(); got != "prod, web" {
		t.Errorf("expected tags pre-filled, got %q", got)
	}
}

// A duplicate must carry over fields the form does not expose and must not
// share mutable state (Options, UseMosh) with the source connection.
func TestNewDuplicateFormModelPreservesAndDeepCopies(t *testing.T) {
	mosh := true
	src := &config.Connection{
		ID:           "gateway",
		Host:         "gw.example.com",
		User:         "ops",
		Port:         22,
		ProxyJump:    "bastion",
		ForwardAgent: true,
		UseMosh:      &mosh,
		Options:      map[string]string{"ServerAliveInterval": "60"},
		Tags:         []string{"infra"},
	}

	m := NewDuplicateFormModel(src, "gateway-copy")

	dup, err := m.GetConnection()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Unexposed fields carried over.
	if dup.ProxyJump != "bastion" {
		t.Errorf("expected ProxyJump preserved, got %q", dup.ProxyJump)
	}
	if !dup.ForwardAgent {
		t.Error("expected ForwardAgent preserved")
	}
	if dup.UseMosh == nil || !*dup.UseMosh {
		t.Error("expected UseMosh preserved")
	}
	if dup.Options["ServerAliveInterval"] != "60" {
		t.Errorf("expected Options preserved, got %v", dup.Options)
	}

	// Mutating the duplicate must not affect the source (no aliasing).
	dup.Options["ServerAliveInterval"] = "99"
	*dup.UseMosh = false
	dup.Tags[0] = "mutated"

	if src.Options["ServerAliveInterval"] != "60" {
		t.Errorf("source Options aliased: %v", src.Options)
	}
	if src.UseMosh == nil || !*src.UseMosh {
		t.Error("source UseMosh aliased")
	}
	if src.Tags[0] != "infra" {
		t.Errorf("source Tags aliased: %v", src.Tags)
	}
}
