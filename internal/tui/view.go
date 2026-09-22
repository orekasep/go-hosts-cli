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

	// 1. Header (exactly 1 line)
	headerView := m.renderHeader()
	headerHeight := 1

	// 2. Footer (exactly 2 lines: 1 status line + 1 keybindings line)
	footerView := m.renderFooter()
	footerHeight := 2

	// 3. Body Dimensions (exact lines matching)
	bodyHeight := m.height - headerHeight - footerHeight
	if bodyHeight < 6 {
		bodyHeight = 6
	}

	// Sidebar width: 24 on standard, 20 on small
	sidebarWidth := 24
	if m.width < 80 {
		sidebarWidth = 20
	}
	tableWidth := m.width - sidebarWidth
	if tableWidth < 28 {
		tableWidth = 28
	}

	sidebarView := m.renderSidebar(sidebarWidth, bodyHeight)
	tableView := m.renderTable(tableWidth, bodyHeight)

	mainBody := lipgloss.JoinHorizontal(lipgloss.Top, sidebarView, tableView)

	baseView := lipgloss.JoinVertical(lipgloss.Left, headerView, mainBody, footerView)

	// 4. Modal Overlay (if active)
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

	header := lipgloss.JoinHorizontal(
		lipgloss.Center,
		title,
		"  ",
		privBadge,
		strings.Repeat(" ", gap),
		stats,
	)

	// Clamp to terminal width
	return truncate(header, m.width)
}

func (m Model) renderSidebar(width, height int) string {
	// With Border (2 cols) and Padding (2 cols), inner printable width is width - 4
	innerW := width - 4
	if innerW < 10 {
		innerW = 10
	}
	// With Border (2 lines) and Padding (0 lines), inner printable lines is height - 2
	innerH := height - 2
	if innerH < 4 {
		innerH = 4
	}

	var lines []string
	lines = append(lines, lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary).Render("📁 GROUPS"))

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
		itemText := fmt.Sprintf("%-*s (%d)", maxNameLen, truncate(grpName, maxNameLen), count)
		prefix := "  "
		if idx == m.selectedGrp {
			prefix = "▸ "
		}
		fullItem := truncate(prefix+itemText, innerW)

		if idx == m.selectedGrp {
			if m.focus == FocusSidebar {
				lines = append(lines, SelectedGroupStyle.Render(fullItem))
			} else {
				lines = append(lines, NormalGroupStyle.Render(fullItem))
			}
		} else {
			lines = append(lines, NormalGroupStyle.Render(fullItem))
		}
	}

	// Actions shortcuts section if vertical space permits
	remainingLines := innerH - len(lines)
	if remainingLines >= 8 {
		div := lipgloss.NewStyle().Foreground(ColorMuted).Render(strings.Repeat("─", innerW))
		lines = append(lines, div)
		lines = append(lines, lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary).Render("⚡ ACTIONS"))
		lines = append(lines, truncate(fmt.Sprintf("%s Add Host", KeyBadgeStyle.Render("[a]")), innerW))
		lines = append(lines, truncate(fmt.Sprintf("%s Edit", KeyBadgeStyle.Render("[e]")), innerW))
		lines = append(lines, truncate(fmt.Sprintf("%s Delete", KeyBadgeStyle.Render("[d]")), innerW))
		lines = append(lines, truncate(fmt.Sprintf("%s Toggle", KeyBadgeStyle.Render("[Spc]")), innerW))
		lines = append(lines, truncate(fmt.Sprintf("%s Search", KeyBadgeStyle.Render("[/]")), innerW))

		remainingLines = innerH - len(lines)
		if remainingLines >= 5 {
			lines = append(lines, div)
			lines = append(lines, lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary).Render("🛡️ SYNC"))
			lines = append(lines, truncate(fmt.Sprintf("%s Apply", KeyBadgeStyle.Render("[A]")), innerW))
			lines = append(lines, truncate(fmt.Sprintf("%s Rollback", KeyBadgeStyle.Render("[R]")), innerW))
			lines = append(lines, truncate(fmt.Sprintf("%s Import", KeyBadgeStyle.Render("[m]")), innerW))
		}
	}

	// Ensure slice does not exceed innerH
	if len(lines) > innerH {
		lines = lines[:innerH]
	}

	content := strings.Join(lines, "\n")
	style := SidebarUnfocusedStyle
	if m.focus == FocusSidebar {
		style = SidebarStyle
	}

	return style.
		Width(width - 2).
		Height(innerH).
		Render(content)
}

