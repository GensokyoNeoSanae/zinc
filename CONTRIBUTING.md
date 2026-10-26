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

Requires Go 1.27+ and 7-Zip (`7za.exe`).

```powershell
git clone https://github.com/GensokyoNeoSanae/zinc.git
cd zinc
go build -ldflags "-s -w" -o zinc.exe .
go build -ldflags "-s -w -H=windowsgui" -o zincw.exe .
```

`build.ps1` does the same thing with a single command.

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
