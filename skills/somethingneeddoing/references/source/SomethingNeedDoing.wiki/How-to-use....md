# How to use "X" function
This is kinda just a place for us to write guides to help each other with common questions. Some might be simple, some might be more complicated. But hopefully this serves as a general place for us to figure out how shit works.

## How to read node text
<details> <summary><strong>Click me for details!</strong></summary>

- Open a Ui Debugger  
  - This can be done by a couple of different methods, but I'm just going to suggest using the build in one from dalamud.    
  - Type in the command `/xldata` -> Go to the "Addon Inspector" tab on the left.   
  - `Addon Inspector v2 (Testing)` is also good to use/acceptable  
- Find the Ui element that you're looking for (In my case, I was testing it with Island Sanc since updating that shit rn >.>  
- Find the dropdown Res Node  
- Follow the chain down to the text that you're trying to grab until you eventually find the text that you're looking for.   
- Either type down, or do the smart way of just clicking Text in the GetAddon->GetNode() part of the very useful lua functions to copy it (this is a good feature, please use it)  
- Fix it up to make sure the first line has the AddonName properly in " " and the set of numbers are all in the ( )  
profit  
<img width="1035" height="1128" alt="image" src="https://github.com/user-attachments/assets/f612cd03-263a-411f-8feb-07d68ff691f8" />
<img width="876" height="267" alt="image" src="https://github.com/user-attachments/assets/e7ae88f0-b531-470b-bf1e-02c49fd792f4" />
</details>

## How to setup a config for scripts

<details> <summary><strong>Click me for details!</strong></summary>

Setting up a configuration for a script is important for people to be able to edit settings for your script. It's the bread and butter for your little mini-plugin ~~for everyone who doesn't want to create an actual plugin you gremlins~~.

First things first, you need to have a metadata setup for your script. This goes at the very beginning of it all
```lua
--[=====[
[[SND Metadata]]
author: Ice
version: 1.0.0
description: "Just a funny little script for testing shit"

configs:
  stringConfig:
    default: Custom Default
    description: A general copy/paste of a string configuration for you to use
  inputInt:
    default: 1
    description: A funny little input for ints, you don't need min/max if you don't want it, it's just there for limits you can set.
    min: 0
    max: 100
  inputFloat/Double:
    default: 1.0
    description: I want to float away, I wanna beeee a decimal 
    min: 0.0
    max: 6.9
  boolInput:
    default: true
    description: "I'm just a fancy little checkmark"
  jobDropdownList:
    description: Class selection dropdown
    is_choice: true
    choices: ["Culinarian", "Alchemist", "Carpenter"]
  jobInputList:
    description: input what jobs you would like to do  |  
                 Press enter to add entries to the list
    is_choice: false
    choices: []

[[End Metadata]]
--]=====]

textInput = Config.Get("stringConfig")
intInput  = Config.Get("inputInt")
floatInput = Config.Get("inputFloat/Double")
class     = Config.Get("jobDropdownList")
options   = Config.Get("jobInputList")

Dalamud.Log(textInput)
Dalamud.Log(intInput)
Dalamud.Log(floatInput)
Dalamud.Log(class)
Dalamud.Log (options[0])
```

Lets break it down slightly.
`author`, `version`, `description`, and `plugin dependencies` can all be setup within the settings.  

`configs:` are the bread and butter of it all. It lets you setup all the variables that you want to setup for your script, and throw it right in their face so they can see what all they need to edit.  

-> `Config Name`: What's the name of this config? Give it something useful so people know what exactly it's for.   
-> `default`: what's the default variable here? If it's a number, give it a number. If it's a string, give it a name, bool? True/false   

Here's a couple of default templates you can use for a quick copy/paste into your thing:  
### String
```lua
  stringConfig:
    default: Custom Default
    description: A general copy/paste of a string configuration for you to use
```
<img width="473" height="71" alt="image" src="https://github.com/user-attachments/assets/e789f6b3-596f-4136-bc01-52709b27bfc5" />

### Int, Float, and Double Inputs
```lua
  inputInt:
    default: 1
    description: A funny little input for ints, you don't need min/max if you don't want it, it's just there for limits you can set.
    min: 0
    max: 100
```
```lua
  inputInt:
    default:
    description:
    min:
    max:
```
<img width="747" height="76" alt="image" src="https://github.com/user-attachments/assets/091d1264-c7f9-4a77-bcf4-b6a6239906ad" />