func (m Model) renderTable(width, height int) string {
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

	var lines []string

	titleText := fmt.Sprintf("HOSTS — %s", currentGrpName)
	if m.searchQuery != "" {
		titleText += fmt.Sprintf(" (Filter: '%s')", m.searchQuery)
	}
	lines = append(lines, truncate(lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary).Render(titleText), innerW))

	// Calculate Responsive Column Widths strictly within innerW
	statusW := 6
	ipW := 15
	if innerW < 55 {
		ipW = 12
	}
	availForRest := innerW - statusW - ipW - 4 // 4 spaces between columns
	if availForRest < 8 {
		availForRest = 8
	}

	hostW := int(float64(availForRest) * 0.45)
	aliasW := int(float64(availForRest) * 0.25)
	commentW := availForRest - hostW - aliasW

	if hostW < 8 {
		hostW = 8
	}
	if aliasW < 5 {
		aliasW = 5
	}
	if commentW < 5 {
		commentW = 5
	}

	headerLine := fmt.Sprintf("%-*s %-*s %-*s %-*s %-*s",
		statusW, "STATUS",
		ipW, "IP ADDRESS",
		hostW, "HOSTNAME",
		aliasW, "ALIASES",
		commentW, "COMMENT",
	)
	lines = append(lines, TableHeaderStyle.Width(innerW).Render(truncate(headerLine, innerW)))

	visible := m.GetVisibleEntries()
	maxRows := innerH - len(lines)
	if maxRows < 1 {
		maxRows = 1
	}

	if len(visible) == 0 {
		lines = append(lines, NormalRowStyle.Render(" (No entries. Press [a] to add a new host entry.)"))
	} else {
		// Clamp scrolling offset
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

			rowContent := fmt.Sprintf("%-6s %-*s %-*s %-*s %-*s",
				statusBadge,
				ipW, truncate(e.IP, ipW),
				hostW, truncate(e.Hostname, hostW),
				aliasW, truncate(aliasStr, aliasW),
				commentW, truncate(e.Comment, commentW),
			)

			truncatedRow := truncate(rowContent, innerW)
			if idx == m.selectedRow {
				lines = append(lines, SelectedRowStyle.Width(innerW).Render(truncatedRow))
			} else {
				lines = append(lines, NormalRowStyle.Width(innerW).Render(truncatedRow))
			}
		}
	}

	// Ensure slice does not exceed innerH
	if len(lines) > innerH {
		lines = lines[:innerH]
	}

	content := strings.Join(lines, "\n")
	style := TableUnfocusedContainerStyle
	if m.focus == FocusTable {
		style = TableContainerStyle
	}

	return style.
		Width(width - 2).
		Height(innerH).
		Render(content)
}

func (m Model) renderFooter() string {
	statusStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(m.statusColor)).
		Bold(true)

	statusLine := truncate(" "+m.statusMsg, m.width)

	// Responsive shortcuts bar that never wraps
	var keysText string
	switch {
	case m.width >= 105:
		keysText = " Tab: Switch │ j/k: Nav │ Space: Toggle │ a: Add │ e: Edit │ d: Del │ A: Apply │ R: Rollback │ /: Search │ q: Quit │ ?: Help"
	case m.width >= 75:
		keysText = " Tab: Pane │ j/k: Nav │ Spc: Toggle │ a: Add │ e: Edit │ d: Del │ A: Apply │ R: Undo │ q: Quit"
	default:
		keysText = " Tab: Pane │ j/k: Nav │ Spc: Toggle │ a: Add │ A: Apply │ q: Quit"
	}

	keysLine := StatusbarStyle.Width(m.width).Render(truncate(keysText, m.width))

	return statusStyle.Render(statusLine) + "\n" + keysLine
}

func (m Model) renderModalOverlay() string {
	modalW := 52
	if modalW > m.width-4 {
		modalW = m.width - 4
	}
	if modalW < 30 {
		modalW = 30
	}

	var modalBox string

	switch m.modal {
	case ModalNotice:
		modalBox = m.renderNoticeModal(modalW)
	case ModalHelp:
		modalBox = m.renderHelpModal(modalW)
	case ModalSearch:
		modalBox = m.renderSearchModal(modalW)
	case ModalDelete:
		modalBox = m.renderDeleteModal(modalW)
	case ModalRollback:
		modalBox = m.renderRollbackModal(modalW)
	case ModalEdit:
		modalBox = m.renderEditModal(modalW)
	}

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, modalBox)
}

func (m Model) renderNoticeModal(boxWidth int) string {
	innerW := boxWidth - 4
	lines := []string{
		ModalTitleStyle.Render(" ℹ️ Permission Notice "),
		"",
		truncate(m.noticeMsg, innerW),
		"",
		lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary).Render("[ Press Enter or Esc ]"),
	}
	return ModalBoxStyle.Width(boxWidth - 2).Render(strings.Join(lines, "\n"))
}

