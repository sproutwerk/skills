## Svc

`ECommons.DalamudServices.Svc` is a class of static properties filled by `pi.Create<Svc>()`. Each is
the stock Dalamud service; member documentation → skill `dalamud`, `references/topic-services.md` there.

| `Svc.` | Dalamud interface | `Svc.` | Dalamud interface |
|---|---|---|---|
| `AddonEventManager` | `IAddonEventManager` | `Hook` | `IGameInteropProvider` |
| `AddonLifecycle` | `IAddonLifecycle` | `GameInventory` | `IGameInventory` |
| `AetheryteList` | `IAetheryteList` | `GameLifecycle` | `IGameLifecycle` |
| `AgentLifecycle` | `IAgentLifecycle` | `GamepadState` | `IGamepadState` |
| `PluginInterface` | `IDalamudPluginInterface` | `Gauges` | `IJobGauges` |
| `Buddies` | `IBuddyList` | `KeyState` | `IKeyState` |
| `Chat` | `IChatGui` | `MarketBoard` | `IMarketBoard` |
| `ClientState` | `IClientState` | `NamePlates` | `INamePlateGui` |
| `Commands` | `ICommandManager` | `NotificationManager` | `INotificationManager` |
| `Condition` | `ICondition` | `Objects` | `IObjectTable` |
| `Console` | `IConsole` | `PfGui` | `IPartyFinderGui` |
| `ContextMenu` | `IContextMenu` | `Party` | `IPartyList` |
| `Data` | `IDataManager` | `PlayerState` | `IPlayerState` |
| `DtrBar` | `IDtrBar` | `Log` | `IPluginLog` |
| `DutyState` | `IDutyState` | `SeStringEvaluator` | `ISeStringEvaluator` |
| `Fates` | `IFateTable` | `SigScanner` | `ISigScanner` |
| `FlyText` | `IFlyTextGui` | `Targets` | `ITargetManager` |
| `Framework` | `IFramework` | `Texture` | `ITextureProvider` |
| `GameConfig` | `IGameConfig` | `TextureSubstitution` | `ITextureSubstitutionProvider` |
| `GameGui` | `IGameGui` | `TextureReadback` | `ITextureReadbackProvider` |
| `TitleScreenMenu` | `ITitleScreenMenu` | `Toasts` | `IToastGui` |
| `UnlockState` | `IUnlockState` | `ReliableFileStorage` | `IReliableFileStorage` |

The names are not the interface names minus `I`: it is `Svc.Data`, `Svc.Objects`, `Svc.Targets`,
`Svc.Commands`, `Svc.Hook`, `Svc.Party`, `Svc.Texture`, `Svc.Toasts`, `Svc.Fates`, `Svc.Buddies`.

`Svc.GameNetwork` is different: an ECommons re-implementation (`ECommons.DalamudServices.Legacy.
IGameNetwork`) of the service Dalamud removed. It is created on first access and that access
installs three packet hooks → `references/topic-game-hooks.md`.

