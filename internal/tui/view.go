package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) View() string {
	if m.width <= 0 || m.height <= 0 {
		return "Initializing hostcli..."
	}

	if m.width < 50 || m.height < 10 {
		return fmt.Sprintf("Terminal window too small (%dx%d).\nPlease resize to at least 60x12.", m.width, m.height)
	}

	// 1. Header (fixed 1 line)
	headerView := m.renderHeader()
	headerHeight := lipgloss.Height(headerView)

	// 2. Footer (fixed 2 lines: 1 status line + 1 keybindings line)
	footerView := m.renderFooter()
	footerHeight := lipgloss.Height(footerView)

	// 3. Body Dimensions
	bodyHeight := m.height - headerHeight - footerHeight
	if bodyHeight < 6 {
		bodyHeight = 6
	}

	// Responsive sidebar width: 24 to 28 cols based on screen width
	sidebarWidth := 26
	if m.width < 80 {
		sidebarWidth = 22
	}
	tableWidth := m.width - sidebarWidth
	if tableWidth < 28 {
		tableWidth = 28
	}

	sidebarView := m.renderSidebar(sidebarWidth, bodyHeight)
	tableView := m.renderTable(tableWidth, bodyHeight)

	mainBody := lipgloss.JoinHorizontal(lipgloss.Top, sidebarView, tableView)

	baseView := lipgloss.JoinVertical(lipgloss.Left, headerView, mainBody, footerView)

	// 4. Modal Dialog Overlay
	if m.modal != ModalNone {
		return m.renderModalOverlay()
	}

	return baseView
}

func (m Model) renderHeader() string {
	title := HeaderStyle.Render(" hostcli v1.0.0 ")

	var privBadge string
	if m.privStatus.HasDirectWrite || m.privStatus.IsElevated {
		privBadge = BadgeSuccessStyle.Render("[✓ Direct Write]")
	} else if m.privStatus.CanSudo {
		privBadge = BadgeWarningStyle.Render("[🔒 sudo on Apply]")
	} else {
		privBadge = BadgeDangerStyle.Render("[⚠️ Read-Only]")
	}

	totalEntries := m.hostsFile.TotalEntries()
	enabledEntries := m.hostsFile.EnabledEntries()
	stats := fmt.Sprintf("Hosts: %d total, %d active", totalEntries, enabledEntries)

	usedWidth := lipgloss.Width(title) + 2 + lipgloss.Width(privBadge) + lipgloss.Width(stats)
	gap := m.width - usedWidth
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
	)
}

func (m Model) renderSidebar(width, height int) string {
	var sb strings.Builder

	// Content width inside borders & padding
	innerW := width - 4
	if innerW < 10 {
		innerW = 10
	}
	innerH := height - 2
	if innerH < 4 {
		innerH = 4
	}

	sb.WriteString(lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary).Render("📁 GROUPS") + "\n")

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

		maxNameLen := innerW - 6
		if maxNameLen < 4 {
			maxNameLen = 4
		}
		line := fmt.Sprintf("%-*s (%d)", maxNameLen, truncate(grpName, maxNameLen), count)
		if idx == m.selectedGrp {
			if m.focus == FocusSidebar {
				sb.WriteString(SelectedGroupStyle.MaxWidth(innerW).Render("▸ " + line))
			} else {
				sb.WriteString(NormalGroupStyle.MaxWidth(innerW).Render("▸ " + line))
			}
		} else {
			sb.WriteString(NormalGroupStyle.MaxWidth(innerW).Render("  " + line))
		}
		sb.WriteString("\n")
	}

	if innerH >= 14 {
		div := lipgloss.NewStyle().Foreground(ColorMuted).Render(strings.Repeat("─", innerW))
		sb.WriteString(div + "\n")
		sb.WriteString(lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary).Render("⚡ ACTIONS") + "\n")
		sb.WriteString(fmt.Sprintf("%s Add Host\n", KeyBadgeStyle.Render("[a]")))
		sb.WriteString(fmt.Sprintf("%s Edit\n", KeyBadgeStyle.Render("[e]")))
		sb.WriteString(fmt.Sprintf("%s Delete\n", KeyBadgeStyle.Render("[d]")))
		sb.WriteString(fmt.Sprintf("%s Toggle\n", KeyBadgeStyle.Render("[Spc]")))
		sb.WriteString(fmt.Sprintf("%s Search\n", KeyBadgeStyle.Render("[/]")))

		if innerH >= 19 {
			sb.WriteString(div + "\n")
			sb.WriteString(lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary).Render("🛡️ SYNC") + "\n")
			sb.WriteString(fmt.Sprintf("%s Apply\n", KeyBadgeStyle.Render("[A]")))
			sb.WriteString(fmt.Sprintf("%s Rollback\n", KeyBadgeStyle.Render("[R]")))
			sb.WriteString(fmt.Sprintf("%s Import\n", KeyBadgeStyle.Render("[m]")))
		}
	}

	style := SidebarUnfocusedStyle
	if m.focus == FocusSidebar {
		style = SidebarStyle
	}

	return style.
		Width(innerW).
		Height(innerH).
		MaxWidth(width).
		MaxHeight(height).
		Render(sb.String())
}

