// Copyright (c) 2026 sproutwerk
// SPDX-License-Identifier: MIT

package main

import (
	"fmt"
	"math"
	"net/http"
	"net/http/httptest"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"testing"
)

func TestCoordinateConversion(t *testing.T) {
	tests := []struct {
		name               string
		world, offset, sf  float64
		wantMap            float64
		roundTripTolerance float64
	}{
		{"phantom village x", 29.5763, 0, 400, 6.717103662109375, 1e-9},
		{"phantom village z", 10.7985, 40, 400, 7.141962158203125, 1e-9},
		{"origin, scale 100", 0, 0, 100, 21.5, 1e-9},
		{"negative offset", 100, -70, 180, 12.98947482638889, 1e-9},
		{"odd scale", -733.21, 21.5, 95, 8.330846782483551, 1e-9},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := worldToMap(tt.world, tt.offset, tt.sf)
			if got != tt.wantMap {
				t.Errorf("worldToMap = %v, want %v", got, tt.wantMap)
			}
			if back := mapToWorld(got, tt.offset, tt.sf); math.Abs(back-tt.world) > tt.roundTripTolerance {
				t.Errorf("mapToWorld(worldToMap(%v)) = %v", tt.world, back)
			}
		})
	}
	if got := mapToWorld(13.37, -175, 200); got != 280.89658536585364 {
		t.Errorf("mapToWorld = %v", got)
	}
	// MapMarker pixel (1070, 911) on Map 855, SizeFactor 180.
	if got := formatFixed(1070*41/(1.8*2048)+1, 1); got != "12.9" {
		t.Errorf("pixel formula = %s, want 12.9", got)
	}
	if got := formatFixed(mapToWorld(6.72, 0, 400), 3); got != "29.721" {
		t.Errorf("mapToWorld = %s, want 29.721", got)
	}
}

func TestBuildURL(t *testing.T) {
	tests := []struct {
		name   string
		path   string
		params []param
		want   string
	}{
		{"no params", "/version", nil, "B/version"},
		{"nil dropped, empty kept", "/sheet/Item", []param{{"limit", new("5")}, {"after", nil}, {"fields", new("")}}, "B/sheet/Item?limit=5&fields="},
		{"plus, space and quotes", "/search", []param{kv("query", `+Name~"Hi Potion" -Name="x"`)}, "B/search?query=%2BName~%22Hi%20Potion%22%20-Name%3D%22x%22"},
		{"brackets and at", "/search", []param{kv("fields", "Item[].Name@lang(ja),A.B")}, "B/search?fields=Item%5B%5D.Name%40lang%28ja%29%2CA.B"},
		{"non-ascii", "/search", []param{kv("query", "Name@ja=\"新月\"")}, "B/search?query=Name%40ja%3D%22%E6%96%B0%E6%9C%88%22"},
		{"order kept", "/s", []param{kv("b", "1"), kv("a", "2")}, "B/s?b=1&a=2"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := buildURL("B", tt.path, tt.params...); got != tt.want {
				t.Errorf("got  %s\nwant %s", got, tt.want)
			}
		})
	}
}

func TestNumbersAndQuoting(t *testing.T) {
	nums := map[string]string{"1": "1", "-0": "0", "1.0": "1.0", "1.50": "1.5", "0.8333333": "0.8333333", "1e3": "1000.0", "1e16": "1e+16", "0.00001": "1e-05", "-29.5": "-29.5"}
	for in, want := range nums {
		if got := normalizeNumber(in); got != want {
			t.Errorf("normalizeNumber(%q) = %q, want %q", in, got, want)
		}
	}
	if got := formatFloat(math.Inf(-1)); got != "-Infinity" {
		t.Errorf("formatFloat(-Inf) = %q", got)
	}
	quotes := map[string]string{"abc": `'abc'`, "it's": `"it's"`, `a'b"c`: `'a\'b"c'`, "a\nb\x01": `'a\nb\x01'`, "その他": `'その他'`}
	for in, want := range quotes {
		if got := quoteString(in); got != want {
			t.Errorf("quoteString(%q) = %s, want %s", in, got, want)
		}
	}
	if got := jsonString("a\"b\\c\n<é>\x01"); got != `"a\"b\\c\n<é>\u0001"` {
		t.Errorf("jsonString = %s", got)
	}
}

