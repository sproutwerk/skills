// Copyright (c) 2026 sproutwerk
// SPDX-License-Identifier: MIT

// Command xivapi is a small standard-library client for XIVAPI v2
// (https://v2.xivapi.com). Run it with -h for the subcommands.
package main

import (
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"
)

const usageLine = `usage: xivapi [-h] [--json] [--lang LANG] [--game-version GAME_VERSION]
              [--schema SCHEMA]
              {search,sheet,row,sheets,version,world2map,map2world} ...`

const helpText = usageLine + `

Tiny stdlib-only client for XIVAPI v2 (https://v2.xivapi.com).

Subcommands:
  search  SHEET[,SHEET..] 'QUERY' [-f FIELDS] [-n N|--all] [-t TRANSIENT]
  sheet   SHEET [-f FIELDS] [-n N|--all] [--after ROW[:SUB]] [--rows 1,2,3:0]
  row     SHEET ROW[:SUB] [-f FIELDS] [-t TRANSIENT]
  sheets  [SUBSTRING]
  version
  world2map MAP_ID X Z        # world (X, Z) -> map (x, y)
  map2world MAP_ID MX MY      # map (x, y) -> world (X, Z)

Global flags go BEFORE the subcommand: --json, --lang en|ja|de|fr,
--game-version 7.2|latest|<key>, --schema exdschema@2:rev:<sha>.

--json prints valid JSON on stdout for every subcommand (counts and
version/schema metadata go to stderr):
  search, sheet   JSON array of result/row objects, all pages concatenated
                  (the per-page envelope with version/schema/next is dropped)
  row             the response object as the API returned it
  sheets          JSON array of sheet names (after the optional filter)
  version         the API's "versions" array of {"key","names"} objects
  world2map       {"map","id","offset_x","offset_y","size_factor","x","y"}
  map2world       {"map","id","offset_x","offset_y","size_factor","X","Z"}

` + "`-f ''`" + ` sends an empty fields= (row ids and empty field objects only);
omitting -f sends no fields= at all (full recursive expansion -- avoid).
--all keeps every page in memory and prints once the last page arrived.
Pagination stops with an error at --max-pages (default 1000), on repeated
cursors, or if version/schema changes between pages; partial rows are not emitted.

Examples (from the skill directory, "xivapi" = go run -C scripts/xivapi .):
  xivapi search Item 'Name="Ecliptic Arcanite"' -f Name
  xivapi --json search Item '+Name~"Tomestone" -Name~"Allagan"' -f Name --all
  xivapi search ENpcBase 'ENpcData[]=1770925' -f ''
  xivapi row SpecialShop 1770925 -f 'Name,Item[].Item[].Name,Item[].CurrencyCost'
  xivapi sheet GilShopItem --after 262156:65535 -n 40 -f Item.Name
  xivapi --json world2map 1025 29.5763 10.7985

Set XIVAPI_BASE to override the base URL.

options:
  -h, --help            show this help message and exit
  --json                JSON on stdout for every subcommand (shapes in the
                        description above); goes before the subcommand
  --lang LANG           language: en ja de fr (default en)
  --game-version GAME_VERSION
                        game version name or key (default latest)
  --schema SCHEMA       schema specifier, e.g. exdschema@2:rev:<sha>

row options (search, sheet, row):
  -f, --fields FIELDS   comma-separated field filter (ALWAYS set this; '' =
                        ids only)
  -t, --transient TRANSIENT
                        field filter for the *Transient row

count options (search, sheet):
  -n N                  max rows (>= 1, default 20)
  --all                 fetch every page (held in memory, printed at the end)
  --max-pages MAX_PAGES
                        page budget (default 1000); fail without partial
                        output when exhausted

sheet options:
  --after AFTER         start after ROW[:SUB]
  --rows ROWS           explicit rows, e.g. 1,2,3:0
`

const (
	defaultBase = "https://v2.xivapi.com/api"
	// Cloudflare in front of xivapi answers 403 for some library default
	// User-Agents (Python-urllib/x.y is one), so always send our own.
	userAgent = "ffxivskills-xivapi/0.2"
	// pageSize is the server cap per page for /search and /sheet (default 100).
	pageSize = 500
)

var commands = []string{"search", "sheet", "row", "sheets", "version", "world2map", "map2world"}

