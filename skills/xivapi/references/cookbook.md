# XIVAPI v2 cookbook — verified calls with output

Run 2026-09-23 against `version=541c0c12e07da325` (7.56x1 = `latest`),
`schema=exdschema@2:rev:e773c41a90aed788cf4c1c48469fa85618ef01fb`. Output is
shown as returned (long JSON trimmed with `…`). The `python3 -c` pipes only
reformatted output when this transcript was recorded; nothing in the skill
needs Python, and `jq` does the same reshaping. Every response also carries `"schema"` and `"version"`; omitted here.

```sh
S=https://v2.xivapi.com/api
q(){ curl -s -G "$S/search" --data-urlencode "sheets=$1" --data-urlencode "query=$2" --data-urlencode "fields=$3" --data-urlencode "limit=${4:-100}"; }
```

## Meta

```sh
curl -s $S/version | python3 -c 'import json,sys;print(json.load(sys.stdin)["versions"][-3:])'
# [{'key': '7583112015aaef5d', 'names': ['7.55']}, {'key': 'c3f948214b90e498', 'names': ['7.55x1']}, ...
#  {'key': '541c0c12e07da325', 'names': ['7.56x1', 'latest']}]

curl -s $S/sheet | python3 -c 'import json,sys;print(len(json.load(sys.stdin)["sheets"]))'
# 7912

curl -s $S/openapi.json | python3 -c 'import json,sys;d=json.load(sys.stdin);print(d["info"],list(d["paths"]))'
# {'title': 'boilmaster', 'version': '1-f0d5f0a'} ['/asset', '/asset/map/{territory}/{index}', '/search', '/sheet', '/sheet/{sheet}', '/sheet/{sheet}/{row}', '/version']
```

User-Agent (Cloudflare):

```sh
curl -s -o /dev/null -w '%{http_code}\n' -A 'Python-urllib/3.12' "$S/sheet/Item/48?fields=Name"   # 403
curl -s -o /dev/null -w '%{http_code}\n' -A 'xivapi-skill/0.1'   "$S/sheet/Item/48?fields=Name"   # 200
python3 -c 'import urllib.request;urllib.request.urlopen("https://v2.xivapi.com/api/sheet/Item/48")'  # HTTP Error 403: Forbidden
```

## Items

```sh
curl -s "$S/sheet/Item/48?fields=Name,Icon,ItemUICategory.Name,IsUntradable,StackSize"
# {"row_id":48,"fields":{"Icon":{"id":65122,"path":"ui/icon/065000/065122.tex","path_hr1":"ui/icon/065000/065122_hr1.tex"},
#   "IsUntradable":false,"ItemUICategory":{"value":63,"sheet":"ItemUICategory","row_id":63,"fields":{"Name":"Other"}},
#   "Name":"Allagan Tomestone of Mathematics","StackSize":2000}}

q Item 'Name="Ecliptic Arcanite"' Name
# {"results":[{"score":1.0,"sheet":"Item","row_id":50977,"fields":{"Name":"Ecliptic Arcanite"}}]}

q Item 'Name~"hi-potion"' Name 3      # partial match is case-insensitive; score = needle/field length
# 4552 Hi-Potion 1.0 | 4603 Hi-Potion of Mind 0.5294118 | 4621 Ice Ward Hi-Potion 0.5
q Item 'Name="potion"' Name           # exact match is case-sensitive
# {"results":[]}

q Item '+Name~"Allagan Tomestone of"' Name | python3 -c 'import json,sys;[print(r["row_id"],round(r["score"],3),r["fields"]["Name"]) for r in json.load(sys.stdin)["results"]]'
# 30 0.833 Allagan Tomestone of Law
# 32 0.8 Allagan Tomestone of Lore
# 34 0.741 Allagan Tomestone of Verity
# 38 0.741 Allagan Tomestone of Goetia
# 45 0.741 Allagan Tomestone of Comedy
# 28 0.714 Allagan Tomestone of Poetics
# 37 0.714 Allagan Tomestone of Genesis
# 26 0.69 Allagan Tomestone of Soldiery
# 35 0.69 Allagan Tomestone of Creation
# 40 0.69 Allagan Tomestone of Allegory
# 42 0.69 Allagan Tomestone of Aphorism
# 24 0.667 Allagan Tomestone of Mythology
# 31 0.667 Allagan Tomestone of Esoterics
# 33 0.667 Allagan Tomestone of Scripture
# 36 0.667 Allagan Tomestone of Mendacity
# 43 0.667 Allagan Tomestone of Astronomy
# 44 0.667 Allagan Tomestone of Causality
# 49 0.667 Allagan Tomestone of Mnemonics
# 28146 0.667 Allagan Tomestone of Frivolity
# 23 0.645 Allagan Tomestone of Philosophy
# 41 0.645 Allagan Tomestone of Revelation
# 46 0.645 Allagan Tomestone of Aesthetics
# 47 0.645 Allagan Tomestone of Heliometry
# 48 0.625 Allagan Tomestone of Mathematics
# 39 0.571 Allagan Tomestone of Phantasmagoria

curl -s "$S/sheet/TomestonesItem?fields=Item.Name,Tomestones" | python3 -c 'import json,sys;[print(r["row_id"],r["fields"]["Tomestones"]["value"],r["fields"]["Item"]["value"],r["fields"]["Item"]["fields"]["Name"]) for r in json.load(sys.stdin)["rows"] if r["fields"]["Tomestones"]["value"]]'
# 3 1 28 Allagan Tomestone of Poetics
# 21 4 47 Allagan Tomestone of Heliometry
# 22 2 48 Allagan Tomestone of Mathematics
# 23 3 49 Allagan Tomestone of Mnemonics          (Tomestones row 3 has WeeklyLimit 900)

q Item '+Name~"Tomestone" -Name~"Allagan"' Name 5
# 16541 Micro Tomestone 0.6 | 35831 Discal Tomestone 0.5625 | 38942 Orthos Tomestone 0.5625 | 40321 Hermetic Tomestone 0.5 | 9384 Encrypted Tomestone 0.47

q Item '+(Name="Potion" Name="Hi-Potion") +LevelItem>=1' 'Name,LevelItem@as(raw)'
# 4551 Potion LevelItem 10 score 2.0 | 4552 Hi-Potion LevelItem 25 score 2.0

q Item '+LevelItem>=790 +ItemUICategory=2' 'Name,LevelItem@as(raw)' 5
# 49581 Augmented Bygone Brass Sword 790 | 49658 Grand Champion's Falchion 795 | 50978 Phantom Sword Eclipticum 790 | 51000 Phantom Sword Occultum 795 | 52299 Palazzo Diamond Sword 795

q Item '+ItemUICategory.Name="Other" +Name~"Tomestone"' Name 3     # nested relation in query
# 16541 Micro Tomestone | 35831 Discal Tomestone | 40321 Hermetic Tomestone
```

