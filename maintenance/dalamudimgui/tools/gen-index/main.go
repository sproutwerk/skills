// Command gen-index extracts public method signatures from C# binding sources.
//
// It writes one line per public method, "<Class>.<signature>\t// <file>",
// sorted and de-duplicated. Multi-line parameter lists are joined. Only files
// that are compiled (per the csproj Compile Remove rules) should be passed in.
package main

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/spf13/cobra"
)

func main() {
	if err := newCommand().Execute(); err != nil {
		os.Exit(1)
	}
}

func newCommand() *cobra.Command {
	var (
		base         string
		exclude      string
		skipPointers bool
	)
	cmd := &cobra.Command{
		Use:   "gen-index <out.txt> <class-name-regex> <file-or-dir>...",
		Short: "Extract public method signatures from C# binding sources",
		Long: `Extract public method signatures from C# binding sources.

Emits one line per public method: "<Class>.<signature>    // <relative file>".
Multi-line parameter lists are joined. The class regex must match the whole
class or struct name. Directories are walked for .cs files; explicit files are
read as given. Only files that are compiled (per the csproj Compile Remove
rules) should be passed in.

INDEX_BASE and INDEX_SKIP_POINTERS set the defaults of --base and
--skip-pointers.`,
		Args:          cobra.MinimumNArgs(3),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			classes, err := regexp.Compile(`^(?:` + args[1] + `)$`)
			if err != nil {
				return fmt.Errorf("class regex: %w", err)
			}
			opts := options{classes: classes, skipPointers: skipPointers}
			if exclude != "" {
				if opts.exclude, err = regexp.Compile(exclude); err != nil {
					return fmt.Errorf("exclude regex: %w", err)
				}
			}
			n, err := run(args[0], base, args[2:], opts)
			if err != nil {
				fmt.Fprintln(cmd.ErrOrStderr(), "gen-index:", err)
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), args[0], n)
			return nil
		},
	}
	cmd.Flags().StringVar(&base, "base", os.Getenv("INDEX_BASE"), "write file paths relative to this directory")
	cmd.Flags().StringVar(&exclude, "exclude", "", "drop classes whose name matches this regex")
	cmd.Flags().BoolVar(&skipPointers, "skip-pointers", os.Getenv("INDEX_SKIP_POINTERS") != "", "drop overloads that take or return raw pointers")
	return cmd
}

// run indexes roots into out and returns the number of lines written.
func run(out, base string, roots []string, opts options) (int, error) {
	files, err := collectFiles(roots)
	if err != nil {
		return 0, err
	}
	var rows []row
	for _, f := range files {
		data, err := os.ReadFile(f)
		if err != nil {
			return 0, err
		}
		content := strings.TrimPrefix(string(data), "\uFEFF")
		content = strings.ToValidUTF8(content, "\uFFFD")
		rel, err := relativePath(f, base)
		if err != nil {
			return 0, err
		}
		rows = append(rows, extractRows(rel, content, opts)...)
	}
	rows = sortedUnique(rows)

	w, err := os.Create(out)
	if err != nil {
		return 0, err
	}
	bw := bufio.NewWriter(w)
	for _, r := range rows {
		if _, err := bw.WriteString(r.String() + "\n"); err != nil {
			w.Close()
			return 0, err
		}
	}
	if err := bw.Flush(); err != nil {
		w.Close()
		return 0, err
	}
	return len(rows), w.Close()
}

// collectFiles expands directories to the .cs files under them and keeps
// explicit files as given, sorted by path.
func collectFiles(roots []string) ([]string, error) {
	var files []string
	for _, r := range roots {
		info, err := os.Stat(r)
		if err != nil {
			return nil, err
		}
		if !info.IsDir() {
			files = append(files, r)
			continue
		}
		err = filepath.WalkDir(r, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if !d.IsDir() && strings.HasSuffix(d.Name(), ".cs") {
				files = append(files, path)
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	slices.Sort(files)
	return files, nil
}

func relativePath(file, base string) (string, error) {
	if base == "" {
		return file, nil
	}
	absFile, err := filepath.Abs(file)
	if err != nil {
		return "", err
	}
	absBase, err := filepath.Abs(base)
	if err != nil {
		return "", err
	}
	return filepath.Rel(absBase, absFile)
}