// usageError is a command-line mistake; the process exits 2.
type usageError struct{ msg string }

func (e *usageError) Error() string { return e.msg }

func usagef(format string, a ...any) error { return &usageError{fmt.Sprintf(format, a...)} }

// errHelp asks for the help text; the process exits 0.
var errHelp = errors.New("help requested")

// app carries everything a subcommand needs.
type app struct {
	base   string
	client *http.Client
	stdout io.Writer
	stderr io.Writer

	json        bool
	lang        *string
	gameVersion *string
	schema      *string
}

func main() {
	a := &app{
		base:   defaultBase,
		client: newClient(),
		stdout: os.Stdout,
		stderr: os.Stderr,
	}
	if b, ok := os.LookupEnv("XIVAPI_BASE"); ok {
		a.base = b
	}
	os.Exit(a.run(os.Args[1:]))
}

// newClient bounds connecting and waiting for response headers at 60 s each.
func newClient() *http.Client {
	t := http.DefaultTransport.(*http.Transport).Clone()
	t.DialContext = (&net.Dialer{Timeout: 60 * time.Second, KeepAlive: 30 * time.Second}).DialContext
	t.TLSHandshakeTimeout = 60 * time.Second
	t.ResponseHeaderTimeout = 60 * time.Second
	return &http.Client{Transport: t}
}

// run executes one command line and returns the exit code: 0 on success,
// 1 on request or data errors, 2 on usage errors.
func (a *app) run(args []string) int {
	err := a.dispatch(args)
	if err == nil {
		return 0
	}
	if errors.Is(err, errHelp) {
		fmt.Fprint(a.stdout, helpText)
		return 0
	}
	if ue, ok := errors.AsType[*usageError](err); ok {
		fmt.Fprintf(a.stderr, "%s\nxivapi: error: %s\n", usageLine, ue.msg)
		return 2
	}
	fmt.Fprintln(a.stderr, err)
	return 1
}

func (a *app) dispatch(args []string) error {
	global := &parser{flags: []*flagSpec{
		{names: []string{"--json"}, set: func(string) error { a.json = true; return nil }},
		stringFlag(&a.lang, "--lang"),
		stringFlag(&a.gameVersion, "--game-version"),
		stringFlag(&a.schema, "--schema"),
	}}
	rest, err := global.parseLeading(args)
	if err != nil {
		return err
	}
	if len(rest) == 0 {
		return usagef("the following arguments are required: cmd")
	}
	cmd, rest := rest[0], rest[1:]

	var (
		fields, transient, after, rows *string
		n, maxPages                    = 20, 1000
		all                            bool
	)
	rowFlags := []*flagSpec{
		stringFlag(&fields, "-f", "--fields"),
		stringFlag(&transient, "-t", "--transient"),
	}
	countFlags := []*flagSpec{
		positiveIntFlag(&n, "-n"),
		{names: []string{"--all"}, set: func(string) error { all = true; return nil }},
		positiveIntFlag(&maxPages, "--max-pages"),
	}

	switch cmd {
	case "search":
		p := &parser{flags: slices.Concat(rowFlags, countFlags), positionals: []string{"sheets", "query"}}
		pos, err := p.parse(rest)
		if err != nil {
			return err
		}
		return a.search(pos[0], pos[1], fields, transient, pager{n: n, all: all, maxPages: maxPages})
	case "sheet":
		p := &parser{
			flags:       slices.Concat(rowFlags, countFlags, []*flagSpec{stringFlag(&after, "--after"), stringFlag(&rows, "--rows")}),
			positionals: []string{"sheet"},
		}
		pos, err := p.parse(rest)
		if err != nil {
			return err
		}
		return a.sheet(pos[0], fields, transient, after, rows, pager{n: n, all: all, maxPages: maxPages})
	case "row":
		p := &parser{flags: rowFlags, positionals: []string{"sheet", "row"}}
		pos, err := p.parse(rest)
		if err != nil {
			return err
		}
		return a.row(pos[0], pos[1], fields, transient)
	case "sheets":
		p := &parser{positionals: []string{"filter"}, optional: 1}
		pos, err := p.parse(rest)
		if err != nil {
			return err
		}
		filter := ""
		if len(pos) > 0 {
			filter = pos[0]
		}
		return a.sheets(filter)
	case "version":
		if _, err := (&parser{}).parse(rest); err != nil {
			return err
		}
		return a.version()
	case "world2map", "map2world":
		names := []string{"map", "x", "z"}
		if cmd == "map2world" {
			names = []string{"map", "mx", "my"}
		}
		pos, err := (&parser{positionals: names}).parse(rest)
		if err != nil {
			return err
		}
		mapID, err := strconv.Atoi(strings.TrimSpace(pos[0]))
		if err != nil {
			return usagef("argument %s: invalid int value: %s", names[0], quoteString(pos[0]))
		}
		var c [2]float64
		for i := range c {
			if c[i], err = parseFloat(pos[i+1]); err != nil {
				return usagef("argument %s: invalid float value: %s", names[i+1], quoteString(pos[i+1]))
			}
		}
		return a.convert(cmd == "world2map", mapID, c[0], c[1])
	}
	return usagef("argument cmd: invalid choice: %s (choose from %s)", quoteString(cmd), strings.Join(commands, ", "))
}

