package tui

import (
	"os"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/orekasep/go-hosts-cli/internal/apply"
	"github.com/orekasep/go-hosts-cli/internal/core/config"
	"github.com/orekasep/go-hosts-cli/internal/domain"
	"golang.org/x/term"
)

type FocusArea int

const (
	FocusSidebar FocusArea = iota
	FocusTable
)

type ModalType int

const (
	ModalNone ModalType = iota
	ModalEdit
	ModalDelete
	ModalRollback
	ModalHelp
	ModalSearch
	ModalNotice
)

type FormField int

const (
	FieldIP FormField = iota
	FieldHostname
	FieldAliases
	FieldGroup
	FieldComment
	FieldEnabled
	FieldSave
	FieldCancel
)

type EditForm struct {
	IsNew        bool
	EntryID      string
	OriginalName string
	IPInput      textinput.Model
	HostInput    textinput.Model
	AliasesInput textinput.Model
	GroupInput   textinput.Model
	CommentInput textinput.Model
	Enabled      bool
	FocusIndex   FormField
	ErrorMsg     string
}

type Model struct {
	configPath  string
	hostsFile   *domain.HostsFile
	privStatus  apply.PrivilegeStatus
	width       int
	height      int
	focus       FocusArea
	modal       ModalType
	selectedGrp  int // Index into groups list (0 is always "[All]")
	selectedRow  int // Index of selected entry within visible entries
	scrollOffset int // Offset for table scrolling
	searchInput  textinput.Model
	searchQuery string
	editForm    EditForm
	statusMsg   string
	statusColor string
	groupsList  []string // computed list of available groups
	noticeMsg   string
	version     string
}

func InitialModel(configPath string, version ...string) Model {
	hf, err := config.LoadConfig(configPath)
	if err != nil {
		hf = &domain.HostsFile{Version: 1, Groups: []domain.Group{}}
	}

	sysHosts := config.GetSystemHostsPath()
	priv := apply.CheckPrivileges(sysHosts)

	ti := textinput.New()
	ti.Placeholder = "Search hostname, IP, alias..."
	ti.CharLimit = 100

	w, h, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil || w <= 0 || h <= 0 {
		w, h, err = term.GetSize(int(os.Stdin.Fd()))
		if err != nil || w <= 0 || h <= 0 {
			w = 80
			h = 24
		}
	}

	ver := "1.0.1"
	if len(version) > 0 && version[0] != "" {
		ver = version[0]
	}

	m := Model{
		configPath:  configPath,
		hostsFile:   hf,
		privStatus:  priv,
		width:       w,
		height:      h,
		version:     ver,
		focus:       FocusTable,
		modal:       ModalNone,
		selectedGrp: 0,
		selectedRow: 0,
		searchInput: ti,
		statusMsg:   "Welcome to hostcli! Press '?' for keybindings.",
		statusColor: "#04B575",
	}

	m.refreshGroups()

	if !priv.HasDirectWrite && !priv.CanSudo && !priv.IsElevated {
		m.noticeMsg = "Warning: No direct write access and 'sudo' not detected. Changes will save locally to ~/.hostcli/hosts.yaml, but applying to system hosts may require root permissions."
		m.modal = ModalNotice
	}

	return m
}

func (m *Model) refreshGroups() {
	set := make(map[string]bool)
	var list []string
	list = append(list, "[All Entries]")

	for _, g := range m.hostsFile.Groups {
		if !set[g.Name] && g.Name != "" {
			set[g.Name] = true
			list = append(list, g.Name)
		}
	}
	m.groupsList = list
	if m.selectedGrp >= len(m.groupsList) {
		m.selectedGrp = 0
	}
}

func (m *Model) GetVisibleEntries() []domain.HostEntry {
	var results []domain.HostEntry
	targetGroup := ""
	if m.selectedGrp > 0 && m.selectedGrp < len(m.groupsList) {
		targetGroup = m.groupsList[m.selectedGrp]
	}

	for _, g := range m.hostsFile.Groups {
		if targetGroup != "" && g.Name != targetGroup {
			continue
		}
		for _, e := range g.Entries {
			if m.searchQuery != "" {
				q := m.searchQuery
				if !containsFold(e.Hostname, q) && !containsFold(e.IP, q) && !containsFold(e.Comment, q) && !containsFoldInSlice(e.Aliases, q) {
					continue
				}
			}
			results = append(results, e)
		}
	}
	return results
}

func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

func (m *Model) SaveLocalConfig() error {
	return config.SaveConfig(m.configPath, m.hostsFile)
}
