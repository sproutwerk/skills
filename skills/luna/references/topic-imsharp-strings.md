# ImSharp strings and string handlers

Covers the three UTF-8 interpolated-string handler aliases, what converts into them, and `StringU8`
/ `StringPair` / `InlineStringU8`. Verified against Luna `52eb0fe` and ImSharp `4d09327` (branch
`gc1.88`, ImGui 1.88). Paths are relative to the skill root. Grep `references/api-index.txt` before
using a member that is not shown here.

## Strings

Three aliases (`ImSharp/GlobalAlias.cs`), all `ref struct` interpolated-string
handlers over a distinct shared buffer:

| Alias | Buffer | Used for |
|---|---|---|
| `Utf8LabelHandler` | `LabelStringHandlerBuffer` | labels and IDs |
| `Utf8TextHandler` | `TextStringHandlerBuffer` | text, tooltips |
| `Utf8HintHandler` | `HintStringHandlerBuffer` | hints, previews, printf formats, shortcuts |

Implicit conversions into a handler: `ReadOnlySpan<byte>` / `Span<byte>`
(`"..."u8`), `StringU8`, `InlineStringU8<T>`, `string`, `ReadOnlySpan<char>`,
and `$"…"` interpolation.

- `"..."u8` and `StringU8` are passed by pointer, no copy. Label parameters are
  documented "HAS to be null-terminated"; `u8` literals and `StringU8` built
  with the default `nullTerminated: true` are.
- `string` and `$"…"` are transcoded into the shared buffer for that handler
  kind (`Utf8.TryWrite` into `T.Buffer`). No managed allocation.
- One call can take one interpolated label, one text and one hint at once
  because the buffers differ. Two interpolated strings of the *same* kind alive
  at once overwrite each other; materialize one with `new StringU8($"…")`.
- `HoverUtf8StringHandler` (`Im.Tooltip.OnHover(HoveredFlags flags, $"…")`,
  `Im.Tooltip.ImageOnHover`) only formats when the item is hovered.
- Analyzer `Luna01` warns when a UTF-16 literal is passed where a handler is
  expected; write `"Text"u8`.

`StringU8` (`readonly partial struct`): `StringU8.Empty`, `StringU8.Null`, `StringU8.EmptyLazy` (a `Lazy<StringU8>`),
`IsNull`, `IsEmpty`, `Length`, `Span`, `Memory`; constructors from `byte[]`,
`ReadOnlySpan<byte>` (+ `nullTerminated`), `ReadOnlySpan<char>`, `byte*`, and
`Utf8InterpolatedStringHandler` (so `new StringU8($"{x}")` works); `+`
concatenation, ordering and equality operators, `Contains`/`StartsWith`/`EndsWith`
with `…CaseInsensitive` variants, `CalculateSize()`. Conversion to
`ReadOnlySpan<byte>` is implicit, from it explicit. `StringPair(string)` /
`StringPair(StringU8)` keeps both encodings (`.Utf16`, `.Utf8`), converts
implicitly to either, and is `IEquatable` and `IComparable` (with `==` and `!=`). `InlineStringU8<TBacking>` packs a short string into an
integer.