// ---------- command line ----------

type flagSpec struct {
	names      []string
	takesValue bool
	set        func(string) error
}

func stringFlag(dst **string, names ...string) *flagSpec {
	return &flagSpec{names: names, takesValue: true, set: func(s string) error {
		*dst = new(s)
		return nil
	}}
}

func positiveIntFlag(dst *int, names ...string) *flagSpec {
	return &flagSpec{names: names, takesValue: true, set: func(s string) error {
		n, err := strconv.Atoi(strings.TrimSpace(s))
		if err != nil {
			return fmt.Errorf("not an integer: %s", quoteString(s))
		}
		if n < 1 {
			return fmt.Errorf("must be >= 1 (got %d); use --all for every row", n)
		}
		*dst = n
		return nil
	}}
}

func parseFloat(s string) (float64, error) {
	s = strings.TrimSpace(s)
	if strings.ContainsAny(s, "xXpP_") {
		return 0, strconv.ErrSyntax
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil && !errors.Is(err, strconv.ErrRange) {
		return 0, err
	}
	return f, nil
}

// parser handles flags mixed freely with positionals. Long flags take
// "--name value" or "--name=value" and may be shortened to a unique prefix;
// short flags take "-f value", "-fvalue" or "-f=value". "--" ends the flags.
type parser struct {
	flags       []*flagSpec
	positionals []string // names, for error messages
	optional    int      // how many trailing positionals may be omitted
}

var negativeNumber = regexp.MustCompile(`^-\d+$|^-\d*\.\d+$`)

// lookup resolves a flag token to its spec and any value attached to it.
func (p *parser) lookup(tok string) (spec *flagSpec, attached *string, err error) {
	find := func(name string) *flagSpec {
		for _, f := range p.flags {
			if slices.Contains(f.names, name) {
				return f
			}
		}
		return nil
	}
	if f := find(tok); f != nil {
		return f, nil, nil
	}
	name, val, hasVal := strings.Cut(tok, "=")
	if hasVal {
		if f := find(name); f != nil {
			return f, &val, nil
		}
	}
	if strings.HasPrefix(tok, "--") {
		var matches []*flagSpec
		var matched []string
		for _, f := range p.flags {
			for _, fn := range f.names {
				if strings.HasPrefix(fn, "--") && strings.HasPrefix(fn, name) {
					matches = append(matches, f)
					matched = append(matched, fn)
				}
			}
		}
		switch len(matches) {
		case 0:
			return nil, nil, nil
		case 1:
			if hasVal {
				return matches[0], &val, nil
			}
			return matches[0], nil, nil
		}
		return nil, nil, usagef("ambiguous option: %s could match %s", name, strings.Join(matched, ", "))
	}
	// -fVALUE
	if f := find(tok[:2]); f != nil && f.takesValue {
		return f, new(tok[2:]), nil
	}
	return nil, nil, nil
}

// isFlag reports whether tok reads as a flag rather than a positional.
// Negative numbers and anything containing a space are positionals.
func (p *parser) isFlag(tok string) bool {
	if len(tok) < 2 || tok[0] != '-' {
		return false
	}
	if tok == "-h" || tok == "--help" {
		return true
	}
	if f, _, err := p.lookup(tok); f != nil || err != nil {
		return true
	}
	return !negativeNumber.MatchString(tok) && !strings.Contains(tok, " ")
}

// parse consumes every argument and returns the positionals.
func (p *parser) parse(args []string) ([]string, error) {
	pos, _, err := p.scan(args, false)
	if err != nil {
		return nil, err
	}
	if required := len(p.positionals) - p.optional; len(pos) < required {
		return nil, usagef("the following arguments are required: %s", strings.Join(p.positionals[len(pos):required], ", "))
	}
	if len(pos) > len(p.positionals) {
		return nil, usagef("unrecognized arguments: %s", strings.Join(pos[len(p.positionals):], " "))
	}
	return pos, nil
}

// parseLeading consumes flags up to the first positional and returns the
// remaining arguments, positional included.
func (p *parser) parseLeading(args []string) ([]string, error) {
	_, rest, err := p.scan(args, true)
	return rest, err
}

func (p *parser) scan(args []string, stopAtPositional bool) (pos, rest []string, err error) {
	for i := 0; i < len(args); i++ {
		tok := args[i]
		if tok == "--" {
			if stopAtPositional {
				return pos, args[i+1:], nil
			}
			return append(pos, args[i+1:]...), nil, nil
		}
		if !p.isFlag(tok) {
			if stopAtPositional {
				return pos, args[i:], nil
			}
			pos = append(pos, tok)
			continue
		}
		if tok == "-h" || tok == "--help" {
			return nil, nil, errHelp
		}
		spec, attached, err := p.lookup(tok)
		if err != nil {
			return nil, nil, err
		}
		if spec == nil {
			return nil, nil, usagef("unrecognized arguments: %s", tok)
		}
		label := strings.Join(spec.names, "/")
		var val string
		switch {
		case !spec.takesValue && attached != nil:
			return nil, nil, usagef("argument %s: ignored explicit argument %s", label, quoteString(*attached))
		case !spec.takesValue:
		case attached != nil:
			val = *attached
		case i+1 < len(args) && args[i+1] != "--" && !p.isFlag(args[i+1]):
			i++
			val = args[i]
		default:
			return nil, nil, usagef("argument %s: expected one argument", label)
		}
		if err := spec.set(val); err != nil {
			return nil, nil, usagef("argument %s: %s", label, err)
		}
	}
	return pos, nil, nil
}

// ---------- HTTP ----------

// param is one query parameter. A nil value means "option not given" and is
// left out; an empty value is kept because it has a defined meaning for
// some parameters (fields= returns ids and empty field objects).
type param struct {
	key string
	val *string
}

func kv(key, val string) param { return param{key, &val} }

// buildURL joins base, path and the given parameters in order. Values are
// percent-encoded, with a space as %20 so that a literal "+" in a query
// stays unambiguous.
func buildURL(base, path string, params ...param) string {
	var b strings.Builder
	b.WriteString(base)
	b.WriteString(path)
	sep := byte('?')
	for _, p := range params {
		if p.val == nil {
			continue
		}
		b.WriteByte(sep)
		sep = '&'
		b.WriteString(escape(p.key))
		b.WriteByte('=')
		b.WriteString(escape(*p.val))
	}
	return b.String()
}

func escape(s string) string {
	return strings.ReplaceAll(url.QueryEscape(s), "+", "%20")
}

// common returns the parameters every data request accepts.
func (a *app) common() []param {
	return []param{{"language", a.lang}, {"version", a.gameVersion}, {"schema", a.schema}}
}

// get fetches one JSON document.
func (a *app) get(path string, params ...param) (*value, error) {
	u := buildURL(a.base, path, params...)
	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return nil, fmt.Errorf("request failed: %v\n  %s", err, u)
	}
	req.Header.Set("User-Agent", userAgent)
	resp, err := a.client.Do(req)
	if err != nil {
		if ue, ok := errors.AsType[*url.Error](err); ok {
			err = ue.Err
		}
		return nil, fmt.Errorf("request failed: %v\n  %s", err, u)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("request failed: %v\n  %s", err, u)
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		msg := strings.ToValidUTF8(string(body), "�")
		if d, err := parseJSON(body); err == nil {
			if m := d.get("message"); m != nil {
				msg = text(m)
			}
		}
		return nil, fmt.Errorf("HTTP %d: %s\n  %s", resp.StatusCode, msg, u)
	}
	d, err := parseJSON(body)
	if err != nil {
		head := strings.ToValidUTF8(string(body[:min(len(body), 120)]), "�")
		return nil, fmt.Errorf("invalid JSON in response: %v\n  starts with: %s\n  %s", err, quoteString(head), u)
	}
	return d, nil
}