func TestJSONRoundTrip(t *testing.T) {
	v, err := parseJSON([]byte(`{"b":1,"a":[{"x":1.0,"y":null},[],{}],"s":"é<","t":true}`))
	if err != nil {
		t.Fatal(err)
	}
	want := "{\n \"b\": 1,\n \"a\": [\n  {\n   \"x\": 1.0,\n   \"y\": null\n  },\n  [],\n  {}\n ],\n \"s\": \"é<\",\n \"t\": true\n}"
	if got := indented(v); got != want {
		t.Errorf("indented:\n%s\nwant:\n%s", got, want)
	}
	for _, bad := range []string{``, `{"a":1} x`, `{"a":`, `<html>`} {
		if _, err := parseJSON([]byte(bad)); err == nil {
			t.Errorf("parseJSON(%q) succeeded", bad)
		}
	}
}

func TestFlat(t *testing.T) {
	tests := []struct {
		name, in string
		want     []string // "key=cell"
	}{
		{"scalars keep order", `{"Name":"Potion","LevelItem":10,"IsX":true,"N":null}`, []string{"Name=Potion", "LevelItem=10", "IsX=true", "N="}},
		{"link with name", `{"Cat":{"value":44,"sheet":"C","row_id":44,"fields":{"Name":"Medicine","Order":3}}}`, []string{"Cat=44 'Medicine'", "Cat.Order=3"}},
		{"bare link", `{"Level":{"value":7}}`, []string{"Level=7"}},
		{"link with empty name", `{"P":{"value":1,"sheet":"PlaceName","row_id":1,"fields":{"Name":""}}}`, []string{"P=1"}},
		{"nested label", `{"T":{"value":5,"sheet":"T","row_id":5,"fields":{"PlaceName":{"value":9,"sheet":"P","row_id":9,"fields":{"Name":"Here"}}}}}`, []string{"T=5 \"9 'Here'\""}},
		{"icon", `{"Icon":{"id":65122,"path":"a.tex","path_hr1":"b.tex"}}`, []string{"Icon=icon 65122"}},
		{"array skips empty slots", `{"A":[{"value":0},{"value":3,"sheet":"S","row_id":3,"fields":{"Name":"n"}},{},{"value":0,"sheet":"S","row_id":0,"fields":{"Name":"zero"}}],"B":[1,2]}`,
			[]string{"A[1]=3 'n'", "A[3]=0 'zero'", "B[0]=1", "B[1]=2"}},
		{"plain object", `{"O":{"X":1.5,"Y":"a\tb\nc"}}`, []string{"O.X=1.5", `O.Y=a b\nc`}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v, err := parseJSON([]byte(tt.in))
			if err != nil {
				t.Fatal(err)
			}
			out := newObj()
			flat(v, "", out)
			var got []string
			for _, k := range out.keys {
				got = append(got, k+"="+cell(out.obj[k]))
			}
			if !slices.Equal(got, tt.want) {
				t.Errorf("got  %q\nwant %q", got, tt.want)
			}
		})
	}
}

func TestTable(t *testing.T) {
	v, _ := parseJSON([]byte(`[{"score":1.0,"sheet":"Item","row_id":4551,"fields":{"Name":"Potion"}},{"score":0.5,"sheet":"Item","row_id":2,"fields":{"Name":"新月鉱石","X":1},"transient":{"D":"t"}}]`))
	var b strings.Builder
	table(&b, v.arr, []string{"sheet", "row_id", "subrow_id", "score"})
	want := "sheet  row_id  subrow_id  score  Name    X  transient.D\n" +
		"Item   4551               1.0    Potion                \n" +
		"Item   2                  0.5    新月鉱石    1  t          \n"
	if b.String() != want {
		t.Errorf("got:\n%s\nwant:\n%s", b.String(), want)
	}
	b.Reset()
	table(&b, nil, nil)
	if b.String() != "(no rows)\n" {
		t.Errorf("empty table = %q", b.String())
	}
}

