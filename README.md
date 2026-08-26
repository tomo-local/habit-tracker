# habit-tracker

A CLI tool for tracking daily habits using Google Calendar.

- Displays a GitHub-contributions-style heatmap in the CLI
- Records a habit to your calendar with a single command
- Supports switching between multiple Google accounts
- Manage habits and records from a web UI (`serve`, low priority)

## Install

```sh
brew tap tomo-local/habit-tracker https://github.com/tomo-local/habit-tracker
brew trust tomo-local/habit-tracker   # Homebrew 6.0+: required to trust a new tap on first use
brew install habit-tracker
```

## Commands

```sh
habit-tracker                        # same as view (shows the combined heatmap for all habits)

habit-tracker view                   # interactively select a habit (or "All habits") and show its heatmap
habit-tracker view <habit>           # show a specific habit directly (must be a configured habit)

habit-tracker auth login             # authenticate a Google account and save the token
habit-tracker auth list              # list authenticated accounts
habit-tracker auth switch            # switch the active account
habit-tracker auth remove            # remove an account

habit-tracker setup                  # select a calendar and register habits (interactive)

habit-tracker add                    # interactively select a registered habit and record it
habit-tracker add <habit>            # record a specific habit directly
habit-tracker add -d 60 <habit>      # change the duration (default: 30 min; -d must come before the habit name)
habit-tracker add -D "**note**" <habit>  # set a Markdown event description (-D must come before the habit name)

habit-tracker serve                  # start the web UI in a browser

habit-tracker -h                     # show the command list
habit-tracker <command> -h           # show detailed options and examples for a command
```

## Setup

### 1. Create Google Calendar API credentials

1. Create a project in the [GCP Console](https://console.cloud.google.com/)
2. Under "APIs & Services" → "Library", enable the **Google Calendar API**
3. Under "Credentials" → "Create Credentials" → "OAuth client ID", choose **Desktop app**
   (you'll be asked to configure the consent screen on first use; set User Type to "External" and add yourself as a test user)
4. Download the JSON and place it at `~/.config/habit/credentials.json`
   (the directory can be changed with the `HABIT_CREDENTIALS_DIR` environment variable)

> **Note**: Don't push `credentials.json` to a public repository.

### 2. Authenticate and set up

```sh
habit-tracker auth login   # OAuth via browser -> saves the token
habit-tracker setup        # create a new calendar or select an existing one, then register habits to track
```

### 3. Verify it works

```sh
habit-tracker              # shows the combined heatmap for all habits
```

## Storage

```text
~/.config/habit/
├── token.json       # OAuth token
└── config.json      # habit and calendar settings
```

Paths can be overridden with environment variables:
- `HABIT_CONFIG_DIR` — path to the config directory (default: `~/.config/habit/`)
- `HABIT_CREDENTIALS_DIR` — path to the directory containing `credentials.json`

### config.json

```json
{
  "calendar_id": "abc123@group.calendar.google.com",
  "calendar_name": "habit",
  "habits": ["workout", "reading"]
}
```

> **Note**: config.json should only be modified via `setup`. Manual edits are not supported.

- `habits` — list of habit names to track. Running `view` with no arguments (or the bare `habit-tracker` command) shows the combined "All habits" heatmap
- If you prefer to split habits across separate calendars, `{"calendars": ["workout", "reading"]}` alone is sufficient

## view output

Same layout as a GitHub contributions graph. Columns = weeks (past on the left), rows = days of the week (Sunday on top).

```text
workout  🔥 12 day streak


Sun  □□□□□■□□■□■□□□□■□□■□■□□□□■□□■□■□□□□■□□■□■□□□
Mon  ■□□■□■□■□□■□■□■□□■□■□□■□■□■□□■□■□□■□■□■□□■□■
Tue  □■□□■□■□■□□■□□■□■□□■□■□■□□■□□■□■□□■□■□■□□■□□
Wed  □□■□□■□■□□■□■□□□■□■□□■□■□■□□■□■□□□■□■□□■□■□□
Thu  ■□□■□□■□■□□■□■□■□□■□■□□■□■□■□□■□■□■□□■□■□□□□
Fri  □■□□■□□■□■□□■□■□■□□■□■□□■□■□■□□■□■□■□□■□■□□□
Sat  □□■□□■□□■□■□□■□□□■□□■□■□□■□□□■□□■□■□□■□□□□□□
```

- ■ = done, □ = not done
- Selecting "All habits" from the interactive list shows a combined heatmap and streak across every tracked habit (a day counts if any habit was recorded)
- 🔥 streak: number of consecutive days completed, counting back from today (still counts as ongoing if today isn't done yet but yesterday was)

## add behavior

- No argument: interactively select a registered habit with arrow keys (choose "New habit" at the end of the list to enter a new name)
- `add <habit>`: record a specific habit directly by name
- `-d <minutes>`: specify the duration directly (default: 30 min; skips the interactive prompt when given). `-d` must always come before the habit name (`add -d 60 <habit>`)
- When `-d` is omitted, you'll be prompted for a duration in minutes after selecting the habit (default: 30 min; values outside 1–1440 are rejected)
- `-D <description>`: set the event description in Markdown (default: none). `-D` must always come before the habit name. There is no interactive prompt for this — it's only settable via the flag
- If the habit is already recorded today, nothing happens (duplicate prevention)

## License

[MIT](LICENSE)
