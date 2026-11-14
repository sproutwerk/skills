## Logging

```csharp
log.Information("Loaded {Count} entries for {Player}", count, name);  // Serilog message template
log.Error(ex, "Failed to parse {File}", path);                          // exception first
log.Verbose("tick");                                                    // hidden unless level lowered
```

`IPluginLog`: `Fatal`, `Error`, `Warning`, `Information` (alias `Info`), `Debug`, `Verbose`, each
`(string messageTemplate, params object[] values)` and `(Exception? exception, string
messageTemplate, params object[] values)`; `Write(LogEventLevel, Exception?, string, params
object[])`; `MinimumLogLevel { get; set; }`; `Logger` (Serilog `ILogger`). Entries are tagged with
the plugin name and go to the Dalamud log; view with `/xllog`. Use placeholders, not `$"…"`
interpolation, so the console can filter and structure values. The static `PluginLog` class no
longer exists. Custom Serilog sinks must set the `Dalamud.PluginName` property or console filtering
breaks.

