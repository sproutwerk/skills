## Topic files

Deep, situational material lives in `references/topic-*.md` (first-party, same pins as above). Open the file when
its trigger applies; do not answer those questions from the stubs in this file alone.

| File | Read this when ... |
|---|---|
| `references/topic-project-setup.md` | writing or fixing a csproj beyond the minimal one, choosing manifest keys (`.json` / `.yaml` / csproj), setting `LoadRequiredState` / `LoadSync` / `LoadPriority` / `CanUnloadAsync` / `ApplicableVersion`, build/SDK version errors, custom ClientStructs build, legacy project layouts |
| `references/topic-plugin-interface.md` | you need any `IDalamudPluginInterface` member: paths, load context, other installed plugins, Dalamud version info, opening Dalamud windows, UI language |
| `references/topic-services.md` | you are about to call a member of any service in the catalog below and need its exact name, signature, event list or gotcha; also the members of `IGameObject` / `ICharacter` / `IBattleChara` / `IStatus` |
| `references/topic-hooking.md` | using `[Signature]` options (`ScanType`, `UseFlags`, `DetourName`, `Offset`, `Fallibility`), `Hook<T>` members other than `Enable`/`Original`/`Dispose`, `AsmHook`, or debugging a signature that resolves to the wrong address |
| `references/topic-ipc.md` | exposing or consuming IPC with another plugin (`GetIpcProvider` / `GetIpcSubscriber`), handling `IpcError`s, or sharing data with DataShare (`GetOrCreateData`) |
| `references/topic-sestring-chat.md` | building or parsing an `SeString`, printing coloured text or links to chat, handling `IChatGui` events, filtering by `XivChatType`, reading log messages, porting chat code to API 15 |
| `references/topic-addons.md` | reacting to a native game window (`IAddonLifecycle.RegisterListener`, `AddonEvent`, args types, `PreventOriginal`) or attaching click/hover events to native nodes (`IAddonEventManager`) |
| `references/topic-dev-workflow.md` | loading a dev plugin, the plugin does not show up or reload, attaching a debugger, using `/xldev` / `/xldata` tools, switching Dalamud channels |
| `references/topic-publishing.md` | preparing a D17 submission or update, custom repo JSON, **and before writing any feature that automates gameplay, touches combat, contacts servers, or collects data about other players** (restrictions, AI policy, account-ID policy) |
| `references/topic-migration.md` | porting a plugin across API levels, a plugin breaks after an API bump, a member named in old code does not exist, preparing for API 16 |

