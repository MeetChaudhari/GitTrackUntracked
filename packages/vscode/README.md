# GitTrackUntracked VS Code extension (preview)

This experimental preview extension exposes six commands through the Command
Palette:

- Initialize Personal Vault
- Initialize This Project
- Add Active File
- Status
- Sync
- Restore

It calls the installed native `gitu` binary and shows command output in the
**GitTrackUntracked** output channel. Install `@meetchaudhari/gittrackuntracked` or build
the Go CLI first, then set `gittrackuntracked.binaryPath` if `gitu` is not on
`PATH`.

The extension deliberately has no telemetry and no automatic/background sync.
It is not a secret manager and should not be used in production.
