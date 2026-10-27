# zinc

**English** · [简体中文](README_zh.md)

[![Go 1.27+](https://img.shields.io/badge/Go-1.27%2B-00ADD8?logo=go&logoColor=white)](https://go.dev/)
[![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Platform](https://img.shields.io/badge/platform-Windows%2010%20%7C%2011-0078D4?logo=windows&logoColor=white)](#requirements)
[![Build](https://github.com/GensokyoNeoSanae/zinc/actions/workflows/build.yml/badge.svg)](https://github.com/GensokyoNeoSanae/zinc/actions/workflows/build.yml)

**File-level incremental backup that writes plain 7-Zip archives: one complete, encrypted `.7z` per batch, restorable with nothing but 7-Zip and the password.**

zinc backs up one or more source folders ("repos") into a single output root. Every run packs only the files that changed since the last run, and every batch becomes one or more self-contained `.7z` archives with encrypted headers. There is no proprietary repository format, no chain of deltas, and no dependency on zinc when the day comes to restore.

## Contents

- [Why](#why)
- [Features](#features)
- [How it works](#how-it-works)
- [Requirements](#requirements)
- [Install](#install)
- [Quick start](#quick-start)
- [Commands](#commands)
- [Configuration](#configuration)
- [Restore](#restore)
- [Scheduling and notifications](#scheduling-and-notifications)
- [What lives where](#what-lives-where)
- [Uninstall](#uninstall)
- [Limitations](#limitations)
- [Troubleshooting](#troubleshooting)
- [Building from source](#building-from-source)
- [Contributing](#contributing)
- [License](#license)
- [Acknowledgements](#acknowledgements)

## Why

- **Cloud drives sync folders, not repositories.** Keep every source folder under one output root and one sync client covers all of them.
- **A split volume is still a single archive.** With `-v4g`, one missing or damaged piece can make the whole volume set useless. zinc writes complete packs instead: losing one pack costs one batch.
- **Restoring should not require the backup tool.** Each pack is a plain 7-Zip archive with an encrypted header. `7z x` in filename order is the entire restore procedure.
- **Upload volume should follow change volume.** A file-level index means a photo added today does not re-upload last year's library.
- **It is an archive, not a mirror.** A finished pack is never rewritten or trimmed. That is what keeps restores tool-independent; deletions ride along in a small text record next to the packs instead of editing them, see [Limitations](#limitations) for what that rules out.

## Features

- **File-level incremental** — only new or changed files are packed; nothing is re-compressed.
- **Independent packs** — every batch is one or more complete archives. A corrupt or missing pack only loses that batch.
- **Encrypted headers** — `-mhe=on`, so file names are invisible without the password.
- **Restore without zinc** — 7-Zip plus the password is enough; disaster recovery is documented below.
- **Deletions remembered** — a file removed from the source stops coming back on restore, without rewriting a single old pack: each batch leaves a small text record next to the packs.
- **Single-path restore** — pull one file, one directory or a wildcard out of the pack chain without unpacking everything.
- **Index is a cache, not the source of truth** — lose `index.tsv` and rebuild it from archive headers, no extraction required.
- **Multiple repos** — pictures, videos and documents under one synced root, each with its own size limit, password and exclusions.
- **Scheduling** — daily, weekly or monthly Task Scheduler entries, with Windows toast notifications on completion.
- **Bilingual CLI** — built-in Chinese and English messages, selected automatically from the Windows UI language.
- **No dependencies** — standard library only; four small executables and the 7-Zip you probably already have.

## How it works

A backup run is a straight pipeline:

```
scan source  →  diff against index  →  plan packs  →  7za a (.part)  →  7za t  →  rename  →  update index + manifest + deletion record
```

1. **Scan.** The source tree is enumerated once. File size and modification time come straight from the directory entry, so no per-file `stat` is needed.
2. **Diff.** A file is "changed" when it is missing from the index, or its size or mtime differs. Unchanged files are never touched again. A path that the index knows but the scan no longer finds is not packed: it is a deletion.
3. **Plan.** Changed files are sorted by mtime ascending (ties broken by relative path) and greedily appended to the current pack. As soon as the next file would exceed `max_pack_byte`, the pack is closed and a new one starts.
4. **Pack.** Each pack is written as `7za a -t7z -p<password> -mhe=on -mx<N> -scsUTF-16LE -sccUTF-8` into a `.part` file; the file list is handed over as a UTF-16LE listfile because 7-Zip's UTF-8 listfile parser chokes on non-BMP characters such as emoji in file names. After packing, `7za t` verifies the archive; only then is it renamed to its final name. A failed batch is deleted and the index is left untouched.
5. **Record deletions.** Paths that disappeared go into a plain-text sidecar next to the packs, named after the batch's last pack. Nothing is written into existing packs, and a batch that only deleted files produces the record without a pack.
6. **Record.** The index and the SHA-256 manifest are updated (write to temporary file, then rename). Deleted paths are dropped from the index, so they are not reported as deleted again next time.

### Why is each pack named after a timestamp?

```
YYYYMMDD_hhmmss_SSS.7z
20261025_070012_334.7z
```

The name is not decoration: **it is the restore order.** A full restore walks the packs in name order and lets later packs overwrite earlier ones, so the name has to be strictly increasing and must never change afterwards.

That is why it records the moment **the pack started packing**, fixed width at 19 characters so lexicographic order equals chronological order. Collisions append `_1`, `_2` (compared numerically when sorting), and if the system clock jumps backwards the name falls back to "latest existing pack + 1 ms" with a warning, so the order never stops increasing.

Packing time is used instead of content mtime because restore correctness depends on "the later backup wins". If a file is backed up as a new version and the old version is later copied back (a downgrade, a vault rollback), its content mtime moves backwards. A pack named after content time would then sort before the newest pack, and a full restore would silently produce a mixed, wrong state. Content mtime ranges are still recorded in the index, just not in the filename.

### What is `index.tsv` for?

It is what makes a backup incremental, and it is the only thing zinc has to read to decide what changed. Every run scans the source and asks one question per file: is this path stored already, and did its size or mtime move? `index.tsv` answers that by mapping every backed-up path to the pack that holds it. Without it there is no diff: zinc treats the repo as empty and compresses everything again.

Restore uses it too — `--only` looks the path up here and jumps straight to the pack that holds the newest copy instead of walking the whole chain.

`<output>/index.tsv` is a cache; the packs are authoritative.

```
# zinc-index 2
P	1	20261024_110635_081.7z	20261019	20261023
F	notes/semiconductor.md	2048	20261023T124138.9843416+0800	1
# end	1	1
```

- `P` row — pack id, pack filename, earliest and latest content mtime in that pack
- `F` row — relative path, bytes, mtime as `20261023T124138.9843416+0800`, owning pack id
- `# end` footer — pack count and file count, used to detect truncation

A bad header, a malformed line, a missing footer or a count mismatch is a hard error, never a silent reset. Losing the file costs a rebuild, not the data:

```powershell
zinc reindex pictures         # rebuild by reading archive headers only
zinc reindex pictures --yes   # overwrite an existing index (the previous file is kept as .bak)
```

`reindex` runs `7za l -slt` and never extracts content.

### How are deletions recorded?

A pack only ever adds files, so "this one is gone" cannot live inside a pack. It goes into a deletion record: a plain-text sidecar next to the packs, one relative path per line (followed by a CRC-32 column), named after the batch that noticed the disappearance.

```
<output>\
  20261024_110635_081.7z    ← pack
  20261025_070012_334.7z    ← pack
  20261025_070012_334.rm    ← deletion record of that batch
```

```
# zinc-deleted 2
# sha256 86314d871a3443ad7fe9af59b2b9a71ae82d48f491db9e15479b8c4124bf3281
notes/old-draft.md	c1ebf7ba
attachments/2026/screenshot.png	864b8d6a
```

- **Order comes from the name.** The record is named after the last pack of its batch, so it sorts right behind it. `restore` runs the packs and the records off the same timeline, which is why a file that is deleted and later re-added comes back in the right version.
- **Two layers of checksums, so damage stays local.** Each line ends with the CRC-32 of its path; the second line is the SHA-256 of the whole list, which is what notices a line being added or removed. A flipped character turns exactly one line unusable — you lose one deletion, which shows up as one extra file, never as the wrong file being deleted.
- **A damaged record does not lock the repository.** A bad line is skipped with a warning and the rest still applies; a hand-edited record (lines added or removed, every remaining line still self-consistent) is used as it stands with a warning. Only a broken header or a missing checksum line stops the restore, because then the file is not a deletion record any more — `zinc restore --no-delete` still gets your files back in any case.
- **Notepad-proof.** A UTF-8 BOM, CRLF line endings, blank lines and stray spaces around a path are all accepted. `zinc verify` compares bytes, so an edited record shows up there as changed — that is the alarm, not a failure of the restore path.
- **Only files that really went away.** Files matching an `exclude` pattern are never recorded as deleted, and neither is anything under a directory the scan could not read — a permissions error must not turn into a restore that deletes real files.
- **A missing source is not a mass deletion.** If `--source` does not exist, the backup stops with "source does not exist" instead of recording thousands of deletions. An empty but present folder, however, is read literally: everything in it is gone, and that is recorded.
- **Checked before anything is touched.** Restore validates every path in every record before extracting the first pack: absolute paths, drive letters, `..` segments and NTFS streams make that line unusable (it is skipped and reported), and a structurally broken record stops the restore before anything is written.
- **The past is never rewritten.** Old packs keep every file they ever held, so a manual `7z x` still restores everything. If you want a manual restore to match the last backup exactly, delete the paths listed in the `.rm` files (first tab-separated field) after extracting the last pack.

### Why can't I restore "as of a date"?

Because packs are not divided by day. One pack can hold files collected over several days, so cutting the chain at a date would either pull in files that are too new or drop files that belong. What people actually want from a date filter is covered by two commands: `zinc restore` gives you the latest state, and `zinc verify` tells you whether the packs are still intact. Point-in-time history is what a repository-based tool such as kopia or restic is for.

## Requirements

- **Windows 10 or 11, 64-bit.** zinc uses the Task Scheduler, Windows toast notifications and Win32 console APIs; it does not build or run on other platforms.
- **7-Zip.** zinc calls `7z.exe` or `7za.exe`; the normal 7-Zip installer is enough. The setup finds an existing 7-Zip and records its path in `settings.json`. If yours sits somewhere unusual, point the wizard at it, or later run `zinc config set 7zip_path <path>`.
- **Go 1.27 or newer** — only if you build from source.

## Install

1. Download `zinc-<version>-setup.exe` from the [releases page](https://github.com/GensokyoNeoSanae/zinc/releases) and run it. The first page picks the language: English by default, 简体中文 as the second option. The setup is not code-signed, so SmartScreen may show *Windows protected your PC* — choose **More info → Run anyway**. The wizard needs no admin rights: it defaults to `%LOCALAPPDATA%\Programs\zinc` (the destination page lets you pick another folder, and the setup never reuses the previous one), registers the notification identity and adds the folder to your user `PATH`.
2. Make sure 7-Zip is installed (<https://www.7-zip.org/>). The wizard looks for `7z.exe` / `7za.exe` through 7-Zip's registry entry, on `PATH`, and in the usual install folders, and lets you pick another path; **Auto-detect** next to the path runs that search again. It will not finish without a usable 7-Zip. If you do not have 7-Zip yet, tick **Install 7-Zip automatically** on that page and pick where it should go (default `C:\Program Files\7-Zip`): the wizard downloads the official installer from 7-zip.org and runs it, so Windows asks for administrator approval once.
3. Open a **new** terminal, then pick where backups live:

   ```powershell
   zinc root D:\Backup\zinc        # example path, pick your own
   ```

   Repos then default to `<root>\<name>`.

The wizard installs `zinc.exe`, `zincw.exe`, the two helpers under `tools\`, this README and `LICENSE` into the folder you chose; `settings.json` and `repos.json` are created there as needed and are not part of the backup. The uninstaller (`unins000.exe`) is generated by the setup, so uninstall through **Settings → Apps** instead of copying files around.

If you prefer a portable copy, the release also ships `zinc-<version>-windows-amd64.zip`: unzip it anywhere and run `tools\zinc-register.exe` from that folder. It performs the same registration as the wizard — notification identity, user `PATH`, 7-Zip check — only without an entry in **Settings → Apps**. `zinc init` is the lighter variant: same notification and `PATH` work, but 7-Zip is reported, not required.

## Quick start

```powershell
# create repos (every path in this README is an example, use your own folders)
zinc add pictures --source D:\Storage\Pictures
zinc add videos   --source D:\Storage\Videos

# back up now (prompts for the password twice unless one is configured; skips the prompt if nothing changed)
zinc backup pictures

# back up everything
zinc backup --all

# status
zinc list
zinc status
zinc status pictures

# restore (overwrites same-name files by default)
zinc restore pictures --dest D:\restore\pictures

# only show which packs would be extracted, in which order
zinc restore pictures --check-list
```

## Commands

| Command | Description |
|---|---|
| `zinc init [--force]` | re-register notifications and the user `PATH` entry. The setup already did this: use it when notifications stop working or `zinc` is not found, and with the portable zip |
| `zinc root [<path>]` | show or set the global output root |
| `zinc list` | list repos |
| `zinc add <name> --source <path> [--output <path>]` | create a repo |
| `zinc remove <name>` | remove a repo (keeps the packs) |
| `zinc set <name> [--source <path>] [--output <path>] [--password <pw>] [--max-pack-byte <n>] [--compress <0-9>] [--exclude <glob>]` | modify a repo |
| `zinc config [get\|set\|unset\|path] [<key>] [<value>]` | global settings |
| `zinc backup <name\|--all> [--password <pw>] [--notify] [--notify-empty] [--no-verify]` | back up |
| `zinc restore <name> [--dest <path>] [--password <pw>] [--overwrite\|--no-overwrite] [--no-delete] [--check-list] [--only <path>] [--from <pack>]` | restore everything or one path |
| `zinc reindex <name> [--password <pw>] [--pack-start <name>] [--pack-end <name>] [--yes]` | rebuild the index from the packs |
| `zinc verify <name\|--all> [--sample <pct>]` | SHA-256 check of the packs and of the deletion records |
| `zinc status [<name>]` | show status |
| `zinc schedule <name\|--all> --daily --at HH:MM` | daily schedule |
| `zinc schedule <name\|--all> --weekly 1,3 --at HH:MM` | weekly (1 = Monday) |
| `zinc schedule <name\|--all> --monthly 1,15 --at HH:MM [--yes]` | monthly |
| `zinc unschedule <name>` | delete that repo's tasks |
| `zinc unschedule all` | delete the "back up all" summary task (`zinc-all-*`) |
| `zinc unschedule --all` | delete **all** zinc tasks (careful) |

`--lang zh|en` is accepted before or after the subcommand. `zinc verify --sample 20` hashes a random 20% of the pack files instead of all of them.

### Safety guards

- `add` and `set` store absolute paths and refuse a source/output pair where one directory contains the other.
- Two repos cannot share the same output directory. Outputs that differ only by case get a warning, because Windows treats them as the same directory.
- `restore --dest` refuses the source directory, the repo's output directory (and anything below it), and drive roots.
- `remove` warns when `zinc-<name>-*` scheduled tasks are left behind.

## Configuration

Global settings live in `<program dir>\settings.json`. They are **not** part of the backup and are not synced. Precedence is: command-line flag, then per-repo field, then `settings.json`, then built-in default.

| Key | Default | Meaning |
|---|---|---|
| `default_password` | empty | empty = prompt every time |
| `restore_overwrite` | `true` | overwrite same-name files during restore |
| `restore_wrap_folder` | `true` | wrap a full restore in `restore-<repo>-<time>` |
| `max_pack_byte` | `4G` | pack size limit; accepts `4G`, `50M` or plain bytes |
| `compression_level` | `0` | 7z level 0-9; `0` = store only |
| `7zip_path` | detected at install | path to `7z.exe` / `7za.exe`; empty = search `PATH` and the usual install folders |
| `lang` | empty | `zh` or `en`; empty = detect from Windows |
| `lock_timeout_minutes` | `30` | longest wait for the repo lock |
| `exclude` | `[]` | glob list; `set` appends to it |
| `complete_notify` | `true` | send a toast when a backup finishes |

```powershell
zinc config                       # print everything (password masked)
zinc config get max_pack_byte     # bare value, for scripts
zinc config set max_pack_byte 50M
zinc config set exclude '**/node_modules/**'
zinc config unset exclude         # fall back to the built-in default
zinc config path                  # where the file is
```

`exclude` patterns are globs matched against the relative path: `**` crosses directory levels, `*` stays within one segment, `?` matches a single character. Matching a directory skips the whole subtree.

### Language

Messages are available in Chinese and English. The language is resolved in this order: `--lang zh|en`, the `ZINC_LANG` environment variable, `lang` in `settings.json`, the build-time default (`-X main.defaultLang`), the Windows UI language, and finally English.

```powershell
zinc --lang en list
zinc list --lang zh
zinc config set lang zh
```

### Password

Four sources, in order: `--password`, the repo's `password`, `settings.json` `default_password`, interactive prompt.

```powershell
# set it once for unattended runs; `-` reads the value interactively so it stays out of shell history
zinc config set default_password -
```

Interactive input is visible (so you can spot typos), and `backup` asks twice. Once a default password is configured, the summary line records where the key came from, for example `enc✓(settings)`.

Passwords are stored in plain text in `settings.json` and in the scheduled task commands (`C:\Windows\System32\Tasks\zinc-*`). The encryption is meant to defeat automated cloud-drive content scanning, not a targeted attacker who already has access to your machine.

**Changing the password requires re-running `schedule`**, otherwise the tasks still carry the old one.

## Restore

One rule: all packs are extracted in name (time) ascending order, so the last backup wins. Deletion records run on that same timeline, right after the pack they are named after, so what you get back is the tree as it looked at the last backup — not a union of everything ever packed.

| Flag | Behaviour |
|---|---|
| none | follows `restore_overwrite` (default `true` = overwrite) |
| `--overwrite` | force overwriting |
| `--no-overwrite` | keep both; same-name files are extracted as `_1`, `_2` |
| `--no-delete` | restore files only; skip the deletions recorded by later backups |
| `--notify=false` | do not send the "restore incomplete" toast |
| `--check-list` | only print the order and sizes (packs and deletion records), change nothing; `--dest` may be omitted |
| `--only <path>` | restore just that one file or directory, leave everything else alone; deletion records are not applied |
| `--from <pack>` | with `--only`, take it from a specific pack: full name, unique prefix, or 1-based position in the repo's complete pack list — not the numbering printed by `--check-list` |

Overwrite mode prints a warning when it starts. Restore works through the packs strictly in name order, so a renamed or hand-copied pack changes which version wins — keep the names zinc assigned.

A damaged pack neither stops the restore nor gets thrown away. 7-Zip extracts everything it can; only the files whose CRC failed come out wrong, and it leaves them on disk. zinc reports that pack's 7-Zip error, skips that batch's deletion record (nobody knows what the source looked like at that moment), carries on with the remaining packs, warns at the end and exits non-zero. With `complete_notify` on — the default — it also sends a toast titled "zinc restore incomplete"; `--notify=false` silences it. Run `zinc verify` to see exactly which pack is bad, then fetch that one again.

### Restore wrapper folder

By default the restore target gets one extra level, `restore-<repo>-<YYYYMMDD_hhmmss_SSS>` (timestamp = when this restore started), and every file lands inside it instead of being scattered over `--dest` itself:

```
<dest>\
  restore-pictures-20261025_170012_334\
    2024\...
```

- Turn it off to restore flat into `--dest`: `zinc config set restore_wrap_folder false`
- An interrupted restore leaves its half-finished folder alone; re-running creates a fresh timestamped folder instead of mixing with it
- Full-repo restores only; `--only` single-path restores still land directly in `--dest`

### Restoring a single file

You do not need to unpack the whole repository:

```powershell
# newest version of one file (finds the right pack automatically)
zinc restore docs --only 'notes/chapter1.md' --dest D:\recovered

# a whole directory
zinc restore docs --only 'notes/archive' --dest D:\recovered

# wildcards
zinc restore docs --only '*.md' --dest D:\recovered

# which packs contain it? (lists versions, touches nothing)
zinc restore docs --only 'notes/chapter1.md' --check-list

# pull an older version (pass a pack name printed by --check-list)
zinc restore docs --only 'notes/chapter1.md' --from 20261020_070012_334.7z --dest D:\old
```

How the pack is chosen:

- exact file path → looked up in `index.tsv`, which points straight at the newest pack holding it
- directory or wildcard → every pack containing a match is extracted oldest to newest, so the result matches a full restore
- a missing index is fine: it falls back to scanning pack headers (`7za l`, no extraction)

Wildcards follow 7-Zip's matching rules, where `*` does **not** reliably cross directories: `--only '*.md'` only hits files at the top level. To restore a whole subtree, pass the directory itself (`--only 'notes/archive'`) instead of guessing with a wildcard.

`--only --check-list` needs the password, because listing an archive means decrypting its header.

## Scheduling and notifications

```powershell
# every day at 20:00, back up all repos, notify when done
zinc schedule --all --daily --at 20:00 --password yourpassword

# pictures at 10:00 and 20:00 every day
zinc schedule pictures --daily --at 10:00,20:00 --password yourpassword

# every Monday and Wednesday at 20:00
zinc schedule pictures --weekly 1,3 --at 20:00 --password yourpassword

# 1st and 15th of every month at 20:00 (days 29-31 need --yes, otherwise the task is refused)
zinc schedule pictures --monthly 1,15 --at 20:00 --password yourpassword
```

Tasks are created through Task Scheduler and point at `zincw.exe` (GUI subsystem), so no console window flashes up. Notification rules:

- `zinc backup` pops a toast when it finishes while `complete_notify` is on (the default). For a quiet manual run, pass `--notify=false`.
- Tasks created by `zinc schedule` always pass `--notify` unless the task was created with `--no-notify`; `complete_notify` has no effect on them.
- `--all` sends one summary; single-repo tasks each send one.
- Nothing changed → no popup (`--notify-empty` forces one).
- `zinc restore` only pops up when the result is incomplete — some pack could not be extracted cleanly. A clean restore stays quiet.

Notification details:

Windows notifications require an AUMID (AppUserModelID), which an unpacked executable does not have. The setup creates the three things below while installing; a portable copy creates them on its first `--notify`:

1. `%APPDATA%\Microsoft\Windows\Start Menu\Programs\zinc.lnk` with AUMID `zinc.backup`
2. `HKCU\Software\Classes\AppUserModelId\zinc.backup` (DisplayName = zinc)
3. `.aumid_ok` in the program directory, so the work is not repeated

**All three must exist for notifications to show.** If they stop working, let zinc rebuild them instead of hunting for files:

```powershell
zinc init --force
```

That rewrites the shortcut, the registry entry and the marker in one go, wherever the program directory happens to be.

Icon: drop a `zinc.ico` next to `zinc.exe`, then run `zinc init --force` to pick it up.

### Deleting tasks: `all` is not `--all`

```powershell
zinc unschedule pictures   # only the pictures repo's tasks
zinc unschedule all        # only the "back up all" summary task
zinc unschedule --all      # all zinc tasks, including every repo
```

Typing `--all` when you meant `all` wipes every repo's schedule.

## What lives where

Output root (sync this folder to your cloud drive):

```
<root>\
  pictures\
    20261024_110635_081.7z    ← each pack is a complete archive; restore in name order
    20261025_070012_334.7z
    20261025_070012_334.rm    ← files deleted in that batch (plain text)
    index.tsv                 ← incremental detection cache, rebuild with zinc reindex
    manifest.json             ← SHA-256 ledger for verify
  videos\
```

| File | Location | Upload | Purpose |
|---|---|---|---|
| `*.7z` | `<output>/` | yes | the data; every pack is a complete archive |
| `*.rm` | `<output>/` | yes | deletions of that batch with their checksum; restore needs them to remove files, so upload them with the packs |
| `index.tsv` | `<output>/` | yes | incremental detection cache; rebuild with `zinc reindex` if lost |
| `manifest.json` | `<output>/` | yes | per-batch SHA-256 records for `verify`; must be uploaded, otherwise a new machine cannot verify old packs |

Runtime files (program directory, never synced):

| File | Purpose |
|---|---|
| `zinc.exe` | console build, for manual use |
| `zincw.exe` | GUI build, used by scheduled tasks |
| `tools\zinc-register.exe` | registration helper, run by the setup: 7-Zip check, notification identity, user `PATH` |
| `tools\zinc-unregister.exe` | the reverse: scheduled tasks, notification registration, `PATH` entry |
| `unins000.exe` | uninstaller generated by the setup; **Settings → Apps** runs it |
| `zinc.ico` | optional icon for the notification shortcut |
| `repos.json` | repo registry, created on first `add` |
| `settings.json` | global defaults, written by the setup (7-Zip path) and by `config set` |
| `logs\zinc-YYYYMMDD.log` | one file per day, newest 10 kept |
| `.aumid_ok` | notification registration marker |

### Disaster recovery

A disk failure does not require `zinc.exe`:

1. Download every pack of a repo into one folder
2. Extract them one by one in name order:

```powershell
7z x 20261024_110635_081.7z -pPASSWORD -oRESTORE_DIR
7z x 20261024_120901_552.7z -pPASSWORD -oRESTORE_DIR
# ... keep going in name order; later packs overwrite earlier ones
```

Every pack is self-contained, so a damaged pack only loses that pack and the rest still extract.

Extracting by hand gives you a superset: files deleted later are still there, because a pack cannot un-delete anything. To finish the job without `zinc.exe`, open the `.rm` records in the same order the packs were applied and delete the listed paths yourself — they are plain text, one relative path per line, and the batch names tell you where each one sits in the timeline.

### Moving to a new machine

The backups themselves need nothing: packs are plain archives. Only the settings are machine-bound.

1. Run the setup on the new machine. It brings its own 7-Zip check, notification registration and `PATH` entry — no manual copying of the program folder.
2. Point it at the same backup location with `zinc root <root>`, and copy `repos.json` and `settings.json` from the old install folder into the new one, so every repo keeps its source, output and password. Re-running `add` and `config` by hand does the same thing.
3. Re-create the schedule (`zinc schedule ...`): Task Scheduler entries belong to the machine and never travel.

## Uninstall

Open **Settings → Apps → Installed apps**, find `zinc` and choose Uninstall. That runs Inno Setup's uninstaller (`unins000.exe` in the install folder): it first asks you to confirm the removal, then asks whether to keep your settings, and only then deletes files.

Before removing them it calls `tools\zinc-unregister.exe` to undo the registration:

- stops and removes the scheduled tasks that point at this folder, so no backup keeps running after the program is gone
- removes the notification registration: Start Menu shortcut, `HKCU\Software\Classes\AppUserModelId\zinc.backup`, `.aumid_ok`
- removes the folder from your user `PATH`

The keep-or-delete question comes after you confirm the uninstall: answer **No** to delete `settings.json`, `repos.json` and `logs` along with the program, **Yes** to keep the repo list and default password for a later reinstall. A silent uninstall (`/VERYSILENT`) behaves like **No**. Marker files such as `.aumid_ok` are not user data and are always removed. Backups — the `.7z` packs and manifests in your output directories — are never touched.

If you moved the program folder after installing, the tasks still point at the old one and will not be matched: either keep the folder where it was, or clear them first with `zinc unschedule --all`.

## Limitations

zinc answers "what did this folder look like at each backup", and a restore reproduces the most recent one — deleted files included. What it does not do is jump back to an arbitrary past moment, deduplicate renames, or reclaim space. Folders you reorganize constantly, such as notes, Obsidian vaults or source trees, are still a poor fit; a repository-based tool such as [kopia](https://kopia.io/) or [restic](https://restic.net/) handles those better, at the cost of needing the tool to restore.

- **Deletions are recorded, not erased.** A removed file is not packed again, its index entry is dropped, and the batch writes it to a `.rm` record; the old packs are never rewritten, so their bytes stay. Restore applies the records and gives you the last backup's tree, but the space is only reclaimed by starting a fresh output directory and backing up again.
- **Moves and renames re-pack.** The index tracks relative paths, so a new name or directory is one delete plus one add, and the content is compressed again in full. Renaming a tree of attachments can therefore cost a full re-upload of that tree. (The delete half is just a `.rm` entry; it does not shrink the old packs.)
- **Backups only grow.** Nothing in zinc shrinks a repository: packs are never rewritten, so deleting a source folder frees no space in the packs that already contain it. To actually reclaim it, start a new output directory, back up again, then delete the old one.
- **No import from older volume layouts.** `vol-*.7z.001` archives are not read. Start a fresh repo, or keep the old tree as an archive.
- **Scheduling needs a real terminal.** `schtasks` cannot run in restricted environments such as sandboxes.
- **Plain-text password in tasks.** Unavoidable for unattended runs; the threat model is content scanning, not a targeted attacker.
- **`manifest.json` must be uploaded with the packs.** It records hashes computed at backup time, so `verify` can detect transfer corruption on a new machine. Without it, old batches cannot be verified.
- **`reindex` trusts the local time zone.** `7za l` prints local time without an offset, so a rebuild interprets mtimes in the current zone.
- **`reindex --pack-start` / `--pack-end` replace the whole index with a slice.** Packs outside that range are then treated as new on the next backup and re-packed. Those flags exist for disaster recovery only; day to day, run a plain `zinc reindex <repo>`.
- **Symlinks and junctions are never scanned.** A `--source` that is itself a junction yields an empty scan (it will look like "nothing new" forever), and junction subtrees inside the source are not indexed. Point `--source` at the real path.
- **Empty directories are not backed up.** The pack list contains files only, so a folder with no file in it does not exist after a restore; folders that do hold files are recreated with them.
- **A stale `.lock` blocks future backups.** Locks are never taken over automatically. After a killed process, delete `<output>\.lock` by hand, otherwise every later backup waits the full `lock_timeout_minutes` (default 30) and then fails.
- **A missing index means a full re-pack.** If `index.tsv` is deleted and you do not rebuild it, the next backup treats the repo as empty and compresses everything again (nothing is lost, but the upload is duplicated). Losing it is cheap if you run `zinc reindex <repo>` first. A **corrupt** index behaves differently: that is a hard error, and the backup stops until you run `zinc reindex`.
- **Changing the password orphans old packs.** Old packs still use the old password.
- **Windows only.** 64-bit Windows 10 or 11.

## Troubleshooting

| Symptom | Cause and fix |
|---|---|
| `zinc` is not recognised after install | `PATH` changes apply to newly opened terminals; if it persists, run `zinc init --force` |
| Setup stops with `unable to create the directory ... is-XXXX.tmp` / `Error 5` | The installer cannot write under `%TEMP%` (redirected to another drive, or guarded by security software). Start it from `cmd` with `set TEMP=%LOCALAPPDATA%\Temp` and `set TMP=%LOCALAPPDATA%\Temp`, or point those variables at a writable folder |
| `[warn] 7z.exe / 7za.exe not found` | 7-Zip is not where zinc looks; install it, or run `zinc config set 7zip_path <path to 7z.exe>` |
| Backup waits a long time and fails | A stale `<output>\.lock`; delete it, then run the backup again |
| Notifications stopped appearing | Run `zinc init --force` to rebuild the shortcut, the registry entry and the marker |
| `restore --from` rejects the value | Pass a pack filename printed by `--check-list`; a bare number is looked up in the repo's complete pack list, not in that printout |
| Restore picks the wrong version | Run `zinc restore <repo> --only <path> --check-list` to see which packs hold the file, then use `--from` |
| `restore` reports `N pack(s) failed to extract` | The packs are damaged (bit rot, a broken upload, a half-synced cloud folder). The rest of the repository was still restored. Run `zinc verify` to see every bad pack, delete them from the output folder and fetch the same names from another copy, or extract what you can by hand |
| One file fails to back up | It is locked exclusively (an open Excel workbook, for example). It is skipped **without touching the index**, so the next run retries automatically |

## Building from source

```powershell
git clone https://github.com/GensokyoNeoSanae/zinc.git
cd zinc
go build -o zinc.exe ./cmd/zinc
go build -ldflags "-s -w -H=windowsgui" -o zincw.exe ./cmd/zinc
go build -o zinc-register.exe ./cmd/install
go build -o zinc-unregister.exe ./cmd/uninstall
```

`build.ps1` wraps those four commands. The installer is an [Inno Setup 6](https://jrsoftware.org/isinfo.php) script, `installer/zinc.iss`. Install Inno Setup once; the build script looks for `ISCC.exe` in `_tools\InnoSetup6`, in `Program Files`, or wherever `-Iscc` points:

```powershell
.\build.ps1                          # the four zinc binaries
.\installer\build.ps1 -Version 0.1.0 # dist\zinc-0.1.0-setup.exe
.\package.ps1 -Version 0.1.0         # setup + portable zip
```

`package.ps1` calls both and writes `dist\zinc-<version>-setup.exe` (the wizard) and `dist\zinc-<version>-windows-amd64.zip` (portable). The setup is a single file: Inno compresses the four binaries, the icon, the docs and its own uninstaller into it. The first page picks the language — English by default, 简体中文 as the second option.

`cmd/` holds the four entry points and `internal/app` the shared code. The Inno script does not implement registration itself: it runs `tools\zinc-register.exe` after copying the files and `tools\zinc-unregister.exe` before removing them, so the CLI stays the single source of truth. The two main executables differ only in subsystem: `zinc.exe` is a console program so you can watch it work, `zincw.exe` uses the GUI subsystem so scheduled tasks do not flash a console window (or steal focus). There are no third-party dependencies; `go.mod` only declares the Go version.

`zinc.ico` is generated, not hand-drawn: run `python tools\make-icon.py` (needs Pillow) after changing the colours at the top of that script. It renders the `Zn` mark separately at ten sizes and assembles the ICO itself, so the 16px frame is not a downscaled 256px one.

## Contributing

Issues and pull requests are welcome. Please keep changes focused, run `gofmt` and `go vet`, and build all four executables before opening a PR. See [CONTRIBUTING.md](CONTRIBUTING.md) for details.

## License

[MIT](LICENSE).

## Acknowledgements

zinc started from an idea in [zbak](https://github.com/imaegoo/zbak), a 7-Zip based backup tool for NAS users: a folder tree can be archived as plain 7z files and still be restorable years later without any special software. Thanks to [@iMaeGoo](https://github.com/imaegoo).

Thanks also to the 7-Zip project, whose command-line interface does most of the actual work, and to the Go standard library, which is the only dependency.
