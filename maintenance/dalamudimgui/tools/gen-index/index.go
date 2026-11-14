package main

import (
	"cmp"
	"regexp"
	"slices"
	"strings"
	"unicode"
)

// ws matches the characters Python's str.isspace() accepts, so the
// generated lines stay identical to the ones the earlier Python script wrote.
const (
	wsChars = `\s\v\x1c-\x1f\x{85}\p{Z}`
	ws      = `[` + wsChars + `]`
)

var (
	classRe       = regexp.MustCompile(`\b(?:partial` + ws + `+)?(?:class|struct)` + ws + `+(\w+)`)
	classWordRe   = regexp.MustCompile(`\b(?:class|struct)\b`)
	methodStartRe = regexp.MustCompile(`^` + ws + `*public` + ws + `+(?:(?:static|virtual|abstract|sealed|unsafe|readonly|extern|override|new)` + ws + `+)*[\w<>\[\]\*\?,\.` + wsChars + `]+?` + ws + `+\w+(?:<[^>]+>)?` + ws + `*\(`)
	whereRe       = regexp.MustCompile(`^(` + ws + `*where` + ws + `+[^{=]+?)(?:\{|=>|$)`)
	spaceRunRe    = regexp.MustCompile(ws + `+`)
	publicRe      = regexp.MustCompile(`^public` + ws + `+`)
	methodNameRe  = regexp.MustCompile(`(\w+)(?:<[^>]+>)?` + ws + `*\(`)
)

// options control which signatures a file contributes to the index.
type options struct {
	classes      *regexp.Regexp // required: the full class or struct name must match
	exclude      *regexp.Regexp // optional: classes whose full name matches are dropped
	skipPointers bool           // drop signatures that take or return raw pointers
}

// row is one index line: "<Class>.<signature>\t// <file>".
type row struct {
	class, signature, file string
}

func (r row) String() string {
	return r.class + "." + r.signature + "\t// " + r.file
}

// splitLines mirrors Python's str.splitlines for the separators that occur in
// C# sources: LF, CRLF and a lone CR.
func splitLines(text string) []string {
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")
	text = strings.TrimSuffix(text, "\n")
	if text == "" {
		return nil
	}
	return strings.Split(text, "\n")
}

func isSpace(r rune) bool {
	return unicode.IsSpace(r) || (r >= 0x1c && r <= 0x1f)
}

func strip(s string) string {
	return strings.TrimFunc(s, isSpace)
}

// parenDepth is the number of unclosed parentheses in s.
func parenDepth(s string) int {
	return strings.Count(s, "(") - strings.Count(s, ")")
}

// extractRows returns every public method signature in one C# source file.
// file is the path written into the index; content is the file's text with any
// byte order mark removed.
func extractRows(file, content string, opts options) []row {
	var rows []row
	class := ""
	lines := splitLines(content)
	for i := 0; i < len(lines); i++ {
		line := lines[i]
		if m := classRe.FindStringSubmatch(line); m != nil && classWordRe.MatchString(line) {
			before, _, _ := strings.Cut(line, "class")
			before, _, _ = strings.Cut(before, "struct")
			trimmed := strip(line)
			if !strings.Contains(before, "(") && !strings.HasPrefix(trimmed, "//") && !strings.HasPrefix(trimmed, "*") {
				class = m[1]
			}
		}
		if !methodStartRe.MatchString(line) || strings.Contains(line, " class ") || strings.Contains(line, " delegate ") {
			continue
		}
		sig := strip(line)
		depth := parenDepth(sig)
		for depth > 0 && i+1 < len(lines) {
			i++
			s := strip(lines[i])
			sig += " " + s
			depth += parenDepth(s)
		}
		sig = normalizeSignature(sig)
		if opts.skipPointers && strings.Contains(sig, "*") {
			continue
		}
		if class == "" || !opts.classes.MatchString(class) {
			continue
		}
		if opts.exclude != nil && opts.exclude.MatchString(class) {
			continue
		}
		rows = append(rows, row{class: class, signature: sig, file: file})
	}
	return rows
}

// normalizeSignature cuts the body or expression after the parameter list,
// keeps a generic constraint clause, collapses whitespace and drops the
// leading "public".
func normalizeSignature(sig string) string {
	end := strings.LastIndex(sig, ")")
	depth := 0
	for idx, ch := range []byte(sig) {
		switch ch {
		case '(':
			depth++
		case ')':
			depth--
			if depth == 0 {
				end = idx
				goto found
			}
		}
	}
found:
	tail := sig[end+1:]
	sig = sig[:end+1]
	if m := whereRe.FindStringSubmatch(tail); m != nil {
		sig += strings.TrimRightFunc(m[1], isSpace)
	}
	sig = spaceRunRe.ReplaceAllString(sig, " ")
	sig = publicRe.ReplaceAllString(sig, "")
	return strings.ReplaceAll(sig, "( ", "(")
}

// sortKey orders rows by class, then by the lower-cased method name, then by
// the full signature, which is the order the index has always used.
func sortKey(r row) (string, string, string) {
	name := r.signature
	if m := methodNameRe.FindStringSubmatch(r.signature); m != nil {
		name = strings.ToLower(m[1])
	}
	return r.class, name, r.signature
}

func compareRows(a, b row) int {
	ac, an, as := sortKey(a)
	bc, bn, bs := sortKey(b)
	return cmp.Or(
		cmp.Compare(ac, bc),
		cmp.Compare(an, bn),
		cmp.Compare(as, bs),
		cmp.Compare(a.file, b.file),
	)
}

// sortedUnique removes duplicate rows and sorts the rest.
func sortedUnique(rows []row) []row {
	slices.SortFunc(rows, compareRows)
	return slices.CompactFunc(rows, func(a, b row) bool { return a == b })
}