func (m Model) renderHelpModal(boxWidth int) string {
	innerW := boxWidth - 4
	lines := []string{
		ModalTitleStyle.Render(" 📖 Keybindings "),
		"",
		truncate("  Tab / h / l   Switch Groups & Table", innerW),
		truncate("  j / k / ↑ / ↓ Navigate rows", innerW),
		truncate("  Space         Toggle enabled / disabled", innerW),
		truncate("  a             Add new host entry", innerW),
		truncate("  e / Enter     Edit selected host entry", innerW),
		truncate("  d             Delete selected host entry", innerW),
		truncate("  /             Fuzzy search & filter", innerW),
		truncate("  A             Apply enabled to system", innerW),
		truncate("  R             Rollback system /etc/hosts", innerW),
		truncate("  m             Import current system hosts", innerW),
		truncate("  q / Ctrl+C    Quit hostcli", innerW),
		"",
		lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary).Render("[ Press Esc or ? to Close ]"),
	}
	return ModalBoxStyle.Width(boxWidth - 2).Render(strings.Join(lines, "\n"))
}

func (m Model) renderSearchModal(boxWidth int) string {
	lines := []string{
		ModalTitleStyle.Render(" 🔍 Search & Filter "),
		"",
		m.searchInput.View(),
		"",
		lipgloss.NewStyle().Foreground(ColorMuted).Render("Enter: Search  │  Esc: Clear & Close"),
	}
	return ModalBoxStyle.Width(boxWidth - 2).Render(strings.Join(lines, "\n"))
}

func (m Model) renderDeleteModal(boxWidth int) string {
	visible := m.GetVisibleEntries()
	hostName := "selected host"
	if len(visible) > 0 && m.selectedRow < len(visible) {
		hostName = visible[m.selectedRow].Hostname
	}
	innerW := boxWidth - 4
	lines := []string{
		ModalTitleStyle.Render(" ⚠️ Confirm Delete "),
		"",
		truncate(fmt.Sprintf("Delete host '%s'?", hostName), innerW),
		"",
		lipgloss.NewStyle().Foreground(ColorDanger).Bold(true).Render("[y] Delete") + "    " +
			lipgloss.NewStyle().Foreground(ColorWhite).Render("[n / Esc] Cancel"),
	}
	return ModalBoxStyle.Width(boxWidth - 2).Render(strings.Join(lines, "\n"))
}

func (m Model) renderRollbackModal(boxWidth int) string {
	innerW := boxWidth - 4
	lines := []string{
		ModalTitleStyle.Render(" 🛡️ System Hosts Rollback "),
		"",
		truncate("  [1] Strip Managed Block", innerW),
		truncate("      (Removes # BEGIN/END lines)", innerW),
		"",
		truncate("  [2] Restore Pristine Original Backup", innerW),
		truncate("      (Restores hosts.original.bak)", innerW),
		"",
		lipgloss.NewStyle().Foreground(ColorMuted).Render("Press 1, 2, or Esc to Cancel"),
	}
	return ModalBoxStyle.Width(boxWidth - 2).Render(strings.Join(lines, "\n"))
}

func (m Model) renderEditModal(boxWidth int) string {
	form := m.editForm
	title := " ✏️ Edit Host "
	if form.IsNew {
		title = " ➕ Add Host "
	}

	innerW := boxWidth - 4
	if innerW < 24 {
		innerW = 24
	}

	var lines []string
	lines = append(lines, ModalTitleStyle.Render(title))
	lines = append(lines, "")

	if form.ErrorMsg != "" {
		lines = append(lines, lipgloss.NewStyle().Foreground(ColorDanger).Bold(true).Render("Error: "+form.ErrorMsg))
	}

	labelW := 11
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
		lines = append(lines, truncate(fmt.Sprintf("%-*s: %s", labelW, lblStyle.Render(f.Label), f.View), innerW))
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
	lines = append(lines, truncate(fmt.Sprintf("%-*s: %s", labelW, enableStyle.Render("Status"), enableStyle.Render(enableBox+" (Spc to toggle)")), innerW))
	lines = append(lines, "")

	// Action Buttons
	saveBtn := "[ Save (Enter) ]"
	cancelBtn := "[ Cancel (Esc) ]"

	if form.FocusIndex == FieldSave {
		saveBtn = lipgloss.NewStyle().Bold(true).Background(ColorPrimary).Foreground(ColorWhite).Render(saveBtn)
	}
	if form.FocusIndex == FieldCancel {
		cancelBtn = lipgloss.NewStyle().Bold(true).Background(ColorHighlight).Foreground(ColorWhite).Render(cancelBtn)
	}

	lines = append(lines, fmt.Sprintf("  %s    %s", saveBtn, cancelBtn))

	return ModalBoxStyle.Width(boxWidth - 2).Render(strings.Join(lines, "\n"))
}

func truncate(s string, maxLen int) string {
	if maxLen <= 0 {
		return ""
	}
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + ".."
}
