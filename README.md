A lightweight Go CLI tool for Clash of Clans clan management and roster analytics.

---

## Building

### Requirements
- Go 1.22+ compiler

### Compilation
```bash
git clone https://github.com/qiqioxxa/clan-manager.git
cd clan-manager
go build -o clan-manager .
```

---

## Usage

### `roster`
Displays clan roster, supports sorting, snapshot creation, and snapshot diffs

- `-token <string>` — Supercell API bearer token
- `-tag <string>` — Clan tag
- `-view <all|progression|activity>` — Roster view (default: `all`)
- `-sort <name|th|herosum|equipsum|lab|cw|cwl|donations|clangames|capital|score>` — Sort column (default: `score`)
- `-r` — Reverse sorting order
- `-snapshot` — Save a local snapshot of the current roster
- `-diff <depth|date>` — Compare current roster with a snapshot by depth (e.g. `1`) or date (`YYYY-MM-DD`)

### `cwlgroup`
Displays Clan War League group stats

- `-token <string>` — Supercell API bearer token
- `-tag <string>` — Clan tag
- `-sort <name|stars|attacks|score|listed|present>` — Sort column (default: `score`)
- `-r` — Reverse sorting order

### `cwllayout`
Displays Clan War League layout details for each clan by day

- `-token <string>` — Supercell API bearer token
- `-tag <string>` — Clan tag
