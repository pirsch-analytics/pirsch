package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"sort"
	"strings"
)

const (
	snowplowList = "pkg/tracker/referrer/mapping-snowplow.json"
	mappingList  = "pkg/tracker/referrer/mapping.json"
)

type domain struct {
	Domains []string `json:"domains"`
}

type list map[string]map[string]domain

// run this script from the root directory to generate the list.go
func main() {
	downloadSnowplowList()
	fromSnowplow := loadList(snowplowList)
	mapping := loadList(mappingList)
	groups := make(map[string]string)
	addGroups(groups, fromSnowplow)
	addGroups(groups, mapping)
	keys := getKeys(groups)
	writeList(groups, keys)
	formatCode()
	log.Println("Done!")
}

func downloadSnowplowList() {
	log.Println("Downloading snowplow list")
	resp, err := http.Get("https://s3-eu-west-1.amazonaws.com/snowplow-hosted-assets/third-party/referer-parser/referers-latest.json")

	if err != nil {
		log.Fatal(err)
	}

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		log.Fatal(err)
	}

	if err := os.WriteFile(snowplowList, body, 0644); err != nil {
		log.Fatal(err)
	}
}

func loadList(file string) list {
	content, err := os.ReadFile(file)

	if err != nil {
		log.Fatal(err)
	}

	var l list

	if err := json.Unmarshal(content, &l); err != nil {
		log.Fatal(err)
	}

	return l
}

func addGroups(groups map[string]string, l list) {
	for key := range l {
		for name, domains := range l[key] {
			for _, domain := range domains.Domains {
				domain = strings.ToLower(domain)

				if strings.HasPrefix(domain, "www.") {
					domain = domain[4:]
				}

				groups[domain] = name
			}
		}
	}
}

func getKeys(groups map[string]string) []string {
	keys := make([]string, 0, len(groups))

	for k := range groups {
		keys = append(keys, k)
	}

	sort.Strings(keys)
	return keys
}

func writeList(groups map[string]string, keys []string) {
	log.Println("Writing list")
	var out strings.Builder
	out.WriteString(`package referrer

var (
	groups = map[string]string{
`)

	for _, key := range keys {
		out.WriteString(fmt.Sprintf(`"%s": "%s",`, key, groups[key]))
		out.WriteRune('\n')
	}

	out.WriteString(`}
)`)

	if err := os.WriteFile("pkg/tracker/referrer/list.go", []byte(out.String()), 0644); err != nil {
		log.Fatal(err)
	}
}

func formatCode() {
	log.Println("Formatting code")
	cmd := exec.Command("go", "fmt", "./...")

	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}
}
