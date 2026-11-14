package main

import (
	"regexp"
	"strings"
	"unicode"
)

// wsChars lists the characters Python's str.isspace() accepts, so the
// generated lines stay identical to the ones the earlier Python script wrote.
const (
	wsChars = `\s\v\x1c-\x1f\x{85}\p{Z}`
	ws      = `[` + wsChars + `]`
)

var (
	startRe      = regexp.MustCompile(`^` + ws + `*(?:\[[^\]]*\]` + ws + `*)*(?:public|protected)` + ws)
	methodImplRe = regexp.MustCompile(`\[MethodImpl[^\]]*\]` + ws + `*`)
	spaceRunRe   = regexp.MustCompile(ws + `+`)
)

func isSpace(r rune) bool {
	return unicode.IsSpace(r) || (r >= 0x1c && r <= 0x1f)
}

func strip(s string) string {
	return strings.TrimFunc(s, isSpace)
}

// declaration is one public or protected declaration with the line it starts on.
type declaration struct {
	line int // 1-based
	text string
}

// declarations finds every public or protected declaration in a C# source.
// A parameter list that continues on following lines is joined into one line,
// [MethodImpl] attributes are dropped and runs of whitespace collapse to one
// space.
func declarations(content string) []declaration {
	lines := strings.Split(content, "\n")
	var out []declaration
	for i := 0; i < len(lines); i++ {
		if !startRe.MatchString(lines[i]) {
			continue
		}
		start := i
		decl := strip(lines[i])
		for strings.Count(decl, "(") > strings.Count(decl, ")") && i+1 < len(lines) {
			i++
			decl += " " + strip(lines[i])
		}
		decl = methodImplRe.ReplaceAllString(decl, "")
		decl = spaceRunRe.ReplaceAllString(decl, " ")
		out = append(out, declaration{line: start + 1, text: decl})
	}
	return out
}
