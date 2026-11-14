## WindowSystem and Window

Complete minimal plugin UI:

```csharp
public sealed class Plugin : IDalamudPlugin
{
    private readonly IDalamudPluginInterface pi;
    private readonly WindowSystem windows = new("MyPlugin");   // namespace = ID scope, must be unique to you
    private readonly MainWindow main;
    private readonly ConfigWindow config;

    public Plugin(IDalamudPluginInterface pi, ITextureProvider textures)
    {
        this.pi = pi;
        this.main = new MainWindow(textures);
        this.config = new ConfigWindow();          // another Window subclass, same shape as MainWindow
        this.windows.AddWindow(this.main);
        this.windows.AddWindow(this.config);
        pi.UiBuilder.Draw += this.windows.Draw;
        pi.UiBuilder.OpenMainUi += this.main.Toggle;
        pi.UiBuilder.OpenConfigUi += this.config.Toggle;
    }

    public void Dispose()
    {
        this.pi.UiBuilder.Draw -= this.windows.Draw;
        this.pi.UiBuilder.OpenMainUi -= this.main.Toggle;
        this.pi.UiBuilder.OpenConfigUi -= this.config.Toggle;
        this.windows.RemoveAllWindows();        // does not dispose the windows
    }
}

public sealed class MainWindow : Window
{
    private readonly ITextureProvider textures;
    private string filter = string.Empty;

    public MainWindow(ITextureProvider textures)
        : base("My Plugin###MyPluginMain", ImGuiWindowFlags.NoScrollbar | ImGuiWindowFlags.NoScrollWithMouse)
    {
        this.textures = textures;
        this.Size = new Vector2(420, 300);                 // unscaled; WindowSystem multiplies by GlobalScale
        this.SizeCondition = ImGuiCond.FirstUseEver;
        this.SizeConstraints = new WindowSizeConstraints
        {
            MinimumSize = new Vector2(300, 200),
            MaximumSize = new Vector2(float.MaxValue),
        };
    }

    public override void Draw()                            // no Begin/End here
    {
        ImGui.SetNextItemWidth(200 * ImGuiHelpers.GlobalScale);
        ImGui.InputTextWithHint("##filter", "Filter...", ref this.filter, 64);
        using var child = ImRaii.Child("list"u8, Vector2.Zero, true);
        if (!child)
            return;
        var icon = this.textures.GetFromGameIcon(new GameIconLookup(60001)).GetWrapOrEmpty();
        ImGui.Image(icon.Handle, ImGuiHelpers.ScaledVector2(32));
        ImGui.SameLine();
        ImGui.Text($"Filter is {this.filter.Length} chars");
    }
}
```

- `Draw()` runs between the system's `Begin`/`End`, only if Begin returned true. `Update()`
  runs every frame while open; `PreOpenCheck()` always; `DrawConditions()` returning `false`
  skips a frame without `OnClose`.
- `Size`/`SizeConstraints` are **unscaled** (multiplied by `GlobalScale`); `Position` is not
  scaled. `IsOpen` defaults to `false`.
- There is no `ImRaii.Window`. Without `WindowSystem` you lose size scaling, pinning,
  persistence, the close hotkey, sounds and the exception guard.

Lifecycle order, all properties, title bar buttons, `PreDraw`/`PostDraw` styling
→ `references/topic-windows.md`