Language:

```sh
curl -s -G "$S/sheet/Item/48" --data-urlencode 'fields=Name,Name@lang(de)' --data-urlencode 'language=ja'
# {"row_id":48,"fields":{"Name":"アラガントームストーン:数理","Name@lang(de)":"Allagischer Stein der Mathematik"}}
q Item 'Name@ja="新月鉱石"' 'Name,Name@lang(ja)'
# 50977 {"Name":"Ecliptic Arcanite","Name@lang(ja)":"新月鉱石"}
curl -s -G "$S/search" --data-urlencode sheets=Item --data-urlencode 'query=Name="新月鉱石"' --data-urlencode fields=Name --data-urlencode language=ja
# 50977 {"Name":"新月鉱石"}
curl -s "$S/sheet/SpecialShopItemCategory/4?fields=Name&language=en"
# {"row_id":4,"fields":{"Name":"その他"}}        <- single-language sheet, ja only
```

## NPCs and shops

```sh
q ENpcResident 'Singular="Ermina"' Singular,Title
# {"results":[{"score":1.0,"sheet":"ENpcResident","row_id":1053728,"fields":{"Singular":"Ermina","Title":""}},
#             {"score":1.0,"sheet":"ENpcResident","row_id":1053904,"fields":{"Singular":"Ermina","Title":"Tomestone Exchange"}}]}

curl -s -G "$S/sheet/ENpcBase/1053904" --data-urlencode 'fields=ENpcData@as(raw)'
# {"row_id":1053904,"fields":{"ENpcData@as(raw)":[1704275,1770925,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0]}}

curl -s -G "$S/sheet/ENpcBase/1053904" --data-urlencode 'fields=ENpcData[].Name'
# {"row_id":1053904,"fields":{"ENpcData":[{"value":1704275,"sheet":"Story","row_id":1704275,"fields":{}},
#   {"value":1770925,"sheet":"SpecialShop","row_id":1770925,"fields":{"Name":"Phantom Weapon Materials Exchange"}},{"value":0},…]}}

curl -s -G "$S/sheet/ENpcBase/1053611" --data-urlencode 'fields=ENpcData[].Name'      # Jeffroy
# [{'value': 1704274, 'sheet': 'Story', 'row_id': 1704274, 'fields': {}},
#  {'value': 721825, 'sheet': 'CustomTalk', 'row_id': 721825, 'fields': {'Name': 'CtsMkdEntrance_00929'}},
#  {'value': 71047, 'sheet': 'Quest', 'row_id': 71047, 'fields': {'Name': 'Occult Reunion'}}]

q ENpcBase 'ENpcData[]=1770925' ''            # shop -> NPC
# {"results":[{"score":1.0,"sheet":"ENpcBase","row_id":1053904,"fields":{}}]}
q ENpcBase 'ENpcData[1]=1770925' ''           # array index allowed in queries
# 1053904

curl -s -G "$S/sheet/SpecialShop/1770925" --data-urlencode 'fields=Name,UseCurrencyType,Item[].Item[].Name,Item[].ReceiveCount,Item[].ItemCost@as(raw),Item[].CurrencyCost,Item[].CostType' \
 | python3 -c 'import json,sys;d=json.load(sys.stdin);f=d["fields"];print(f["Name"],"UseCurrencyType",f["UseCurrencyType"],len(f["Item"]),"slots");[print(json.dumps(e)) for e in f["Item"][:4]]'
# Phantom Weapon Materials Exchange UseCurrencyType 4 60 slots
# {"CostType": [2, 0, 0], "CurrencyCost": [500, 0, 0], "Item": [{"value": 47750, "sheet": "Item", "row_id": 47750, "fields": {"Name": "Arcanite"}}, {"value": 0, "sheet": "Item", "row_id": 0, "fields": {"Name": ""}}], "ItemCost@as(raw)": [2, 0, 0], "ReceiveCount": [1, 1]}
# {… "Item": [{"value": 46850, … "Waxing Arcanite"} …], "ItemCost@as(raw)": [2, 0, 0], "CurrencyCost": [500, 0, 0]}
# {… "Item": [{"value": 50058, … "Waning Arcanite"} …], "ItemCost@as(raw)": [2, 0, 0], "CurrencyCost": [500, 0, 0]}
# {… "Item": [{"value": 50977, … "Ecliptic Arcanite"} …], "ItemCost@as(raw)": [2, 0, 0], "CurrencyCost": [500, 0, 0]}
# UseCurrencyType=4 => ItemCost is a TomestonesItem.Tomestones slot: 2 = Item 48 Mathematics. Expanded, ItemCost[0] shows Item 2 "Fire Shard" — wrong.
# Garland confirms: npc/1053904 shops[0].entries[0] = {"item":[{"id":"47750","amount":1}],"currency":[{"id":"48","amount":500}]}

curl -s -G "$S/sheet/SpecialShop/1770925" --data-urlencode 'fields=Item[].ReceiveItems[].Item.Name'   # EXDSchema relation names: silently empty
# {"row_id":1770925,"fields":{"Item":[{},{},{},…]}}
curl -s -G "$S/sheet/SpecialShop/1770925" --data-urlencode 'fields=Item[0].Item[0].Name'
# {"code":400,"message":"invalid request: Failed to deserialize query string: fields: invalid request: error Eof at: [0].Item[0].Name"}

q SpecialShop 'Item[].Item[]=50977' Name       # item -> shop
# 1770925 Phantom Weapon Materials Exchange

# GilShop: GilShopItem is a subrow sheet keyed by the GilShop row id
q ENpcBase '+ENpcData[]>=262144 +ENpcData[]<=262200' 'ENpcData[].Name' 3
# 1000185 [('GilShop', 262157, '')] | 1000232 [('GilShop', 262185, '')] | 1000236 [('GilShop', 262182, '')]
curl -s "$S/sheet/ENpcResident/1000185?fields=Singular,Title"
# {"row_id":1000185,"fields":{"Singular":"Tanie","Title":"Florist"}}
curl -s "$S/sheet/GilShopItem?after=262156:65535&limit=40&fields=Item.Name,Item.PriceMid" | python3 -c 'import json,sys;[print(r["subrow_id"],r["fields"]["Item"]["fields"]["Name"],r["fields"]["Item"]["fields"]["PriceMid"]) for r in json.load(sys.stdin)["rows"] if r["row_id"]==262157]' | head -5
# 0 Carnation 13
# 1 Cloves 16
# 2 Fish Meal 21
# 3 Krakka Root Seeds 3480
# 4 Oldrose Seeds 300
curl -s "$S/sheet/GilShopItem?rows=262144&fields=Item.Name"          # rows= gives subrow 0 only
# {"rows":[{"row_id":262144,"subrow_id":0,"fields":{"Item":{…"Name":"Potion of Strength"}}}]}
curl -s "$S/sheet/GilShopItem/262144:1?fields=Item.Name"
# {"row_id":262144,"subrow_id":1,"fields":{"Item":{"value":4595,…"Name":"Potion of Dexterity"}}}
```

