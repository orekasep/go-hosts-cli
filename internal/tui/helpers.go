package tui

import "strings"

func containsFold(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}

func containsFoldInSlice(slice []string, substr string) bool {
	q := strings.ToLower(substr)
	for _, item := range slice {
		if strings.Contains(strings.ToLower(item), q) {
			return true
		}
	}
	return false
}
