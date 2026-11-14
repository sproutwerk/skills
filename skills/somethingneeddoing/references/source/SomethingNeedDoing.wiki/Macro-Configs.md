Macros can contain config values that can be rendered in the settings tab. These will be saved along side the macro itself, but not overwritten when the macro updates from an outside source (e.g. github) as is the case if you just have local variables inside the macro.

**For Lua macros:**
```lua
--[=====[
[[SND Metadata]]
author: croizat
version: 1.0.0
description: A macro with configurable parameters
configs:
  loopCount:
    default: 5
    description: Number of times to loop the process
    type: int
    min: 1
    max: 100
    required: true
  targetQuality:
    default: 100
    description: Target quality percentage
    type: int
    min: 0
    max: 100
    required: false
  useHQMats:
    default: true
    description: Whether to use HQ materials
    type: bool
    required: false
  itemName:
    default: Iron Ore
    description: Name of the item to craft
    type: string
    required: false
[[End Metadata]]
--]=====]

local loopCount = Config.Get("loopCount")
local targetQuality = Config.Get("targetQuality")
local useHQMats = Config.Get("useHQMats")
local itemName = Config.Get("itemName")
    
yield("/echo Example macro config values:")
yield("/echo Loop count: " .. loopCount)
yield("/echo Target quality: " .. targetQuality)
yield("/echo Use HQ materials: " .. tostring(useHQMats))
yield("/echo Item name: " .. itemName)
```