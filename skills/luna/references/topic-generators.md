# Luna.Generators: source generators and analyzers

Covers every `Luna.Generators` attribute with what it emits, and the `Luna01`-`Luna04` analyzers.
Verified against Luna `52eb0fe` and ImSharp `4d09327` (branch `gc1.88`, ImGui 1.88). Paths are
relative to the skill root. Grep `references/api-index.txt` before using a member that is not shown
here.

Enable with the analyzer `ProjectReference` shown in
[adding Luna to a plugin](guide-adding-luna-to-a-plugin.md).
Attributes live in namespace `Luna.Generators`, are emitted `internal` with `[Embedded]` into the
consuming assembly, so add `using Luna.Generators;`. `LunaGenerator` only emits the
`EmbeddedAttribute` definition.

| Attribute | On | Emits |
|---|---|---|
| `[NamedEnum(Method = "ToName", Utf8 = true, Utf16 = true, Unknown = "Unknown", Namespace = null, Class = null)]` + `[Name(Name = null, Omit = false)]` on members | enum | Extension class `<Enum>Extensions` with `ToName()` (string) and `ToNameU8()`. File `NamedEnum.<Enum>.g.cs`. |
| `[TooltipEnum(Method = "Tooltip", Utf16 = false, Unknown = "", Namespace, Class)]` + `[Tooltip(Tooltip = null, Omit = false)]` | enum | Tooltip lookup method, UTF-8 by default. |
| `[AssociatedEnum<T>(ForwardMethod, BackwardMethod, ForwardDefaultValue, BackwardDefaultValue, Namespace, Class)]` + `[Associate<T>(T Value, Associate = true, DefaultName = false)]` | enum | Conversions to and from enum `T` (`ToTest2()`, custom back method), matching by member name unless overridden. |
| `[StrongType<T>(FieldName = "Value", Flags = StrongTypeFlag.Default)]` | `partial struct` | A typed ID over number `T`; `StrongTypeFlag` selects equality, comparison, `++`/`--`, arithmetic with self/base, conversions, Newtonsoft and System.Text.Json converters, `Zero`/`One`. |
| `[ConfigProperty(PropertyName, EventName, SaveMethodName = "Save", SkipSave)]` | field in a `partial` class | A property that calls the save method and an optional event on change, plus partial hook `On<Name>Changing(newValue, oldValue)`. |
| `[GenerateArities(maximumArity, IncludeZeroArity = …)]` | generic method, class, struct, delegate | Copies with 1..N (optionally 0) type parameters. |
| `[AdapterMethod(methodId, DisposeOnFailure, AlwaysAlive, SubscribeEvent, UnsubscribeEvent)]` | method, property, event | Dispatch code exposing the member through Dalamud's `IIdDataShareAdapter`. |
| `[GeneratedIpcSubscriber(LazySubscribers = false)]`, `[GeneratedIpcProvider]`, `[Ipc("Label")]`, `[EraseType]` / `[EraseType<T>]` (`Marshal`, `MarshalBack`), `[ProviderIpcContext]` | members of a `partial interface` describing an IPC contract | Call-gate subscriber and provider implementations. See `Luna.Tests/IpcTests.cs`. |

Analyzers (all warnings, category "Optimization"): `Luna01` UTF-16 literal
passed to a `Utf8StringHandler`; `Luna02` `Enum.GetValues<T>()` — use the
`T.Values` extension property from `ImSharp/Extensions/EnumExtensions.cs`;
`Luna03` `ToString()` result passed to a handler — pass an interpolated string;
`Luna04` UTF-16 literal passed to a `StringU8` constructor.
