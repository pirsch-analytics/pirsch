package pirsch

import (
	"strings"
)

const (
	uaSystemLeftDelimiter  = '('
	uaSystemRightDelimiter = ')'
	systemDelimiter        = ";"
)

// UserAgent contains information extracted from the User-Agent header.
type UserAgent struct {
	Browser        string
	BrowserVersion string
	OS             string
	OSVersion      string
}

// ParseUserAgent parses given User-Agent header and returns the extracted information.
func ParseUserAgent(ua string) UserAgent {
	// TODO
	return UserAgent{}
}

func parseUserAgent(ua string) ([]string, []string) {
	// remove leading spaces, single and double quotes
	ua = strings.Trim(ua, " '\"")

	if ua == "" {
		return nil, nil
	}

	var system, versions []string
	systemStart := strings.IndexRune(ua, uaSystemLeftDelimiter)
	systemEnd := strings.IndexRune(ua, uaSystemRightDelimiter)

	if systemStart > -1 && systemEnd > -1 && systemStart < systemEnd {
		systemParts := strings.Split(ua[systemStart+1:systemEnd], systemDelimiter)
		versions = strings.Fields(ua[:systemStart] + " " + ua[systemEnd+1:])
		system = make([]string, 0, len(systemParts))

		for i := range systemParts {
			systemParts[i] = strings.TrimSpace(systemParts[i])

			if systemParts[i] != "" {
				system = append(system, systemParts[i])
			}
		}
	}

	if systemStart > -1 && systemEnd > -1 {
		versions = strings.Fields(ua[:systemStart] + " " + ua[systemEnd+1:])
	} else {
		versions = strings.Fields(ua)
	}

	return system, versions
}
