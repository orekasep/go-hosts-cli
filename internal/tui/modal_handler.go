package tui

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/orekasep/go-hosts-cli/internal/apply"
	"github.com/orekasep/go-hosts-cli/internal/domain"
)

func (m *Model) handleModalKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.modal {
	case ModalNotice, ModalHelp:
		switch msg.String() {
		case "esc", "enter", "q", "?":
			m.modal = ModalNone
			return *m, nil
		}
		return *m, nil

	case ModalSearch:
		switch msg.String() {
		case "enter":
			m.searchQuery = strings.TrimSpace(m.searchInput.Value())
			m.modal = ModalNone
			m.selectedRow = 0
			return *m, nil
		case "esc":
			m.modal = ModalNone
			return *m, nil
		default:
			var cmd tea.Cmd
			m.searchInput, cmd = m.searchInput.Update(msg)
			return *m, cmd
		}

	case ModalDelete:
		switch msg.String() {
		case "y", "Y", "enter":
			visible := m.GetVisibleEntries()
			if len(visible) > 0 && m.selectedRow < len(visible) {
				toDelete := visible[m.selectedRow]
				// Remove from underlying group
				for i := range m.hostsFile.Groups {
					var filtered []domain.HostEntry
					for _, e := range m.hostsFile.Groups[i].Entries {
						if e.ID != toDelete.ID {
							filtered = append(filtered, e)
						}
					}
					m.hostsFile.Groups[i].Entries = filtered
				}
				_ = m.SaveLocalConfig()
				m.refreshGroups()
				if m.selectedRow > 0 {
					m.selectedRow--
				}
				m.statusMsg = fmt.Sprintf("Removed host '%s'. Changes saved locally.", toDelete.Hostname)
				m.statusColor = "#FFAF00"
			}
			m.modal = ModalNone
			return *m, nil
		case "n", "N", "esc", "q":
			m.modal = ModalNone
			return *m, nil
		}
		return *m, nil

	case ModalRollback:
		switch msg.String() {
		case "1", "s":
			m.modal = ModalNone
			return m.triggerRollback(false)
		case "2", "o":
			m.modal = ModalNone
			return m.triggerRollback(true)
		case "esc", "q":
			m.modal = ModalNone
			return *m, nil
		}
		return *m, nil

	case ModalEdit:
		return m.handleEditFormKey(msg)
	}

	return *m, nil
}

func (m *Model) triggerRollback(restoreOriginal bool) (Model, tea.Cmd) {
	if m.privStatus.HasDirectWrite || m.privStatus.IsElevated {
		res, err := apply.Rollback(restoreOriginal, false)
		if err != nil {
			m.statusMsg = fmt.Sprintf("Rollback failed: %v", err)
			m.statusColor = "#FF4672"
		} else {
			m.statusMsg = fmt.Sprintf("✓ %s", res.Message)
			m.statusColor = "#04B575"
		}
		return *m, nil
	}

	if m.privStatus.CanSudo {
		exe, err := os.Executable()
		if err != nil {
			exe = "hostcli"
		}
		args := []string{"rollback"}
		if restoreOriginal {
			args = append(args, "--original")
		}
		c := exec.Command("sudo", append([]string{exe}, args...)...)
		c.Stdin = os.Stdin
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr

		return *m, tea.ExecProcess(c, func(err error) tea.Msg {
			return rollbackFinishedMsg{err: err}
		})
	}

	m.statusMsg = "Rollback failed: sudo required but unavailable."
	m.statusColor = "#FF4672"
	return *m, nil
}