func TestPagerLimit(t *testing.T) {
	tests := []struct {
		pg        pager
		have      int
		limit     int
		satisfied bool
	}{
		{pager{n: 20}, 0, 20, false},
		{pager{n: 1200}, 0, 500, false},
		{pager{n: 1200}, 1000, 200, false},
		{pager{n: 1200}, 1200, 0, true},
		{pager{n: 20, all: true}, 5000, 500, false},
	}
	for _, tt := range tests {
		if got := tt.pg.limit(tt.have); got != tt.limit {
			t.Errorf("%+v.limit(%d) = %d, want %d", tt.pg, tt.have, got, tt.limit)
		}
		if got := tt.pg.satisfied(tt.have); got != tt.satisfied {
			t.Errorf("%+v.satisfied(%d) = %v", tt.pg, tt.have, got)
		}
	}
}

// harness runs the CLI against a test server and records each request.
type harness struct {
	t        *testing.T
	srv      *httptest.Server
	requests []url.Values
	paths    []string
}

func newHarness(t *testing.T, handler func(n int, r *http.Request) (status int, body string)) *harness {
	h := &harness{t: t}
	h.srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if ua := r.Header.Get("User-Agent"); ua != userAgent {
			t.Errorf("User-Agent = %q, want %q", ua, userAgent)
		}
		h.requests = append(h.requests, r.URL.Query())
		h.paths = append(h.paths, r.URL.Path)
		status, body := handler(len(h.requests), r)
		w.WriteHeader(status)
		fmt.Fprint(w, body)
	}))
	t.Cleanup(h.srv.Close)
	return h
}

func (h *harness) run(args ...string) (code int, stdout, stderr string) {
	var out, errb strings.Builder
	a := &app{base: h.srv.URL, client: h.srv.Client(), stdout: &out, stderr: &errb}
	code = a.run(args)
	return code, out.String(), errb.String()
}

const metaJSON = `"version":"v1","schema":"s1"`

func rowsJSON(from, count int, sub bool) string {
	parts := make([]string, count)
	for i := range count {
		if sub {
			parts[i] = fmt.Sprintf(`{"row_id":7,"subrow_id":%d,"fields":{}}`, from+i)
		} else {
			parts[i] = fmt.Sprintf(`{"row_id":%d,"fields":{}}`, from+i)
		}
	}
	return strings.Join(parts, ",")
}

func resultsJSON(count int) string {
	parts := make([]string, count)
	for i := range count {
		parts[i] = fmt.Sprintf(`{"score":1.0,"sheet":"Item","row_id":%d,"fields":{"Name":"n"}}`, i)
	}
	return strings.Join(parts, ",")
}

