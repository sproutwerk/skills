// Copyright (c) 2026 sproutwerk
// SPDX-License-Identifier: MIT

package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"slices"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

// kind is the JSON type of a value.
type kind uint8

const (
	kNull kind = iota
	kBool
	kNum
	kStr
	kArr
	kObj
)

// value is a decoded JSON value that keeps object keys in document order.
// The table and row views list fields in the order the API returned them,
// which a Go map would lose.
type value struct {
	kind kind
	b    bool
	s    string // string content, or the number literal
	arr  []*value
	keys []string
	obj  map[string]*value
}

var null = &value{kind: kNull}

func strVal(s string) *value       { return &value{kind: kStr, s: s} }
func numVal(lit string) *value     { return &value{kind: kNum, s: lit} }
func arrVal(items []*value) *value { return &value{kind: kArr, arr: items} }
func newObj() *value               { return &value{kind: kObj, obj: map[string]*value{}} }

// set stores a key. A repeated key keeps its first position and takes the
// last value.
func (v *value) set(k string, x *value) *value {
	if _, ok := v.obj[k]; !ok {
		v.keys = append(v.keys, k)
	}
	v.obj[k] = x
	return v
}

// get returns the member k of an object, or nil when v is not an object or
// has no such key.
func (v *value) get(k string) *value {
	if v == nil || v.kind != kObj {
		return nil
	}
	return v.obj[k]
}

// absent reports whether v is missing or JSON null.
func absent(v *value) bool { return v == nil || v.kind == kNull }

// truthy reports whether v is neither missing, null, false, zero nor empty.
func truthy(v *value) bool {
	if v == nil {
		return false
	}
	switch v.kind {
	case kBool:
		return v.b
	case kNum:
		f, ok := v.float()
		return !ok || f != 0
	case kStr:
		return v.s != ""
	case kArr:
		return len(v.arr) > 0
	case kObj:
		return len(v.keys) > 0
	}
	return false
}

// float returns the numeric value of a number.
func (v *value) float() (float64, bool) {
	if v == nil || v.kind != kNum {
		return 0, false
	}
	f, err := strconv.ParseFloat(v.s, 64)
	if err != nil && !errors.Is(err, strconv.ErrRange) {
		return 0, false
	}
	return f, true
}

// isZero reports whether v equals 0 (the number zero or false).
func isZero(v *value) bool {
	if v == nil {
		return false
	}
	if v.kind == kBool {
		return !v.b
	}
	f, ok := v.float()
	return ok && f == 0
}

// parseJSON decodes exactly one JSON document.
func parseJSON(data []byte) (*value, error) {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()
	v, err := decodeValue(dec)
	if err != nil {
		return nil, err
	}
	if _, err := dec.Token(); !errors.Is(err, io.EOF) {
		return nil, fmt.Errorf("extra data after the JSON document at offset %d", dec.InputOffset())
	}
	return v, nil
}

func decodeValue(dec *json.Decoder) (*value, error) {
	tok, err := dec.Token()
	if err != nil {
		if errors.Is(err, io.EOF) {
			return nil, io.ErrUnexpectedEOF
		}
		return nil, err
	}
	switch t := tok.(type) {
	case nil:
		return null, nil
	case bool:
		return &value{kind: kBool, b: t}, nil
	case json.Number:
		return numVal(normalizeNumber(string(t))), nil
	case string:
		return strVal(t), nil
	case json.Delim:
		switch t {
		case '[':
			v := &value{kind: kArr}
			for dec.More() {
				item, err := decodeValue(dec)
				if err != nil {
					return nil, err
				}
				v.arr = append(v.arr, item)
			}
			_, err := dec.Token()
			return v, err
		case '{':
			v := newObj()
			for dec.More() {
				kt, err := dec.Token()
				if err != nil {
					return nil, err
				}
				key, ok := kt.(string)
				if !ok {
					return nil, fmt.Errorf("object key is not a string at offset %d", dec.InputOffset())
				}
				item, err := decodeValue(dec)
				if err != nil {
					return nil, err
				}
				v.set(key, item)
			}
			_, err := dec.Token()
			return v, err
		}
	}
	return nil, fmt.Errorf("unexpected token %v at offset %d", tok, dec.InputOffset())
}

// normalizeNumber rewrites a JSON number literal into the canonical form
// used for output: integers stay as written, anything with a fraction or an
// exponent becomes the shortest float that round-trips (1.50 -> 1.5, 1.0
// stays 1.0).
func normalizeNumber(lit string) string {
	if !strings.ContainsAny(lit, ".eE") {
		if lit == "-0" {
			return "0"
		}
		return lit
	}
	f, err := strconv.ParseFloat(lit, 64)
	if err != nil && !errors.Is(err, strconv.ErrRange) {
		return lit
	}
	return formatFloat(f)
}

