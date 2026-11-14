## AddonLifecycle and AddonEventManager

`addonLifecycle.RegisterListener(AddonEvent.PostSetup, "SelectYesno", OnSetup)` with handler
`(AddonEvent type, AddonArgs args)`; unregister in `Dispose`. An addon that is already open when you
register gets no `PostSetup`. `IAddonEventManager.AddEvent` takes a **node**, never a component (a
component address crashes). Events, args types, examples → `references/topic-addons.md`

