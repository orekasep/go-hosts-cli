package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) View() string {
	if m.width == 0 {
		return "Initializing hostcli..."
	}

	// 1. Render Top Header
	header := m.renderHeader()

	// 2. Render Main Body (Sidebar + Table)
	bodyHeight := m.height - 5
	if bodyHeight < 10 {
		bodyHeight = 10
	}

	sidebarWidth := 28
	tableWidth := m.width - sidebarWidth - 4
	if tableWidth < 40 {
		tableWidth = 40
	}

	sidebarView := m.renderSidebar(sidebarWidth, bodyHeight)
	tableView := m.renderTable(tableWidth, bodyHeight)

	mainBody := lipgloss.JoinHorizontal(lipgloss.Top, sidebarView, tableView)

	// 3. Render Status & Footer
	footer := m.renderFooter()

	baseView := lipgloss.JoinVertical(lipgloss.Left, header, mainBody, footer)

	// 4. If a modal is open, overlay it on top of baseView
	if m.modal != ModalNone {
		return m.renderModalOverlay(baseView)
	}

	return baseView
}

func (m Model) renderHeader() string {
	title := HeaderStyle.Render(" hostcli v1.0.0 ")

	var privBadge string
	if m.privStatus.HasDirectWrite || m.privStatus.IsElevated {
		privBadge = BadgeSuccessStyle.Render("[✓ Direct Write Access]")
	} else if m.privStatus.CanSudo {
		privBadge = BadgeWarningStyle.Render("[🔒 Unprivileged: sudo prompted on Apply]")
	} else {
		privBadge = BadgeDangerStyle.Render("[⚠️ Sudo not found: Read-Only Mode]")
	}

	totalEntries := m.hostsFile.TotalEntries()
	enabledEntries := m.hostsFile.EnabledEntries()
	stats := fmt.Sprintf("Hosts: %d total, %d active", totalEntries, enabledEntries)

	gap := m.width - lipgloss.Width(title) - lipgloss.Width(privBadge) - lipgloss.Width(stats) - 4
	if gap < 2 {
		gap = 2
	}

	return lipgloss.JoinHorizontal(
		lipgloss.Center,
		title,
		"  ",
		privBadge,
		strings.Repeat(" ", gap),
		stats,
	) + "\n"
}

func (m Model) renderSidebar(width, height int) string {
	var sb strings.Builder

	sb.WriteString(lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary).Render("📁 GROUPS"))
	sb.WriteString("\n")

	for idx, grpName := range m.groupsList {
		count := 0
		if idx == 0 {
			count = m.hostsFile.TotalEntries()
		} else {
			for _, g := range m.hostsFile.Groups {
				if g.Name == grpName {
					count = len(g.Entries)
					break
				}
			}
		}

		line := fmt.Sprintf("%-16s (%d)", truncate(grpName, 15), count)
		if idx == m.selectedGrp {
			if m.focus == FocusSidebar {
				sb.WriteString(SelectedGroupStyle.Render("▸ " + line))
			} else {
				sb.WriteString(NormalGroupStyle.Render("▸ " + line))
			}
		} else {
			sb.WriteString(NormalGroupStyle.Render("  " + line))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("\n" + lipgloss.NewStyle().Foreground(ColorMuted).Render(strings.Repeat("─", width-4)) + "\n")
	sb.WriteString(lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary).Render("⚡ QUICK ACTIONS") + "\n")
	sb.WriteString(fmt.Sprintf("%s Add Host\n", KeyBadgeStyle.Render("[a]")))
	sb.WriteString(fmt.Sprintf("%s Edit Selected\n", KeyBadgeStyle.Render("[e]")))
	sb.WriteString(fmt.Sprintf("%s Remove Host\n", KeyBadgeStyle.Render("[d]")))
	sb.WriteString(fmt.Sprintf("%s Toggle Active\n", KeyBadgeStyle.Render("[Space]")))
	sb.WriteString(fmt.Sprintf("%s Search / Filter\n", KeyBadgeStyle.Render("[/]")))

	sb.WriteString("\n" + lipgloss.NewStyle().Foreground(ColorMuted).Render(strings.Repeat("─", width-4)) + "\n")
	sb.WriteString(lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary).Render("🛡️ SYSTEM SYNC") + "\n")
	sb.WriteString(fmt.Sprintf("%s Apply to /etc/hosts\n", KeyBadgeStyle.Render("[A]")))
	sb.WriteString(fmt.Sprintf("%s Rollback Original\n", KeyBadgeStyle.Render("[R]")))
	sb.WriteString(fmt.Sprintf("%s Import System Hosts\n", KeyBadgeStyle.Render("[m]")))

	style := SidebarUnfocusedStyle
	if m.focus == FocusSidebar {
		style = SidebarStyle
	}

	return style.Width(width).Height(height).Render(sb.String())
}

