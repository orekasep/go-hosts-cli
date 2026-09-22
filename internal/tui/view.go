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
	headerHeight := 1

	// 2. Footer (fixed 2 lines: 1 status line + 1 keybindings line)
	footerView := m.renderFooter()
	footerHeight := 2

	// 3. Body Dimensions (reserve 1 line buffer so terminal cursor never triggers scroll)
	bodyHeight := m.height - headerHeight - footerHeight - 1
	if bodyHeight < 6 {
		bodyHeight = 6
	}

	// Sidebar width
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
	ver := m.version
	if ver == "" {
		ver = "1.0.1"
	}
	title := HeaderStyle.Render(fmt.Sprintf(" hostcli v%s ", ver))
	titleW := lipgloss.Width(title)

	var privBadge string
	if m.privStatus.HasDirectWrite || m.privStatus.IsElevated {
		privBadge = BadgeSuccessStyle.Render("[✓ Direct Write]")
	} else if m.privStatus.CanSudo {
		privBadge = BadgeWarningStyle.Render("[🔒 sudo on Apply]")
	} else {
		privBadge = BadgeDangerStyle.Render("[⚠️ Read-Only]")
	}
	privW := lipgloss.Width(privBadge)

	totalEntries := m.hostsFile.TotalEntries()
	enabledEntries := m.hostsFile.EnabledEntries()
	stats := fmt.Sprintf("Hosts: %d total, %d active", totalEntries, enabledEntries)
	statsW := len(stats)

	gap := m.width - titleW - 2 - privW - statsW
	if gap < 2 {
		// If screen is narrow, omit privBadge
		if m.width > titleW+statsW+4 {
			gap = m.width - titleW - statsW - 2
			return title + strings.Repeat(" ", gap) + stats
		}
		return title
	}

	return title + "  " + privBadge + strings.Repeat(" ", gap) + stats
}

