# ccf — Claude Code fork

Pin a [Claude Code](https://docs.claude.com/en/docs/claude-code) session once, then start a fresh
fork of it whenever you need one.

You build up a session that knows your project: the architecture, the conventions, the tricky
parts. Every new task can start from that context instead of from zero, and the pinned session
itself never changes, because each run is a `claude --resume <id> --fork-session`.

Works on macOS, Linux and Windows. Single binary, no runtime dependencies.

## Install

macOS / Linux:

```sh
curl -fsSL https://raw.githubusercontent.com/savchenko-dev/ccf/main/install.sh | sh
```

Windows (PowerShell):

```powershell
irm https://raw.githubusercontent.com/savchenko-dev/ccf/main/install.ps1 | iex
```

With Go 1.22+:

```sh
go install github.com/savchenko-dev/ccf@latest
```

Or download an archive from [Releases](https://github.com/savchenko-dev/ccf/releases).

## Usage

```text
ccf [name] [claude args...]   fork the pinned session (pins it first if none exist)
ccf pin [name] [session-id]   pin a session (default: name "main", latest session here)
ccf ls                        list pins for the current directory
ccf rm <name>                 remove a pin
ccf help | version
```

A typical flow:

```sh
claude                 # build up a session with good project context, then exit
ccf pin                # pin the latest session in this directory as "main"
ccf                    # every new task: a fresh fork of "main"
ccf --model opus       # extra args are passed through to claude
```

Several pins per project:

```sh
ccf pin backend        # pin the latest session as "backend"
ccf pin ui 1f3c…       # pin a specific session id as "ui"
ccf backend            # fork "backend"
ccf                    # with several pins, pick one interactively
```

The interactive picker uses [fzf](https://github.com/junegunn/fzf) if it is on `PATH`, and falls
back to a numbered prompt otherwise.

## How it works

Claude Code keeps sessions under `~/.claude/projects/<project>/<session-id>.jsonl`, where
`<project>` is the working directory with every non-alphanumeric character replaced by `-`.
`ccf` stores pins in a plain text file next to them, `<project>/.ccf-pins`, one `name session-id`
per line. Pins are therefore per directory, and you can edit the file by hand.

`CLAUDE_CONFIG_DIR` is respected if you have moved the Claude Code config directory.

## License

MIT
