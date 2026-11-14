Dalamud topic: SeString types and builders, payload classes, chat events (API 15 shape), `XivChatType` and log messages. Verified against `goatcorp/Dalamud` `43b3411` (2026-11-10, `DalamudVersion 15.0.3.9`, API 15), `goatcorp/dalamud-docs` `3c7597d` (2026-11-13), `goatcorp/SamplePlugin` `b8477da` (2026-08-15) and `aers/FFXIVClientStructs` `80cb4a7` (2026-11-10). Reviewed 2026-11-13.

Paths such as `references/...` and `references/topic-*.md` are relative to the skill root (the directory that holds `SKILL.md`), not to this file. Part of the `dalamud` skill; items marked *(unverified)* are not confirmed from the pinned checkouts.

## SeString and chat

Two families coexist in API 15:

| Type | Namespace | Nature |
|---|---|---|
| `SeString`, `Payload` subclasses, `SeStringBuilder` | `Dalamud.Game.Text.SeStringHandling[.Payloads]` | Mutable, class based. Used by most API 15 service signatures. **Removed in API 16.** |
| `ReadOnlySeString`, `ReadOnlySeStringSpan` | `Lumina.Text.ReadOnly` | Immutable structs; what sheets return. `ToString()` gives plain text (API 13+), `ToMacroString()` the macro form. |
| `Lumina.Text.SeStringBuilder`, `RentedSeStringBuilder` | `Lumina.Text`, `Dalamud.Utility` | Builder for the read-only family; `using var b = new RentedSeStringBuilder();` pools it. |

`SeString`: `TextValue` (text payloads only), `Payloads`, implicit from `string`, `Parse(byte[] | ReadOnlySpan<byte> |
byte*)`, `Encode()`, `EncodeWithNullTerminator()`, `Append(…)`, static `CreateItemLink(uint itemId, bool isHq |
ItemKind kind, string? displayNameOverride)`, `CreateMapLink(uint territoryId, uint mapId, float xCoord, float
yCoord)`, `CreateMapLink(string placeName, float x, float y)`. Bridges (`Dalamud.Utility.SeStringExtensions`):
`readOnly.ToDalamudString()`, `builder.AppendMacroString("…")`, `ross.ContainsText("x"u8)`, `ReplaceText`.

Helpers for the read-only family, backported from API 16 after 15.0.3.9 (tag `15.0.4.0`; absent from earlier API 15
runtimes), all in `Dalamud.Utility`: on `ReadOnlySePayload` and `ReadOnlySePayloadSpan`, `IsMacro(MacroCode)`,
`IsLink(LinkMacroPayloadType)`, `IsLinkTerminator()`, `TryReadLink(out linkType, out expr1 … out expr4)` and
`TryReadDalamudLink(out DalamudLinkPayload)`; on `ReadOnlySeString` and `ReadOnlySeStringSpan`,
`ContainsText(ReadOnlySpan<char>)`, and on the span `ReplaceText(ReadOnlySpan<byte> toFind, ReadOnlySpan<byte>
replacement)`; on `Lumina.Text.SeStringBuilder`, `PushDalamudLink`, `PushItemLink`, `PushPartyFinderLink` and
`PushPartyFinderSearchConditionsLink`. `SeIconChar.Hydaelyn` replaces the misspelt `Hyadelyn` (`[Obsolete]`) in the
same range.

```csharp
var msg = new SeStringBuilder()
    .AddUiForeground("[MyPlugin] ", 45)      // text wrapped in a UIColor row id
    .AddText("Found ")
    .AddItemLink(itemId, isHq: false)
    .AddItalics(" nearby")
    .Build();
chat.Print(msg);
chat.Print(new XivChatEntry { Message = msg, Type = XivChatType.Echo });
```

Builder methods: `AddText`, `AddUiForeground(ushort)` / `AddUiForegroundOff()`, `AddUiGlow` / `AddUiGlowOff()`,
`AddIcon(BitmapFontIcon)`, `AddItemLink`, `AddItemLinkRaw`, `AddMapLink`, `AddQuestLink`, `AddStatusLink`,
`AddPartyFinderLink`, `AddItalics` / `AddItalicsOn/Off`, `Add(Payload)`, `Append`, `Build()`, `BuiltString`,
`Encode()`. Payload classes: `TextPayload`, `PlayerPayload`, `ItemPayload`, `MapLinkPayload`, `StatusPayload`,
`QuestPayload`, `IconPayload`, `UIForegroundPayload`, `UIGlowPayload`, `EmphasisItalicPayload`,
`AutoTranslatePayload`, `DalamudLinkPayload`, `PartyFinderPayload`, `NewLinePayload`, `SeHyphenPayload`, `RawPayload`.

Chat events (API 15 shape):

```csharp
chat.ChatMessage += OnChat;

private void OnChat(IHandleableChatMessage message)
{
    if (message.LogKind != XivChatType.TellIncoming) return;
    var sender = message.Sender.Payloads.OfType<PlayerPayload>().FirstOrDefault();
    log.Debug("{Name}@{World}: {Text}", sender?.PlayerName, sender?.World.Value.Name, message.Message.TextValue);
    // message.Message = …;       // IMutableChatMessage: rewrite
    // message.PreventOriginal();  // IHandleableChatMessage: hide from the log
}
```

`IChatMessage`: `LogKind` (`XivChatType`), `SourceKind`, `TargetKind` (`XivChatRelationKind`), `Sender`, `Message`
(`SeString`), `OriginalSender`, `OriginalMessage` (`ReadOnlySeString`), `Timestamp`, `IsHandled`. `XivChatType` values
are `LogKind` row ids (`Say = 10`, `Shout = 11`, `TellOutgoing = 12`, `TellIncoming = 13`, `Party = 14`, `FreeCompany
= 24`, `Echo = 56`, `SystemMessage = 57`, `ErrorMessage = 60`, …); since API 15 the source/target relation bits are
split out, so old masks and values above 110 no longer apply. For system messages prefer the `LogMessage` event
(`ILogMessage`: `LogMessageId`, `GameData`, `Parameters`, `TryGetIntParameter`, `TryGetStringParameter`,
`SourceEntity`, `TargetEntity`, `PreventOriginal()`).

Macro/payload encoding reference (macro codes `0x06`-`0x61`, expressions, global parameters):
`references/docs/docs/plugin-development/sestring.mdx` (2456 lines).

Porting to the read-only family ahead of API 16: `references/docs/docs/versions/v16/sestring.md` maps each `SeString`,
`SeStringBuilder` and payload member to its replacement and covers chat links, map links and ImGui rendering. API 16
is not finalized, so recheck it against the target branch.