func (m Model) renderSidebar(width, height int) string {
	// Inner printable bounds inside borders (2) and padding (2)
	innerW := width - 4
	if innerW < 10 {
		innerW = 10
	}
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

		countStr := fmt.Sprintf("(%d)", count)
		maxNameLen := innerW - len(countStr) - 3 // space for prefix (2) and 1 space
		if maxNameLen < 4 {
			maxNameLen = 4
		}
		cleanName := truncatePlain(grpName, maxNameLen)
		itemText := fmt.Sprintf("%-*s %s", maxNameLen, cleanName, countStr)
		prefix := "  "
		if idx == m.selectedGrp {
			prefix = "▸ "
		}
		fullItem := truncatePlain(prefix+itemText, innerW)

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
		lines = append(lines, fmt.Sprintf("%s Add Host", KeyBadgeStyle.Render("[a]")))
		lines = append(lines, fmt.Sprintf("%s Edit", KeyBadgeStyle.Render("[e]")))
		lines = append(lines, fmt.Sprintf("%s Delete", KeyBadgeStyle.Render("[d]")),)
		lines = append(lines, fmt.Sprintf("%s Toggle", KeyBadgeStyle.Render("[Spc]")))
		lines = append(lines, fmt.Sprintf("%s Search", KeyBadgeStyle.Render("[/]")))

		remainingLines = innerH - len(lines)
		if remainingLines >= 5 {
			lines = append(lines, div)
			lines = append(lines, lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary).Render("🛡️ SYNC"))
			lines = append(lines, fmt.Sprintf("%s Apply", KeyBadgeStyle.Render("[A]")))
			lines = append(lines, fmt.Sprintf("%s Rollback", KeyBadgeStyle.Render("[R]")))
			lines = append(lines, fmt.Sprintf("%s Import", KeyBadgeStyle.Render("[m]")))
		}
	}

	// Clamp to innerH lines
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

	titleText := truncatePlain(fmt.Sprintf("HOSTS — %s", currentGrpName), innerW)
	if m.searchQuery != "" {
		titleText = truncatePlain(fmt.Sprintf("HOSTS — %s (Filter: '%s')", currentGrpName, m.searchQuery), innerW)
	}
	lines = append(lines, lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary).Render(titleText))

	// Calculate Responsive Column Widths strictly within innerW
	statusW := 6
	ipW := 15
	if innerW < 55 {
		ipW = 12
	}
	availForRest := innerW - statusW - ipW - 4 // 4 spaces between 5 columns
	if availForRest < 8 {
		availForRest = 8
	}

	hostW := int(float64(availForRest) * 0.40)
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

	// Adjust if rounding exceeded innerW
	for statusW+ipW+hostW+aliasW+commentW+4 > innerW && commentW > 3 {
		commentW--
	}
	for statusW+ipW+hostW+aliasW+commentW+4 > innerW && hostW > 6 {
		hostW--
	}

	headerLine := fmt.Sprintf("%-*s %-*s %-*s %-*s %-*s",
		statusW, "STATUS",
		ipW, "IP ADDRESS",
		hostW, "HOSTNAME",
		aliasW, "ALIASES",
		commentW, "COMMENT",
	)
	lines = append(lines, TableHeaderStyle.Width(innerW).Render(truncatePlain(headerLine, innerW)))

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
			statusText := "○ OFF"
			var statusStyled string
			if e.Enabled {
				statusText = "● ON "
				statusStyled = BadgeSuccessStyle.Render(statusText)
			} else {
				statusStyled = BadgeDangerStyle.Render(statusText)
			}

			ipClean := truncatePlain(e.IP, ipW)
			hostClean := truncatePlain(e.Hostname, hostW)
			aliasClean := truncatePlain(strings.Join(e.Aliases, ","), aliasW)
			if aliasClean == "" {
				aliasClean = "-"
			}
			commentClean := truncatePlain(e.Comment, commentW)

			rowContent := fmt.Sprintf("%s %-*s %-*s %-*s %-*s",
				statusStyled,
				ipW, ipClean,
				hostW, hostClean,
				aliasW, aliasClean,
				commentW, commentClean,
			)

			if idx == m.selectedRow && m.focus == FocusTable {
				lines = append(lines, SelectedRowStyle.Render(rowContent))
			} else {
				lines = append(lines, NormalRowStyle.Render(rowContent))
			}
		}
	}

	// Clamp to innerH
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

	statusLine := truncatePlain(" "+m.statusMsg, m.width)

	var keysText string
	switch {
	case m.width >= 105:
		keysText = " Tab: Switch │ j/k: Nav │ Space: Toggle │ a: Add │ e: Edit │ d: Del │ A: Apply │ R: Rollback │ /: Search │ q: Quit │ ?: Help"
	case m.width >= 75:
		keysText = " Tab: Pane │ j/k: Nav │ Spc: Toggle │ a: Add │ e: Edit │ d: Del │ A: Apply │ R: Undo │ q: Quit"
	default:
		keysText = " Tab: Pane │ j/k: Nav │ Spc: Toggle │ a: Add │ A: Apply │ q: Quit"
	}

	keysLine := StatusbarStyle.Width(m.width).Render(truncatePlain(keysText, m.width))

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
		truncatePlain(m.noticeMsg, innerW),
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
		truncatePlain("  Tab / h / l   Switch Groups & Table", innerW),
		truncatePlain("  j / k / ↑ / ↓ Navigate rows", innerW),
		truncatePlain("  Space         Toggle enabled / disabled", innerW),
		truncatePlain("  a             Add new host entry", innerW),
		truncatePlain("  e / Enter     Edit selected host entry", innerW),
		truncatePlain("  d             Delete selected host entry", innerW),
		truncatePlain("  /             Fuzzy search & filter", innerW),
		truncatePlain("  A             Apply enabled to system", innerW),
		truncatePlain("  R             Rollback system /etc/hosts", innerW),
		truncatePlain("  m             Import current system hosts", innerW),
		truncatePlain("  q / Ctrl+C    Quit hostcli", innerW),
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
		truncatePlain(fmt.Sprintf("Delete host '%s'?", hostName), innerW),
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
		truncatePlain("  [1] Strip Managed Block", innerW),
		truncatePlain("      (Removes # BEGIN/END lines)", innerW),
		"",
		truncatePlain("  [2] Restore Pristine Original Backup", innerW),
		truncatePlain("      (Restores hosts.original.bak)", innerW),
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
		lines = append(lines, fmt.Sprintf("%-*s: %s", labelW, lblStyle.Render(f.Label), f.View))
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
	lines = append(lines, fmt.Sprintf("%-*s: %s", labelW, enableStyle.Render("Status"), enableStyle.Render(enableBox+" (Spc to toggle)")))
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

func truncatePlain(s string, maxLen int) string {
	if maxLen <= 0 {
		return ""
	}
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	if maxLen <= 2 {
		return string(runes[:maxLen])
	}
	return string(runes[:maxLen-2]) + ".."
}