func TestSearchPagination(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		pages      []string // response bodies in order; the last repeats
		wantCode   int
		wantStderr string
		wantCalls  int
		check      func(t *testing.T, reqs []url.Values, stdout string)
	}{
		{
			name:       "single page, no cursor",
			args:       []string{"search", "Item", `Name="Potion"`, "-f", "Name"},
			pages:      []string{`{` + metaJSON + `,"results":[` + resultsJSON(1) + `]}`},
			wantStderr: "1 results  version=v1 schema=s1\n",
			wantCalls:  1,
			check: func(t *testing.T, reqs []url.Values, stdout string) {
				q := reqs[0]
				if q.Get("sheets") != "Item" || q.Get("query") != `Name="Potion"` || q.Get("limit") != "20" || q.Get("fields") != "Name" {
					t.Errorf("request = %v", q)
				}
				if q.Has("transient") || q.Has("language") || q.Has("cursor") {
					t.Errorf("unset options were sent: %v", q)
				}
				if want := "sheet  row_id  subrow_id  score  Name\nItem   0                  1.0    n   \n"; stdout != want {
					t.Errorf("stdout = %q", stdout)
				}
			},
		},
		{
			name:       "-n satisfied reports next cursor",
			args:       []string{"search", "Item", "q", "-n", "2", "-f", ""},
			pages:      []string{`{"next":"c1",` + metaJSON + `,"results":[` + resultsJSON(2) + `]}`},
			wantStderr: "2 results  version=v1 schema=s1 next=c1\n",
			wantCalls:  1,
			check: func(t *testing.T, reqs []url.Values, _ string) {
				if !reqs[0].Has("fields") || reqs[0].Get("fields") != "" || reqs[0].Get("limit") != "2" {
					t.Errorf("request = %v", reqs[0])
				}
			},
		},
		{
			name: "--all follows cursors",
			args: []string{"--json", "--lang", "ja", "search", "Item", "q", "--all"},
			pages: []string{
				`{"next":"c1",` + metaJSON + `,"results":[` + resultsJSON(2) + `]}`,
				`{"next":"c2",` + metaJSON + `,"results":[` + resultsJSON(2) + `]}`,
				`{` + metaJSON + `,"results":[` + resultsJSON(1) + `]}`,
			},
			wantStderr: "5 results  version=v1 schema=s1\n",
			wantCalls:  3,
			check: func(t *testing.T, reqs []url.Values, stdout string) {
				if reqs[1].Get("cursor") != "c1" || reqs[2].Get("cursor") != "c2" || reqs[1].Has("query") || reqs[1].Has("sheets") {
					t.Errorf("cursor requests = %v", reqs[1:])
				}
				if reqs[0].Has("fields") || reqs[0].Get("limit") != "500" || reqs[2].Get("language") != "ja" {
					t.Errorf("requests = %v", reqs)
				}
				v, err := parseJSON([]byte(stdout))
				if err != nil || len(v.arr) != 5 {
					t.Errorf("stdout is not a 5-element JSON array: %v", err)
				}
			},
		},
		{
			name:       "repeated cursor",
			args:       []string{"search", "Item", "q", "--all"},
			pages:      []string{`{"next":"c1",` + metaJSON + `,"results":[` + resultsJSON(1) + `]}`},
			wantCode:   1,
			wantStderr: "paging stalled: /search repeated cursor c1 after 2 results\n",
			wantCalls:  2,
		},
		{
			name: "page budget",
			args: []string{"search", "Item", "q", "--all", "--max-pages", "2"},
			pages: []string{
				`{"next":"c1",` + metaJSON + `,"results":[` + resultsJSON(1) + `]}`,
				`{"next":"c2",` + metaJSON + `,"results":[` + resultsJSON(1) + `]}`,
			},
			wantCode:   1,
			wantStderr: "page budget exhausted: --max-pages=2; request fewer rows or raise the budget\n",
			wantCalls:  2,
		},
		{
			name: "metadata change",
			args: []string{"search", "Item", "q", "--all"},
			pages: []string{
				`{"next":"c1",` + metaJSON + `,"results":[` + resultsJSON(1) + `]}`,
				`{"version":"v2","schema":"s1","results":[]}`,
			},
			wantCode:   1,
			wantStderr: "page metadata changed: ('v1', 's1') -> ('v2', 's1'); pin --game-version and --schema\n",
			wantCalls:  2,
		},
		{
			name:       "cursor without rows stops",
			args:       []string{"search", "Item", "q", "--all"},
			pages:      []string{`{"next":"c1",` + metaJSON + `,"results":[]}`},
			wantStderr: "0 results  version=v1 schema=s1\n",
			wantCalls:  1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHarness(t, func(n int, r *http.Request) (int, string) {
				if r.URL.Path != "/search" {
					t.Errorf("path = %s", r.URL.Path)
				}
				return 200, tt.pages[min(n, len(tt.pages))-1]
			})
			code, stdout, stderr := h.run(tt.args...)
			if code != tt.wantCode || stderr != tt.wantStderr || len(h.requests) != tt.wantCalls {
				t.Errorf("code=%d calls=%d stderr=%q", code, len(h.requests), stderr)
			}
			if tt.wantCode != 0 && stdout != "" {
				t.Errorf("partial output on failure: %q", stdout)
			}
			if tt.check != nil {
				tt.check(t, h.requests, stdout)
			}
		})
	}
}

