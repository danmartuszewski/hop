package picker

import (
	"testing"

	"github.com/danmartuszewski/hop/internal/config"
	"github.com/danmartuszewski/hop/internal/fuzzy"
)

func TestSelectConnection_SingleMatch(t *testing.T) {
	matches := []fuzzy.Match{
		{
			Connection: &config.Connection{
				ID:   "server1",
				Host: "server1.example.com",
				User: "admin",
			},
			Score: 100,
		},
	}

	conn, err := SelectConnection(matches)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if conn.ID != "server1" {
		t.Errorf("expected ID 'server1', got '%s'", conn.ID)
	}
}

func TestSelectConnection_EmptyMatches(t *testing.T) {
	matches := []fuzzy.Match{}

	_, err := SelectConnection(matches)
	if err == nil {
		t.Fatal("expected error for empty matches")
	}

	expected := "no matches to select from"
	if err.Error() != expected {
		t.Errorf("expected error '%s', got '%s'", expected, err.Error())
	}
}
