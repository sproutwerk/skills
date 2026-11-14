// Copyright (c) 2026 sproutwerk. SPDX-License-Identifier: MIT

// Command audits rewrites the audit table in the repository README from the
// results skills.sh publishes for each skill.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"html"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

const (
	startMark = "<!-- audits:start -->"
	endMark   = "<!-- audits:end -->"
)

// auditors lists the label skills.sh shows for each audit and the path
// segment of its detail page.
var auditors = []struct{ label, slug string }{
	{"Gen Agent Trust Hub", "agent-trust-hub"},
	{"Socket", "socket"},
	{"Snyk", "snyk"},
}

// colors are the Tailwind 500 shades skills.sh uses for each status.
var colors = map[string]string{"Pass": "22c55e", "Warn": "f59e0b", "Fail": "ef4444"}

var (
	hidden = regexp.MustCompile(`(?is)<(script|style|svg)\b.*?</(script|style|svg)>`)
	tag    = regexp.MustCompile(`(?s)<[^>]*>`)
)

// parseStatuses returns the status shown beside each auditor label on a
// skill page. An auditor with no recognised status is absent from the map.
func parseStatuses(page string) map[string]string {
	text := tag.ReplaceAllString(hidden.ReplaceAllString(page, "\n"), "\n")
	var lines []string
	for _, line := range strings.Split(html.UnescapeString(text), "\n") {
		if line = strings.TrimSpace(line); line != "" {
			lines = append(lines, line)
		}
	}
	found := map[string]string{}
	for i := 0; i+1 < len(lines); i++ {
		for _, a := range auditors {
			if lines[i] != a.label {
				continue
			}
			switch status := strings.ToLower(lines[i+1]); status {
			case "pass", "warn", "fail":
				found[a.label] = strings.ToUpper(status[:1]) + status[1:]
			}
		}
	}
	return found
}

// table renders the README section for the given skills and their statuses.
// A skill with no entry in results is reported as not listed.
func table(repo string, skills []string, results map[string]map[string]string, date string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Results from [skills.sh](https://skills.sh/%s), checked %s.\n\n", repo, date)
	b.WriteString("| Skill |")
	for _, a := range auditors {
		b.WriteString(" " + a.label + " |")
	}
	b.WriteString("\n| --- |" + strings.Repeat(" --- |", len(auditors)) + "\n")
	for _, skill := range skills {
		statuses, listed := results[skill]
		if !listed {
			fmt.Fprintf(&b, "| %s |%s\n", skill, strings.Repeat(" not listed |", len(auditors)))
			continue
		}
		fmt.Fprintf(&b, "| [%s](https://skills.sh/%s/%s) |", skill, repo, skill)
		for _, a := range auditors {
			status, ok := statuses[a.label]
			if !ok {
				b.WriteString(" pending |")
				continue
			}
			fmt.Fprintf(&b, " [![%s](https://img.shields.io/badge/%s-%s)](https://skills.sh/%s/%s/security/%s) |",
				status, status, colors[status], repo, skill, a.slug)
		}
		b.WriteString("\n")
	}
	return b.String()
}

// replaceSection swaps the text between the two markers for body.
func replaceSection(readme, body string) (string, error) {
	start := strings.Index(readme, startMark)
	end := strings.Index(readme, endMark)
	if start < 0 || end < start {
		return "", fmt.Errorf("README lacks %s and %s markers", startMark, endMark)
	}
	return readme[:start+len(startMark)] + "\n" + body + readme[end:], nil
}

func skillNames(path string) ([]string, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var config struct {
		Groupings []struct {
			Skills []string `json:"skills"`
		} `json:"groupings"`
	}
	if err := json.Unmarshal(raw, &config); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	var names []string
	for _, group := range config.Groupings {
		names = append(names, group.Skills...)
	}
	if len(names) == 0 {
		return nil, fmt.Errorf("%s lists no skills", path)
	}
	return names, nil
}

func fetch(client *http.Client, url string) (string, int, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", 0, err
	}
	req.Header.Set("User-Agent", "skills-audit-table/1")
	resp, err := client.Do(req)
	if err != nil {
		return "", 0, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	return string(body), resp.StatusCode, err
}

func run() error {
	repo := flag.String("repo", "sproutwerk/skills", "owner/name on skills.sh")
	config := flag.String("config", "../../skills.sh.json", "path to skills.sh.json")
	readmePath := flag.String("readme", "../../README.md", "README to rewrite")
	printOnly := flag.Bool("print", false, "print the table and leave the README unchanged")
	flag.Parse()

	skills, err := skillNames(*config)
	if err != nil {
		return err
	}
	client := &http.Client{Timeout: 60 * time.Second}
	results := map[string]map[string]string{}
	for _, skill := range skills {
		page, status, err := fetch(client, "https://skills.sh/"+*repo+"/"+skill)
		if err != nil {
			return fmt.Errorf("%s: %w", skill, err)
		}
		switch {
		case status == http.StatusNotFound:
			continue
		case status != http.StatusOK:
			return fmt.Errorf("%s: HTTP %d", skill, status)
		}
		results[skill] = parseStatuses(page)
	}
	body := table(*repo, skills, results, time.Now().UTC().Format("2006-01-02"))
	if *printOnly {
		fmt.Print(body)
		return nil
	}
	readme, err := os.ReadFile(*readmePath)
	if err != nil {
		return err
	}
	updated, err := replaceSection(string(readme), body)
	if err != nil {
		return err
	}
	return os.WriteFile(*readmePath, []byte(updated), 0o644)
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "audits:", err)
		os.Exit(1)
	}
}