## Aetherytes

```sh
curl -s -G "$S/sheet/Aetheryte/239" --data-urlencode 'fields=PlaceName.Name,AethernetName.Name,Territory@as(raw),IsAetheryte,Invisible,AethernetGroup,Level[].X,Level[].Y,Level[].Z,Level[].Territory@as(raw),Map@as(raw),Order'
# {"row_id":239,"fields":{"AethernetGroup":14,"AethernetName":{"value":4757,"sheet":"PlaceName","row_id":4757,"fields":{"Name":"Phantom Village"}},
#   "Invisible":true,"IsAetheryte":false,
#   "Level":[{"value":12250673,"sheet":"Level","row_id":12250673,"fields":{"Territory@as(raw)":1278,"X":43.3027,"Y":0.0199971,"Z":-1.07516}},{"value":0},{"value":0},{"value":0}],
#   "Map@as(raw)":0,"Order":0,"PlaceName":{"value":0,"sheet":"PlaceName","row_id":0,"fields":{"Name":""}},"Territory@as(raw)":1278}}

curl -s -G "$S/sheet/Aetheryte/216" --data-urlencode 'fields=PlaceName.Name,AethernetName.Name,Territory@as(raw),IsAetheryte,Invisible,AethernetGroup,Level[].X,Map@as(raw)'
# {"row_id":216,"fields":{"AethernetGroup":14,"AethernetName":{…"Name":"Tuliyollal Aetheryte Plaza"},"Invisible":false,"IsAetheryte":true,
#   "Level":[{"value":10566589},{"value":10566590},{"value":10566591},{"value":10566592}],     <- links don't resolve
#   "Map@as(raw)":855,"PlaceName":{"value":4504,…"Name":"Tuliyollal"},"Territory@as(raw)":1185},"transient":{"Unknown0":false}}
curl -s "$S/sheet/Level/10566589?fields=X"
# {"code":404,"message":"not found: the Excel row (none)/10566589:0 could not be found"}
# (also absent from ffxiv-datamining csv/en/Level.csv — the EXD Level sheet simply doesn't contain it)

q Aetheryte 'Territory=1278' 'PlaceName.Name,AethernetName.Name,IsAetheryte,AethernetGroup,Level[].X,Level[].Y,Level[].Z'
# 239 score 1.0 PlaceName '' AethernetName 'Phantom Village' IsAetheryte False group 14 Level (43.3027, 0.0199971, -1.07516)

q Aetheryte 'PlaceName.Name="Tuliyollal" +IsAetheryte=true' 'PlaceName.Name,Territory@as(raw),AethernetGroup'
# 216 score 2.0 Tuliyollal Territory 1185 group 14   ...followed by EVERY IsAetheryte=true row at score 1.0 (bare clause is optional) + "next" cursor
q Aetheryte '+IsAetheryte=true +Territory=1185' PlaceName.Name
# {"results":[{"score":2.0,"sheet":"Aetheryte","row_id":216,"fields":{"PlaceName":{…"Name":"Tuliyollal"}}}]}

q Aetheryte '+IsAetheryte=true' PlaceName.Name 500 | python3 -c 'import json,sys;d=json.load(sys.stdin);print(len(d["results"]),d.get("next"))'
# 108 None
```

