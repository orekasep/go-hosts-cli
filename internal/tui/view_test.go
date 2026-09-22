package tui

import (
	"fmt"
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/orekasep/go-hosts-cli/internal/domain"
)

func TestPrintViewLines(t *testing.T) {
	m := InitialModel("dummy.yaml")
	m.hostsFile = &domain.HostsFile{
		Version: 1,
		Groups: []domain.Group{
			{
				Name: "work",
				Entries: []domain.HostEntry{
					{ID: "1", IP: "192.168.1.10", Hostname: "api.local", Enabled: true, Comment: "Gateway"},
					{ID: "2", IP: "10.0.0.1", Hostname: "db.local", Enabled: false, Comment: "Database"},
				},
			},
		},
	}
	m.refreshGroups()
	m.width = 80
	m.height = 24

	rendered := m.View()
	lines := strings.Split(rendered, "\n")
	fmt.Printf("Total lines: %d (m.height: %d)\n", len(lines), m.height)
	for i, l := range lines {
		w := lipgloss.Width(l)
		fmt.Printf("Line %2d (w=%2d): %s\n", i+1, w, l)
		if w > m.width {
			t.Errorf("Line %d exceeds width %d (got %d): %s", i+1, m.width, w, l)
		}
	}
	if len(lines) > m.height || len(lines) < m.height-1 {
		t.Errorf("expected %d or %d lines, got %d", m.height-1, m.height, len(lines))
	}

	m.openAddForm()
	modalRendered := m.View()
	modalLines := strings.Split(modalRendered, "\n")
	fmt.Printf("\n--- MODAL ---\nTotal modal lines: %d (expected %d)\n", len(modalLines), m.height)
	for i, l := range modalLines {
		w := lipgloss.Width(l)
		if w > m.width {
			t.Errorf("Modal line %d exceeds width %d (got %d): %s", i+1, m.width, w, l)
		}
	}
	if len(modalLines) > m.height {
		t.Errorf("expected modal lines <= %d, got %d", m.height, len(modalLines))
	}
}

func TestViewDimensionsOnWindowResize(t *testing.T) {
	sizes := []struct {
		w, h int
	}{
		{80, 24},
		{100, 30},
		{120, 40},
		{65, 16},
	}

	for _, sz := range sizes {
		m := InitialModel("dummy.yaml")
		m.hostsFile = &domain.HostsFile{
			Version: 1,
			Groups: []domain.Group{
				{
					Name: "work",
					Entries: []domain.HostEntry{
						{ID: "1", IP: "192.168.1.10", Hostname: "api.local", Enabled: true, Comment: "Gateway"},
						{ID: "2", IP: "10.0.0.1", Hostname: "db.local", Enabled: false, Comment: "Database"},
					},
				},
			},
		}
		m.refreshGroups()
		m.width = sz.w
		m.height = sz.h

		// Test Main Screen View
		rendered := m.View()
		renderedW := lipgloss.Width(rendered)
		renderedH := lipgloss.Height(rendered)

		if renderedW > sz.w {
			t.Errorf("Screen width overflow for %dx%d: rendered width is %d", sz.w, sz.h, renderedW)
		}
		if renderedH > sz.h {
			t.Errorf("Screen height overflow for %dx%d: expected <= %d, got %d", sz.w, sz.h, sz.h, renderedH)
		}

		// Test Edit Modal View
		m.openAddForm()
		modalRendered := m.View()
		modalW := lipgloss.Width(modalRendered)
		modalH := lipgloss.Height(modalRendered)

		if modalW > sz.w {
			t.Errorf("Modal width overflow for %dx%d: rendered width is %d", sz.w, sz.h, modalW)
		}
		if modalH > sz.h {
			t.Errorf("Modal height overflow for %dx%d: expected <= %d, got %d", sz.w, sz.h, sz.h, modalH)
		}
	}
}