func (m Model) renderTable(width, height int) string {
	var sb strings.Builder

	currentGrpName := "[All Entries]"
	if m.selectedGrp < len(m.groupsList) {
		currentGrpName = m.groupsList[m.selectedGrp]
	}

	titleText := fmt.Sprintf("HOST ENTRIES — Group: %s", currentGrpName)
	if m.searchQuery != "" {
		titleText += fmt.Sprintf(" (Filter: '%s')", m.searchQuery)
	}

	sb.WriteString(lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary).Render(titleText) + "\n\n")

	// Table Header
	headerLine := fmt.Sprintf("  %-7s  %-16s  %-22s  %-15s  %s",
		"STATUS", "IP ADDRESS", "HOSTNAME", "ALIASES", "COMMENT")
	sb.WriteString(TableHeaderStyle.Width(width - 4).Render(headerLine) + "\n")

	visible := m.GetVisibleEntries()
	if len(visible) == 0 {
		sb.WriteString("\n" + NormalRowStyle.Render("   (No entries found. Press [a] to add a new host entry.)\n"))
	} else {
		for idx, e := range visible {
			statusBadge := BadgeDangerStyle.Render("○ OFF")
			if e.Enabled {
				statusBadge = BadgeSuccessStyle.Render("● ON ")
			}

			aliasStr := strings.Join(e.Aliases, ",")
			if len(aliasStr) == 0 {
				aliasStr = "-"
			}

			rowContent := fmt.Sprintf("  %s  %-16s  %-22s  %-15s  %s",
				statusBadge,
				truncate(e.IP, 16),
				truncate(e.Hostname, 22),
				truncate(aliasStr, 15),
				truncate(e.Comment, 25),
			)

			if idx == m.selectedRow {
				indicator := "▸ "
				if m.focus != FocusTable {
					indicator = "  "
				}
				sb.WriteString(SelectedRowStyle.Width(width - 4).Render(indicator+rowContent[2:]) + "\n")
			} else {
				sb.WriteString(NormalRowStyle.Width(width - 4).Render(rowContent) + "\n")
			}
		}
	}

	style := TableUnfocusedContainerStyle
	if m.focus == FocusTable {
		style = TableContainerStyle
	}

	return style.Width(width).Height(height).Render(sb.String())
}

func (m Model) renderFooter() string {
	statusStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(m.statusColor)).
		Bold(true)

	statusLine := statusStyle.Render(m.statusMsg)
	keysLine := lipgloss.NewStyle().Foreground(lipgloss.Color("#AAAAAA")).
		Render("Tab: Switch Pane │ j/k: Navigate │ Space: Toggle │ a: Add │ e: Edit │ d: Del │ A: Apply │ R: Rollback │ q: Quit │ ?: Help")

	return "\n" + statusLine + "\n" + StatusbarStyle.Width(m.width).Render(keysLine)
}