func TestSheetPagination(t *testing.T) {
	t.Run("after cursor from last row, with subrows", func(t *testing.T) {
		h := newHarness(t, func(n int, r *http.Request) (int, string) {
			limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
			count := limit
			if n == 3 {
				count = 10
			}
			return 200, `{` + metaJSON + `,"rows":[` + rowsJSON((n-1)*500, count, true) + `]}`
		})
		code, _, stderr := h.run("sheet", "GilShopItem", "--after", "6:65535", "--all", "-f", "")
		if code != 0 || stderr != "1010 rows version=v1 schema=s1\n" {
			t.Fatalf("code=%d stderr=%q", code, stderr)
		}
		var afters []string
		for _, q := range h.requests {
			afters = append(afters, q.Get("after"))
		}
		if want := []string{"6:65535", "7:499", "7:999"}; !slices.Equal(afters, want) || h.paths[0] != "/sheet/GilShopItem" {
			t.Errorf("after = %v, want %v (path %s)", afters, want, h.paths[0])
		}
	})
	t.Run("-n spans pages and stops", func(t *testing.T) {
		h := newHarness(t, func(n int, r *http.Request) (int, string) {
			limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
			return 200, `{` + metaJSON + `,"rows":[` + rowsJSON((n-1)*500+1, limit, false) + `]}`
		})
		code, _, stderr := h.run("sheet", "Item", "-n", "600")
		if code != 0 || stderr != "600 rows version=v1 schema=s1\n" || len(h.requests) != 2 {
			t.Fatalf("code=%d calls=%d stderr=%q", code, len(h.requests), stderr)
		}
		if h.requests[0].Has("after") || h.requests[0].Get("limit") != "500" || h.requests[1].Get("after") != "500" || h.requests[1].Get("limit") != "100" {
			t.Errorf("requests = %v", h.requests)
		}
	})
	t.Run("stalled", func(t *testing.T) {
		h := newHarness(t, func(int, *http.Request) (int, string) {
			return 200, `{` + metaJSON + `,"rows":[` + rowsJSON(1, 500, false) + `]}`
		})
		code, stdout, stderr := h.run("sheet", "Item", "--all")
		if code != 1 || stdout != "" || stderr != "paging stalled: /sheet/Item page ended at after=500 again (1000 rows so far)\n" {
			t.Errorf("code=%d stderr=%q", code, stderr)
		}
	})
	t.Run("--rows is one request", func(t *testing.T) {
		h := newHarness(t, func(int, *http.Request) (int, string) {
			return 200, `{` + metaJSON + `,"rows":[` + rowsJSON(1, 2, false) + `]}`
		})
		code, stdout, stderr := h.run("--schema", "exdschema@2:latest", "sheet", "Item", "--rows", "1,2", "-t", "Description")
		q := h.requests[0]
		if code != 0 || stderr != "2 rows version=v1 schema=s1\n" || q.Get("rows") != "1,2" || q.Has("limit") || q.Get("transient") != "Description" || q.Get("schema") != "exdschema@2:latest" {
			t.Errorf("code=%d stderr=%q request=%v", code, stderr, q)
		}
		if want := "row_id  subrow_id\n1                \n2                \n"; stdout != want {
			t.Errorf("stdout = %q", stdout)
		}
	})
}