func (m Model) renderTable(width, height int) string {
	var sb strings.Builder

	innerW := width - 4
	if innerW < 20 {
		innerW = 20
	}
	innerH := height - 2
	if innerH < 4 {
		innerH = 4
	}

	currentGrpName := "[All Entries]"
	if m.selectedGrp < len(m.groupsList) {
		currentGrpName = m.groupsList[m.selectedGrp]
	}

	titleText := fmt.Sprintf("HOSTS — %s", currentGrpName)
	if m.searchQuery != "" {
		titleText += fmt.Sprintf(" (Filter: '%s')", m.searchQuery)
	}

	sb.WriteString(lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary).Render(titleText) + "\n")

	// Calculate Responsive Column Widths
	statusW := 6
	ipW := 15
	if innerW < 60 {
		ipW = 13
	}
	remaining := innerW - statusW - ipW - 6 // spacing
	if remaining < 10 {
		remaining = 10
	}

	hostW := int(float64(remaining) * 0.45)
	aliasW := int(float64(remaining) * 0.25)
	commentW := remaining - hostW - aliasW

	if hostW < 10 {
		hostW = 10
	}
	if aliasW < 6 {
		aliasW = 6
	}
	if commentW < 6 {
		commentW = 6
	}

	headerLine := fmt.Sprintf(" %-*s %-*s %-*s %-*s %-*s",
		statusW, "STATUS",
		ipW, "IP ADDRESS",
		hostW, "HOSTNAME",
		aliasW, "ALIASES",
		commentW, "COMMENT",
	)
	sb.WriteString(TableHeaderStyle.Width(innerW).MaxWidth(innerW).Render(headerLine) + "\n")

	visible := m.GetVisibleEntries()
	maxRows := innerH - 2
	if maxRows < 1 {
		maxRows = 1
	}

	if len(visible) == 0 {
		sb.WriteString("\n" + NormalRowStyle.Render("  (No host entries. Press [a] to add a new host entry.)\n"))
	} else {
		// Calculate scrolling window
		startIdx := m.scrollOffset
		if startIdx >= len(visible) {
			startIdx = 0
		}
		endIdx := startIdx + maxRows
		if endIdx > len(visible) {
			endIdx = len(visible)
		}

		for idx := startIdx; idx < endIdx; idx++ {
			e := visible[idx]
			statusBadge := BadgeDangerStyle.Render("○ OFF")
			if e.Enabled {
				statusBadge = BadgeSuccessStyle.Render("● ON ")
			}

			aliasStr := strings.Join(e.Aliases, ",")
			if len(aliasStr) == 0 {
				aliasStr = "-"
			}

			rowContent := fmt.Sprintf(" %-6s %-*s %-*s %-*s %-*s",
				statusBadge,
				ipW, truncate(e.IP, ipW),
				hostW, truncate(e.Hostname, hostW),
				aliasW, truncate(aliasStr, aliasW),
				commentW, truncate(e.Comment, commentW),
			)

			if idx == m.selectedRow {
				indicator := "▸"
				if m.focus != FocusTable {
					indicator = " "
				}
				sb.WriteString(SelectedRowStyle.Width(innerW).MaxWidth(innerW).Render(indicator+rowContent[1:]) + "\n")
			} else {
				sb.WriteString(NormalRowStyle.Width(innerW).MaxWidth(innerW).Render(rowContent) + "\n")
			}
		}

		// Scroll indicator if more entries exist
		if len(visible) > maxRows {
			info := fmt.Sprintf("  [%d/%d hosts]", m.selectedRow+1, len(visible))
			sb.WriteString(lipgloss.NewStyle().Foreground(ColorMuted).Render(info))
		}
	}

	style := TableUnfocusedContainerStyle
	if m.focus == FocusTable {
		style = TableContainerStyle
	}

	return style.
		Width(innerW).
		Height(innerH).
		MaxWidth(width).
		MaxHeight(height).
		Render(sb.String())
}

func (m Model) renderFooter() string {
	statusStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(m.statusColor)).
		Bold(true)

	statusLine := statusStyle.Render(" " + m.statusMsg)
	keysLine := lipgloss.NewStyle().Foreground(lipgloss.Color("#AAAAAA")).
		Render(" Tab: Switch │ j/k: Nav │ Space: Toggle │ a: Add │ e: Edit │ d: Del │ A: Apply │ R: Rollback │ q: Quit │ ?: Help")

	return statusLine + "\n" + StatusbarStyle.Width(m.width).Render(keysLine)
}

