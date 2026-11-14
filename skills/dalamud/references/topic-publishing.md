Dalamud topic: publishing to the official repo (DalamudPluginsD17), custom repositories, restrictions, backend-server rules, the AI usage policy and the account-ID / Content-ID policy. Verified against `goatcorp/Dalamud` `43b3411` (2026-11-10, `DalamudVersion 15.0.3.9`, API 15), `goatcorp/dalamud-docs` `3c7597d` (2026-11-13), `goatcorp/SamplePlugin` `b8477da` (2026-08-15) and `aers/FFXIVClientStructs` `80cb4a7` (2026-11-10). Reviewed 2026-11-13.

Paths such as `references/...` and `references/topic-*.md` are relative to the skill root (the directory that holds `SKILL.md`), not to this file. Part of the `dalamud` skill; items marked *(unverified)* are not confirmed from the pinned checkouts.

## Publishing

Official repo = PR to `goatcorp/DalamudPluginsD17`; built by Plogon from an exact commit with no network access.

```toml
# testing/live/MyPlugin/manifest.toml  (new plugins MUST start in testing/live)
[plugin]
repository = "https://github.com/you/MyPlugin.git"
commit = "765d9bb434ac99a27e9a3f2ba0a555b55fe6269d"
owners = ["you"]
project_path = "MyPlugin"
changelog = "Added the thing"
```

- Beside it `images/icon.png` (1:1, 64x64 to 512x512) and optional `image1.png` … `image5.png`.
- One plugin per PR, each PR from its own branch. Update = new PR changing `commit`. Moving between tracks (`stable/`,
  `testing/live/`) = copy the folder; the highest `AssemblyVersion` reachable wins.
- Changelog precedence: `manifest.toml` `changelog` → PR description → manifest `Changelog`. Only the first and last
  show in the installer. PR description starting with `nofranz` suppresses the Discord post.
- Source must be open. New plugins need 4 of 6 approval-team votes (any member can veto); updates need one reviewer.
  Expect a week or more.
- Testing builds must set `TestingDalamudApiLevel` (API 14).
- Custom repositories: a JSON array of store entries served over HTTP GET (`DownloadLinkInstall`,
  `DownloadLinkUpdate`, `DownloadLinkTesting`, `IsHide`, `IsTestingExclusive`, `TestingAssemblyVersion`, `LastUpdate`,
  …); minimal support from the project.
- Kill switch: PR to `DalamudAssets` `bannedplugin.json` (`Name`, `AssemblyVersion`, `Reason`); unban by releasing a
  higher version.

### Restrictions (official repo)

A plugin must not:

- talk to the game servers automatically (polling, requests without direct user interaction) or out of spec (things a
  player could not do);
- augment, alter or interfere with combat, except re-presenting information about your own party/alliance that is
  already available, and then only after talking to the approval team first;
- touch Square Enix's revenue (Mog Station items, avoiding Fantasia);
- provide parsing, DPS meters, raid logging, FFLogs integration;
- collect account IDs of other characters, in any form;
- hard-depend on a plugin that violates the guidelines;
- be useful only out of spec (out-of-bounds etc.);
- give any advantage in PvP or competitive content.

Standing rejections: emote looping, cutscene or dialog skipping, automated crafting, auto-roll on loot, friend login
alerts, AOE markers for untelegraphed mechanics, AOE recolouring, camera zoom beyond game limits, more XIVCombo-style
combos. Unsure → ask in the Dalamud Discord before writing code.

Backend servers (technical-considerations, binding): send the minimum; hash local player identifiers (Content ID,
name) client-side; extra telemetry is opt-in, topical, for the public interest, with a pseudo-random identifier the
user can reset locally; never expose who else uses the plugin; HTTPS/TLS with a trusted CA; connect by DNS hostname,
not IP. Recommended: allow a user-defined server, open-source the server, dual-stack, retry logic, version checks.
`Dalamud.Networking.Http.HappyEyeballsCallback` exists for `SocketsHttpHandler.ConnectCallback`.

### AI usage policy (official repo submissions)

Applies to D17 submissions, not to contributions to Dalamud itself.

Disclosure levels (state yours in the PR description): **None** and **Hint** (autocomplete only) need no disclosure;
**Assist** (human-led, AI for specific tasks), **Pair** (roughly equal), **Copilot** (AI implements, human plans and
reviews) and **Auto** (AI acts autonomously) must be disclosed.

- The author must personally test the plugin and be able to explain every implementation choice. "The AI did it" is
  not an answer.
- Entirely AI-generated submissions are auto-rejected; twice = ban. Undisclosed AI use in a demonstrably AI-written
  submission = ban.
- **Copilot** and **Auto** submissions are reviewed at the team's discretion and may be closed without a full review
  when they are larger or more complex than the author's shown understanding; resubmitting at a smaller scope is
  allowed.
- AI disclosures and pull request descriptions must be written by a human. Do not draft them for the user.
- AI-generated assets must be disclosed in the plugin description; a hand-made icon is preferred; AI audio should be
  toggleable.
- AI translations are acceptable as placeholders.

When this skill is used to write plugin code, tell the user that a D17 submission needs the matching disclosure level,
their own testing, and a pull request description they wrote themselves.

### Account IDs and Content IDs

Dalamud's API does not expose Account IDs; ClientStructs does (`Character.AccountId`). News posts 2025-01-10 and
2026-07-13: the project does not approve of tools that correlate characters to accounts, and the restriction above
forbids collecting account IDs of other players regardless of purpose. Content IDs of other players are permanent
identifiers: do not upload or persist them unhashed. Do not write code that harvests either.
