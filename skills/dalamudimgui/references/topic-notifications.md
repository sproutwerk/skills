# Notifications

Covers `INotificationManager`, `Notification` and the live `IActiveNotification`.
Verified against goatcorp/Dalamud `master` @ `43b3411` (2026-11-10, API level 15, Dear ImGui 1.88 docking).
Paths are relative to the skill root (`SKILL.md`, `references/topic-*.md`, `references/...`); a bare
`index/...` or `src/...` path means `references/index/...` or `references/src/...`.

```csharp
var n = notifications.AddNotification(new Notification
{
    Title = "Export",
    Content = "Exporting 120 items...",
    Type = NotificationType.Info,
    InitialDuration = TimeSpan.FromSeconds(10),
    Progress = 0f,
    Minimized = false,
});
n.Progress = 0.5f;                       // IActiveNotification is live-editable
n.Click += _ => this.main.IsOpen = true;
n.DrawActions += args => { if (ImGui.Button("Open folder"u8)) Open(); };
n.DismissNow();
```

`INotificationManager.AddNotification(Notification)` returns a live
`IActiveNotification`. Other `Notification` properties: `MinimizedText`,
`Icon` (`INotificationIcon.From(FontAwesomeIcon | SeIconChar)`,
`FromGame(path)`, `FromFile(path)`), `IconTexture`, `HardExpiry`,
`ExtensionDurationSinceLastInterest`, `ShowIndeterminateIfNoExpiry`,
`RespectUiHidden`, `UserDismissable`. `Type` is `None`, `Success`, `Warning`,
`Error` or `Info`. Active extras: `Id`, `EffectiveExpiry`, `DismissReason`,
`ExtendBy(TimeSpan)`, event `Dismiss`. `DrawActions` runs inside the
notification window during draw; treat it like `Draw`.
