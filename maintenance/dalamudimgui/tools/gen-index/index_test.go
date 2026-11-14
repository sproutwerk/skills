package main

import (
	"regexp"
	"strings"
	"testing"
)

const sample = "\uFEFF" + `namespace Demo;

// public class Commented
public static unsafe partial class ImGui
{
    public static void Begin(ImU8String name,
        ref bool open,
        ImGuiWindowFlags flags = ImGuiWindowFlags.None) => BeginImpl(name, ref open, flags);

    public static T* Alloc<T>(int count) where T : unmanaged { return null; }

    public static int Count<T>( ReadOnlySpan<T> span ) where T : struct => span.Length;

    public static int Count<T>( ReadOnlySpan<T> span ) where T : struct => span.Length;

    public class Nested { }

    public delegate void Callback(int x);

    private static void Hidden() { }
}

public readonly struct ImVector
{
    public readonly int Size => 0;

    public unsafe ImVector(int size) { }

    public int Capacity()
    {
        return 0;
    }
}
`

func rowsOf(t *testing.T, opts options) []string {
	t.Helper()
	if opts.classes == nil {
		opts.classes = regexp.MustCompile(`^(?:.*)$`)
	}
	content := strings.TrimPrefix(sample, "\uFEFF")
	var out []string
	for _, r := range sortedUnique(extractRows("Demo/File.cs", content, opts)) {
		out = append(out, r.String())
	}
	return out
}

func TestExtractRows(t *testing.T) {
	got := rowsOf(t, options{})
	want := []string{
		"ImGui.static T* Alloc<T>(int count) where T : unmanaged\t// Demo/File.cs",
		"ImGui.static void Begin(ImU8String name, ref bool open, ImGuiWindowFlags flags = ImGuiWindowFlags.None)\t// Demo/File.cs",
		"ImGui.static int Count<T>(ReadOnlySpan<T> span ) where T : struct\t// Demo/File.cs",
		"ImVector.int Capacity()\t// Demo/File.cs",
		"ImVector.unsafe ImVector(int size)\t// Demo/File.cs",
	}
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Errorf("rows mismatch\n got:\n%s\nwant:\n%s", strings.Join(got, "\n"), strings.Join(want, "\n"))
	}
}

func TestClassFilterAndExclude(t *testing.T) {
	got := rowsOf(t, options{classes: regexp.MustCompile(`^(?:ImGui)$`)})
	for _, r := range got {
		if !strings.HasPrefix(r, "ImGui.") {
			t.Errorf("class filter let through %q", r)
		}
	}
	if len(got) != 3 {
		t.Errorf("got %d ImGui rows, want 3", len(got))
	}

	got = rowsOf(t, options{exclude: regexp.MustCompile(`^(ImGui)$`)})
	for _, r := range got {
		if strings.HasPrefix(r, "ImGui.") {
			t.Errorf("exclude kept %q", r)
		}
	}
	if len(got) != 2 {
		t.Errorf("got %d rows after exclude, want 2", len(got))
	}
}

func TestSkipPointers(t *testing.T) {
	for _, r := range rowsOf(t, options{skipPointers: true}) {
		if strings.Contains(r, "*") {
			t.Errorf("pointer overload kept: %q", r)
		}
	}
}

func TestNormalizeSignature(t *testing.T) {
	cases := map[string]string{
		"public static void A(int x) { }":                      "static void A(int x)",
		"public static void A( int x ) => Run(x)":              "static void A(int x )",
		"public  static   T  B<T>(T x)   where T : class => x": "static T B<T>(T x) where T : class",
		"public void C(Func<int, int> f) where T : new() { }":  "void C(Func<int, int> f) where T : new()",
		"public void D(int x) where T : I = default":           "void D(int x)",
	}
	for in, want := range cases {
		if got := normalizeSignature(in); got != want {
			t.Errorf("normalizeSignature(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestSortOrder(t *testing.T) {
	rows := []row{
		{"Z", "void b()", "f"},
		{"A", "static void Zeta()", "f"},
		{"A", "static int alpha(int x)", "f"},
		{"A", "static int Alpha()", "f"},
	}
	got := sortedUnique(rows)
	want := []row{
		{"A", "static int Alpha()", "f"},
		{"A", "static int alpha(int x)", "f"},
		{"A", "static void Zeta()", "f"},
		{"Z", "void b()", "f"},
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("position %d: got %v, want %v", i, got[i], want[i])
		}
	}
}