func (m Model) renderModalOverlay() string {
	modalW := 54
	if modalW > m.width-4 {
		modalW = m.width - 4
	}
	if modalW < 30 {
		modalW = 30
	}

	var modalContent string

	switch m.modal {
	case ModalNotice:
		modalContent = ModalBoxStyle.Width(modalW).Render(
			ModalTitleStyle.Render(" ℹ️ Permission Notice ") + "\n\n" +
				m.noticeMsg + "\n\n" +
				lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary).Render("[ Press Enter or Esc to Continue ]"),
		)

	case ModalHelp:
		modalContent = ModalBoxStyle.Width(modalW).Render(
			ModalTitleStyle.Render(" 📖 Keybindings ") + "\n\n" +
				"  Tab / h / l   Switch Groups & Table\n" +
				"  j / k / ↑ / ↓ Navigate rows\n" +
				"  Space         Toggle enabled / disabled\n" +
				"  a             Add new host entry\n" +
				"  e / Enter     Edit selected host entry\n" +
				"  d             Delete selected host entry\n" +
				"  /             Fuzzy search & filter hosts\n" +
				"  A             Apply enabled entries to system\n" +
				"  R             Rollback system /etc/hosts\n" +
				"  m             Import current system hosts\n" +
				"  ?             Toggle help\n" +
				"  q / Ctrl+C    Quit hostcli\n\n" +
				lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary).Render("[ Press Esc or ? to Close ]"),
		)

	case ModalSearch:
		modalContent = ModalBoxStyle.Width(modalW).Render(
			ModalTitleStyle.Render(" 🔍 Search & Filter ") + "\n\n" +
				m.searchInput.View() + "\n\n" +
				lipgloss.NewStyle().Foreground(ColorMuted).Render("Enter: Search  │  Esc: Clear & Close"),
		)

	case ModalDelete:
		visible := m.GetVisibleEntries()
		hostName := "selected host"
		if len(visible) > 0 && m.selectedRow < len(visible) {
			hostName = visible[m.selectedRow].Hostname
		}
		modalContent = ModalBoxStyle.Width(modalW).Render(
			ModalTitleStyle.Render(" ⚠️ Confirm Delete ") + "\n\n" +
				fmt.Sprintf("Delete host '%s'?\n\n", hostName) +
				lipgloss.NewStyle().Foreground(ColorDanger).Bold(true).Render("[y] Delete") +
				"    " +
				lipgloss.NewStyle().Foreground(ColorWhite).Render("[n / Esc] Cancel"),
		)

	case ModalRollback:
		modalContent = ModalBoxStyle.Width(modalW).Render(
			ModalTitleStyle.Render(" 🛡️ System Hosts Rollback ") + "\n\n" +
				"  [1] Strip Managed Block\n" +
				"      (Removes # BEGIN HOSTCLI ... # END lines)\n\n" +
				"  [2] Restore Pristine Original Backup\n" +
				"      (Restores hosts.original.bak)\n\n" +
				lipgloss.NewStyle().Foreground(ColorMuted).Render("Press 1, 2, or Esc to Cancel"),
		)

	case ModalEdit:
		modalContent = m.renderEditModal(modalW)
	}

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, modalContent)
}

func (m Model) renderEditModal(boxWidth int) string {
	form := m.editForm
	title := " ✏️ Edit Host "
	if form.IsNew {
		title = " ➕ Add Host "
	}

	var sb strings.Builder
	sb.WriteString(ModalTitleStyle.Render(title) + "\n\n")

	if form.ErrorMsg != "" {
		sb.WriteString(lipgloss.NewStyle().Foreground(ColorDanger).Bold(true).Render("Error: "+form.ErrorMsg) + "\n")
	}

	labelW := 12
	fields := []struct {
		Label string
		View  string
		Index FormField
	}{
		{"IP Address", form.IPInput.View(), FieldIP},
		{"Hostname", form.HostInput.View(), FieldHostname},
		{"Aliases", form.AliasesInput.View(), FieldAliases},
		{"Group", form.GroupInput.View(), FieldGroup},
		{"Comment", form.CommentInput.View(), FieldComment},
	}

	for _, f := range fields {
		lblStyle := lipgloss.NewStyle().Foreground(ColorMuted)
		if form.FocusIndex == f.Index {
			lblStyle = lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary)
		}
		sb.WriteString(fmt.Sprintf("%-*s: %s\n", labelW, lblStyle.Render(f.Label), f.View))
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
	sb.WriteString(fmt.Sprintf("%-*s: %s\n\n", labelW, enableStyle.Render("Status"), enableStyle.Render(enableBox+" (Spc to toggle)")))

	// Action Buttons
	saveBtn := "[ Save (Enter) ]"
	cancelBtn := "[ Cancel (Esc) ]"

	if form.FocusIndex == FieldSave {
		saveBtn = lipgloss.NewStyle().Bold(true).Background(ColorPrimary).Foreground(ColorWhite).Render(saveBtn)
	}
	if form.FocusIndex == FieldCancel {
		cancelBtn = lipgloss.NewStyle().Bold(true).Background(ColorHighlight).Foreground(ColorWhite).Render(cancelBtn)
	}

	sb.WriteString(fmt.Sprintf("  %s    %s\n", saveBtn, cancelBtn))

	innerW := boxWidth - 4
	if innerW < 24 {
		innerW = 24
	}

	return ModalBoxStyle.Width(innerW).Render(sb.String())
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + ".."
}