// list returns the array member key of a response.
func list(d *value, key string) ([]*value, error) {
	v := d.get(key)
	if v == nil || v.kind != kArr {
		return nil, fmt.Errorf("unexpected response: no %q array", key)
	}
	return v.arr, nil
}

// ---------- output ----------

func (a *app) printJSON(v *value) { fmt.Fprintln(a.stdout, indented(v)) }

func (a *app) emit(rows []*value, idCols []string, meta string) {
	if a.json {
		a.printJSON(arrVal(rows))
	} else {
		table(a.stdout, rows, idCols)
	}
	fmt.Fprintln(a.stderr, meta)
}

// ---------- pagination ----------

type pager struct {
	n        int
	all      bool
	maxPages int
}

// limit is the page size to request once have rows are in hand; 0 or less
// means the request is satisfied.
func (p pager) limit(have int) int {
	if p.all {
		return pageSize
	}
	return min(pageSize, p.n-have)
}

func (p pager) satisfied(have int) bool { return !p.all && have >= p.n }

func (p pager) exhausted() error {
	return fmt.Errorf("page budget exhausted: --max-pages=%d; request fewer rows or raise the budget", p.maxPages)
}

// pageMeta is the game version and schema a page was read with.
type pageMeta struct{ version, schema string }

func (m pageMeta) String() string { return "version=" + m.version + " schema=" + m.schema }

