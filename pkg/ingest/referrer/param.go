package referrer

import (
	"net/http"
	"strings"
)

var (
	queryParams = []queryParam{
		{"ref", false},
		{"referer", false},
		{"referrer", false},
		{"source", true},
		{"utm_source", true},
	}
)

type queryParam struct {
	Param        string
	PreferHeader bool
}

func fromHeaderOrQuery(request *http.Request) string {
	fromHeader := strings.TrimSpace(request.Header.Get("Referer"))

	if index := strings.IndexAny(fromHeader, "\n\r"); index > 0 {
		fromHeader = strings.TrimSpace(fromHeader[:index])
	}

	for _, param := range queryParams {
		referrer := request.URL.Query().Get(param.Param)

		if referrer != "" && (!param.PreferHeader || param.PreferHeader && fromHeader == "") {
			return referrer
		}
	}

	return fromHeader
}
