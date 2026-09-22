package tui

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/orekasep/go-hosts-cli/internal/domain"
)

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
			t.Errorf("Screen height overflow for %dx%d: rendered height is %d", sz.w, sz.h, renderedH)
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
			t.Errorf("Modal height overflow for %dx%d: rendered height is %d", sz.w, sz.h, modalH)
		}
	}
}