MapMarker for main aetheryte positions (map pixels, 0–2048):

```sh
q MapMarker '+DataType=3 +DataKey=216' 'X,Y,Icon.id,DataType,DataKey@as(raw),PlaceNameSubtext.Name'
# {"results":[{"score":2.0,"sheet":"MapMarker","row_id":620,"subrow_id":9,"fields":{"DataKey@as(raw)":216,"DataType":3,
#   "Icon":{"id":60453,…},"PlaceNameSubtext":{…"Name":"Tuliyollal Aetheryte Plaza"},"X":1070,"Y":911}}]}
curl -s "$S/sheet/Map/855?fields=Id,MapMarkerRange,SizeFactor,OffsetX,OffsetY,PlaceName.Name"
# {"row_id":855,"fields":{"Id":"y6t1/00","MapMarkerRange":620,"OffsetX":50,"OffsetY":-70,"PlaceName":{…"Name":"Tuliyollal"},"SizeFactor":180}}
# map coord = px*41/(1.8*2048)+1 -> (12.9, 11.1); world = (px-1024)/1.8 - Offset -> X -24.4, Z 7.2
# MapMarker rows for a map: after=<MapMarkerRange-1>:65535, keep row_id==MapMarkerRange (Map 1025 -> 660: 7 subrows)
```

## Territories and maps

