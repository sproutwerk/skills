// Command sigs prints the public and protected declarations of C# files as
// "path:line: declaration", one per line, with multi-line parameter lists
// joined.
package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

func main() {
	if err := newCommand().Execute(); err != nil {
		os.Exit(1)
	}
}

func newCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "sigs <file.cs>...",
		Short: "Print public and protected C# declarations as path:line: declaration",
		Long: `Print the public and protected declarations of C# files.

Each declaration is written as "path:line: declaration", where path is the
argument as given and line is the 1-based line the declaration starts on.
Parameter lists that span several lines are joined, [MethodImpl] attributes
are removed and runs of whitespace collapse to one space.`,
		Args:          cobra.MinimumNArgs(1),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			w := bufio.NewWriter(cmd.OutOrStdout())
			for _, path := range args {
				data, err := os.ReadFile(path)
				if err != nil {
					fmt.Fprintln(cmd.ErrOrStderr(), "sigs:", err)
					return err
				}
				content := strings.TrimPrefix(string(data), "\uFEFF")
				content = strings.ToValidUTF8(content, "\uFFFD")
				for _, d := range declarations(content) {
					fmt.Fprintf(w, "%s:%d: %s\n", path, d.line, d.text)
				}
			}
			return w.Flush()
		},
	}
}
