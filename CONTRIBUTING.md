# Contributing

Thanks for taking a look at zinc. Bug reports, feature ideas and pull requests are all welcome.

## Reporting a bug

Open an [issue](https://github.com/GensokyoNeoSanae/zinc/issues/new/choose) and include:

- Windows version (`winver`) and whether it is 64-bit
- The exact command you ran, and the output (the summary line and any `[warn]`, `[skip]`, `[locked]` lines help a lot)
- The relevant part of `<program dir>\logs\zinc-YYYYMMDD.log`
- Whether it reproduces on a fresh, small source directory

Please do not paste real passwords. Replace them with `xxxx`.

## Building

Requires Go 1.27+ and 7-Zip (`7z.exe` or `7za.exe`).

```powershell
git clone https://github.com/GensokyoNeoSanae/zinc.git
cd zinc
.\build.ps1
```

`build.ps1` produces the four binaries (`zinc.exe`, `zincw.exe`, `zinc-install.exe`, `zinc-uninstall.exe`); `package.ps1` also assembles the release zip under `dist/`.

Layout: `cmd/` holds the four entry points and `internal/app` holds all shared code. The installer and uninstaller are thin wrappers around `registerSetup` and `UninstallMain`; anything that writes to the registry, the Start Menu or the user `PATH` belongs there, not in the CLI.

## Installer

`installer/` contains the two GUI wizards (C# WinForms, compiled by the .NET Framework `csc.exe` that ships with Windows, so no SDK download is needed):

- `Setup.cs` — welcome/license, install folder, 7-Zip, progress, finish. It embeds the binaries in `payload.zip` and writes the Add/Remove Programs entry.
- `Uninstall.cs` — the uninstall wizard, copied into the install folder by the setup.

`.\installer\build.ps1` compiles both; `.\package.ps1` produces the single-file setup and the portable zip. The wizards deliberately do not implement registration themselves — they call `zinc-install.exe` / `zinc-uninstall.exe`, so the CLI stays the single source of truth for the `PATH` entry, the notification registration and scheduled tasks. If you change one of those mechanics, change it in `internal/app`, not in the wizard.

Before opening a pull request:

```powershell
gofmt -l .
go vet ./...
```

Both should come back clean, and both executables should build.

## Scope

zinc is deliberately small: one package, standard library only, Windows only. Changes that add a third-party dependency, a background service, a GUI, or cross-platform support are unlikely to be merged. Changes that make restore more robust, make the CLI clearer, or fix a real data-loss path are very welcome.

## Style

- Match the surrounding code; comments are in Chinese, identifiers in English.
- User-facing strings go through `messages.go` (one entry per language), never inline in command code.
- Anything that touches packs, the index or the manifest should be tested against a scratch output directory before and after the change. A restore of a small tree, compared with the source, is the cheapest useful test.

## Commit messages

Short imperative subject line, optionally with a body explaining the reason. For example:

```
fix lock timeout message on stale .lock

A killed backup left .lock behind, and the next run waited the full
timeout before failing. Print the lock path and the age of the lock file
so it is obvious what to delete.
```