```sh
curl -s -G "$S/sheet/TerritoryType/1278" --data-urlencode 'fields=Name,Bg,PlaceName.Name,PlaceNameRegion.Name,PlaceNameZone.Name,Map.Id,Map.OffsetX,Map.OffsetY,Map.SizeFactor,Map.PlaceName.Name,ContentFinderCondition@as(raw),Aetheryte@as(raw),ExVersion.Name,TerritoryIntendedUse@as(raw),IsPvpZone,Mount'
# {"row_id":1278,"fields":{"Aetheryte@as(raw)":216,"Bg":"ex5/03_ocn_o6/evt/o6e1/level/o6e1","ContentFinderCondition@as(raw)":0,
#   "ExVersion":{"value":5,…"Name":"Dawntrail"},"IsPvpZone":false,
#   "Map":{"value":1025,"sheet":"Map","row_id":1025,"fields":{"Id":"o6e1/00","OffsetX":0,"OffsetY":40,"PlaceName":{…"Name":"Phantom Village"},"SizeFactor":400}},
#   "Mount":false,"Name":"o6e1_2","PlaceName":{"value":4757,…"Name":"Phantom Village"},"PlaceNameRegion":{"value":2405,…"Name":"???"},
#   "PlaceNameZone":{"value":519,…"Name":""},"TerritoryIntendedUse@as(raw)":7},"transient":{"OffsetZ":-10000}}

curl -s -G "$S/sheet/Map/1025" --data-urlencode 'fields=Id,OffsetX,OffsetY,SizeFactor,PlaceName.Name,TerritoryType@as(raw),MapMarkerRange,MapIndex'
# {"row_id":1025,"fields":{"Id":"o6e1/00","MapIndex":0,"MapMarkerRange":660,"OffsetX":0,"OffsetY":40,"PlaceName":{…"Name":"Phantom Village"},"SizeFactor":400,"TerritoryType@as(raw)":1269}}
#   ^ Map.TerritoryType is 1269, not 1278: several TerritoryType rows share Map 1025

q TerritoryType 'PlaceName.Name="Tuliyollal"' 'Name,Bg,PlaceName.Name,Map@as(raw),TerritoryIntendedUse@as(raw)'
# 1185 {"Bg":"ex5/02_ykt_y6/twn/y6t1/level/y6t1","Map@as(raw)":855,"Name":"y6t1","PlaceName":{…"Tuliyollal"},"TerritoryIntendedUse@as(raw)":0}

q ContentFinderCondition 'Name~"Vanguard"' 'Name,TerritoryType@as(raw),ContentType.Name,ClassJobLevelRequired,ItemLevelRequired'
# 831 {"ClassJobLevelRequired":97,"ContentType":{…"Name":"Dungeons"},"ItemLevelRequired":645,"Name":"Vanguard","TerritoryType@as(raw)":1198}

q PlaceName 'Name="Phantom Village"' Name
# 4757

curl -s -o map.jpg -w '%{http_code} %{content_type} %{size_download}\n' "$S/asset/map/o6e1/00"
# 200 image/jpeg 388107     (2048x2048; format=png -> 6215440 bytes)
```

## Level (object placements)

```sh
curl -s -G "$S/sheet/Level/12250673" --data-urlencode 'fields=X,Y,Z,Yaw,Radius,Type,Object@as(raw),Territory@as(raw),Map@as(raw)'
# {"row_id":12250673,"fields":{"Map@as(raw)":1025,"Object@as(raw)":0,"Radius":1.0,"Territory@as(raw)":1278,"Type":40,"X":43.3027,"Y":0.0199971,"Yaw":-1.55241,"Z":-1.07516}}

q Level '+Territory=1278 +Type=8 +Object=1053904' 'X,Y,Z'        # Ermina: no placement row
# {"results":[]}
q Level '+Territory=1278 +Type=8' 'X,Y,Z,Object@as(raw)' 500 | python3 -c 'import json,sys;d=json.load(sys.stdin);print(len(d["results"]))'
# 41
q Level 'Object=1053905' 'X,Z,Type,Territory@as(raw)'
# {"results":[{"score":1.0,"sheet":"Level","row_id":11559029,"fields":{"Territory@as(raw)":1278,"Type":8,"X":29.5763,"Z":10.7985}}]}
curl -s "$S/sheet/ENpcResident?rows=1053611,1053803,1053906,1053907,1053909,1053905&fields=Singular,Title" | python3 -c 'import json,sys;[print(r["row_id"],r["fields"]["Singular"],"|",r["fields"]["Title"]) for r in json.load(sys.stdin)["rows"]]'
# 1053611 Jeffroy | Expedition Coordinator
# 1053803 Ketenramm |
# 1053906 Lydirceil |
# 1053907 Gerolt |
# 1053909 Dodokkuli |
# 1053905 Dodokkuli | Phantom Weapon Exchange

q Level '+Territory=1278 +Type=8' 'X,Z,Object.ENpcData@as(raw)' 1      # Object resolves via the Type switch (8 -> ENpcBase)
# {"score":2.0,"sheet":"Level","row_id":11505102,"fields":{"Object":{"value":1053611,"sheet":"ENpcBase","row_id":1053611,"fields":{"ENpcData@as(raw)":[1704274,721825,71047,0,…]}},"X":-77.9279,"Z":-15.3653}}

curl -s "$S/sheet/Level?limit=3&fields=X,Type,Territory@as(raw)"
# {"rows":[{"row_id":1034660,"fields":{"Territory@as(raw)":1,"Type":8,"X":8.29875}},{"row_id":1034661,…},{"row_id":1034662,…}]}
curl -s "$S/sheet/Level?rows=12250673,10566589&fields=X"           # one missing row fails the whole call
# {"code":404,"message":"not found: the Excel row (none)/10566589:0 could not be found"}
```

