A lightweight Go CLI tool for Clash of Clans clan management and roster analytics.

---

## Building

### Requirements
- Go 1.25+ compiler

### Compilation
```bash
git clone https://github.com/qiqioxxa/clan-manager.git
cd clan-manager
go build -o clan-manager .
```

---

## API authentication

Pass a Supercell API token with `-token`, or omit it and configure developer portal credentials in `.env`:

```dotenv
COC_DEV_EMAIL=you@example.com
COC_DEV_PASSWORD=your-password
```

You can copy `.env.example` as a starting point. When no `-token` is provided, the CLI logs in to [developer.clashofclans.com](https://developer.clashofclans.com), reuses a token for the current IP if available, or creates one. If a supplied API token receives `403 Forbidden`, the CLI attempts to refresh it using .env credentials.

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