// formatFloat prints a float the way the JSON output writes it: shortest
// round-trip digits, always with a fraction or an exponent, and
// NaN/Infinity for non-finite values.
func formatFloat(f float64) string {
	switch {
	case math.IsNaN(f):
		return "NaN"
	case math.IsInf(f, 1):
		return "Infinity"
	case math.IsInf(f, -1):
		return "-Infinity"
	}
	if a := math.Abs(f); f != 0 && (a < 1e-4 || a >= 1e16) {
		return strconv.FormatFloat(f, 'e', -1, 64)
	}
	s := strconv.FormatFloat(f, 'f', -1, 64)
	if !strings.Contains(s, ".") {
		s += ".0"
	}
	return s
}

// formatFixed prints f with prec decimals; non-finite values are nan, inf
// and -inf.
func formatFixed(f float64, prec int) string {
	switch {
	case math.IsNaN(f):
		return "nan"
	case math.IsInf(f, 1):
		return "inf"
	case math.IsInf(f, -1):
		return "-inf"
	}
	return strconv.FormatFloat(f, 'f', prec, 64)
}

// text is the plain rendering of a scalar: strings unquoted, booleans as
// True/False, null as None. Containers fall back to quoted.
func text(v *value) string {
	if v == nil {
		return "None"
	}
	if v.kind == kStr {
		return v.s
	}
	return quoted(v)
}

// quoted renders v for diagnostics and link labels, with strings in single
// quotes.
func quoted(v *value) string {
	if v == nil {
		return "None"
	}
	switch v.kind {
	case kBool:
		if v.b {
			return "True"
		}
		return "False"
	case kNum:
		return v.s
	case kStr:
		return quoteString(v.s)
	case kArr:
		parts := make([]string, len(v.arr))
		for i, item := range v.arr {
			parts[i] = quoted(item)
		}
		return "[" + strings.Join(parts, ", ") + "]"
	case kObj:
		parts := make([]string, len(v.keys))
		for i, k := range v.keys {
			parts[i] = quoteString(k) + ": " + quoted(v.obj[k])
		}
		return "{" + strings.Join(parts, ", ") + "}"
	}
	return "None"
}

// quoteString wraps s in single quotes, or in double quotes when s holds a
// single quote and no double quote. Control and non-printable characters
// are escaped.
func quoteString(s string) string {
	q := '\''
	if strings.ContainsRune(s, '\'') && !strings.ContainsRune(s, '"') {
		q = '"'
	}
	var b strings.Builder
	b.WriteRune(q)
	for _, r := range s {
		switch {
		case r == q || r == '\\':
			b.WriteByte('\\')
			b.WriteRune(r)
		case r == '\n':
			b.WriteString(`\n`)
		case r == '\r':
			b.WriteString(`\r`)
		case r == '\t':
			b.WriteString(`\t`)
		case unicode.IsPrint(r):
			b.WriteRune(r)
		case r < 0x100:
			fmt.Fprintf(&b, `\x%02x`, r)
		case r < 0x10000:
			fmt.Fprintf(&b, `\u%04x`, r)
		default:
			fmt.Fprintf(&b, `\U%08x`, r)
		}
	}
	b.WriteRune(q)
	return b.String()
}

// jsonString encodes s as a JSON string without escaping non-ASCII or HTML
// characters.
func jsonString(s string) string {
	var b strings.Builder
	b.WriteByte('"')
	for _, r := range s {
		switch r {
		case '"':
			b.WriteString(`\"`)
		case '\\':
			b.WriteString(`\\`)
		case '\n':
			b.WriteString(`\n`)
		case '\r':
			b.WriteString(`\r`)
		case '\t':
			b.WriteString(`\t`)
		case '\b':
			b.WriteString(`\b`)
		case '\f':
			b.WriteString(`\f`)
		default:
			if r < 0x20 {
				fmt.Fprintf(&b, `\u%04x`, r)
			} else {
				b.WriteRune(r)
			}
		}
	}
	b.WriteByte('"')
	return b.String()
}

// encode writes v as JSON. indent < 0 gives the compact single-line form;
// otherwise each nesting level is indented by one space, starting at indent.
func encode(b *strings.Builder, v *value, indent int) {
	if v == nil {
		b.WriteString("null")
		return
	}
	nl := func(level int) {
		if indent >= 0 {
			b.WriteByte('\n')
			b.WriteString(strings.Repeat(" ", level))
		}
	}
	next := indent
	if indent >= 0 {
		next = indent + 1
	}
	sep := ","
	if indent < 0 {
		sep = ", "
	}
	switch v.kind {
	case kNull:
		b.WriteString("null")
	case kBool:
		b.WriteString(strconv.FormatBool(v.b))
	case kNum:
		b.WriteString(v.s)
	case kStr:
		b.WriteString(jsonString(v.s))
	case kArr:
		if len(v.arr) == 0 {
			b.WriteString("[]")
			return
		}
		b.WriteByte('[')
		for i, item := range v.arr {
			if i > 0 {
				b.WriteString(sep)
			}
			nl(next)
			encode(b, item, next)
		}
		nl(indent)
		b.WriteByte(']')
	case kObj:
		if len(v.keys) == 0 {
			b.WriteString("{}")
			return
		}
		b.WriteByte('{')
		for i, k := range v.keys {
			if i > 0 {
				b.WriteString(sep)
			}
			nl(next)
			b.WriteString(jsonString(k))
			b.WriteString(": ")
			encode(b, v.obj[k], next)
		}
		nl(indent)
		b.WriteByte('}')
	}
}

