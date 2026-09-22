package tui

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/orekasep/go-hosts-cli/internal/apply"
	"github.com/orekasep/go-hosts-cli/internal/core/config"
	"github.com/orekasep/go-hosts-cli/internal/core/migrate"
	"github.com/orekasep/go-hosts-cli/internal/domain"
)

type applyFinishedMsg struct {
	err error
	msg string
}

type rollbackFinishedMsg struct {
	err error
	msg string
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case applyFinishedMsg:
		if msg.err != nil {
			m.statusMsg = fmt.Sprintf("Apply error: %v", msg.err)
			m.statusColor = "#FF4672"
		} else {
			m.statusMsg = "✓ System hosts updated successfully!"
			m.statusColor = "#04B575"
		}
		return m, nil

	case rollbackFinishedMsg:
		if msg.err != nil {
			m.statusMsg = fmt.Sprintf("Rollback error: %v", msg.err)
			m.statusColor = "#FF4672"
		} else {
			m.statusMsg = "✓ System hosts rolled back successfully!"
			m.statusColor = "#04B575"
		}
		return m, nil

	case tea.KeyMsg:
		// When modal is active, delegate key handling to modal handler
		if m.modal != ModalNone {
			return m.handleModalKey(msg)
		}

		// Normal navigation key handling
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "?":
			m.modal = ModalHelp
			return m, nil

		case "/":
			m.modal = ModalSearch
			m.searchInput.SetValue(m.searchQuery)
			m.searchInput.Focus()
			return m, textinput.Blink

		case "tab":
			if m.focus == FocusSidebar {
				m.focus = FocusTable
			} else {
				m.focus = FocusSidebar
			}
			return m, nil

		case "h", "left":
			m.focus = FocusSidebar
			return m, nil

		case "l", "right":
			m.focus = FocusTable
			return m, nil

		case "j", "down":
			if m.focus == FocusSidebar {
				if len(m.groupsList) > 0 {
					m.selectedGrp = (m.selectedGrp + 1) % len(m.groupsList)
					m.selectedRow = 0
				}
			} else {
				visible := m.GetVisibleEntries()
				if len(visible) > 0 && m.selectedRow < len(visible)-1 {
					m.selectedRow++
				}
			}
			return m, nil

		case "k", "up":
			if m.focus == FocusSidebar {
				if len(m.groupsList) > 0 {
					m.selectedGrp = (m.selectedGrp - 1 + len(m.groupsList)) % len(m.groupsList)
					m.selectedRow = 0
				}
			} else {
				if m.selectedRow > 0 {
					m.selectedRow--
				}
			}
			return m, nil

		case " ":
			// Space toggles enabled on selected row
			visible := m.GetVisibleEntries()
			if len(visible) > 0 && m.selectedRow < len(visible) {
				selected := visible[m.selectedRow]
				// Find and toggle in underlying model
				for i := range m.hostsFile.Groups {
					for j := range m.hostsFile.Groups[i].Entries {
						if m.hostsFile.Groups[i].Entries[j].ID == selected.ID {
							m.hostsFile.Groups[i].Entries[j].Enabled = !m.hostsFile.Groups[i].Entries[j].Enabled
							_ = m.SaveLocalConfig()
							stateStr := "enabled"
							if !m.hostsFile.Groups[i].Entries[j].Enabled {
								stateStr = "disabled"
							}
							m.statusMsg = fmt.Sprintf("Entry '%s' %s. Local YAML updated.", selected.Hostname, stateStr)
							m.statusColor = "#FFAF00"
							return m, nil
						}
					}
				}
			}
			return m, nil

		case "a":
			m.openAddForm()
			return m, textinput.Blink

		case "e", "enter":
			if m.focus == FocusTable {
				visible := m.GetVisibleEntries()
				if len(visible) > 0 && m.selectedRow < len(visible) {
					m.openEditForm(visible[m.selectedRow])
					return m, textinput.Blink
				}
			}
			return m, nil

		case "d":
			visible := m.GetVisibleEntries()
			if len(visible) > 0 && m.selectedRow < len(visible) {
				m.modal = ModalDelete
			}
			return m, nil

		case "A":
			// Apply changes
			return m.triggerApply()

		case "R":
			// Rollback modal
			m.modal = ModalRollback
			return m, nil

		case "m":
			// Migrate/Import from system hosts
			return m.triggerMigrate()
		}
	}

	return m, nil
}

