package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

type domain struct {
	Domains []string `json:"domains"`
}

type database map[string]map[string]domain

// run this script from the root directory to generate the referrer_list.go
func main() {
	log.Println("Downloading database")
	resp, err := http.Get("https://s3-eu-west-1.amazonaws.com/snowplow-hosted-assets/third-party/referer-parser/referers-latest.json")

	if err != nil {
		log.Fatal(err)
	}

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		log.Fatal(err)
	}

	log.Println("Processing database")
	var db database

	if err := json.Unmarshal(body, &db); err != nil {
		log.Fatal(err)
	}

	log.Println("Writing list")
	var out strings.Builder
	out.WriteString(`package pirsch

var (
	referrerGroups = map[string]string{
`)

	for key := range db {
		for name, domains := range db[key] {
			for _, domain := range domains.Domains {
				out.WriteString(fmt.Sprintf(`"%s": "%s",`, strings.ToLower(domain), name))
				out.WriteRune('\n')
			}
		}
	}

	out.WriteString(`}
)`)

	if err := os.WriteFile("referrer_list.go", []byte(out.String()), 0644); err != nil {
		log.Fatal(err)
	}

	log.Println("Formatting code")
	cmd := exec.Command("go", "fmt", "./...")

	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}

	log.Println("Done!")
}