func (m pageMeta) tuple() string {
	return "(" + quoteString(m.version) + ", " + quoteString(m.schema) + ")"
}

// pageMetadata reads a page's version and schema and fails when they differ
// from the previous page's.
func pageMetadata(d *value, previous *pageMeta) (*pageMeta, error) {
	v, s := d.get("version"), d.get("schema")
	if v == nil || s == nil {
		return nil, errors.New(`unexpected response: no "version"/"schema"`)
	}
	current := &pageMeta{text(v), text(s)}
	if previous != nil && *current != *previous {
		return nil, fmt.Errorf("page metadata changed: %s -> %s; pin --game-version and --schema", previous.tuple(), current.tuple())
	}
	return current, nil
}

// each calls fetch once per page until it reports done, and fails when the
// page budget runs out first.
func (p pager) each(fetch func() (done bool, err error)) error {
	for range p.maxPages {
		if done, err := fetch(); err != nil || done {
			return err
		}
	}
	return p.exhausted()
}

// ---------- subcommands ----------

func (a *app) search(sheets, query string, fields, transient *string, pg pager) error {
	var (
		results []*value
		cursor  string
		meta    string
		md      *pageMeta
		seen    = map[string]bool{}
	)
	err := pg.each(func() (bool, error) {
		limit := pg.limit(len(results))
		if limit <= 0 {
			return true, nil
		}
		var params []param
		if cursor != "" {
			seen[cursor] = true
			params = []param{kv("cursor", cursor)}
		} else {
			params = []param{kv("sheets", sheets), kv("query", query)}
		}
		params = append(params, kv("limit", strconv.Itoa(limit)), param{"fields", fields}, param{"transient", transient})
		d, err := a.get("/search", append(params, a.common()...)...)
		if err != nil {
			return false, err
		}
		if md, err = pageMetadata(d, md); err != nil {
			return false, err
		}
		page, err := list(d, "results")
		if err != nil {
			return false, err
		}
		results = append(results, page...)
		meta = md.String()
		next := d.get("next")
		// A cursor without rows has nothing to page towards.
		if !truthy(next) || len(page) == 0 {
			return true, nil
		}
		if seen[text(next)] {
			return false, fmt.Errorf("paging stalled: /search repeated cursor %s after %d results", text(next), len(results))
		}
		cursor = text(next)
		if pg.satisfied(len(results)) {
			meta += " next=" + cursor
			return true, nil
		}
		return false, nil
	})
	if err != nil {
		return err
	}
	a.emit(results, []string{"sheet", "row_id", "subrow_id", "score"}, fmt.Sprintf("%d results  %s", len(results), meta))
	return nil
}

