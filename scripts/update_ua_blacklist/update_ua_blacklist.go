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

// run this script from the parent directory to update the user_agent_blacklist.go
func main() {
	log.Println("Updating User-Agent blacklist")
	list, err := os.Open("user_agent_blacklist.txt")

	if err != nil {
		log.Fatal(err)
	}

	defer func() {
		if err := list.Close(); err != nil {
			log.Fatal(err)
		}
	}()
	scanner := bufio.NewScanner(list)
	scanner.Split(bufio.ScanLines)
	entries := make(map[string]struct{})

	for scanner.Scan() {
		line := scanner.Text()

		if line != "" && !strings.HasPrefix(line, "#") {
			entries[strings.ToLower(line)] = struct{}{}
		}
	}

	ua := make([]string, 0, len(entries))

	for entry := range entries {
		ua = append(ua, entry)
	}

	sort.Strings(ua)
	var out strings.Builder
	out.WriteString(`package pirsch

var userAgentBlacklist = []string{
`)

	for _, entry := range ua {
		out.WriteString(fmt.Sprintf("\"%s\",\n", entry))
	}

	out.WriteString("}\n")

	if err := os.WriteFile("user_agent_blacklist.go", []byte(out.String()), 0644); err != nil {
		log.Fatal(err)
	}

	cmd := exec.Command("go", "fmt", "./...")

	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}

	log.Println("Done!")
}