func TestCommands(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		status     int
		body       string
		wantCode   int
		wantStdout string
		wantStderr string // "*" skips the comparison; the URL is replaced by $URL
		wantPath   string
	}{
		{
			name: "row text", args: []string{"row", "Action", "3", "-f", "Name", "-t", "Description"},
			body:       `{` + metaJSON + `,"row_id":3,"fields":{"Name":"Sprint","Icon":{"id":104,"path":"p"},"L":[{"value":0},{"value":2,"sheet":"S","row_id":2,"fields":{}}]},"transient":{"Description":"a\nb"}}`,
			wantStdout: "Action/3\n  Name = \"Sprint\"\n  Icon = \"icon 104\"\n  L[1] = \"2\"\n  transient.Description = \"a\\nb\"\n",
			wantPath:   "/sheet/Action/3",
		},
		{
			name: "row subrow json", args: []string{"--json", "row", "GilShopItem", "7:1"},
			body:       `{"row_id":7,"subrow_id":1,"fields":{}}`,
			wantStdout: "{\n \"row_id\": 7,\n \"subrow_id\": 1,\n \"fields\": {}\n}\n",
			wantPath:   "/sheet/GilShopItem/7:1",
		},
		{
			name: "sheets filter", args: []string{"sheets", "shop"},
			body:       `{"sheets":[{"name":"GilShop"},{"name":"Item"},{"name":"SpecialShop"}]}`,
			wantStdout: "GilShop\nSpecialShop\n", wantStderr: "2 sheets\n", wantPath: "/sheet",
		},
		{
			name: "sheets json", args: []string{"--json", "--game-version", "7.2", "sheets"},
			body:       `{"sheets":[{"name":"Item"}]}`,
			wantStdout: "[\n \"Item\"\n]\n", wantStderr: "1 sheets\n",
		},
		{
			name: "version", args: []string{"version"},
			body:       `{"versions":[{"key":"k1","names":["7.57"]},{"key":"k2","names":["7.58","latest"]}]}`,
			wantStdout: "k1 7.57\nk2 7.58 latest\n", wantPath: "/version",
		},
		{
			name: "world2map", args: []string{"world2map", "1025", "29.5763", "10.7985"},
			body:       `{"fields":{"Id":"o6e1/00","OffsetX":0,"OffsetY":40,"SizeFactor":400,"PlaceName":{"value":4757,"sheet":"PlaceName","row_id":4757,"fields":{"Name":"Phantom Village"}}}}`,
			wantStdout: "map x=6.72 y=7.14\n",
			wantStderr: "Map 1025 o6e1/00 'Phantom Village' OffsetX=0 OffsetY=40 SizeFactor=400\n",
			wantPath:   "/sheet/Map/1025",
		},
		{
			name: "map2world json, negative input", args: []string{"--json", "map2world", "1", "-1.5", "7.14"},
			body:       `{"fields":{"Id":"a/00","OffsetX":0,"OffsetY":40,"SizeFactor":400,"PlaceName":{"value":0}}}`,
			wantStdout: "{\n \"map\": 1,\n \"id\": \"a/00\",\n \"offset_x\": 0,\n \"offset_y\": 40,\n \"size_factor\": 400,\n \"X\": -380.8780487804878,\n \"Z\": 10.700487804878037\n}\n",
			wantStderr: "Map 1 a/00 (no PlaceName) OffsetX=0 OffsetY=40 SizeFactor=400\n",
		},
		{
			name: "map without scale", args: []string{"world2map", "1", "0", "0"},
			body:     `{"fields":{"Id":"","OffsetX":0,"OffsetY":0,"SizeFactor":0}}`,
			wantCode: 1, wantStderr: "Map 1: SizeFactor is 0, so the map has no scale (unused/placeholder row?) -- cannot convert\n",
		},
		{
			name: "map schema rename", args: []string{"world2map", "1", "0", "0"},
			body:     `{"fields":{"SizeFactor":100,"Id":"x"}}`,
			wantCode: 1, wantStderr: "Map 1: no numeric OffsetX, OffsetY in response fields ['Id', 'SizeFactor'] (schema rename? compare references/schemas/Map.yml, or pin --schema)\n",
		},
		{
			name: "http error with message", args: []string{"row", "Item", "9"}, status: 404,
			body:     `{"code":404,"message":"not found: the Excel row Item/9:0 could not be found"}`,
			wantCode: 1, wantStderr: "HTTP 404: not found: the Excel row Item/9:0 could not be found\n  $URL/sheet/Item/9\n",
		},
		{
			name: "http error, plain body", args: []string{"version"}, status: 403, body: "error code: 1010",
			wantCode: 1, wantStderr: "HTTP 403: error code: 1010\n  $URL/version\n",
		},
		{
			name: "invalid json", args: []string{"version"}, body: "<html>",
			wantCode: 1, wantStderr: "*",
		},
		{name: "no command", args: nil, wantCode: 2, wantStderr: "*"},
		{name: "unknown command", args: []string{"nope"}, wantCode: 2, wantStderr: "*"},
		{name: "global flag after subcommand", args: []string{"version", "--json"}, wantCode: 2, wantStderr: "*"},
		{name: "missing positional", args: []string{"search", "Item"}, wantCode: 2, wantStderr: "*"},
		{name: "extra positional", args: []string{"row", "Item", "1", "2"}, wantCode: 2, wantStderr: "*"},
		{name: "-n zero", args: []string{"sheet", "Item", "-n", "0"}, wantCode: 2, wantStderr: "*"},
		{name: "-f without value", args: []string{"row", "Item", "1", "-f"}, wantCode: 2, wantStderr: "*"},
		{name: "bad float", args: []string{"world2map", "1", "x", "0"}, wantCode: 2, wantStderr: "*"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHarness(t, func(int, *http.Request) (int, string) {
				if tt.status != 0 {
					return tt.status, tt.body
				}
				return 200, tt.body
			})
			code, stdout, stderr := h.run(tt.args...)
			if code != tt.wantCode {
				t.Errorf("code = %d, want %d (stderr %q)", code, tt.wantCode, stderr)
			}
			if stdout != tt.wantStdout {
				t.Errorf("stdout:\n%q\nwant:\n%q", stdout, tt.wantStdout)
			}
			if want := strings.ReplaceAll(tt.wantStderr, "$URL", h.srv.URL); want != "*" && stderr != want {
				t.Errorf("stderr:\n%q\nwant:\n%q", stderr, want)
			}
			if tt.wantCode == 2 && len(h.requests) != 0 {
				t.Errorf("usage error still sent %d request(s)", len(h.requests))
			}
			if tt.wantPath != "" && (len(h.paths) == 0 || h.paths[0] != tt.wantPath) {
				t.Errorf("paths = %v, want %s", h.paths, tt.wantPath)
			}
		})
	}
}

