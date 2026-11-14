package main

import (
	"strings"
	"testing"
)

const sample = `using System;

namespace Demo;

public static partial class Widgets
{
    [MethodImpl(MethodImplOptions.AggressiveInlining)]
    public static bool Button(ReadOnlySpan<byte> label,
        Vector2 size = default,
        bool   disabled = false)
        => ButtonImpl(label, size, disabled);

    protected  internal   int Count { get; set; }

    [Pure] public static int Add(int a, int b) => a + b;

    private static void Hidden(int a,
        int b) { }

    internal struct Inner { }
}
`

func TestDeclarations(t *testing.T) {
	got := declarations(sample)
	want := []declaration{
		{5, "public static partial class Widgets"},
		{8, "public static bool Button(ReadOnlySpan<byte> label, Vector2 size = default, bool disabled = false)"},
		{13, "protected internal int Count { get; set; }"},
		{15, "[Pure] public static int Add(int a, int b) => a + b;"},
	}
	if len(got) != len(want) {
		t.Fatalf("got %d declarations, want %d:\n%v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("declaration %d: got %d %q, want %d %q", i, got[i].line, got[i].text, want[i].line, want[i].text)
		}
	}
}

func TestDeclarationsCRLF(t *testing.T) {
	got := declarations(strings.ReplaceAll(sample, "\n", "\r\n"))
	if len(got) != 4 || got[1].text != "public static bool Button(ReadOnlySpan<byte> label, Vector2 size = default, bool disabled = false)" {
		t.Errorf("CRLF input changed the result: %v", got)
	}
}

func TestUnclosedParenthesisStopsAtEOF(t *testing.T) {
	got := declarations("public void Open(int a,\n    int b")
	if len(got) != 1 || got[0].text != "public void Open(int a, int b" {
		t.Errorf("got %v", got)
	}
}