## Actions, statuses, BNpcs, quests

```sh
q Action 'Name="Sprint"' 'Name,ClassJob@as(raw),ClassJobLevel,IsPvP,Icon.id' 10 | python3 -c 'import json,sys;[print(r["row_id"],r["fields"]) for r in json.load(sys.stdin)["results"]]'
# 3 {'ClassJob@as(raw)': 0, 'ClassJobLevel': 0, 'Icon': {'id': 104, …}, 'IsPvP': False, 'Name': 'Sprint'}
# 16933 {'ClassJob@as(raw)': -1, … 'Icon': {'id': 405 …}, 'IsPvP': False, 'Name': 'Sprint'}
# 16986, 28268, 28269, 38655 (same as 16933)
# 29057 {'ClassJob@as(raw)': 0, … 'Icon': {'id': 9583 …}, 'IsPvP': True, 'Name': 'Sprint'}

curl -s -G "$S/sheet/Action/3" --data-urlencode 'fields=Name' --data-urlencode 'transient=Description'
# {"row_id":3,"fields":{"Name":"Sprint"},"transient":{"Description":"Increases movement speed.\nDuration: 10s (20s when not in combat)\nAdditional Effect: Grants Jog when the effect expires\n…"}}
curl -s -G "$S/sheet/Action/3" --data-urlencode 'fields=Name' --data-urlencode 'transient=Description@as(html)'
# …"Description@as(html)":"Increases movement speed.<br><span style=\"color:rgba(0,204,34,1);\">Duration:</span> 10s …"

q Status 'Name="Sprint"' 'Name,Description,Icon.id,MaxStacks'
# 50, 481, 1342, 1548, 1938, 4520 — all {"Description":"Movement speed is increased.","Icon":{"id":210101,…},"MaxStacks":0,"Name":"Sprint"}

q BNpcName 'Singular="Ifrit"' Singular,Plural
# {"results":[{"score":1.0,"sheet":"BNpcName","row_id":1185,…},{"score":1.0,"sheet":"BNpcName","row_id":9288,…}]}

q Quest 'Name="The Ultimate Weapon"' 'Name,Id,ClassJobLevel,Expansion.Name,IssuerStart@as(raw),IssuerLocation@as(raw),PreviousQuest[].Name'
# 66060 {"ClassJobLevel":[50,0],"Expansion":{…"A Realm Reborn"},"Id":"ManFst503_00524","IssuerLocation@as(raw)":0,"IssuerStart@as(raw)":1034221,
#        "PreviousQuest":[{"value":69409,…"Name":"Rock the Castrum"},{"value":0},{"value":0}]}
# 70058 {… "Id":"XxcFst503_04522","IssuerLocation@as(raw)":4270118,"IssuerStart@as(raw)":1006657 …}
curl -s -G "$S/sheet/Quest/65537" --data-urlencode 'fields=Name,PreviousQuest@as(raw)'
# {"row_id":65537,"fields":{"Name":"A Good Adventurer Is Hard to Find","PreviousQuest@as(raw)":[65575,0,0]}}
q Quest 'Name="x"' 'PreviousQuest[]@as(raw)'
# {"code":400,"message":"… fields: invalid request: error Eof at: @as(raw)"}

q 'Item,Action,Status' 'Name="Sprint"' Name 10 | python3 -c 'import json,sys;print([(r["sheet"],r["row_id"]) for r in json.load(sys.stdin)["results"]])'
# [('Action', 3), ('Action', 16933), ('Action', 16986), ('Action', 28268), ('Action', 28269), ('Action', 29057), ('Action', 38655), ('Status', 50), ('Status', 481), ('Status', 1342)]
```