func (m *Model) triggerApply() (Model, tea.Cmd) {
	// If direct write access is available (or running as root)
	if m.privStatus.HasDirectWrite || m.privStatus.IsElevated {
		res, err := apply.Apply(m.hostsFile, false)
		if err != nil {
			m.statusMsg = fmt.Sprintf("Apply failed: %v", err)
			m.statusColor = "#FF4672"
		} else {
			m.statusMsg = fmt.Sprintf("✓ %s", res.Message)
			m.statusColor = "#04B575"
		}
		return *m, nil
	}

	// Sudo elevation via tea.ExecProcess
	if m.privStatus.CanSudo {
		exe, err := os.Executable()
		if err != nil {
			exe = "hostcli"
		}
		c := exec.Command("sudo", exe, "apply")
		c.Stdin = os.Stdin
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr

		return *m, tea.ExecProcess(c, func(err error) tea.Msg {
			return applyFinishedMsg{err: err}
		})
	}

	m.statusMsg = "Cannot write to system hosts: sudo not found and no write permission."
	m.statusColor = "#FF4672"
	return *m, nil
}

func (m *Model) triggerMigrate() (Model, tea.Cmd) {
	sysHostsPath := config.GetSystemHostsPath()
	content, err := os.ReadFile(sysHostsPath)
	if err != nil {
		m.statusMsg = fmt.Sprintf("Import failed: %v", err)
		m.statusColor = "#FF4672"
		return *m, nil
	}

	parsed, err := migrate.ParseHostsFile(content)
	if err != nil {
		m.statusMsg = fmt.Sprintf("Parsing error: %v", err)
		m.statusColor = "#FF4672"
		return *m, nil
	}

	m.hostsFile = parsed
	_ = m.SaveLocalConfig()
	m.refreshGroups()
	m.selectedRow = 0
	m.statusMsg = fmt.Sprintf("✓ Successfully imported %d entries from %s", parsed.TotalEntries(), sysHostsPath)
	m.statusColor = "#04B575"
	return *m, nil
}

func (m *Model) openAddForm() {
	currentGroup := "default"
	if m.selectedGrp > 0 && m.selectedGrp < len(m.groupsList) {
		currentGroup = m.groupsList[m.selectedGrp]
	}

	ipIn := textinput.New()
	ipIn.Placeholder = "192.168.1.10 or ::1"
	ipIn.Focus()

	hostIn := textinput.New()
	hostIn.Placeholder = "example.local"

	aliasIn := textinput.New()
	aliasIn.Placeholder = "alias1, alias2"

	grpIn := textinput.New()
	grpIn.SetValue(currentGroup)

	commentIn := textinput.New()
	commentIn.Placeholder = "Optional description"

	m.editForm = EditForm{
		IsNew:        true,
		IPInput:      ipIn,
		HostInput:    hostIn,
		AliasesInput: aliasIn,
		GroupInput:   grpIn,
		CommentInput: commentIn,
		Enabled:      true,
		FocusIndex:   FieldIP,
	}
	m.modal = ModalEdit
}

func (m *Model) openEditForm(entry domain.HostEntry) {
	_, grpName := m.hostsFile.FindEntry(entry.Hostname)

	ipIn := textinput.New()
	ipIn.SetValue(entry.IP)
	ipIn.Focus()

	hostIn := textinput.New()
	hostIn.SetValue(entry.Hostname)

	aliasIn := textinput.New()
	aliasIn.SetValue(strings.Join(entry.Aliases, ", "))

	grpIn := textinput.New()
	grpIn.SetValue(grpName)

	commentIn := textinput.New()
	commentIn.SetValue(entry.Comment)

	m.editForm = EditForm{
		IsNew:        false,
		EntryID:      entry.ID,
		OriginalName: entry.Hostname,
		IPInput:      ipIn,
		HostInput:    hostIn,
		AliasesInput: aliasIn,
		GroupInput:   grpIn,
		CommentInput: commentIn,
		Enabled:      entry.Enabled,
		FocusIndex:   FieldIP,
	}
	m.modal = ModalEdit
}
