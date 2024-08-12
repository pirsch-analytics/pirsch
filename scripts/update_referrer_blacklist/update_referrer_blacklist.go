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
	log.Println("Updating referrer blacklist")
	list, err := os.Open("pkg/tracker/referrer/blacklist.txt")

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

		if strings.HasPrefix(line, `"`) && strings.HasSuffix(line, `"`) {
			line = line[1 : len(line)-1]
		}

		if line != "" && !strings.HasPrefix(line, "#") {
			entries[strings.ToLower(line)] = struct{}{}
		}
	}

	keywords := make([]string, 0, len(entries))

	for entry := range entries {
		keywords = append(keywords, strings.ReplaceAll(entry, `"`, `\"`))
	}

	sort.Strings(keywords)
	var out strings.Builder
	out.WriteString(`package referrer

// Blacklist is a list of referrer keywords to ignore.
var Blacklist = []string{
`)

	for _, entry := range keywords {
		out.WriteString(fmt.Sprintf("\"%s\",\n", entry))
	}

	out.WriteString("}\n")

	if err := os.WriteFile("pkg/tracker/referrer/blacklist.go", []byte(out.String()), 0644); err != nil {
		log.Fatal(err)
	}

	cmd := exec.Command("go", "fmt", "./...")

	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}

	log.Println("Done!")
}
