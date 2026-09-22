package migrate

import (
	"testing"
)

func TestParseHostsFile(t *testing.T) {
	input := []byte(`
# System loopback entries
127.0.0.1 localhost
::1 localhost

# [work/dev]
192.168.1.50 api.dev.local api gateway # API Gateway service
# 192.168.1.51 db.dev.local # Disabled DB

# Personal Labs
10.0.0.5 lab.internal
`)

	hf, err := ParseHostsFile(input)
	if err != nil {
		t.Fatalf("ParseHostsFile returned error: %v", err)
	}

	if hf.TotalEntries() != 5 {
		t.Fatalf("expected 5 total entries, got %d", hf.TotalEntries())
	}

	// Verify loopback auto-grouped to system
	entry, grp := hf.FindEntry("localhost")
	if entry == nil {
		t.Fatalf("localhost entry not found")
	}
	if grp != "system" {
		t.Errorf("expected localhost to be in 'system' group, got %q", grp)
	}

	// Verify api.dev.local in work/dev
	apiEntry, apiGrp := hf.FindEntry("api.dev.local")
	if apiEntry == nil {
		t.Fatalf("api.dev.local not found")
	}
	if apiGrp != "work/dev" {
		t.Errorf("expected 'work/dev' group, got %q", apiGrp)
	}
	if len(apiEntry.Aliases) != 2 || apiEntry.Aliases[0] != "api" || apiEntry.Aliases[1] != "gateway" {
		t.Errorf("unexpected aliases: %v", apiEntry.Aliases)
	}
	if !apiEntry.Enabled {
		t.Errorf("expected api.dev.local to be enabled")
	}

	// Verify disabled entry detected from comment
	dbEntry, _ := hf.FindEntry("db.dev.local")
	if dbEntry == nil {
		t.Fatalf("db.dev.local not found")
	}
	if dbEntry.Enabled {
		t.Errorf("expected db.dev.local to be disabled")
	}
}