func TestArgumentForms(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want map[string]string // expected query parameters
	}{
		{"flags before positionals", []string{"search", "-f", "Name", "Item", "q"}, map[string]string{"fields": "Name", "query": "q"}},
		{"equals and attached", []string{"search", "Item", "q", "--fields=A,B", "-tD", "-n5"}, map[string]string{"fields": "A,B", "transient": "D", "limit": "5"}},
		{"unique prefix", []string{"--game=7.2", "search", "Item", "q", "--trans", "D"}, map[string]string{"version": "7.2", "transient": "D"}},
		{"query with leading minus and a space", []string{"search", "Item", `-Name~"a b"`}, map[string]string{"query": `-Name~"a b"`}},
		{"double dash", []string{"search", "--", "Item", "-Name=1"}, map[string]string{"query": "-Name=1"}},
		{"plus query survives", []string{"search", "Item", `+Name~"x" +LevelItem>=1`}, map[string]string{"query": `+Name~"x" +LevelItem>=1`}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHarness(t, func(int, *http.Request) (int, string) {
				return 200, `{` + metaJSON + `,"results":[]}`
			})
			if code, _, stderr := h.run(tt.args...); code != 0 {
				t.Fatalf("code = %d, stderr %q", code, stderr)
			}
			for k, want := range tt.want {
				if got := h.requests[0].Get(k); got != want {
					t.Errorf("%s = %q, want %q", k, got, want)
				}
			}
		})
	}
	h := newHarness(t, func(int, *http.Request) (int, string) { return 500, "" })
	if code, stdout, _ := h.run("-h"); code != 0 || !strings.HasPrefix(stdout, "usage: xivapi") || len(h.requests) != 0 {
		t.Errorf("-h: code=%d stdout starts %q", code, stdout[:min(len(stdout), 20)])
	}
}