// indented returns v as JSON indented by one space per level.
func indented(v *value) string {
	var b strings.Builder
	encode(&b, v, 0)
	return b.String()
}

// compact returns v as single-line JSON.
func compact(v *value) string {
	var b strings.Builder
	encode(&b, v, -1)
	return b.String()
}

// labelKeys are the fields that can name a linked row, in priority order.
var labelKeys = []string{"Name", "Singular", "PlaceName", "Id"}

func isLabelKey(k string) bool { return slices.Contains(labelKeys, k) }

func joinPath(prefix, k string) string {
	if prefix == "" {
		return k
	}
	return prefix + "." + k
}

// flat flattens a fields object into out as dotted.path -> scalar. A link
// collapses to its row id, followed by the quoted label when a Name-like
// field was selected. Icons collapse to "icon <id>".
func flat(v *value, prefix string, out *value) {
	if v == nil {
		v = null
	}
	switch v.kind {
	case kObj:
		val, isLink := v.obj["value"]
		if isLink {
			_, hasSheet := v.obj["sheet"]
			isLink = hasSheet || len(v.keys) == 1
		}
		_, hasID := v.obj["id"]
		_, hasPath := v.obj["path"]
		switch {
		case isLink:
			sub := v.obj["fields"]
			if sub == nil || sub.kind != kObj {
				sub = newObj()
			}
			var label *value
			for _, k := range labelKeys {
				lv, ok := sub.obj[k]
				if !ok {
					continue
				}
				label = lv
				if lv.kind == kObj {
					inner := newObj()
					flat(lv, "", inner)
					if label = inner.obj[""]; label == nil {
						label = strVal("")
					}
				}
				break
			}
			s := text(val)
			if !absent(label) && (label.kind != kStr || label.s != "") {
				s += " " + quoted(label)
			}
			out.set(prefix, strVal(s))
			for _, k := range sub.keys {
				if isLabelKey(k) && !absent(label) {
					continue
				}
				flat(sub.obj[k], joinPath(prefix, k), out)
			}
		case hasID && hasPath:
			out.set(prefix, strVal("icon "+text(v.obj["id"])))
		default:
			for _, k := range v.keys {
				flat(v.obj[k], joinPath(prefix, k), out)
			}
		}
	case kArr:
		for i, item := range v.arr {
			if item.kind == kObj {
				// Skip empty link slots and empty objects.
				if isZero(item.obj["value"]) && !truthy(item.get("fields").get("Name")) {
					continue
				}
				if len(item.keys) == 0 {
					continue
				}
			}
			flat(item, fmt.Sprintf("%s[%d]", prefix, i), out)
		}
	default:
		out.set(prefix, v)
	}
}

// cell renders one table cell.
func cell(v *value) string {
	if absent(v) {
		return ""
	}
	switch v.kind {
	case kBool:
		return strconv.FormatBool(v.b)
	case kStr:
		return strings.ReplaceAll(strings.ReplaceAll(v.s, "\n", `\n`), "\t", " ")
	}
	return text(v)
}

func padRight(s string, width int) string {
	if n := utf8.RuneCountInString(s); n < width {
		return s + strings.Repeat(" ", width-n)
	}
	return s
}

// table prints rows as aligned columns: the id columns first, then every
// flattened field in first-seen order.
func table(w io.Writer, rows []*value, idCols []string) {
	if len(rows) == 0 {
		fmt.Fprintln(w, "(no rows)")
		return
	}
	recs := make([]*value, len(rows))
	cols := newObj()
	for _, c := range idCols {
		cols.set(c, null)
	}
	for i, r := range rows {
		rec := newObj()
		for _, c := range idCols {
			x := r.get(c)
			if x == nil {
				x = null
			}
			rec.set(c, x)
		}
		if f := r.get("fields"); f != nil {
			flat(f, "", rec)
		}
		if tr := r.get("transient"); truthy(tr) {
			flat(tr, "transient", rec)
		}
		for _, k := range rec.keys {
			cols.set(k, null)
		}
		recs[i] = rec
	}
	widths := make([]int, len(cols.keys))
	for i, c := range cols.keys {
		widths[i] = utf8.RuneCountInString(c)
		for _, rec := range recs {
			widths[i] = max(widths[i], utf8.RuneCountInString(cell(rec.obj[c])))
		}
	}
	line := make([]string, len(cols.keys))
	for i, c := range cols.keys {
		line[i] = padRight(c, widths[i])
	}
	fmt.Fprintln(w, strings.Join(line, "  "))
	for _, rec := range recs {
		for i, c := range cols.keys {
			line[i] = padRight(cell(rec.obj[c]), widths[i])
		}
		fmt.Fprintln(w, strings.Join(line, "  "))
	}
}
