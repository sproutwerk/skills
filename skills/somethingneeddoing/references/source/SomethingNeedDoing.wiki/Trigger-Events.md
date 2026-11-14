# Trigger Events

Both native macros and Lua scripts can be run at any time via the UI or commands, but they can also be set to run automatically on certain **Trigger Events**. These come in two forms:

## Macro-Level Trigger Events

Macro-level trigger events are assigned in the metadata of a macro and cause the **entire macro** to be executed when the selected event occurs.

### Available Trigger Events

See the auto-generated docs inside the plugin

## Function-Level Trigger Events

Function-level trigger events are exclusive to Lua scripts and will trigger an **individual function** while the script is running. They are created by naming functions with the trigger event name as a prefix.

### Function Naming Convention

Functions must start with the trigger event name (case-insensitive):

```lua
function OnTerritoryChange()
    -- This function will be called when territory changes
    print("Territory changed!")
end

function OnCombatStart()
    -- This function will be called when combat starts
    print("Combat started!")
end
```

### OnAddonEvent Function Naming

For addon events, the function name must follow this pattern:

```
OnAddonEvent_AddonName_EventType
```

Example:

```lua
function OnAddonEvent_SelectYesno_PostSetup()
    -- This function will be called when the SelectYesno addon's PostSetup event fires
    print("Yes/No dialog appeared!")
end
```

### OnChatMessage Filters

The `OnChatMessage` trigger can be configured with filters to only fire when specific chat messages are received. This is basically a requirement as without any filters, the trigger fires too often and will lag the game to unplayable levels.

#### Macro-Level Chat Filters

For macro-level triggers, add a `chat_message_filter` section to your metadata:

```yaml
triggers:
  - OnChatMessage
chat_message_filter:
  channels:
    - Say
    - Shout
    - Party
  message_contains: "ready"
```

#### Function-Level Chat Filters

For function-level triggers, add a `function_chat_filters` section to your metadata, keyed by function name:

```yaml
function_chat_filters:
  OnChatMessage:
    channels:
      - Say
    message_contains: "hello"
  OnChatMessagePartyOnly:
    channels:
      - Party
    message_regex: ".*ready.*"
```

#### Available Filter Options

- **`channels`**: List of `[XivChatType](https://github.com/goatcorp/Dalamud/blob/c1df0da9beaec38c7132e519c60c04e56ec0caef/Dalamud/Game/Text/XivChatType.cs#L6)` values. If not specified, all channels are allowed.
- **`message_contains`**: String that the message must contain (case-insensitive). If not specified, no message content filter is applied.
- **`sender_contains`**: String that the sender name must contain (case-insensitive). If not specified, no sender filter is applied.
- **`message_regex`**: Regex pattern that the message must match (case-insensitive). If not specified, no regex filter is applied.

All filter conditions are combined with AND logic - all specified filters must pass for the event to fire.

#### Example Usage

**Macro-level filter:**
```lua
--[=====[
[[SND Metadata]]
triggers:
  - OnChatMessage
chat_message_filter:
  channels:
    - Say
    - Shout
  message_contains: "ready"
[[End Metadata]]
--]=====]

-- This macro will only start when Say/Shout messages contain "ready"
local message = TriggerData.message
print("Ready message received: " .. message)
```

**Function-level filters:**
```lua
--[=====[
[[SND Metadata]]
function_chat_filters:
  OnChatMessage:
    channels:
      - Say
    message_contains: "hello"
  OnChatMessageParty:
    channels:
      - Party
    message_regex: ".*ready.*"
[[End Metadata]]
--]=====]

function OnChatMessage()
    -- Only fires for Say channel messages containing "hello"
    Engines.Run("/echo Hello detected: " .. TriggerData.message)
end

function OnChatMessageParty()
    -- Only fires for Party channel messages matching ".*ready.*"
    Engines.Run("/p Ready!")
end

-- Keep script running
while true do
    yield("/wait 1")
end
```

### Function Name Flexibility

The function name must _start with_ the trigger event name but doesn't have to be exactly that, allowing for descriptive function names:

```lua
function OnTerritoryChange_CheckForDiadem()
    -- This function will still be triggered by OnTerritoryChange
    if TriggerData.territoryType == 901 then
        print("Entered Diadem!")
    end
end
```

## TriggerData Access

When a trigger event occurs, you get access to the `TriggerData` table, which contains:

- **eventType** - The type of trigger event that fired
- **timestamp** - When the event occurred
- **Additional properties** - Varies by event type

### Event-Specific Data

Different trigger events provide different data in `TriggerData`:

```lua
function OnTerritoryChange()
    local changeTime = TriggerData.timestamp
    local territoryId = TriggerData.territoryType
    print("Territory changed to", territoryId, "at", changeTime)
end

function OnChatMessage()
    local messageType = TriggerData.type
    local sender = TriggerData.sender
    local message = TriggerData.message
    print("Chat:", messageType, "from", sender, ":", message)
end
```

## Important Limitations

### ⚠️ No Yielding in Function-Level Triggers

Function-level trigger functions **cannot use `yield()`**. As a workaround, you can use the `Engines` module to call another engine, in this case the native, but doing so executes the command immediately, meaning `waits` are not possible. This is a fundamental problem of triggers sharing state with the main macro, and it's either have one or the other.

```lua
-- ❌ WRONG
function OnTerritoryChange()
    yield("/e Territory changed!")  -- This will fail!
end

-- ✅ CORRECT
function OnTerritoryChange()
    Engines.Native.Run("/e Territory changed!")
end
```