func (m Model) renderModalOverlay(baseView string) string {
	var modalContent string

	switch m.modal {
	case ModalNotice:
		modalContent = ModalBoxStyle.Width(60).Render(
			ModalTitleStyle.Render(" ℹ️ Permission Notice ") + "\n\n" +
				m.noticeMsg + "\n\n" +
				lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary).Render("[ Press Enter or Esc to Continue ]"),
		)

	case ModalHelp:
		modalContent = ModalBoxStyle.Width(62).Render(
			ModalTitleStyle.Render(" 📖 hostcli Keybindings ") + "\n\n" +
				"  Tab / h / l   Switch between Groups and Host Table\n" +
				"  j / k / ↑ / ↓ Move cursor selection\n" +
				"  Space         Toggle host enabled / disabled\n" +
				"  a             Add new host entry\n" +
				"  e / Enter     Edit selected host entry\n" +
				"  d             Delete selected host entry\n" +
				"  /             Fuzzy search & filter hosts\n" +
				"  A             Apply enabled entries to system /etc/hosts\n" +
				"  R             Rollback / restore system /etc/hosts\n" +
				"  m             Import current system hosts to hostcli\n" +
				"  ?             Toggle this help modal\n" +
				"  q / Ctrl+C    Quit hostcli\n\n" +
				lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary).Render("[ Press Esc or ? to Close ]"),
		)

	case ModalSearch:
		modalContent = ModalBoxStyle.Width(50).Render(
			ModalTitleStyle.Render(" 🔍 Search & Filter ") + "\n\n" +
				m.searchInput.View() + "\n\n" +
				lipgloss.NewStyle().Foreground(ColorMuted).Render("Enter: Apply Search  │  Esc: Clear & Close"),
		)

	case ModalDelete:
		visible := m.GetVisibleEntries()
		hostName := "selected host"
		if len(visible) > 0 && m.selectedRow < len(visible) {
			hostName = visible[m.selectedRow].Hostname
		}
		modalContent = ModalBoxStyle.Width(50).Render(
			ModalTitleStyle.Render(" ⚠️ Confirm Delete ") + "\n\n" +
				fmt.Sprintf("Are you sure you want to delete '%s'?\n\n", hostName) +
				lipgloss.NewStyle().Foreground(ColorDanger).Bold(true).Render("[y] Yes, Delete") +
				"    " +
				lipgloss.NewStyle().Foreground(ColorWhite).Render("[n / Esc] Cancel"),
		)

	case ModalRollback:
		modalContent = ModalBoxStyle.Width(55).Render(
			ModalTitleStyle.Render(" 🛡️ System Hosts Rollback ") + "\n\n" +
				"Choose a rollback option:\n\n" +
				"  [1] Strip Managed Block\n" +
				"      (Removes # BEGIN HOSTCLI ... # END HOSTCLI lines)\n\n" +
				"  [2] Restore Pristine Original Backup\n" +
				"      (Restores pre-migration hosts.original.bak)\n\n" +
				lipgloss.NewStyle().Foreground(ColorMuted).Render("Press 1, 2, or Esc to Cancel"),
		)

	case ModalEdit:
		modalContent = m.renderEditModal()
	}

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, modalContent)
}

func (m Model) renderEditModal() string {
	form := m.editForm
	title := " ✏️ Edit Host Entry "
	if form.IsNew {
		title = " ➕ Add New Host Entry "
	}

	var sb strings.Builder
	sb.WriteString(ModalTitleStyle.Render(title) + "\n\n")

	if form.ErrorMsg != "" {
		sb.WriteString(lipgloss.NewStyle().Foreground(ColorDanger).Bold(true).Render("Error: "+form.ErrorMsg) + "\n\n")
	}

	fields := []struct {
		Label string
		View  string
		Index FormField
	}{
		{"IP Address", form.IPInput.View(), FieldIP},
		{"Hostname", form.HostInput.View(), FieldHostname},
		{"Aliases (comma-separated)", form.AliasesInput.View(), FieldAliases},
		{"Group", form.GroupInput.View(), FieldGroup},
		{"Comment", form.CommentInput.View(), FieldComment},
	}

	for _, f := range fields {
		lblStyle := lipgloss.NewStyle().Foreground(ColorMuted)
		if form.FocusIndex == f.Index {
			lblStyle = lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary)
		}
		sb.WriteString(lblStyle.Render(f.Label) + "\n")
		sb.WriteString(f.View + "\n\n")
	}

	// Enabled Checkbox
	enableBox := "[ ] Disabled"
	if form.Enabled {
		enableBox = "[x] Enabled"
	}
	enableStyle := lipgloss.NewStyle().Foreground(ColorMuted)
	if form.FocusIndex == FieldEnabled {
		enableStyle = lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary)
	}
	sb.WriteString(enableStyle.Render(enableBox+" (Space to toggle)") + "\n\n")

	// Buttons
	saveBtn := "[ Save (Enter) ]"
	cancelBtn := "[ Cancel (Esc) ]"

	if form.FocusIndex == FieldSave {
		saveBtn = lipgloss.NewStyle().Bold(true).Background(ColorPrimary).Foreground(ColorWhite).Render(saveBtn)
	}
	if form.FocusIndex == FieldCancel {
		cancelBtn = lipgloss.NewStyle().Bold(true).Background(ColorHighlight).Foreground(ColorWhite).Render(cancelBtn)
	}

	sb.WriteString(fmt.Sprintf("%s    %s\n", saveBtn, cancelBtn))

	return ModalBoxStyle.Width(54).Render(sb.String())
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}