func (a *app) sheet(sheet string, fields, transient, after, rowList *string, pg pager) error {
	path := "/sheet/" + sheet
	idCols := []string{"row_id", "subrow_id"}
	tail := append([]param{{"fields", fields}, {"transient", transient}}, a.common()...)
	if rowList != nil && *rowList != "" {
		d, err := a.get(path, append([]param{{"rows", rowList}}, tail...)...)
		if err != nil {
			return err
		}
		rows, err := list(d, "rows")
		if err != nil {
			return err
		}
		md, err := pageMetadata(d, nil)
		if err != nil {
			return err
		}
		a.emit(rows, idCols, fmt.Sprintf("%d rows %s", len(rows), md))
		return nil
	}
	var (
		rows []*value
		meta string
		md   *pageMeta
		seen = map[string]bool{}
	)
	err := pg.each(func() (bool, error) {
		limit := pg.limit(len(rows))
		if limit <= 0 {
			return true, nil
		}
		if after != nil {
			seen[*after] = true
		}
		d, err := a.get(path, append([]param{kv("limit", strconv.Itoa(limit)), {"after", after}}, tail...)...)
		if err != nil {
			return false, err
		}
		if md, err = pageMetadata(d, md); err != nil {
			return false, err
		}
		page, err := list(d, "rows")
		if err != nil {
			return false, err
		}
		rows = append(rows, page...)
		meta = md.String()
		if len(page) < limit {
			return true, nil
		}
		last := page[len(page)-1]
		next := text(last.get("row_id"))
		if sub := last.get("subrow_id"); !absent(sub) {
			next += ":" + text(sub)
		}
		if seen[next] {
			return false, fmt.Errorf("paging stalled: %s page ended at after=%s again (%d rows so far)", path, next, len(rows))
		}
		after = &next
		return pg.satisfied(len(rows)), nil
	})
	if err != nil {
		return err
	}
	a.emit(rows, idCols, fmt.Sprintf("%d rows %s", len(rows), meta))
	return nil
}

func (a *app) row(sheet, row string, fields, transient *string) error {
	d, err := a.get("/sheet/"+sheet+"/"+row, append([]param{{"fields", fields}, {"transient", transient}}, a.common()...)...)
	if err != nil {
		return err
	}
	if a.json {
		a.printJSON(d)
		return nil
	}
	f := d.get("fields")
	if d.get("row_id") == nil || f == nil {
		return errors.New(`unexpected response: no "row_id"/"fields"`)
	}
	head := sheet + "/" + text(d.get("row_id"))
	if sub := d.get("subrow_id"); !absent(sub) {
		head += ":" + text(sub)
	}
	fmt.Fprintln(a.stdout, head)
	out := newObj()
	flat(f, "", out)
	for _, k := range out.keys {
		fmt.Fprintf(a.stdout, "  %s = %s\n", k, compact(out.obj[k]))
	}
	if tr := d.get("transient"); truthy(tr) {
		out = newObj()
		flat(tr, "", out)
		for _, k := range out.keys {
			fmt.Fprintf(a.stdout, "  transient.%s = %s\n", k, compact(out.obj[k]))
		}
	}
	return nil
}

func (a *app) sheets(filter string) error {
	d, err := a.get("/sheet", param{"version", a.gameVersion})
	if err != nil {
		return err
	}
	all, err := list(d, "sheets")
	if err != nil {
		return err
	}
	needle := strings.ToLower(filter)
	var names []*value
	var lines []string
	for _, s := range all {
		name := s.get("name")
		if name == nil {
			return errors.New(`unexpected response: sheet entry without "name"`)
		}
		if strings.Contains(strings.ToLower(text(name)), needle) {
			names = append(names, name)
			lines = append(lines, text(name))
		}
	}
	if a.json {
		a.printJSON(arrVal(names))
	} else {
		fmt.Fprintln(a.stdout, strings.Join(lines, "\n"))
	}
	fmt.Fprintf(a.stderr, "%d sheets\n", len(names))
	return nil
}

func (a *app) version() error {
	d, err := a.get("/version")
	if err != nil {
		return err
	}
	versions, err := list(d, "versions")
	if err != nil {
		return err
	}
	if a.json {
		a.printJSON(arrVal(versions))
		return nil
	}
	for _, v := range versions {
		names, err := list(v, "names")
		if err != nil || v.get("key") == nil {
			return errors.New(`unexpected response: version entry without "key"/"names"`)
		}
		parts := make([]string, len(names))
		for i, n := range names {
			parts[i] = text(n)
		}
		fmt.Fprintln(a.stdout, text(v.get("key")), strings.Join(parts, " "))
	}
	return nil
}