func (m *Model) handleEditFormKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.modal = ModalNone
		return *m, nil

	case "tab", "down":
		m.editForm.FocusIndex = (m.editForm.FocusIndex + 1) % 8
		m.updateFormFocus()
		return *m, nil

	case "shift+tab", "up":
		m.editForm.FocusIndex = (m.editForm.FocusIndex - 1 + 8) % 8
		m.updateFormFocus()
		return *m, nil

	case " ":
		if m.editForm.FocusIndex == FieldEnabled {
			m.editForm.Enabled = !m.editForm.Enabled
			return *m, nil
		}

	case "enter":
		if m.editForm.FocusIndex == FieldCancel {
			m.modal = ModalNone
			return *m, nil
		}
		if m.editForm.FocusIndex == FieldSave || m.editForm.FocusIndex == FieldComment {
			return m.commitEditForm()
		}
	}

	// Update currently active text input
	var cmd tea.Cmd
	switch m.editForm.FocusIndex {
	case FieldIP:
		m.editForm.IPInput, cmd = m.editForm.IPInput.Update(msg)
	case FieldHostname:
		m.editForm.HostInput, cmd = m.editForm.HostInput.Update(msg)
	case FieldAliases:
		m.editForm.AliasesInput, cmd = m.editForm.AliasesInput.Update(msg)
	case FieldGroup:
		m.editForm.GroupInput, cmd = m.editForm.GroupInput.Update(msg)
	case FieldComment:
		m.editForm.CommentInput, cmd = m.editForm.CommentInput.Update(msg)
	}

	return *m, cmd
}

func (m *Model) updateFormFocus() {
	m.editForm.IPInput.Blur()
	m.editForm.HostInput.Blur()
	m.editForm.AliasesInput.Blur()
	m.editForm.GroupInput.Blur()
	m.editForm.CommentInput.Blur()

	switch m.editForm.FocusIndex {
	case FieldIP:
		m.editForm.IPInput.Focus()
	case FieldHostname:
		m.editForm.HostInput.Focus()
	case FieldAliases:
		m.editForm.AliasesInput.Focus()
	case FieldGroup:
		m.editForm.GroupInput.Focus()
	case FieldComment:
		m.editForm.CommentInput.Focus()
	}
}

func (m *Model) commitEditForm() (tea.Model, tea.Cmd) {
	ip := strings.TrimSpace(m.editForm.IPInput.Value())
	host := strings.TrimSpace(m.editForm.HostInput.Value())
	groupName := strings.TrimSpace(m.editForm.GroupInput.Value())
	comment := strings.TrimSpace(m.editForm.CommentInput.Value())

	if groupName == "" {
		groupName = "default"
	}

	// Parse aliases
	rawAliases := strings.Split(m.editForm.AliasesInput.Value(), ",")
	var aliases []string
	for _, a := range rawAliases {
		trimmed := strings.TrimSpace(a)
		if trimmed != "" {
			aliases = append(aliases, trimmed)
		}
	}

	entry := domain.HostEntry{
		ID:       m.editForm.EntryID,
		IP:       ip,
		Hostname: host,
		Aliases:  aliases,
		Enabled:  m.editForm.Enabled,
		Comment:  comment,
	}

	if err := domain.ValidateEntry(&entry); err != nil {
		m.editForm.ErrorMsg = err.Error()
		return *m, nil
	}

	// Remove old entry if editing
	if !m.editForm.IsNew {
		for i := range m.hostsFile.Groups {
			var filtered []domain.HostEntry
			for _, e := range m.hostsFile.Groups[i].Entries {
				if e.ID != entry.ID {
					filtered = append(filtered, e)
				}
			}
			m.hostsFile.Groups[i].Entries = filtered
		}
	} else {
		entry.ID = domain.NewULID()
	}

	// Place into designated group
	targetGrp := m.hostsFile.GetOrCreateGroup(groupName)
	targetGrp.Entries = append(targetGrp.Entries, entry)

	_ = m.SaveLocalConfig()
	m.refreshGroups()
	m.modal = ModalNone
	m.statusMsg = fmt.Sprintf("Saved host entry '%s'. Remember to press [A] to apply to system.", entry.Hostname)
	m.statusColor = "#04B575"

	return *m, nil
}
