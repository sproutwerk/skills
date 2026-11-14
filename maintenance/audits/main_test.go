// Copyright (c) 2026 sproutwerk. SPDX-License-Identifier: MIT

package main

import (
	"strings"
	"testing"
	"time"
)

const page = `<html><script>var Snyk = "Fail"</script><main>
<a><span class="x">Gen Agent Trust Hub</span><span class="y">Pass</span></a>
<a><span>Socket</span><span>Warn</span></a>
<a><span>Snyk</span><span>FAIL</span></a>
<footer>Socket</footer><p>Docs</p></main></html>`

func TestParseStatuses(t *testing.T) {
	got := parseStatuses(page)
	want := map[string]string{"Gen Agent Trust Hub": "Pass", "Socket": "Warn", "Snyk": "Fail"}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for label, status := range want {
		if got[label] != status {
			t.Errorf("%s: got %q, want %q", label, got[label], status)
		}
	}
	if got := parseStatuses("<p>Snyk</p><p>Docs</p>"); len(got) != 0 {
		t.Errorf("unrecognised status parsed as %v", got)
	}
}

func TestParseAnalyzed(t *testing.T) {
	for _, label := range []string{"Analyzed", "Analyzed At"} {
		at, ok := parseAnalyzed("<div><span>" + label + "</span><span>Nov 14, 2026, 09:37 PM</span></div>")
		if want := time.Date(2026, 11, 14, 21, 37, 0, 0, time.UTC); !ok || !at.Equal(want) {
			t.Errorf("%s: got %v %v, want %v", label, at, ok, want)
		}
	}
	if _, ok := parseAnalyzed("<p>Analyzed</p><p>soon</p>"); ok {
		t.Error("unparseable time accepted")
	}
}

func TestStale(t *testing.T) {
	commit := time.Date(2026, 11, 14, 10, 0, 30, 0, time.UTC)
	for scan, want := range map[time.Time]bool{
		time.Date(2026, 11, 14, 9, 37, 0, 0, time.UTC): true,
		time.Date(2026, 11, 14, 10, 0, 0, 0, time.UTC): true,
		time.Date(2026, 11, 14, 10, 1, 0, 0, time.UTC): false,
	} {
		if got := stale(scan, commit); got != want {
			t.Errorf("scan %v: got %v, want %v", scan, got, want)
		}
	}
}

func TestTable(t *testing.T) {
	out := table("o/r", []string{"a", "b", "c"}, map[string]map[string]string{
		"a": {"Gen Agent Trust Hub": "Pass", "Socket": "Pass", "Snyk": "Warn"},
		"b": {},
	}, "2026-01-02")
	for _, want := range []string{
		"checked 2026-01-02",
		"| [a](https://skills.sh/o/r/a) | [![Pass](https://img.shields.io/badge/Pass-22c55e)](https://skills.sh/o/r/a/security/agent-trust-hub) | [![Pass](https://img.shields.io/badge/Pass-22c55e)](https://skills.sh/o/r/a/security/socket) | [![Warn](https://img.shields.io/badge/Warn-f59e0b)](https://skills.sh/o/r/a/security/snyk) |",
		"| [b](https://skills.sh/o/r/b) | pending | pending | pending |",
		"| c | not listed | not listed | not listed |",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("table lacks %q:\n%s", want, out)
		}
	}
}

func TestReplaceSection(t *testing.T) {
	got, err := replaceSection("top\n"+startMark+"\nold\n"+endMark+"\nend\n", "new\n")
	if err != nil {
		t.Fatal(err)
	}
	if want := "top\n" + startMark + "\nnew\n" + endMark + "\nend\n"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
	if _, err := replaceSection("no markers", "new\n"); err == nil {
		t.Error("missing markers accepted")
	}
}
