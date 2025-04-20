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
	log.Println("Updating browser blacklist")
	list, err := os.Open("pkg/tracker/ua/browser_blacklist.txt")

	if err != nil {
		log.Fatal(err)
	}

	defer func() {
		if err := list.Close(); err != nil {
			log.Fatal(err)
		}
	}()

	entries := readLines(list)
	browser := make([]string, 0, len(entries))

	for entry := range entries {
		browser = append(browser, strings.ReplaceAll(entry, `"`, `\"`))
	}

	sort.Strings(browser)
	var out strings.Builder
	out.WriteString(`package ua

// BrowserBlacklist is a list of User-Agents to ignore.
var BrowserBlacklist = []string{
`)

	for _, entry := range browser {
		out.WriteString(fmt.Sprintf("\"%s\",\n", entry))
	}

	out.WriteString("}\n")

	if err := os.WriteFile("pkg/tracker/ua/browser_blacklist.go", []byte(out.String()), 0644); err != nil {
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
