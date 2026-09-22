package domain

import (
	"crypto/rand"
	"time"

	"github.com/oklog/ulid/v2"
)

// HostEntry represents an individual host mapping.
type HostEntry struct {
	ID       string   `yaml:"id" json:"id"`
	IP       string   `yaml:"ip" json:"ip"`
	Hostname string   `yaml:"hostname" json:"hostname"`
	Aliases  []string `yaml:"aliases,omitempty" json:"aliases,omitempty"`
	Enabled  bool     `yaml:"enabled" json:"enabled"`
	Comment  string   `yaml:"comment,omitempty" json:"comment,omitempty"`
}

// Group represents a logical collection of host entries.
type Group struct {
	Name    string      `yaml:"name" json:"name"`
	Entries []HostEntry `yaml:"entries" json:"entries"`
}

// HostsFile is the top-level YAML configuration schema.
type HostsFile struct {
	Version int     `yaml:"version" json:"version"`
	Groups  []Group `yaml:"groups" json:"groups"`
}

// NewULID generates a fresh, collision-resistant ULID string.
func NewULID() string {
	entropy := ulid.Monotonic(rand.Reader, 0)
	id := ulid.MustNew(ulid.Timestamp(time.Now()), entropy)
	return id.String()
}

// FindEntry searches for an entry by hostname or alias across all groups.
func (hf *HostsFile) FindEntry(name string) (*HostEntry, string) {
	for i := range hf.Groups {
		for j := range hf.Groups[i].Entries {
			entry := &hf.Groups[i].Entries[j]
			if entry.Hostname == name {
				return entry, hf.Groups[i].Name
			}
			for _, alias := range entry.Aliases {
				if alias == name {
					return entry, hf.Groups[i].Name
				}
			}
		}
	}
	return nil, ""
}

// GetOrCreateGroup returns a pointer to the group with the given name,
// creating it if it doesn't already exist.
func (hf *HostsFile) GetOrCreateGroup(groupName string) *Group {
	if groupName == "" {
		groupName = "default"
	}
	for i := range hf.Groups {
		if hf.Groups[i].Name == groupName {
			return &hf.Groups[i]
		}
	}
	newGroup := Group{
		Name:    groupName,
		Entries: []HostEntry{},
	}
	hf.Groups = append(hf.Groups, newGroup)
	return &hf.Groups[len(hf.Groups)-1]
}

// TotalEntries returns the count of all entries across all groups.
func (hf *HostsFile) TotalEntries() int {
	count := 0
	for _, g := range hf.Groups {
		count += len(g.Entries)
	}
	return count
}

// EnabledEntries returns the count of enabled entries across all groups.
func (hf *HostsFile) EnabledEntries() int {
	count := 0
	for _, g := range hf.Groups {
		for _, e := range g.Entries {
			if e.Enabled {
				count++
			}
		}
	}
	return count
}
