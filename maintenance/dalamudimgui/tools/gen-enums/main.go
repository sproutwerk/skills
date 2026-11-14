// Command gen-enums writes one line per C# enum found under the Enums directories of a
// Dalamud bindings checkout, with the enum's members and values, sorted for grep.
package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/spf13/cobra"
)

var (
	enumDecl    = regexp.MustCompile(`public enum (\w+)(?:\s*:\s*(\w+))?`)
	uncheckedRe = regexp.MustCompile(`(?m)^\s*(\w+)\s*=\s*unchecked\((?:\(\w+\))?(-?\w+)\)`)
	plainRe     = regexp.MustCompile(`(?m)^\s*(\w+)\s*=\s*([^,\n]+),`)
)

func main() {
	cmd := &cobra.Command{
		Use:   "gen-enums <bindings-root> <out>",
		Short: "Index every enum under the Enums directories of a bindings tree",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			rows, err := collect(args[0])
			if err != nil {
				return err
			}
			if err := os.WriteFile(args[1], []byte(strings.Join(rows, "\n")+"\n"), 0o644); err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), args[1], len(rows))
			return nil
		},
		SilenceUsage: true,
	}
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}

// collect walks root and returns one "<lib>.<Enum>: a=1, b=2" row per enum declaration
// found in a file whose directory path contains an Enums segment, sorted.
func collect(root string) ([]string, error) {
	var rows []string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		dir := filepath.Dir(path)
		if !strings.Contains(dir, string(os.PathSeparator)+"Enums") {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		text := strings.TrimPrefix(string(data), "\uFEFF")
		m := enumDecl.FindStringSubmatchIndex(text)
		if m == nil {
			return nil
		}
		name := text[m[2]:m[3]]
		body := text[m[1]:]
		members := uncheckedRe.FindAllStringSubmatch(body, -1)
		if len(members) == 0 {
			members = plainRe.FindAllStringSubmatch(body, -1)
		}
		rel, err := filepath.Rel(root, dir)
		if err != nil {
			return err
		}
		lib := strings.ReplaceAll(strings.Split(rel, string(os.PathSeparator))[0], "Dalamud.Bindings.", "")
		parts := make([]string, 0, len(members))
		for _, mem := range members {
			parts = append(parts, mem[1]+"="+strings.TrimSpace(mem[2]))
		}
		rows = append(rows, lib+"."+name+": "+strings.Join(parts, ", "))
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(rows)
	return rows, nil
}