## Pagination, limits, sheet listing

```sh
q Item 'Name~"Potion"' Name 3
# [(4551, 'Potion'), (4554, 'X-Potion'), (4552, 'Hi-Potion')] next 1a53a40e-dca1-48cc-a3be-cdd4e5e5d4cc
curl -s -G "$S/search" --data-urlencode 'cursor=1a53a40e-dca1-48cc-a3be-cdd4e5e5d4cc' --data-urlencode 'limit=3'
# [(13637, 'Max-Potion'), (4553, 'Mega-Potion'), (24521, 'Potion Rack')] next d7c1b7b8-…
curl -s -G "$S/search" --data-urlencode 'cursor=…' --data-urlencode 'limit=2' --data-urlencode 'fields=Name,LevelItem@as(raw)'   # fields may change per page
# [(13637, {'LevelItem@as(raw)': 150, 'Name': 'Max-Potion'}), (4553, {'LevelItem@as(raw)': 45, 'Name': 'Mega-Potion'})]

q Item 'Name~"a"' Name 1000 | python3 -c 'import json,sys;d=json.load(sys.stdin);print(len(d["results"]),d["next"])'
# 500 3e98b3c9-…                                   (search page cap 500; limit=300 -> 300, limit=501 -> 500; default 100)
curl -s "$S/sheet/Item?limit=1000&fields=Name" | python3 -c 'import json,sys;print(len(json.load(sys.stdin)["rows"]))'
# 500                                              (sheet page cap)

curl -s "$S/sheet/Item?rows=48,50977&fields=Name"
# {"rows":[{"row_id":48,"fields":{"Name":"Allagan Tomestone of Mathematics"}},{"row_id":50977,"fields":{"Name":"Ecliptic Arcanite"}}]}
curl -s "$S/sheet/Item?limit=2&after=48&fields=Name"
# {"rows":[{"row_id":49,"fields":{"Name":"Allagan Tomestone of Mnemonics"}},{"row_id":50,"fields":{"Name":""}}]}
curl -s "$S/sheet/Item?limit=2&fields=Name"
# {"rows":[{"row_id":0,"fields":{"Name":""}},{"row_id":1,"fields":{"Name":"Gil"}}]}
```

## Errors

```sh
curl -s "$S/sheet/NopeSheet"
# {"code":404,"message":"not found: the Excel sheet \"NopeSheet\" could not be found"}
q Item 'Name=Ecliptic' Name
# {"code":400,"message":"invalid request: Failed to deserialize query string: query: invalid request: error Char at: Ecliptic"}
q Story 'Name~"a"' ''                    # Story has no Name column
# {"code":400,"message":"invalid request: malformed search query: no queries could be resolved"}
q NopeSheet 'Name="x"' ''
# {"code":400,"message":"invalid request: malformed search query: no queries could be resolved"}
curl -s "$S/search?sheets=Item&query=+Name=%22Potion%22&fields=Name"          # raw + decodes to space
# {"code":400,"message":"… query: invalid request: error AlphaNumeric at:  Name=\"Potion\""}
curl -s "$S/search?sheets=Item&query=%2BName=%22Potion%22&fields=Name"
# {"results":[{"score":1.0,"sheet":"Item","row_id":4551,"fields":{"Name":"Potion"}}]}
curl -s "$S/sheet/Item/48?fields=Nope,Name"
# {"row_id":48,"fields":{"Name":"Allagan Tomestone of Mathematics"}}         (unknown field silently dropped)
curl -s "$S/asset?path=ui/icon/065000/065122.tex"
# {"code":400,"message":"invalid request: Failed to deserialize query string: missing field `format`"}
curl -s "$S/asset?path=ui/icon/065000/999999.tex&format=png"
# {"code":404,"message":"not found: source file \"ui/icon/065000/999999.tex\" does not exist"}
```

## Version / schema pinning

