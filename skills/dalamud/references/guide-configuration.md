## Configuration

```csharp
[Serializable]
public class Configuration : IPluginConfiguration
{
    public int Version { get; set; } = 1;
    public bool Enabled { get; set; } = true;

    public void Save() => Plugin.PluginInterface.SavePluginConfig(this);
}

var config = pluginInterface.GetPluginConfig() as Configuration ?? new Configuration();
```

- `IPluginConfiguration` (`Dalamud.Configuration`) has one member: `int Version { get; set; }`.
  Migrations are yours to write.
- Stored as `<InternalName>.json` in Dalamud's plugin config folder (`pluginInterface.ConfigFile`),
  Newtonsoft.Json, indented, `TypeNameHandling.Objects` with simple assembly names. The root `$type`
  carries your class and assembly name, so **renaming or moving the config class breaks loading of
  existing files**. `GetPluginConfig()` returns `null` when the file is missing.
- `SavePluginConfig` writes synchronously through reliable storage. Do not call it every frame; save
  on change or debounce.
- Own files go under `pluginInterface.ConfigDirectory` (a folder named `<InternalName>` beside the
  config file). Non-serialisable members need `[JsonIgnore]` (Newtonsoft's, not System.Text.Json's).

