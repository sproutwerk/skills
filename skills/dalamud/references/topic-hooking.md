Dalamud topic: `Hook<T>` members, `AsmHook`, the `[Signature]` attribute property table and signature-scan behaviour. Verified against `goatcorp/Dalamud` `43b3411` (2026-11-10, `DalamudVersion 15.0.3.9`, API 15), `goatcorp/dalamud-docs` `3c7597d` (2026-11-13), `goatcorp/SamplePlugin` `b8477da` (2026-08-15) and `aers/FFXIVClientStructs` `80cb4a7` (2026-11-10). Reviewed 2026-11-13.

Paths such as `references/...` and `references/topic-*.md` are relative to the skill root (the directory that holds `SKILL.md`), not to this file. Part of the `dalamud` skill; items marked *(unverified)* are not confirmed from the pinned checkouts.

## Hooking and signatures

The hook priority order and the worked `Hook<T>` + `[Signature]` example are in
`references/guide-hooking-and-signatures.md`. `IGameInteropProvider` and `ISigScanner` member lists are in `references/topic-services.md`.

`Hook<T>` (`Dalamud.Hooking`): `Enable()`, `Disable()`, `Dispose()`, `Original`, `OriginalDisposeSafe` (callable after
dispose), `IsEnabled`, `IsDisposed`, `Address`, `BackendName`. The static `Hook<T>.FromAddress` etc. are `internal`:
create hooks only through `IGameInteropProvider`. `AsmHook` is public (`new AsmHook(nint address, byte[] assembly,
string name, AsmHookBehaviour)`); its `string[]` overload dies in API 16.

`[Signature]` (`SignatureAttribute`, fields **or** properties):

| Property | Default | Meaning |
|---|---|---|
| ctor `signature` | | IDA-style, `??` wildcards |
| `ScanType` | `ScanType.Text` | `Text` → `TryScanText`; `StaticAddress` → `TryGetStaticAddressFromSig` |
| `UseFlags` | `Auto` | `Pointer` (nint, pointer, function pointer, delegate), `Hook` (`Hook<T>`), `Offset` (primitive read at sig + `Offset`). Auto picks by member type. |
| `DetourName` | `null` | Method to use for a `Hook<T>` member. If null, exactly one method in the class must match the delegate. |
| `Offset` | 0 | For `UseFlags.Offset` |
| `Fallibility` | `Auto` | Auto = `Fallible` if the member is nullable, else `Infallible`. Infallible throws `SignatureException`; fallible logs a warning and leaves the member unset. |

`ScanText` follows the rel32 automatically when the match starts with `E8` or `E9` and throws if the target is outside
`.text`. Hooks created by `[Signature]` are not enabled for you; call `Enable()` (upstream examples always do). A
managed exception escaping a detour crashes the game; wrap the body in try/catch. Several plugins may hook one
function; the last to enable runs first, so call `Original` and avoid changing arguments.