```sh
curl -s -G "$S/sheet/Item/48" --data-urlencode fields=Name --data-urlencode version=7.2
# {"schema":"exdschema@2:rev:468f06ee4899b492745e99e77173ea3e4cf1ea53","version":"74f6e4449ea4d505","row_id":48,…}
curl -s -G "$S/sheet/Item/48" --data-urlencode fields=Name --data-urlencode version=nope
# {"code":400,"message":"invalid request: unknown version \"nope\""}
curl -s -G "$S/sheet/Item/48" --data-urlencode fields=Name --data-urlencode 'schema=exdschema@2:latest'
# {"schema":"exdschema@2:rev:f3cacf3bb7bcea2e1dcd2ce1204f8be22a6c9922","version":"541c0c12e07da325",…}   (newer than the default e773c41a)
curl -s -G "$S/sheet/Item/48" --data-urlencode fields=Name --data-urlencode 'schema=exdschema@2:rev:468f06ee4899b492745e99e77173ea3e4cf1ea53'
# {"schema":"exdschema@2:rev:468f06ee…","version":"541c0c12e07da325",…}
curl -s -G "$S/sheet/Item/48" --data-urlencode fields=Name --data-urlencode 'schema=exdschema@2'
# {"code":500,"message":"internal server error"}
curl -s -G "$S/sheet/Item/48" --data-urlencode fields=Name --data-urlencode 'schema=nope'
# {"code":400,"message":"invalid request: unknown schema source \"nope\""}

git ls-remote --heads https://github.com/xivdev/EXDSchema | tail -3
# 83e965d0… refs/heads/ver/2026.07.16.0001.0000
# e773c41a… refs/heads/ver/2026.09.01.0000.0000      <- server default
# f3cacf3b… refs/heads/latest
```

## Assets

```sh
curl -s -o icon.png -w '%{http_code} %{content_type} %{size_download}\n' "$S/asset?path=ui/icon/065000/065122_hr1.tex&format=png"
# 200 image/png 10244        -> PNG 72 x 72 RGBA
curl -s -o icon.png -w '%{http_code} %{size_download}\n' "$S/asset/ui/icon/065000/065122.tex?format=png"
# 200 3056                   -> 36 x 36
curl -s -o /dev/null -w '%{content_type} %{size_download}\n' "$S/asset?path=ui/icon/065000/065122.tex&format=jpg"
# image/jpeg 1371
```

## Alternatives

```sh
curl -s https://garlandtools.org/db/doc/item/en/3/48.json | python3 -c 'import json,sys;d=json.load(sys.stdin);print(list(d),json.dumps(d["item"])[:300])'
# ['item', 'partials'] {"name": "Allagan Tomestone of Mathematics", "description": "…", "id": 48, "patch": 7.2, "patchCategory": 5, "price": 2, "ilvl": 1, "category": 63, … "icon": "t/65122", "tradeCurrency": [{"shop": "Allagan Tomestones of Aphorism (DoW)", "npcs": [], "listings": [{"item": [{"id": "34850", "amount": 1}], "currency": [{"id": "48", "amount": 3…
curl -s https://garlandtools.org/db/doc/npc/en/2/1053904.json | head -c 400
# {"npc":{"name":"Ermina","id":1053904,"patch":7.25,"title":"Tomestone Exchange","appearance":{…},"alts":[1053728],"appalts":[1053728],"trade":1,"shops":[{"name":"Phantom Weapon Materials Exchange","entries":[{"item":[{"id":"47750","amount":1}],"currency":[{"id":"48","amount":500}]},…

curl -s 'https://universalis.app/api/v2/Aether/4551?listings=1&fields=itemID,lastUploadTime,minPrice,listings.pricePerUnit,listings.worldName'
# {"itemID":4551,"lastUploadTime":1790222001615,"listings":[{"pricePerUnit":1,"worldName":"Gilgamesh"}],"minPrice":1}
curl -s 'https://universalis.app/api/v2/Aether/50977?listings=1'      # untradable item
# {"type":"https://tools.ietf.org/html/rfc9110#section-15.5.5","title":"Not Found","status":404,…}

curl -sL 'https://raw.githubusercontent.com/xivapi/ffxiv-datamining/master/csv/en/Aetheryte.csv' | grep -n '^239,'
# 241:239,,,0,0,0,1,0,0,Aetheryte,12250673,0,0,0,70847,0,4757,1278,0,0,0,0,14,0,False,True
curl -sL -o /dev/null -w '%{http_code}\n' 'https://raw.githubusercontent.com/xivapi/ffxiv-datamining/master/csv/Aetheryte.csv'
# 404                       (language subdir is required: csv/en, csv/ja, csv/de, csv/fr)

curl -s -o /dev/null -w '%{http_code}\n' -A 'xivapi-skill/0.1' 'https://ffxiv.consolegameswiki.com/mediawiki/api.php?action=parse&page=Ecliptic_Arcanite&prop=wikitext&format=json'   # 200
curl -s -o /dev/null -w '%{http_code}\n' -A 'Mozilla/5.0 (Macintosh) AppleWebKit/537.36 Chrome/128 Safari/537.36' 'https://ffxiv.gamerescape.com/wiki/Ecliptic_Arcanite'           # 403
curl -s -o /dev/null -w '%{http_code} %{content_type}\n' 'https://xivapi.com/item/48'    # 200 application/json (legacy v1)
```