// mapRow fetches a Map row and checks the numbers the conversion needs.
// PlaceName is a link and may come back as just {"value": N} for sparse or
// missing targets, so its Name is only a label.
func (a *app) mapRow(mapID int) (f *value, offsetX, offsetY, sizeFactor float64, err error) {
	d, err := a.get("/sheet/Map/"+strconv.Itoa(mapID), append([]param{kv("fields", "Id,OffsetX,OffsetY,SizeFactor,PlaceName.Name")}, a.common()...)...)
	if err != nil {
		return nil, 0, 0, 0, err
	}
	f = d.get("fields")
	if !truthy(f) || f.kind != kObj {
		f = newObj()
	}
	var missing []string
	nums := map[string]float64{}
	for _, k := range []string{"OffsetX", "OffsetY", "SizeFactor"} {
		n, ok := f.get(k).float()
		if !ok {
			missing = append(missing, k)
		}
		nums[k] = n
	}
	if len(missing) > 0 {
		have := make([]*value, len(f.keys))
		for i, k := range slices.Sorted(slices.Values(f.keys)) {
			have[i] = strVal(k)
		}
		return nil, 0, 0, 0, fmt.Errorf("Map %d: no numeric %s in response fields %s (schema rename? compare references/schemas/Map.yml, or pin --schema)",
			mapID, strings.Join(missing, ", "), quoted(arrVal(have)))
	}
	if nums["SizeFactor"] == 0 {
		return nil, 0, 0, 0, fmt.Errorf("Map %d: SizeFactor is 0, so the map has no scale (unused/placeholder row?) -- cannot convert", mapID)
	}
	shown := "(no PlaceName)"
	if label := f.get("PlaceName").get("fields").get("Name"); truthy(label) {
		shown = quoted(label)
	}
	fmt.Fprintf(a.stderr, "Map %d %s %s OffsetX=%s OffsetY=%s SizeFactor=%s\n",
		mapID, text(f.get("Id")), shown, text(f.get("OffsetX")), text(f.get("OffsetY")), text(f.get("SizeFactor")))
	return f, nums["OffsetX"], nums["OffsetY"], nums["SizeFactor"], nil
}

// worldToMap converts one world coordinate (X or Z) to the map coordinate
// shown in game. The float64 conversions keep each product a separate
// operation so results are the same on every architecture.
func worldToMap(world, offset, sizeFactor float64) float64 {
	c := sizeFactor / 100
	pixel := float64((world+offset)*c) + 1024
	return float64(pixel*41)/float64(c*2048) + 1
}

// mapToWorld is the inverse of worldToMap.
func mapToWorld(m, offset, sizeFactor float64) float64 {
	c := sizeFactor / 100
	pixel := float64((m-1)*c*2048)/41 - 1024
	return pixel/c - offset
}

func (a *app) convert(toMap bool, mapID int, first, second float64) error {
	f, offsetX, offsetY, sizeFactor, err := a.mapRow(mapID)
	if err != nil {
		return err
	}
	conv, keys := mapToWorld, [2]string{"X", "Z"}
	if toMap {
		conv, keys = worldToMap, [2]string{"x", "y"}
	}
	x, y := conv(first, offsetX, sizeFactor), conv(second, offsetY, sizeFactor)
	if a.json {
		id := f.get("Id")
		if id == nil {
			id = null
		}
		a.printJSON(newObj().
			set("map", numVal(strconv.Itoa(mapID))).
			set("id", id).
			set("offset_x", f.get("OffsetX")).
			set("offset_y", f.get("OffsetY")).
			set("size_factor", f.get("SizeFactor")).
			set(keys[0], numVal(formatFloat(x))).
			set(keys[1], numVal(formatFloat(y))))
		return nil
	}
	if toMap {
		fmt.Fprintf(a.stdout, "map x=%s y=%s\n", formatFixed(x, 2), formatFixed(y, 2))
	} else {
		fmt.Fprintf(a.stdout, "world X=%s Z=%s\n", formatFixed(x, 3), formatFixed(y, 3))
	}
	return nil
}