```lua
  inputFloat/Double:
    default: 1.0
    description: I want to float away, I wanna beeee a decimal 
    min: 0.0
    max: 6.9
```
```lua
  inputFloat/Double:
    default:
    description:
    min:
    max:
```
<img width="426" height="75" alt="image" src="https://github.com/user-attachments/assets/c32a9b66-c55a-48d8-8cda-d95a2508d11e" />

### Bool
```lua
  boolInput:
    default: true
    description: "I'm just a fancy little checkmark"
```
```lua
  boolInput:
    default: true
    description:
```
<img width="279" height="76" alt="image" src="https://github.com/user-attachments/assets/b5ada697-8915-472b-904b-9d4f7f76a198" />


### Dropdown | List Inputs  
These are special in the sense that `type` and `is_choice` are important here for two reasons.   
-> `type` defines it to be a list, and it required for it to function in either form.  
-> `is_choice` is a bool that lets you either  
  -> lets you define a pre-set list of options that someone can choose from
  -> let users input multiple items into  
Both have their use cases. 

```lua
  jobDropdownList:
    description: Class selection dropdown
    choices: ["Culinarian", "Alchemist", "Carpenter"]
    is_choice: true
```
```lua
  jobDropdownList:
    description: 
    is_choice: true
    choices: []
```
<img width="285" height="143" alt="image" src="https://github.com/user-attachments/assets/cf1f1f62-2a0e-42e8-ab8b-84237131447c" />


### List Input
```lua
  jobInputList:
    description: input what jobs you would like to do  |  
                 Press enter to add entries to the list
    is_choice: false
    choices: []
```
```lua
  jobInputList:
    description:
    is_choice: false
    choices: []
```
<img width="555" height="162" alt="image" src="https://github.com/user-attachments/assets/8175bf92-2d15-4ffe-a8a3-17369fbfe388" />

## Grabbing values from the user inputs:
Grabbing the variable that the user inputted is simple. All you need to do is assign it to a variable of your choosing, and use `Config.Get("jobInputList")`. **Make sure that the config you're grabbing from matches exactly**.  
List can just be quickly assigned and can be found using basic lua functions of sorting through the list if you're either iterating them, or sorting through them. 

```lua
textInput = Config.Get("stringConfig")
intInput  = Config.Get("inputInt")
floatInput = Config.Get("inputFloat/Double")
class     = Config.Get("jobDropdownList")
options   = Config.Get("jobInputList")
```

</details>

## Callbacks and How to use them (Stolen from Wayback Machine)
<details> <summary><strong>Click me for details!</strong></summary>
NOTE: Pulled from prior notes - Images are likely broken and will need to be re-added.

Alright, this is a lot to explain in itself so bear with me as I try to explain this.

Disclaimer. These can crash your game if you try and fire/activate it while that addon isn't active. No
Callback -> /callback is a way to directly interact with certain elements of the game. The values of these can range -a lot-
Simpletweaks is also virtually required to be able to tell what you're looking for. Once installed, type in tweaks Debug (yes, it's case sensitive)

This will pop up the following window, you want to go Addon Logging, make sure you're on the Callbacks tab, and Enable Logging

image

From here, it's just clicking on the ui element that you're trying to click on. In the example below. Lets say you wanted to make it to where it would click on Recommended Gear for you. With the logging enabled, you would get the following:

image

and the command would look like: /callback Character true 12

-> Character is the addon in this case
-> true is the Update Visibility
-> 12 is the value(s) that are necessary for this to go off properly.

There is A lot of different ones that can be used. Some that honestly I'm still trying to figure out how to make work.

Another good example for people who want to really dwelve/see how strong this is.

Lets say you're buying an item from a vendor (Sabina in Idyllshire is going to be my dummy for this)

If you click on a shop item from here (using the top of the list as an example here) image

You'll get the following
image

-> ShopExchangeItem is the addon name
-> true for the visibility
-> 0 is a value that doesn't really matter to us here/doesn't effect the rest of them
-> the 2nd 0 pertains to which item you're selecting. Top of the list usually starts at 0, and works it's way up
-> the 3rd number (1 in this case`) pertains to How many of those items you're buying. And this can in fact bypass how many items a vendor is normally locked to

image image

Here, I just interacted w/ the vendor like normal, just turning in 1 items worth of materials to get gear from them

Now if I were to change that 1 to a 5 now on an item that I can actually buy and run the following: /callback ShopExchangeItem true 0 3 5

image image

Now I have 5 unique pieces of armor
image

(Yes, you can have multiple unique items, no, you can't equip them as far as I can tell)

But these are the 2 kinds of extreme's. It's mainly a lot of poking around and seeing how certain addon's fire off and how they interact, and a lot of trial and error. But congrats, you're one step closer to learning how plugin makers interact with menu's!
</details>
