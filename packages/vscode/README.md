# GitTrackUntracked VS Code extension (preview)

This experimental preview extension adds a **GitTrackUntracked** icon to the
VS Code Activity Bar. Its **Private Local Files** view provides a compact
dashboard for the current workspace:

- vault/project readiness and sync state;
- registered local files, which can be opened from the view;
- buttons to initialize the vault or project, add files or folders, sync,
  restore, and refresh.

The same actions are also available through the Command Palette:

- Initialize Personal Vault
- Initialize This Project
- Add Active File
- Add Files or Folders
- Status
- Sync
- Restore
- Refresh

It calls the installed native `gitu` binary and shows command output in the
**GitTrackUntracked** output channel. Install `gittrackuntracked` or build
the Go CLI first, then set `gittrackuntracked.binaryPath` if `gitu` is not on
`PATH`.

Restoring always starts in safe mode: files already present in the workspace
are not overwritten unless the user deliberately chooses **Restore and
overwrite**. The extension deliberately has no telemetry and no
automatic/background sync. It is not a secret manager and should not be used
in production.
