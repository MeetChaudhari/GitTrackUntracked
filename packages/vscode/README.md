# GitTrackUntracked VS Code extension (preview)

This is an unpublished, experimental extension source package. It exposes six
commands through the Command Palette:

- Initialize Personal Vault
- Initialize This Project
- Add Active File
- Status
- Sync
- Restore

It calls the installed native `gitu` binary and shows command output in the
**GitTrackUntracked** output channel. It never reads file contents itself. Set
`gittrackuntracked.binaryPath` if `gitu` is not on `PATH`.

The extension deliberately has no telemetry and no automatic/background sync.
Do not publish it before following the release checklist.
