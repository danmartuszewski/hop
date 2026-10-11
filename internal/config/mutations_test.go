package config

import (
	"path/filepath"
	"reflect"
	"testing"
)

func TestConnectionMutationsPreserveGroups(t *testing.T) {
	for _, operation := range []string{"rename", "delete"} {
		t.Run(operation, func(t *testing.T) {
			cfg := &Config{
				Version:     1,
				Connections: []Connection{{ID: "web", Host: "web.example.com"}, {ID: "db", Host: "db.example.com"}},
				Groups:      map[string][]string{"production": {"web", "db", "web"}, "web-only": {"web"}, "database": {"db"}},
			}
			want := map[string][]string{"production": {"renamed", "db", "renamed"}, "web-only": {"renamed"}, "database": {"db"}}
			if operation == "rename" {
				conn := cfg.FindConnection("web").Clone()
				conn.ID = "renamed"
				if !cfg.UpdateConnection("web", conn) {
					t.Fatal("rename failed")
				}
			} else {
				if !cfg.DeleteConnection("web") {
					t.Fatal("delete failed")
				}
				want["production"] = []string{"db"}
				want["web-only"] = []string{}
			}
			path := filepath.Join(t.TempDir(), "config.yaml")
			if err := cfg.Save(path); err != nil {
				t.Fatal(err)
			}
			reloaded, err := Load(path)
			if err != nil {
				t.Fatal(err)
			}
			if err := reloaded.Validate(); err != nil {
				t.Fatalf("mutation left an invalid config: %v", err)
			}
			if !reflect.DeepEqual(reloaded.Groups, want) {
				t.Fatalf("groups = %v, want %v", reloaded.Groups, want)
			}
		})
	}
}

func TestUpdateConnectionRejectsDuplicateID(t *testing.T) {
	cfg := &Config{
		Connections: []Connection{{ID: "web", Host: "web.example.com"}, {ID: "db", Host: "db.example.com"}},
		Groups:      map[string][]string{"production": {"web", "db"}},
	}
	if cfg.UpdateConnection("web", Connection{ID: "db", Host: "changed.example.com"}) {
		t.Fatal("rename to an existing ID succeeded")
	}
	if cfg.FindConnection("web").Host != "web.example.com" || cfg.FindConnection("db").Host != "db.example.com" {
		t.Fatal("rejected rename changed connections")
	}
	if !reflect.DeepEqual(cfg.Groups["production"], []string{"web", "db"}) {
		t.Fatal("rejected rename changed group members")
	}
}
