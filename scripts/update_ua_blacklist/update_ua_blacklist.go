package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"sort"
	"strings"
)

// run this script from the root directory to update the blacklist.go
func main() {
	log.Println("Updating User-Agent blacklist")
	list, err := os.Open("pkg/tracker/ua/blacklist.txt")

	if err != nil {
		log.Fatal(err)
	}

	defer func() {
		if err := list.Close(); err != nil {
			log.Fatal(err)
		}
	}()

	regexList, err := os.Open("pkg/tracker/ua/blacklist_regex.txt")

	if err != nil {
		log.Fatal(err)
	}

	defer func() {
		if err := regexList.Close(); err != nil {
			log.Fatal(err)
		}
	}()

	entries := readLines(list)
	regexEntries := readLines(regexList)
	ua := make([]string, 0, len(entries))
	uaRegex := make([]string, 0, len(regexEntries))

	for entry := range entries {
		ua = append(ua, strings.ReplaceAll(entry, `"`, `\"`))
	}

	for entry := range regexEntries {
		uaRegex = append(uaRegex, fmt.Sprintf(`regexp.MustCompile("%s")`, strings.ReplaceAll(entry, `"`, `\"`)))
	}

	sort.Strings(ua)
	sort.Strings(uaRegex)
	var out strings.Builder
	out.WriteString(`package ua

import "regexp"

// Blacklist is a list of User-Agents to ignore.
var Blacklist = []string{
`)

	for _, entry := range ua {
		out.WriteString(fmt.Sprintf("\"%s\",\n", entry))
	}

	out.WriteString("}\n\n")
	out.WriteString(`// RegexBlacklist is a list of User-Agents to ignore.
var RegexBlacklist = []*regexp.Regexp{
`)

	for _, entry := range uaRegex {
		out.WriteString(fmt.Sprintf("%s,\n", entry))
	}

	out.WriteString("}\n")

	if err := os.WriteFile("pkg/tracker/ua/blacklist.go", []byte(out.String()), 0644); err != nil {
		log.Fatal(err)
	}

	cmd := exec.Command("go", "fmt", "./...")

	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}

	log.Println("Done!")
}

func readLines(list *os.File) map[string]struct{} {
	scanner := bufio.NewScanner(list)
	scanner.Split(bufio.ScanLines)
	entries := make(map[string]struct{})

	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, `"`) && strings.HasSuffix(line, `"`) {
			line = line[1 : len(line)-1]
		}

		if line != "" && !strings.HasPrefix(line, "#") {
			entries[strings.ToLower(line)] = struct{}{}
		}
	}

	return entries
}
