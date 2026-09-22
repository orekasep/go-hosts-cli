package etchosts

import (
	"bytes"
	"errors"
)

// BeginMarker and EndMarker define the demarcation for hostcli entries in system hosts.
const (
	BeginMarker = "# BEGIN HOSTCLI"
	EndMarker   = "# END HOSTCLI"
)

var (
	ErrMissingEndMarker     = errors.New("hostcli markers malformed: missing END")
	ErrMissingBeginMarker   = errors.New("hostcli markers malformed: missing BEGIN")
	ErrDuplicateBeginMarker = errors.New("hostcli markers malformed: duplicate BEGIN")
	ErrDuplicateEndMarker   = errors.New("hostcli markers malformed: duplicate END")
	ErrEndBeforeBegin       = errors.New("hostcli markers malformed: END before BEGIN")
)

// ExtractManagedBlock splits etcHosts content into:
// - preamble: everything before the BEGIN HOSTCLI line
// - managed: the BEGIN line through the END line inclusive
// - suffix: everything after the END line
func ExtractManagedBlock(etcHosts []byte) (preamble, managed, suffix []byte, err error) {
	beginBytes := []byte(BeginMarker)
	endBytes := []byte(EndMarker)

	var beginStart, endLineEndExclusive int = -1, -1
	var beginCount, endCount int

	i := 0
	for i <= len(etcHosts) {
		lineStart := i
		j := i
		for j < len(etcHosts) && etcHosts[j] != '\n' {
			j++
		}
		line := etcHosts[lineStart:j]
		trimmed := rtrimASCIIWhitespace(line)

		var nextLineStart int
		if j < len(etcHosts) {
			nextLineStart = j + 1
		} else {
			nextLineStart = j
		}

		if bytes.Equal(trimmed, beginBytes) {
			beginCount++
			if beginCount == 1 {
				beginStart = lineStart
			}
		}
		if bytes.Equal(trimmed, endBytes) {
			endCount++
			if endCount == 1 {
				endLineEndExclusive = nextLineStart
			}
		}

		if j == len(etcHosts) {
			break
		}
		i = nextLineStart
	}

	switch {
	case beginCount == 0 && endCount == 0:
		return etcHosts, nil, nil, nil
	case beginCount > 1:
		return nil, nil, nil, ErrDuplicateBeginMarker
	case endCount > 1:
		return nil, nil, nil, ErrDuplicateEndMarker
	case beginCount == 1 && endCount == 0:
		return nil, nil, nil, ErrMissingEndMarker
	case beginCount == 0 && endCount == 1:
		return nil, nil, nil, ErrMissingBeginMarker
	case endLineEndExclusive <= beginStart:
		return nil, nil, nil, ErrEndBeforeBegin
	}

	preamble = etcHosts[:beginStart]
	managed = etcHosts[beginStart:endLineEndExclusive]
	suffix = etcHosts[endLineEndExclusive:]
	return preamble, managed, suffix, nil
}

// ReplaceManagedBlock replaces any existing hostcli block in etcHosts with newManaged.
// If no existing block is found, newManaged is appended cleanly at the end.
func ReplaceManagedBlock(etcHosts []byte, newManaged []byte) ([]byte, error) {
	preamble, managed, suffix, err := ExtractManagedBlock(etcHosts)
	if err != nil {
		return nil, err
	}

	if managed == nil {
		if len(preamble) == 0 {
			out := make([]byte, len(newManaged))
			copy(out, newManaged)
			return out, nil
		}
		needsSep := preamble[len(preamble)-1] != '\n'
		size := len(preamble) + len(newManaged)
		if needsSep {
			size++
		}
		out := make([]byte, 0, size)
		out = append(out, preamble...)
		if needsSep {
			out = append(out, '\n')
		}
		out = append(out, newManaged...)
		return out, nil
	}

	out := make([]byte, 0, len(preamble)+len(newManaged)+len(suffix))
	out = append(out, preamble...)
	out = append(out, newManaged...)
	out = append(out, suffix...)
	return out, nil
}

// StripManagedBlock cleanly removes the managed block from etcHosts.
func StripManagedBlock(etcHosts []byte) ([]byte, error) {
	preamble, managed, suffix, err := ExtractManagedBlock(etcHosts)
	if err != nil {
		return nil, err
	}
	if managed == nil {
		return etcHosts, nil // nothing to strip
	}

	trimmedPreamble := bytes.TrimRight(preamble, "\r\n")
	trimmedSuffix := bytes.TrimLeft(suffix, "\r\n")

	if len(trimmedPreamble) == 0 && len(trimmedSuffix) == 0 {
		return []byte(""), nil
	}
	if len(trimmedSuffix) == 0 {
		return append(trimmedPreamble, '\n'), nil
	}
	if len(trimmedPreamble) == 0 {
		return append(trimmedSuffix, '\n'), nil
	}

	var out bytes.Buffer
	out.Write(trimmedPreamble)
	out.WriteByte('\n')
	out.Write(trimmedSuffix)
	if !bytes.HasSuffix(trimmedSuffix, []byte("\n")) {
		out.WriteByte('\n')
	}
	return out.Bytes(), nil
}

func rtrimASCIIWhitespace(line []byte) []byte {
	n := len(line)
	for n > 0 {
		c := line[n-1]
		if c == ' ' || c == '\t' || c == '\r' || c == '\v' || c == '\f' {
			n--
			continue
		}
		break
	}
	return line[:n]
}
