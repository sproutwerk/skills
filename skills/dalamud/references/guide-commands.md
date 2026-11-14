## Commands

```csharp
commandManager.AddHandler("/myplugin", new CommandInfo(OnCommand)
{
    HelpMessage = "Open the main window. /myplugin config → settings",
    ShowInHelp = true,        // default true
    DisplayOrder = -1,        // default
    AllowedInMacros = true,   // default
});
commandManager.RemoveHandler("/myplugin");

private void OnCommand(string command, string arguments) { /* arguments is the raw remainder */ }
```

`AddHandler` returns `false` (and logs) when the command is already registered; it does not throw.
The key is the full command including `/`. `ProcessCommand(string content)` dispatches a Dalamud
command line; `Commands` lists all registered. There is no argument parser; split `arguments`
yourself. Dalamud's own: `/xlplugins`, `/xlsettings`, `/xllog`, `/xldev`, `/xldata`, `/xlstats`,
`/xlhelp`.

