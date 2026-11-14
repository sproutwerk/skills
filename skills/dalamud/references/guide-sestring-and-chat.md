## SeString and chat

API 15 services mostly take the mutable `SeString` / `SeStringBuilder`
(`Dalamud.Game.Text.SeStringHandling`), which is **removed in API 16**; sheets return
`ReadOnlySeString`. `IChatGui` events take `IChatMessage`-family objects since API 15. Builders,
payloads, chat events, `XivChatType` → `references/topic-sestring-chat.md`

