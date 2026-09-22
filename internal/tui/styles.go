package tui

import "github.com/charmbracelet/lipgloss"

var (
	// Theme Colors
	ColorPrimary   = lipgloss.Color("#7D56F4") // Vibrant Purple
	ColorSecondary = lipgloss.Color("#5A32E6")
	ColorSuccess   = lipgloss.Color("#04B575") // Vibrant Green
	ColorWarning   = lipgloss.Color("#FFAF00") // Warm Amber
	ColorDanger    = lipgloss.Color("#FF4672") // Coral Red
	ColorMuted     = lipgloss.Color("#626262") // Slate Gray
	ColorHighlight = lipgloss.Color("#383838")
	ColorBgDark    = lipgloss.Color("#1E1E2E")
	ColorWhite     = lipgloss.Color("#FAFAFA")

	// Header Styles
	HeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorWhite).
			Background(ColorPrimary).
			Padding(0, 1)

	BadgeSuccessStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(ColorSuccess)

	BadgeDangerStyle = lipgloss.NewStyle().
				Foreground(ColorMuted)

	BadgeWarningStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(ColorWarning)

	// Panel Styles
	SidebarStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorPrimary).
			Padding(1, 1)

	SidebarUnfocusedStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(ColorMuted).
				Padding(1, 1)

	TableContainerStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(ColorPrimary).
				Padding(0, 1)

	TableUnfocusedContainerStyle = lipgloss.NewStyle().
					Border(lipgloss.RoundedBorder()).
					BorderForeground(ColorMuted).
					Padding(0, 1)

	// Table Element Styles
	TableHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(ColorPrimary).
				BorderStyle(lipgloss.NormalBorder()).
				BorderBottom(true).
				BorderForeground(ColorMuted)

	SelectedRowStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(ColorWhite).
				Background(ColorHighlight)

	NormalRowStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#D9D9D9"))

	// Group Item Styles
	SelectedGroupStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(ColorWhite).
				Background(ColorSecondary).
				Padding(0, 1)

	NormalGroupStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#C0C0C0")).
				Padding(0, 1)

	// Statusbar & Cheatsheet Styles
	StatusbarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#EEEEEE")).
			Background(lipgloss.Color("#2E2E3E")).
			Padding(0, 1)

	KeyBadgeStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorPrimary)

	// Modal Dialog Styles
	ModalBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.DoubleBorder()).
			BorderForeground(ColorPrimary).
			Padding(1, 2).
			Background(ColorBgDark)

	ModalTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorWhite).
			Background(ColorPrimary).
			Padding(0, 1).
			MarginBottom(1)
)
